package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/argoproj-labs/argocd-notifications/expr"
	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/subscriptions"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"
	"github.com/argoproj-labs/argocd-notifications/shared/k8s"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod           = 60 * time.Second
	notifiedHistoryMaxSize = 100
)

var (
	notifiedAnnotationKey = "notified." + subscriptions.AnnotationPrefix
)

type NotificationController interface {
	Run(ctx context.Context, processors int)
	Init(ctx context.Context) error
}

func NewController(
	client dynamic.Interface,
	namespace string,
	cfg settings.Config,
	appLabelSelector string,
	metricsRegistry *controllerRegistry,
) (NotificationController, error) {
	appClient := k8s.NewAppClient(client, namespace)
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	appInformer := newInformer(appClient, appLabelSelector)

	appInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					queue.Add(key)
				}
			},
		},
	)
	appProjInformer := newInformer(k8s.NewAppProjClient(client, namespace), "")

	return &notificationController{
		appClient:       appClient,
		appInformer:     appInformer,
		appProjInformer: appProjInformer,
		refreshQueue:    queue,
		cfg:             cfg,
		metricsRegistry: metricsRegistry,
	}, nil
}

func newInformer(resClient dynamic.ResourceInterface, selector string) cache.SharedIndexInformer {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (object runtime.Object, err error) {
				options.LabelSelector = selector
				return resClient.List(context.Background(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = selector
				return resClient.Watch(context.Background(), options)
			},
		},
		&unstructured.Unstructured{},
		resyncPeriod,
		cache.Indexers{},
	)
	return informer
}

type notificationController struct {
	appClient       dynamic.ResourceInterface
	appInformer     cache.SharedIndexInformer
	appProjInformer cache.SharedIndexInformer
	refreshQueue    workqueue.RateLimitingInterface
	cfg             settings.Config
	metricsRegistry *controllerRegistry
}

func (c *notificationController) Init(ctx context.Context) error {
	go c.appInformer.Run(ctx.Done())
	go c.appProjInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.appInformer.HasSynced, c.appProjInformer.HasSynced) {
		return errors.New("Timed out waiting for caches to sync")
	}
	return nil
}

func (c *notificationController) Run(ctx context.Context, processors int) {
	defer runtimeutil.HandleCrash()
	defer c.refreshQueue.ShutDown()

	log.Warn("Controller is running.")
	for i := 0; i < processors; i++ {
		go wait.Until(func() {
			for c.processQueueItem() {
			}
		}, time.Second, ctx.Done())
	}
	<-ctx.Done()
	log.Warn("Controller has stopped.")
}

func ensureAnnotations(obj *unstructured.Unstructured) {
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{})
	}
}

func (c *notificationController) processApp(app *unstructured.Unstructured, logEntry *log.Entry) error {
	refreshed := false
	ensureAnnotations(app)

	state := triggers.NewState(app.GetAnnotations()[notifiedAnnotationKey])
	// changes state of specified trigger/destination and returns if state has changed or not
	setAlreadyNotified := func(trigger string, result triggers.ConditionResult, dest services.Destination, isNotified bool) (bool, error) {
		changed := state.SetAlreadyNotified(trigger, result, dest, isNotified)
		// if state changes reload application
		if changed && !refreshed {
			refreshedApp, err := c.appClient.Get(context.Background(), app.GetName(), v1.GetOptions{})
			if err != nil {
				return false, err
			}
			ensureAnnotations(refreshedApp)
			app.GetAnnotations()[notifiedAnnotationKey] = refreshedApp.GetAnnotations()[notifiedAnnotationKey]

			state = triggers.NewState(app.GetAnnotations()[notifiedAnnotationKey])
			refreshed = true
			return state.SetAlreadyNotified(trigger, result, dest, isNotified), nil
		}
		return changed, nil
	}

	subscriptions := c.getSubscriptions(app)
	if len(subscriptions) == 0 {
		return nil
	}

	for trigger, destinations := range subscriptions {

		res, err := c.cfg.API.RunTrigger(trigger, expr.Spawn(app, c.cfg.ArgoCDService, map[string]interface{}{"app": app.Object}))
		if err != nil {
			logEntry.Debugf("Failed to execute condition of trigger %s: %v", trigger, err)
		}
		logEntry.Infof("Trigger %s result: %v", trigger, res)

		for _, cr := range res {
			if !cr.Triggered {
				for _, to := range destinations {
					if _, err := setAlreadyNotified(trigger, cr, to, false); err != nil {
						return err
					}
				}
				continue
			}

			for _, to := range destinations {
				if changed, err := setAlreadyNotified(trigger, cr, to, true); err != nil {
					return err
				} else if !changed {
					logEntry.Infof("Notification about condition '%s.%s' already sent to '%v'", trigger, cr.Key, to)
					continue // move to the next recipient
				}

				logEntry.Infof("Sending notification about condition '%s.%s' to '%v'", trigger, cr.Key, to)
				vars := expr.Spawn(app, c.cfg.ArgoCDService, map[string]interface{}{
					"app":     app.Object,
					"context": legacy.InjectLegacyVar(c.cfg.Context, to.Service),
				})

				if err := c.cfg.API.Send(vars, cr.Templates, to); err != nil {
					logEntry.Errorf("Failed to notify recipient %s defined in app %s/%s: %v",
						to, app.GetNamespace(), app.GetName(), err)
					_ = state.SetAlreadyNotified(trigger, cr, to, false)
					c.metricsRegistry.IncDeliveriesCounter(trigger, to.Service, false)
				} else {
					logEntry.Debugf("Notification %s was sent", to.Recipient)
					c.metricsRegistry.IncDeliveriesCounter(trigger, to.Service, true)
				}
			}
		}
	}

	state.Truncate(notifiedHistoryMaxSize)

	annotations := app.GetAnnotations()

	if len(state) == 0 {
		delete(annotations, notifiedAnnotationKey)
	} else {
		stateJson, err := json.Marshal(state)
		if err != nil {
			return err
		}
		annotations[notifiedAnnotationKey] = string(stateJson)
	}

	app.SetAnnotations(annotations)
	return nil
}

func (c *notificationController) getAppProj(app *unstructured.Unstructured) *unstructured.Unstructured {
	projName, ok, err := unstructured.NestedString(app.Object, "spec", "project")
	if !ok || err != nil {
		return nil
	}
	projObj, ok, err := c.appProjInformer.GetIndexer().GetByKey(fmt.Sprintf("%s/%s", app.GetNamespace(), projName))
	if !ok || err != nil {
		return nil
	}
	proj, ok := projObj.(*unstructured.Unstructured)
	if !ok {
		return nil
	}
	ensureAnnotations(proj)
	return proj
}

func (c *notificationController) getSubscriptions(app *unstructured.Unstructured) pkg.Subscriptions {
	res := c.cfg.GetGlobalSubscriptions(app.GetLabels())

	res.Merge(subscriptions.Annotations(app.GetAnnotations()).GetAll(c.cfg.DefaultTriggers...))
	res.Merge(legacy.GetSubscriptions(app.GetAnnotations(), c.cfg.DefaultTriggers...))

	if proj := c.getAppProj(app); proj != nil {
		res.Merge(subscriptions.Annotations(proj.GetAnnotations()).GetAll(c.cfg.DefaultTriggers...))
		res.Merge(legacy.GetSubscriptions(proj.GetAnnotations(), c.cfg.DefaultTriggers...))
	}

	return res.Dedup()
}

// Checks if the application SyncStatus has been refreshed by Argo CD after an operation has completed
func (c *notificationController) isAppSyncStatusRefreshed(app *unstructured.Unstructured, logEntry *log.Entry) bool {
	_, ok, err := unstructured.NestedMap(app.Object, "status", "operationState")
	if !ok || err != nil {
		logEntry.Debug("No OperationState found, SyncStatus is assumed to be up-to-date")
		return true
	}

	phase, ok, err := unstructured.NestedString(app.Object, "status", "operationState", "phase")
	if !ok || err != nil {
		logEntry.Debug("No OperationPhase found, SyncStatus is assumed to be up-to-date")
		return true
	}
	switch phase {
	case "Failed", "Error", "Succeeded":
		finishedAtRaw, ok, err := unstructured.NestedString(app.Object, "status", "operationState", "finishedAt")
		if !ok || err != nil {
			logEntry.Debugf("No FinishedAt found for completed phase '%s', SyncStatus is assumed to be out-of-date", phase)
			return false
		}
		finishedAt, err := time.Parse(time.RFC3339, finishedAtRaw)
		if err != nil {
			logEntry.Warnf("Failed to parse FinishedAt '%s'", finishedAtRaw)
			return false
		}
		var reconciledAt, observedAt time.Time
		reconciledAtRaw, ok, err := unstructured.NestedString(app.Object, "status", "reconciledAt")
		if ok && err == nil {
			reconciledAt, _ = time.Parse(time.RFC3339, reconciledAtRaw)
		}
		observedAtRaw, ok, err := unstructured.NestedString(app.Object, "status", "observedAt")
		if ok && err == nil {
			observedAt, _ = time.Parse(time.RFC3339, observedAtRaw)
		}
		if finishedAt.After(reconciledAt) && finishedAt.After(observedAt) {
			logEntry.Debugf("SyncStatus out-of-date (FinishedAt=%v, ReconciledAt=%v, Observed=%v", finishedAt, reconciledAt, observedAt)
			return false
		}
		logEntry.Debugf("SyncStatus up-to-date (FinishedAt=%v, ReconciledAt=%v, Observed=%v", finishedAt, reconciledAt, observedAt)
	default:
		logEntry.Debugf("Found phase '%s', SyncStatus is assumed to be up-to-date", phase)
	}

	return true
}

func (c *notificationController) processQueueItem() (processNext bool) {
	key, shutdown := c.refreshQueue.Get()
	if shutdown {
		processNext = false
		return
	}
	processNext = true
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic: %+v\n%s", r, debug.Stack())
		}
		c.refreshQueue.Done(key)
	}()

	obj, exists, err := c.appInformer.GetIndexer().GetByKey(key.(string))
	if err != nil {
		log.Errorf("Failed to get app '%s' from appInformer index: %+v", key, err)
		return
	}
	if !exists {
		// This happens after app was deleted, but the work queue still had an entry for it.
		return
	}
	app, ok := obj.(*unstructured.Unstructured)
	if !ok {
		log.Errorf("Failed to get app '%s' from appInformer index: %+v", key, err)
		return
	}
	appCopy := app.DeepCopy()
	logEntry := log.WithField("app", key)
	logEntry.Info("Start processing")
	subscriptions := c.getSubscriptions(appCopy)
	if len(subscriptions) == 0 {
		return
	}
	if refreshed := c.isAppSyncStatusRefreshed(appCopy, logEntry); !refreshed {
		logEntry.Info("Processing skipped, sync status out of date")
		return
	}
	err = c.processApp(appCopy, logEntry)
	if err != nil {
		logEntry.Errorf("Failed to process: %v", err)
		return
	}
	if !reflect.DeepEqual(app.GetAnnotations(), appCopy.GetAnnotations()) {
		annotationsPatch := make(map[string]interface{})
		for k, v := range appCopy.GetAnnotations() {
			annotationsPatch[k] = v
		}
		for k := range app.GetAnnotations() {
			if _, ok = appCopy.GetAnnotations()[k]; !ok {
				annotationsPatch[k] = nil
			}
		}

		patchData, err := json.Marshal(map[string]map[string]interface{}{
			"metadata": {"annotations": annotationsPatch},
		})
		if err != nil {
			logEntry.Errorf("Failed to marshal app patch: %v", err)
			return
		}
		_, err = c.appClient.Patch(context.Background(), app.GetName(), types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			logEntry.Errorf("Failed to patch app: %v", err)
			return
		}
	}
	logEntry.Info("Processing completed")

	return
}

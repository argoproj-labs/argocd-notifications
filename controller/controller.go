package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod      = 60 * time.Second
	annotationPostfix = "argocd-notifications.argoproj.io"
)

var (
	recipientsAnnotation = "recipients." + annotationPostfix
)

type NotificationController interface {
	Run(ctx context.Context, processors int)
	Init(ctx context.Context) error
}

func NewController(client dynamic.Interface,
	namespace string,
	triggers map[string]triggers.Trigger,
	notifiers map[string]notifiers.Notifier,
	context map[string]string,
	appLabelSelector string,
) (NotificationController, error) {
	appClient := createAppClient(client, namespace)
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
	appProjInformer := newInformer(createAppProjClient(client, namespace), "")

	return &notificationController{
		appClient:       appClient,
		appInformer:     appInformer,
		appProjInformer: appProjInformer,
		refreshQueue:    queue,
		triggers:        triggers,
		notifiers:       notifiers,
		context:         context,
	}, nil
}

func createAppClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	appResource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	resClient := client.Resource(appResource).Namespace(namespace)
	return resClient
}

func newInformer(resClient dynamic.ResourceInterface, selector string) cache.SharedIndexInformer {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (object runtime.Object, err error) {
				options.LabelSelector = selector
				return resClient.List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = selector
				return resClient.Watch(options)
			},
		},
		&unstructured.Unstructured{},
		resyncPeriod,
		cache.Indexers{},
	)
	return informer
}

func createAppProjClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	appResource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "appprojects"}
	resClient := client.Resource(appResource).Namespace(namespace)
	return resClient
}

type notificationController struct {
	appClient       dynamic.ResourceInterface
	appInformer     cache.SharedIndexInformer
	appProjInformer cache.SharedIndexInformer
	refreshQueue    workqueue.RateLimitingInterface
	triggers        map[string]triggers.Trigger
	notifiers       map[string]notifiers.Notifier
	context         map[string]string
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

func (c *notificationController) notify(title string, body string, recipient string) error {
	parts := strings.Split(recipient, ":")
	if len(parts) < 2 {
		return fmt.Errorf("%s is not valid recipient. Expected recipient format is <type>:<name>", recipient)
	}
	notifier, ok := c.notifiers[parts[0]]
	if !ok {
		return fmt.Errorf("%s is not valid recipient type.", parts[0])
	}
	return notifier.Send(title, body, parts[1])
}

func getRecipientsFromAnnotations(annotations map[string]string) []string {
	recipients := make([]string, 0)
	for _, recipient := range strings.Split(annotations[recipientsAnnotation], ",") {
		if recipient = strings.TrimSpace(recipient); recipient != "" {
			recipients = append(recipients, recipient)
		}
	}
	return recipients
}

func (c *notificationController) getRecipients(app *unstructured.Unstructured) map[string]bool {
	recipients := make(map[string]bool)
	if annotations := app.GetAnnotations(); annotations != nil {
		for _, recipient := range getRecipientsFromAnnotations(annotations) {
			recipients[recipient] = true
		}
	}
	projName, ok, err := unstructured.NestedString(app.Object, "spec", "project")
	if !ok && err != nil {
		return recipients
	}
	projObj, ok, err := c.appProjInformer.GetIndexer().GetByKey(fmt.Sprintf("%s/%s", app.GetNamespace(), projName))
	if ok && err != nil {
		return recipients
	}
	proj, ok := projObj.(*unstructured.Unstructured)
	if !ok {
		return recipients
	}
	if annotations := proj.GetAnnotations(); annotations != nil {
		for _, recipient := range getRecipientsFromAnnotations(annotations) {
			recipients[recipient] = true
		}
	}
	return recipients
}

func (c *notificationController) processApp(app *unstructured.Unstructured, logEntry *log.Entry) error {
	refreshed := false
	annotations := app.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for triggerKey, t := range c.triggers {
		triggered, err := t.Triggered(app)
		if err != nil {
			logEntry.Debugf("Failed to execute condition of trigger %: %v", triggerKey, err)
		}
		recipients := c.getRecipients(app)
		triggerAnnotation := fmt.Sprintf("%s.%s", triggerKey, annotationPostfix)
		logEntry.Infof("Trigger %s result: %v", triggerKey, triggered)
		if triggered {
			_, alreadyNotified := annotations[triggerAnnotation]
			// informer might have stale data, so we cannot trust it and should reload app state to avoid sending notification twice
			if !alreadyNotified && !refreshed {
				refreshedApp, err := c.appClient.Get(app.GetName(), v1.GetOptions{})
				if err != nil {
					return err
				}
				if refreshedApp.GetAnnotations() != nil {
					for k, v := range refreshedApp.GetAnnotations() {
						annotations[k] = v
					}
				}
				_, alreadyNotified = annotations[triggerAnnotation]

				refreshed = true
			}

			if !alreadyNotified {
				logEntry.Infof("Sending %s notification", triggerKey)
				title, body, err := t.FormatNotification(app, c.context)
				if err != nil {
					return err
				}
				successful := true
				for recipient := range recipients {
					if err = c.notify(title, body, recipient); err != nil {
						logEntry.Errorf("Failed to notify recipient %s defined in app %s/%s: %v",
							recipient, app.GetNamespace(), app.GetName(), err)
						successful = false
					}
				}
				if successful {
					annotations[triggerAnnotation] = time.Now().Format(time.RFC3339)
				}
			} else {
				logEntry.Infof("%s notification already sent", triggerKey)
			}
		} else {
			delete(annotations, triggerAnnotation)
		}
	}
	app.SetAnnotations(annotations)
	return nil
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
		_, err = c.appClient.Patch(app.GetName(), types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			logEntry.Errorf("Failed to patch app: %v", err)
			return
		}
	}
	logEntry.Info("Processing completed")

	return
}

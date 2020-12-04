package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/shared/clients"
	sharedrecipients "github.com/argoproj-labs/argocd-notifications/shared/recipients"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"

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
	resyncPeriod     = 60 * time.Second
	notificationType = "notificationType"
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
	subscriptions settings.DefaultSubscriptions,
	appLabelSelector string,
	metricsRegistry *controllerRegistry,
) (NotificationController, error) {
	appClient := clients.NewAppClient(client, namespace)
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
	appProjInformer := newInformer(clients.NewAppProjClient(client, namespace), "")

	return &notificationController{
		subscriptions:   subscriptions,
		appClient:       appClient,
		appInformer:     appInformer,
		appProjInformer: appProjInformer,
		refreshQueue:    queue,
		triggers:        triggers,
		notifiers:       notifiers,
		context:         context,
		metricsRegistry: metricsRegistry,
	}, nil
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

type notificationController struct {
	appClient       dynamic.ResourceInterface
	appInformer     cache.SharedIndexInformer
	appProjInformer cache.SharedIndexInformer
	refreshQueue    workqueue.RateLimitingInterface
	triggers        map[string]triggers.Trigger
	notifiers       map[string]notifiers.Notifier
	context         map[string]string
	subscriptions   settings.DefaultSubscriptions
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

func (c *notificationController) getRecipients(app *unstructured.Unstructured, trigger string) map[string]bool {
	recipients := make(map[string]bool)
	for _, r := range c.subscriptions.GetRecipients(trigger, app.GetLabels()) {
		recipients[r] = true
	}
	if annotations := app.GetAnnotations(); annotations != nil {
		for _, recipient := range sharedrecipients.GetRecipientsFromAnnotations(annotations, trigger) {
			recipients[recipient] = true
		}
	}
	projName, ok, err := unstructured.NestedString(app.Object, "spec", "project")
	if !ok || err != nil {
		return recipients
	}
	projObj, ok, err := c.appProjInformer.GetIndexer().GetByKey(fmt.Sprintf("%s/%s", app.GetNamespace(), projName))
	if !ok || err != nil {
		return recipients
	}
	proj, ok := projObj.(*unstructured.Unstructured)
	if !ok {
		return recipients
	}
	if annotations := proj.GetAnnotations(); annotations != nil {
		for _, recipient := range sharedrecipients.GetRecipientsFromAnnotations(annotations, trigger) {
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
			logEntry.Debugf("Failed to execute condition of trigger %s: %v", triggerKey, err)
		}
		recipients := c.getRecipients(app, triggerKey)
		logEntry.Infof("Trigger %s result: %v", triggerKey, triggered)
		c.metricsRegistry.IncTriggerEvaluationsCounter(triggerKey, triggered)
		trackTriggerKey := fmt.Sprintf("%s.%s", triggerKey, sharedrecipients.AnnotationPostfix)
		if !triggered {
			for recipient := range recipients {
				deleteNotifiedTracking(annotations, trackTriggerKey, recipient)
			}
			app.SetAnnotations(annotations)
			continue
		}

		// informer might have stale data, so we cannot trust it and should reload app state to avoid sending notification twice
		if triggered && !refreshed {
			refreshedApp, err := c.appClient.Get(app.GetName(), v1.GetOptions{})
			if err != nil {
				return err
			}
			annotations = refreshedApp.GetAnnotations()
			if annotations == nil {
				annotations = map[string]string{}
			}
			refreshed = true
		}

		for recipient := range recipients {
			if checkAlreadyNotified(annotations, trackTriggerKey, recipient) {
				logEntry.Infof("%s notification already sent", triggerKey)
				continue // move to the next recipient
			}
			successful := true

			parts := strings.Split(recipient, ":")
			if len(parts) < 2 {
				return fmt.Errorf("%s is not valid recipient. Expected recipient format is <type>:<name>", recipient)
			}
			notifierType := parts[0]
			notifier, ok := c.notifiers[notifierType]
			if !ok {
				return fmt.Errorf("%s is not valid recipient type.", notifierType)
			}

			logEntry.Infof("Sending %s notification", triggerKey)
			ctx := sharedrecipients.CopyStringMap(c.context)
			ctx[notificationType] = notifierType
			notification, err := t.FormatNotification(app, ctx)
			if err != nil {
				return err
			}
			if err = notifier.Send(*notification, parts[1]); err != nil {
				logEntry.Errorf("Failed to notify recipient %s defined in app %s/%s: %v",
					recipient, app.GetNamespace(), app.GetName(), err)
				successful = false
				c.metricsRegistry.IncDeliveriesCounter(t.GetTemplateName(), notifierType, false)
			} else {
				c.metricsRegistry.IncDeliveriesCounter(t.GetTemplateName(), notifierType, true)
			}

			if successful {
				logEntry.Debugf("Notification %s was sent", recipient)
				insertNotifiedTracking(annotations, trackTriggerKey, recipient, time.Now().Format(time.RFC3339))
			}
		}

	}
	app.SetAnnotations(annotations)
	return nil
}

func mapToString(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=%s\n", key, value)
	}
	return b.String()
}

//From annotations get recipientTimestampMap by trackTriggerKey. E.g. on-sync-succeeded.argocd-notifications.argoproj.io
/*
metadata:
  annotations:
    on-sync-succeeded.argocd-notifications.argoproj.io: |
      slack:family=2020-11-23T14:29:19-08:00
*/
func getMapByTrackTrigger(annotations map[string]string, trackTriggerKey string) map[string]string {
	recipientTimestampMap := map[string]string{}
	recipientTimestampString, ok := annotations[trackTriggerKey]
	if !ok {
		return recipientTimestampMap
	}

	for _, value := range strings.Split(recipientTimestampString, "\n") {
		parts := strings.Split(value, "=")
		if len(parts) == 2 {
			recipientTimestampMap[parts[0]] = parts[1]
		}
	}
	return recipientTimestampMap
}

func checkAlreadyNotified(annotations map[string]string, trackTriggerKey string, recipient string) bool {
	recipientTimestampMap := getMapByTrackTrigger(annotations, trackTriggerKey)
	_, ok := recipientTimestampMap[recipient]
	return ok
}

func insertNotifiedTracking(annotations map[string]string, trackTriggerKey string, recipient string, recipientTimestamp string) {
	recipientTimestampMap := getMapByTrackTrigger(annotations, trackTriggerKey)
	recipientTimestampMap[recipient] = recipientTimestamp
	annotations[trackTriggerKey] = mapToString(recipientTimestampMap)
}

func deleteNotifiedTracking(annotations map[string]string, trackTriggerKey string, recipient string) {
	recipientTimestampMap := getMapByTrackTrigger(annotations, trackTriggerKey)
	_, ok := recipientTimestampMap[recipient]
	if ok {
		delete(recipientTimestampMap, recipient)
		recipientTimestampString := mapToString(recipientTimestampMap)
		if recipientTimestampString == "" {
			delete(annotations, trackTriggerKey)
		} else {
			annotations[trackTriggerKey] = recipientTimestampString
		}
	}
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
		_, err = c.appClient.Patch(app.GetName(), types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			logEntry.Errorf("Failed to patch app: %v", err)
			return
		}
	}
	logEntry.Info("Processing completed")

	return
}

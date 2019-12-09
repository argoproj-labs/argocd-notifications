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

	"k8s.io/apimachinery/pkg/types"

	"github.com/argoproj-labs/argocd-notifications/notifiers"

	"github.com/argoproj-labs/argocd-notifications/triggers"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

type Config struct {
	Triggers  []triggers.NotificationTrigger  `json:"triggers"`
	Templates []triggers.NotificationTemplate `json:"templates"`
	Context   map[string]string               `json:"context"`
}

type NotificationController interface {
	Run(ctx context.Context, processors int) error
}

func NewController(client dynamic.Interface,
	namespace string,
	triggers map[string]triggers.Trigger,
	notifiers map[string]notifiers.Notifier,
	context map[string]string,
) (NotificationController, error) {
	resource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	resClient := client.Resource(resource).Namespace(namespace)
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (object runtime.Object, err error) {
				return resClient.List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return resClient.Watch(options)
			},
		},
		&unstructured.Unstructured{},
		resyncPeriod,
		cache.Indexers{},
	)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(
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
	return &notificationController{
		resClient:    resClient,
		informer:     informer,
		refreshQueue: queue,
		triggers:     triggers,
		notifiers:    notifiers,
		context:      context,
	}, nil
}

type notificationController struct {
	resClient    dynamic.ResourceInterface
	informer     cache.SharedIndexInformer
	refreshQueue workqueue.RateLimitingInterface
	triggers     map[string]triggers.Trigger
	notifiers    map[string]notifiers.Notifier
	context      map[string]string
}

func (c *notificationController) Run(ctx context.Context, processors int) error {
	defer runtimeutil.HandleCrash()
	defer c.refreshQueue.ShutDown()
	go c.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return errors.New("Timed out waiting for caches to sync")
	}
	for i := 0; i < processors; i++ {
		go wait.Until(func() {
			for c.processQueueItem() {
			}
		}, time.Second, ctx.Done())
	}
	<-ctx.Done()
	return nil
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

func (c *notificationController) processApp(app *unstructured.Unstructured) error {
	annotations := app.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for triggerKey, t := range c.triggers {
		triggered, err := t.Triggered(app)
		if err != nil {
			log.Debug("Failed to execute condition of trigger %s for app %s/%s: %+v", triggerKey, app.GetNamespace(), app.GetName(), err)
		}
		triggerAnnotation := fmt.Sprintf("%s.%s", triggerKey, annotationPostfix)
		if triggered {
			if _, alreadyNotified := annotations[triggerAnnotation]; !alreadyNotified {
				title, body, err := t.FormatNotification(app, c.context)
				if err != nil {
					return err
				}
				successful := true
				for _, recipient := range strings.Split(annotations[recipientsAnnotation], ",") {
					if recipient = strings.TrimSpace(recipient); recipient != "" {
						if err = c.notify(title, body, recipient); err != nil {
							log.Errorf("Failed to notify recipient %s defined in app %s/%s: %v", recipient, app.GetNamespace(), app.GetName(), err)
							successful = false
						}
					}
				}
				if successful {
					annotations[triggerAnnotation] = time.Now().Format(time.RFC3339)
				}
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

	obj, exists, err := c.informer.GetIndexer().GetByKey(key.(string))
	if err != nil {
		log.Errorf("Failed to get app '%s' from informer index: %+v", key, err)
		return
	}
	if !exists {
		// This happens after app was deleted, but the work queue still had an entry for it.
		return
	}
	app, ok := obj.(*unstructured.Unstructured)
	if !ok {
		log.Errorf("Failed to get app '%s' from informer index: %+v", key, err)
		return
	}
	appCopy := app.DeepCopy()
	err = c.processApp(appCopy)
	if err != nil {
		log.Errorf("Failed to process app '%s': %+v", key, err)
		return
	}
	if !reflect.DeepEqual(app.GetAnnotations(), appCopy.GetAnnotations()) {
		patchData, err := json.Marshal(map[string]map[string]interface{}{
			"metadata": {"annotations": appCopy.GetAnnotations()},
		})
		if err != nil {
			log.Errorf("Failed to marshal app '%s' patch: %+v", key, err)
			return
		}
		_, err = c.resClient.Patch(app.GetName(), types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			log.Errorf("Failed to patch app '%s': %+v", key, err)
			return
		}
	}

	return
}

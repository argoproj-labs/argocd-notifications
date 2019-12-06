package controller

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

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
	resyncPeriod = 60 * time.Second
)

func NewController(client dynamic.Interface, namespace string, config Config) NotificationController {
	resource := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (object runtime.Object, err error) {
				return client.Resource(resource).Namespace(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.Resource(resource).Namespace(namespace).Watch(options)
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
		client:       client,
		informer:     informer,
		refreshQueue: queue,
		config:       config,
	}
}

type notificationController struct {
	client       dynamic.Interface
	informer     cache.SharedIndexInformer
	refreshQueue workqueue.RateLimitingInterface
	config       Config
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

	app, exists, err := c.informer.GetIndexer().GetByKey(key.(string))
	if err != nil {
		log.Errorf("Failed to get app '%s' from informer index: %+v", key, err)
		return
	}
	if !exists {
		// This happens after app was deleted, but the work queue still had an entry for it.
		return
	}
	println(fmt.Sprintf("%v", app))
	return
}

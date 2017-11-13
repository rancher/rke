package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	resyncPeriod = 5 * time.Minute
)

type HandlerFunc func(key string) error

type GenericController interface {
	Informer() cache.SharedIndexInformer
	AddHandler(handler HandlerFunc)
	Enqueue(namespace, name string)
	Start(threadiness int, ctx context.Context) error
}

type genericController struct {
	sync.Mutex
	informer cache.SharedIndexInformer
	handlers []HandlerFunc
	queue    workqueue.RateLimitingInterface
	name     string
	running  bool
}

func NewGenericController(name string, objectClient *clientbase.ObjectClient) GenericController {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  objectClient.List,
			WatchFunc: objectClient.Watch,
		},
		objectClient.Factory.Object(), resyncPeriod, cache.Indexers{})

	return &genericController{
		informer: informer,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			name),
		name: name,
	}
}

func (g *genericController) Informer() cache.SharedIndexInformer {
	return g.informer
}

func (g *genericController) Enqueue(namespace, name string) {
	if namespace == "" {
		g.queue.Add(name)
	} else {
		g.queue.Add(namespace + "/" + name)
	}
}

func (g *genericController) AddHandler(handler HandlerFunc) {
	g.handlers = append(g.handlers, handler)
}

func (g *genericController) Start(threadiness int, ctx context.Context) error {
	g.Lock()
	defer g.Unlock()

	if !g.running {
		go g.run(threadiness, ctx)
	}

	g.running = true
	return nil
}

func (g *genericController) queueObject(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		g.queue.Add(key)
	}
}

func (g *genericController) run(threadiness int, ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer g.queue.ShutDown()

	g.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: g.queueObject,
		UpdateFunc: func(_, obj interface{}) {
			g.queueObject(obj)
		},
		DeleteFunc: g.queueObject,
	})

	logrus.Infof("Starting %s Controller", g.name)

	go g.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), g.informer.HasSynced) {
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(g.runWorker, time.Second, ctx.Done())
	}

	<-ctx.Done()
	logrus.Infof("Shutting down %s controller", g.name)
}

func (g *genericController) runWorker() {
	for g.processNextWorkItem() {
	}
}

func (g *genericController) processNextWorkItem() bool {
	key, quit := g.queue.Get()
	if quit {
		return false
	}
	defer g.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := g.syncHandler(key.(string))
	if err == nil {
		g.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))
	g.queue.AddRateLimited(key)

	return true
}

func (g *genericController) syncHandler(s string) error {
	var errs []error
	for _, handler := range g.handlers {
		if err := handler(s); err != nil {
			errs = append(errs, err)
		}
	}
	return types.NewErrors(errs)
}

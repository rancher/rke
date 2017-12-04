package v1

import (
	"context"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	ClusterNodeGroupVersionKind = schema.GroupVersionKind{
		Version: "v1",
		Group:   "cluster.cattle.io",
		Kind:    "ClusterNode",
	}
	ClusterNodeResource = metav1.APIResource{
		Name:         "clusternodes",
		SingularName: "clusternode",
		Namespaced:   false,
		Kind:         ClusterNodeGroupVersionKind.Kind,
	}
)

type ClusterNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterNode
}

type ClusterNodeHandlerFunc func(key string, obj *ClusterNode) error

type ClusterNodeLister interface {
	List(namespace string, selector labels.Selector) (ret []*ClusterNode, err error)
	Get(namespace, name string) (*ClusterNode, error)
}

type ClusterNodeController interface {
	Informer() cache.SharedIndexInformer
	Lister() ClusterNodeLister
	AddHandler(handler ClusterNodeHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ClusterNodeInterface interface {
	ObjectClient() *clientbase.ObjectClient
	Create(*ClusterNode) (*ClusterNode, error)
	Get(name string, opts metav1.GetOptions) (*ClusterNode, error)
	Update(*ClusterNode) (*ClusterNode, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterNodeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterNodeController
}

type clusterNodeLister struct {
	controller *clusterNodeController
}

func (l *clusterNodeLister) List(namespace string, selector labels.Selector) (ret []*ClusterNode, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ClusterNode))
	})
	return
}

func (l *clusterNodeLister) Get(namespace, name string) (*ClusterNode, error) {
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    ClusterNodeGroupVersionKind.Group,
			Resource: "clusterNode",
		}, name)
	}
	return obj.(*ClusterNode), nil
}

type clusterNodeController struct {
	controller.GenericController
}

func (c *clusterNodeController) Lister() ClusterNodeLister {
	return &clusterNodeLister{
		controller: c,
	}
}

func (c *clusterNodeController) AddHandler(handler ClusterNodeHandlerFunc) {
	c.GenericController.AddHandler(func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*ClusterNode))
	})
}

type clusterNodeFactory struct {
}

func (c clusterNodeFactory) Object() runtime.Object {
	return &ClusterNode{}
}

func (c clusterNodeFactory) List() runtime.Object {
	return &ClusterNodeList{}
}

func (s *clusterNodeClient) Controller() ClusterNodeController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterNodeControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterNodeGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterNodeController{
		GenericController: genericController,
	}

	s.client.clusterNodeControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type clusterNodeClient struct {
	client       *Client
	ns           string
	objectClient *clientbase.ObjectClient
	controller   ClusterNodeController
}

func (s *clusterNodeClient) ObjectClient() *clientbase.ObjectClient {
	return s.objectClient
}

func (s *clusterNodeClient) Create(o *ClusterNode) (*ClusterNode, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ClusterNode), err
}

func (s *clusterNodeClient) Get(name string, opts metav1.GetOptions) (*ClusterNode, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ClusterNode), err
}

func (s *clusterNodeClient) Update(o *ClusterNode) (*ClusterNode, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ClusterNode), err
}

func (s *clusterNodeClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterNodeClient) List(opts metav1.ListOptions) (*ClusterNodeList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterNodeList), err
}

func (s *clusterNodeClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

func (s *clusterNodeClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

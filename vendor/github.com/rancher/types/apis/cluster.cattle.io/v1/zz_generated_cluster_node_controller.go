package v1

import (
	"context"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type ClusterNodeController interface {
	Informer() cache.SharedIndexInformer
	AddHandler(handler ClusterNodeHandlerFunc)
	Enqueue(namespace, name string)
	Start(ctx context.Context, threadiness int) error
}

type ClusterNodeInterface interface {
	Create(*ClusterNode) (*ClusterNode, error)
	Get(name string, opts metav1.GetOptions) (*ClusterNode, error)
	Update(*ClusterNode) (*ClusterNode, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterNodeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterNodeController
}

type clusterNodeController struct {
	controller.GenericController
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

	return c
}

type clusterNodeClient struct {
	client       *Client
	ns           string
	objectClient *clientbase.ObjectClient
	controller   ClusterNodeController
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

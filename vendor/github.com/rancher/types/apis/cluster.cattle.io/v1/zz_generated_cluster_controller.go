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
	ClusterGroupVersionKind = schema.GroupVersionKind{
		Version: "v1",
		Group:   "cluster.cattle.io",
		Kind:    "Cluster",
	}
	ClusterResource = metav1.APIResource{
		Name:         "clusters",
		SingularName: "cluster",
		Namespaced:   false,
		Kind:         ClusterGroupVersionKind.Kind,
	}
)

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster
}

type ClusterHandlerFunc func(key string, obj *Cluster) error

type ClusterController interface {
	Informer() cache.SharedIndexInformer
	AddHandler(handler ClusterHandlerFunc)
	Enqueue(namespace, name string)
	Start(ctx context.Context, threadiness int) error
}

type ClusterInterface interface {
	Create(*Cluster) (*Cluster, error)
	Get(name string, opts metav1.GetOptions) (*Cluster, error)
	Update(*Cluster) (*Cluster, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterController
}

type clusterController struct {
	controller.GenericController
}

func (c *clusterController) AddHandler(handler ClusterHandlerFunc) {
	c.GenericController.AddHandler(func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*Cluster))
	})
}

type clusterFactory struct {
}

func (c clusterFactory) Object() runtime.Object {
	return &Cluster{}
}

func (c clusterFactory) List() runtime.Object {
	return &ClusterList{}
}

func (s *clusterClient) Controller() ClusterController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterController{
		GenericController: genericController,
	}

	s.client.clusterControllers[s.ns] = c

	return c
}

type clusterClient struct {
	client       *Client
	ns           string
	objectClient *clientbase.ObjectClient
	controller   ClusterController
}

func (s *clusterClient) Create(o *Cluster) (*Cluster, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Cluster), err
}

func (s *clusterClient) Get(name string, opts metav1.GetOptions) (*Cluster, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Cluster), err
}

func (s *clusterClient) Update(o *Cluster) (*Cluster, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Cluster), err
}

func (s *clusterClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterClient) List(opts metav1.ListOptions) (*ClusterList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterList), err
}

func (s *clusterClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

func (s *clusterClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

package v3

import (
	"context"
	"sync"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	MachinesGetter
	MachineDriversGetter
	MachineTemplatesGetter
	ProjectsGetter
	ProjectRoleTemplatesGetter
	PodSecurityPolicyTemplatesGetter
	ClusterRoleTemplatesGetter
	ClusterRoleTemplateBindingsGetter
	ProjectRoleTemplateBindingsGetter
	ClustersGetter
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	machineControllers                    map[string]MachineController
	machineDriverControllers              map[string]MachineDriverController
	machineTemplateControllers            map[string]MachineTemplateController
	projectControllers                    map[string]ProjectController
	projectRoleTemplateControllers        map[string]ProjectRoleTemplateController
	podSecurityPolicyTemplateControllers  map[string]PodSecurityPolicyTemplateController
	clusterRoleTemplateControllers        map[string]ClusterRoleTemplateController
	clusterRoleTemplateBindingControllers map[string]ClusterRoleTemplateBindingController
	projectRoleTemplateBindingControllers map[string]ProjectRoleTemplateBindingController
	clusterControllers                    map[string]ClusterController
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		configConfig := dynamic.ContentConfig()
		config.NegotiatedSerializer = configConfig.NegotiatedSerializer
	}

	restClient, err := rest.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient: restClient,

		machineControllers:                    map[string]MachineController{},
		machineDriverControllers:              map[string]MachineDriverController{},
		machineTemplateControllers:            map[string]MachineTemplateController{},
		projectControllers:                    map[string]ProjectController{},
		projectRoleTemplateControllers:        map[string]ProjectRoleTemplateController{},
		podSecurityPolicyTemplateControllers:  map[string]PodSecurityPolicyTemplateController{},
		clusterRoleTemplateControllers:        map[string]ClusterRoleTemplateController{},
		clusterRoleTemplateBindingControllers: map[string]ClusterRoleTemplateBindingController{},
		projectRoleTemplateBindingControllers: map[string]ProjectRoleTemplateBindingController{},
		clusterControllers:                    map[string]ClusterController{},
	}, nil
}

func (c *Client) RESTClient() rest.Interface {
	return c.restClient
}

func (c *Client) Sync(ctx context.Context) error {
	return controller.Sync(ctx, c.starters...)
}

func (c *Client) Start(ctx context.Context, threadiness int) error {
	return controller.Start(ctx, threadiness, c.starters...)
}

type MachinesGetter interface {
	Machines(namespace string) MachineInterface
}

func (c *Client) Machines(namespace string) MachineInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &MachineResource, MachineGroupVersionKind, machineFactory{})
	return &machineClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type MachineDriversGetter interface {
	MachineDrivers(namespace string) MachineDriverInterface
}

func (c *Client) MachineDrivers(namespace string) MachineDriverInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &MachineDriverResource, MachineDriverGroupVersionKind, machineDriverFactory{})
	return &machineDriverClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type MachineTemplatesGetter interface {
	MachineTemplates(namespace string) MachineTemplateInterface
}

func (c *Client) MachineTemplates(namespace string) MachineTemplateInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &MachineTemplateResource, MachineTemplateGroupVersionKind, machineTemplateFactory{})
	return &machineTemplateClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ProjectsGetter interface {
	Projects(namespace string) ProjectInterface
}

func (c *Client) Projects(namespace string) ProjectInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ProjectResource, ProjectGroupVersionKind, projectFactory{})
	return &projectClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ProjectRoleTemplatesGetter interface {
	ProjectRoleTemplates(namespace string) ProjectRoleTemplateInterface
}

func (c *Client) ProjectRoleTemplates(namespace string) ProjectRoleTemplateInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ProjectRoleTemplateResource, ProjectRoleTemplateGroupVersionKind, projectRoleTemplateFactory{})
	return &projectRoleTemplateClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type PodSecurityPolicyTemplatesGetter interface {
	PodSecurityPolicyTemplates(namespace string) PodSecurityPolicyTemplateInterface
}

func (c *Client) PodSecurityPolicyTemplates(namespace string) PodSecurityPolicyTemplateInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &PodSecurityPolicyTemplateResource, PodSecurityPolicyTemplateGroupVersionKind, podSecurityPolicyTemplateFactory{})
	return &podSecurityPolicyTemplateClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ClusterRoleTemplatesGetter interface {
	ClusterRoleTemplates(namespace string) ClusterRoleTemplateInterface
}

func (c *Client) ClusterRoleTemplates(namespace string) ClusterRoleTemplateInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ClusterRoleTemplateResource, ClusterRoleTemplateGroupVersionKind, clusterRoleTemplateFactory{})
	return &clusterRoleTemplateClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ClusterRoleTemplateBindingsGetter interface {
	ClusterRoleTemplateBindings(namespace string) ClusterRoleTemplateBindingInterface
}

func (c *Client) ClusterRoleTemplateBindings(namespace string) ClusterRoleTemplateBindingInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ClusterRoleTemplateBindingResource, ClusterRoleTemplateBindingGroupVersionKind, clusterRoleTemplateBindingFactory{})
	return &clusterRoleTemplateBindingClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ProjectRoleTemplateBindingsGetter interface {
	ProjectRoleTemplateBindings(namespace string) ProjectRoleTemplateBindingInterface
}

func (c *Client) ProjectRoleTemplateBindings(namespace string) ProjectRoleTemplateBindingInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ProjectRoleTemplateBindingResource, ProjectRoleTemplateBindingGroupVersionKind, projectRoleTemplateBindingFactory{})
	return &projectRoleTemplateBindingClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ClustersGetter interface {
	Clusters(namespace string) ClusterInterface
}

func (c *Client) Clusters(namespace string) ClusterInterface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &ClusterResource, ClusterGroupVersionKind, clusterFactory{})
	return &clusterClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

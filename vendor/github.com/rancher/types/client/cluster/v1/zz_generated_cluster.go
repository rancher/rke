package client

import (
	"github.com/rancher/norman/types"
)

const (
	ClusterType             = "cluster"
	ClusterFieldAKSConfig   = "aksConfig"
	ClusterFieldAPIVersion  = "apiVersion"
	ClusterFieldAnnotations = "annotations"
	ClusterFieldCreated     = "created"
	ClusterFieldGKEConfig   = "gkeConfig"
	ClusterFieldKind        = "kind"
	ClusterFieldLabels      = "labels"
	ClusterFieldName        = "name"
	ClusterFieldNamespace   = "namespace"
	ClusterFieldRKEConfig   = "rkeConfig"
	ClusterFieldRemoved     = "removed"
	ClusterFieldUuid        = "uuid"
)

type Cluster struct {
	types.Resource
	AKSConfig   *AKSConfig        `json:"aksConfig,omitempty"`
	APIVersion  string            `json:"apiVersion,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Created     string            `json:"created,omitempty"`
	GKEConfig   *GKEConfig        `json:"gkeConfig,omitempty"`
	Kind        string            `json:"kind,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	RKEConfig   *RKEConfig        `json:"rkeConfig,omitempty"`
	Removed     string            `json:"removed,omitempty"`
	Uuid        string            `json:"uuid,omitempty"`
}
type ClusterCollection struct {
	types.Collection
	Data   []Cluster `json:"data,omitempty"`
	client *ClusterClient
}

type ClusterClient struct {
	apiClient *Client
}

type ClusterOperations interface {
	List(opts *types.ListOpts) (*ClusterCollection, error)
	Create(opts *Cluster) (*Cluster, error)
	Update(existing *Cluster, updates interface{}) (*Cluster, error)
	ByID(id string) (*Cluster, error)
	Delete(container *Cluster) error
}

func newClusterClient(apiClient *Client) *ClusterClient {
	return &ClusterClient{
		apiClient: apiClient,
	}
}

func (c *ClusterClient) Create(container *Cluster) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoCreate(ClusterType, container, resp)
	return resp, err
}

func (c *ClusterClient) Update(existing *Cluster, updates interface{}) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoUpdate(ClusterType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterClient) List(opts *types.ListOpts) (*ClusterCollection, error) {
	resp := &ClusterCollection{}
	err := c.apiClient.Ops.DoList(ClusterType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ClusterCollection) Next() (*ClusterCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterClient) ByID(id string) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoByID(ClusterType, id, resp)
	return resp, err
}

func (c *ClusterClient) Delete(container *Cluster) error {
	return c.apiClient.Ops.DoResourceDelete(ClusterType, &container.Resource)
}

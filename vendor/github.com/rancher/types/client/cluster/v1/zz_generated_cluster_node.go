package client

import (
	"github.com/rancher/norman/types"
)

const (
	ClusterNodeType               = "clusterNode"
	ClusterNodeFieldAPIVersion    = "apiVersion"
	ClusterNodeFieldAnnotations   = "annotations"
	ClusterNodeFieldConfigSource  = "configSource"
	ClusterNodeFieldCreated       = "created"
	ClusterNodeFieldExternalID    = "externalID"
	ClusterNodeFieldKind          = "kind"
	ClusterNodeFieldLabels        = "labels"
	ClusterNodeFieldName          = "name"
	ClusterNodeFieldNamespace     = "namespace"
	ClusterNodeFieldPodCIDR       = "podCIDR"
	ClusterNodeFieldProviderID    = "providerID"
	ClusterNodeFieldRemoved       = "removed"
	ClusterNodeFieldTaints        = "taints"
	ClusterNodeFieldUnschedulable = "unschedulable"
	ClusterNodeFieldUuid          = "uuid"
)

type ClusterNode struct {
	types.Resource
	APIVersion    string            `json:"apiVersion,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	ConfigSource  *NodeConfigSource `json:"configSource,omitempty"`
	Created       string            `json:"created,omitempty"`
	ExternalID    string            `json:"externalID,omitempty"`
	Kind          string            `json:"kind,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Name          string            `json:"name,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	PodCIDR       string            `json:"podCIDR,omitempty"`
	ProviderID    string            `json:"providerID,omitempty"`
	Removed       string            `json:"removed,omitempty"`
	Taints        []Taint           `json:"taints,omitempty"`
	Unschedulable bool              `json:"unschedulable,omitempty"`
	Uuid          string            `json:"uuid,omitempty"`
}
type ClusterNodeCollection struct {
	types.Collection
	Data   []ClusterNode `json:"data,omitempty"`
	client *ClusterNodeClient
}

type ClusterNodeClient struct {
	apiClient *Client
}

type ClusterNodeOperations interface {
	List(opts *types.ListOpts) (*ClusterNodeCollection, error)
	Create(opts *ClusterNode) (*ClusterNode, error)
	Update(existing *ClusterNode, updates interface{}) (*ClusterNode, error)
	ByID(id string) (*ClusterNode, error)
	Delete(container *ClusterNode) error
}

func newClusterNodeClient(apiClient *Client) *ClusterNodeClient {
	return &ClusterNodeClient{
		apiClient: apiClient,
	}
}

func (c *ClusterNodeClient) Create(container *ClusterNode) (*ClusterNode, error) {
	resp := &ClusterNode{}
	err := c.apiClient.Ops.DoCreate(ClusterNodeType, container, resp)
	return resp, err
}

func (c *ClusterNodeClient) Update(existing *ClusterNode, updates interface{}) (*ClusterNode, error) {
	resp := &ClusterNode{}
	err := c.apiClient.Ops.DoUpdate(ClusterNodeType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterNodeClient) List(opts *types.ListOpts) (*ClusterNodeCollection, error) {
	resp := &ClusterNodeCollection{}
	err := c.apiClient.Ops.DoList(ClusterNodeType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ClusterNodeCollection) Next() (*ClusterNodeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterNodeCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterNodeClient) ByID(id string) (*ClusterNode, error) {
	resp := &ClusterNode{}
	err := c.apiClient.Ops.DoByID(ClusterNodeType, id, resp)
	return resp, err
}

func (c *ClusterNodeClient) Delete(container *ClusterNode) error {
	return c.apiClient.Ops.DoResourceDelete(ClusterNodeType, &container.Resource)
}

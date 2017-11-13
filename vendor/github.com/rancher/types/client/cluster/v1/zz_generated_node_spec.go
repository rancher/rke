package client

const (
	NodeSpecType               = "nodeSpec"
	NodeSpecFieldConfigSource  = "configSource"
	NodeSpecFieldExternalID    = "externalID"
	NodeSpecFieldPodCIDR       = "podCIDR"
	NodeSpecFieldProviderID    = "providerID"
	NodeSpecFieldTaints        = "taints"
	NodeSpecFieldUnschedulable = "unschedulable"
)

type NodeSpec struct {
	ConfigSource  *NodeConfigSource `json:"configSource,omitempty"`
	ExternalID    string            `json:"externalID,omitempty"`
	PodCIDR       string            `json:"podCIDR,omitempty"`
	ProviderID    string            `json:"providerID,omitempty"`
	Taints        []Taint           `json:"taints,omitempty"`
	Unschedulable bool              `json:"unschedulable,omitempty"`
}

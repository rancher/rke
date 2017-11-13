package client

const (
	NodeStatusType                 = "nodeStatus"
	NodeStatusFieldAddresses       = "addresses"
	NodeStatusFieldAllocatable     = "allocatable"
	NodeStatusFieldCapacity        = "capacity"
	NodeStatusFieldConditions      = "conditions"
	NodeStatusFieldDaemonEndpoints = "daemonEndpoints"
	NodeStatusFieldImages          = "images"
	NodeStatusFieldNodeInfo        = "nodeInfo"
	NodeStatusFieldPhase           = "phase"
	NodeStatusFieldVolumesAttached = "volumesAttached"
	NodeStatusFieldVolumesInUse    = "volumesInUse"
)

type NodeStatus struct {
	Addresses       []NodeAddress       `json:"addresses,omitempty"`
	Allocatable     map[string]string   `json:"allocatable,omitempty"`
	Capacity        map[string]string   `json:"capacity,omitempty"`
	Conditions      []NodeCondition     `json:"conditions,omitempty"`
	DaemonEndpoints NodeDaemonEndpoints `json:"daemonEndpoints,omitempty"`
	Images          []ContainerImage    `json:"images,omitempty"`
	NodeInfo        NodeSystemInfo      `json:"nodeInfo,omitempty"`
	Phase           string              `json:"phase,omitempty"`
	VolumesAttached []AttachedVolume    `json:"volumesAttached,omitempty"`
	VolumesInUse    []string            `json:"volumesInUse,omitempty"`
}

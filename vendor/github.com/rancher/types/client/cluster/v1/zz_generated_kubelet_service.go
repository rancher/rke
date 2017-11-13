package client

const (
	KubeletServiceType                     = "kubeletService"
	KubeletServiceFieldClusterDNSServer    = "clusterDNSServer"
	KubeletServiceFieldClusterDomain       = "clusterDomain"
	KubeletServiceFieldInfraContainerImage = "infraContainerImage"
)

type KubeletService struct {
	ClusterDNSServer    string `json:"clusterDNSServer,omitempty"`
	ClusterDomain       string `json:"clusterDomain,omitempty"`
	InfraContainerImage string `json:"infraContainerImage,omitempty"`
}

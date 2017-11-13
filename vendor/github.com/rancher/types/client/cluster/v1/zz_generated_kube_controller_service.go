package client

const (
	KubeControllerServiceType                       = "kubeControllerService"
	KubeControllerServiceFieldClusterCIDR           = "clusterCIDR"
	KubeControllerServiceFieldServiceClusterIPRange = "serviceClusterIPRange"
)

type KubeControllerService struct {
	ClusterCIDR           string `json:"clusterCIDR,omitempty"`
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
}

package client

const (
	KubeAPIServiceType                       = "kubeAPIService"
	KubeAPIServiceFieldServiceClusterIPRange = "serviceClusterIPRange"
)

type KubeAPIService struct {
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
}

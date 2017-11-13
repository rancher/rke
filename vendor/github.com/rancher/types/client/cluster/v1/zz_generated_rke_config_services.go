package client

const (
	RKEConfigServicesType                = "rkeConfigServices"
	RKEConfigServicesFieldEtcd           = "etcd"
	RKEConfigServicesFieldKubeAPI        = "kubeAPI"
	RKEConfigServicesFieldKubeController = "kubeController"
	RKEConfigServicesFieldKubelet        = "kubelet"
	RKEConfigServicesFieldKubeproxy      = "kubeproxy"
	RKEConfigServicesFieldScheduler      = "scheduler"
)

type RKEConfigServices struct {
	Etcd           ETCDService           `json:"etcd,omitempty"`
	KubeAPI        KubeAPIService        `json:"kubeAPI,omitempty"`
	KubeController KubeControllerService `json:"kubeController,omitempty"`
	Kubelet        KubeletService        `json:"kubelet,omitempty"`
	Kubeproxy      KubeproxyService      `json:"kubeproxy,omitempty"`
	Scheduler      SchedulerService      `json:"scheduler,omitempty"`
}

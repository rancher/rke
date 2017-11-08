package services

type Services struct {
	Etcd           Etcd           `yaml:"etcd"`
	KubeAPI        KubeAPI        `yaml:"kube-api"`
	KubeController KubeController `yaml:"kube-controller"`
	Scheduler      Scheduler      `yaml:"scheduler"`
	Kubelet        Kubelet        `yaml:"kubelet"`
	Kubeproxy      Kubeproxy      `yaml:"kubeproxy"`
}

type Etcd struct {
	Image string `yaml:"image"`
}

type KubeAPI struct {
	Image                 string `yaml:"image"`
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range"`
}

type KubeController struct {
	Image                 string `yaml:"image"`
	ClusterCIDR           string `yaml:"cluster_cidr"`
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range"`
}

type Kubelet struct {
	Image               string `yaml:"image"`
	ClusterDomain       string `yaml:"cluster_domain"`
	InfraContainerImage string `yaml:"infra_container_image"`
	ClusterDNSServer    string `yaml:"cluster_dns_server"`
}

type Kubeproxy struct {
	Image string `yaml:"image"`
}

type Scheduler struct {
	Image string `yaml:"image"`
}

package services

import (
	"fmt"
	"net"
)

type Container struct {
	Services Services `yaml:"services"`
}

type Services struct {
	Etcd           Etcd           `yaml:"etcd"`
	KubeAPI        KubeAPI        `yaml:"kube-api"`
	KubeController KubeController `yaml:"kube-controller"`
	Scheduler      Scheduler      `yaml:"scheduler"`
	Kubelet        Kubelet        `yaml:"kubelet"`
	Kubeproxy      Kubeproxy      `yaml:"kubeproxy"`
}

const (
	ETCDRole    = "etcd"
	ControlRole = "controlplane"
	WorkerRole  = "worker"

	KubeAPIContainerName        = "kube-api"
	KubeletContainerName        = "kubelet"
	KubeproxyContainerName      = "kube-proxy"
	KubeControllerContainerName = "kube-controller"
	SchedulerContainerName      = "scheduler"
	EtcdContainerName           = "etcd"
)

func GetKubernetesServiceIp(serviceClusterRange string) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(serviceClusterRange)
	if err != nil {
		return nil, fmt.Errorf("Failed to get kubernetes service IP: %v", err)
	}
	ip = ip.Mask(ipnet.Mask)
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return ip, nil
}

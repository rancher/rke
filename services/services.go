package services

import (
	"fmt"
	"net"
)

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
	NginxProxyContainerName     = "nginx-proxy"
)

func GetKubernetesServiceIP(serviceClusterRange string) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(serviceClusterRange)
	if err != nil {
		return nil, fmt.Errorf("Failed to get kubernetes service IP from Kube API option [service_cluster_ip_range]: %v", err)
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

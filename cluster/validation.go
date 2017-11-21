package cluster

import (
	"fmt"
	"strings"
)

func (c *Cluster) ValidateCluster() error {
	// make sure cluster has at least one controlplane/etcd host
	if len(c.ControlPlaneHosts) == 0 {
		return fmt.Errorf("Cluster must have at least one control plane host")
	}
	if len(c.EtcdHosts) == 0 {
		return fmt.Errorf("Cluster must have at least one etcd host")
	}

	// validate services options
	err := validateServicesOption(c)
	if err != nil {
		return err
	}
	return nil
}

func validateServicesOption(c *Cluster) error {
	servicesOptions := map[string]string{
		"etcd_image":                               c.Services.Etcd.Image,
		"kube_api_image":                           c.Services.KubeAPI.Image,
		"kube_api_service_cluster_ip_range":        c.Services.KubeAPI.ServiceClusterIPRange,
		"kube_controller_image":                    c.Services.KubeController.Image,
		"kube_controller_service_cluster_ip_range": c.Services.KubeController.ServiceClusterIPRange,
		"kube_controller_cluster_cidr":             c.Services.KubeController.ClusterCIDR,
		"scheduler_image":                          c.Services.Scheduler.Image,
		"kubelet_image":                            c.Services.Kubelet.Image,
		"kubelet_cluster_dns_service":              c.Services.Kubelet.ClusterDNSServer,
		"kubelet_cluster_domain":                   c.Services.Kubelet.ClusterDomain,
		"kubelet_infra_container_image":            c.Services.Kubelet.InfraContainerImage,
		"kubeproxy_image":                          c.Services.Kubeproxy.Image,
	}
	for optionName, OptionValue := range servicesOptions {
		if len(OptionValue) == 0 {
			return fmt.Errorf("%s can't be empty", strings.Join(strings.Split(optionName, "_"), " "))
		}
	}
	return nil
}

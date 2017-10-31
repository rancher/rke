package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

type KubeController struct {
	Version               string `yaml:"version"`
	Image                 string `yaml:"image"`
	ClusterCIDR           string `yaml:"cluster_cider"`
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range"`
}

func runKubeController(host hosts.Host, kubeControllerService KubeController) error {
	imageCfg, hostCfg := buildKubeControllerConfig(kubeControllerService)
	err := docker.DoRunContainer(imageCfg, hostCfg, KubeControllerContainerName, &host, ControlRole)
	if err != nil {
		return err
	}
	return nil
}

func buildKubeControllerConfig(kubeControllerService KubeController) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeControllerService.Image + ":" + kubeControllerService.Version,
		Cmd: []string{"/hyperkube",
			"controller-manager",
			"--address=0.0.0.0",
			"--cloud-provider=",
			"--kubeconfig=" + pki.KubeControllerConfigPath,
			"--enable-hostpath-provisioner=false",
			"--node-monitor-grace-period=40s",
			"--pod-eviction-timeout=5m0s",
			"--v=2",
			"--allocate-node-cidrs=true",
			"--cluster-cidr=" + kubeControllerService.ClusterCIDR,
			"--service-cluster-ip-range=" + kubeControllerService.ServiceClusterIPRange,
			"--service-account-private-key-file=" + pki.KubeAPIKeyPath,
			"--root-ca-file=" + pki.CACertPath,
		},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}
	return imageCfg, hostCfg
}

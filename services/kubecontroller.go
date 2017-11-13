package services

import (
	"github.com/alena1108/cluster-controller/client/v1"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

func runKubeController(host hosts.Host, kubeControllerService v1.KubeControllerService) error {
	imageCfg, hostCfg := buildKubeControllerConfig(kubeControllerService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, KubeControllerContainerName, host.Hostname, ControlRole)
}

func buildKubeControllerConfig(kubeControllerService v1.KubeControllerService) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeControllerService.Image,
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
	imageCfg.Cmd = append(imageCfg.Cmd, kubeControllerService.ExtraArgs...)
	return imageCfg, hostCfg
}

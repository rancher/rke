package services

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/hosts"
)

type KubeController struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
	ClusterCIDR string `yaml:"cluster_cider"`
	ServiceClusterIPRange	string `yaml:"service_cluster_ip_range"`
}

func runKubeController(host hosts.Host, kubeControllerService KubeController) error {
	isRunning, err := IsContainerRunning(host, KubeControllerContainerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[ControlPlane] Kube-Controller is already running on host [%s]", host.Hostname)
		return nil
	}
	err = runKubeControllerContainer(host, kubeControllerService)
	if err != nil {
		return err
	}
	return nil
}

func runKubeControllerContainer(host hosts.Host, kubeControllerService KubeController) error {
	logrus.Debugf("[ControlPlane] Pulling Kube Controller Image on host [%s]", host.Hostname)
	err := PullImage(host, kubeControllerService.Image+":"+kubeControllerService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully pulled Kube Controller image on host [%s]", host.Hostname)

	err = doRunKubeController(host, kubeControllerService)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully ran Kube Controller container on host [%s]", host.Hostname)
	return nil
}

func doRunKubeController(host hosts.Host, kubeControllerService KubeController) error {
	imageCfg := &container.Config{
		Image: kubeControllerService.Image + ":" + kubeControllerService.Version,
		Cmd: []string{"/hyperkube",
			"controller-manager",
			"--address=0.0.0.0",
			"--cloud-provider=",
			"--master=http://" + host.IP + ":8080",
			"--enable-hostpath-provisioner=false",
			"--node-monitor-grace-period=40s",
			"--pod-eviction-timeout=5m0s",
			"--v=2",
			"--allocate-node-cidrs=true",
			"--cluster-cidr=" + kubeControllerService.ClusterCIDR,
			"--service-cluster-ip-range=" + kubeControllerService.ServiceClusterIPRange},
	}
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, KubeControllerContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create Kube Controller container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Kube Controller container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[ControlPlane] Successfully started Kube Controller container: %s", resp.ID)
	return nil
}

package services

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/hosts"
)

type Kubeproxy struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

func runKubeproxy(host hosts.Host, masterHost hosts.Host, kubeproxyService Kubeproxy) error {
	isRunning, err := IsContainerRunning(host, KubeproxyContainerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[WorkerPlane] Kubeproxy is already running on host [%s]", host.Hostname)
		return nil
	}
	err = runKubeproxyContainer(host, masterHost, kubeproxyService)
	if err != nil {
		return err
	}
	return nil
}

func runKubeproxyContainer(host hosts.Host, masterHost hosts.Host, kubeproxyService Kubeproxy) error {
	logrus.Debugf("[WorkerPlane] Pulling KubeProxy Image on host [%s]", host.Hostname)
	err := PullImage(host, kubeproxyService.Image+":"+kubeproxyService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[WorkerPlane] Successfully pulled KubeProxy image on host [%s]", host.Hostname)

	err = doRunKubeProxy(host, masterHost, kubeproxyService)
	if err != nil {
		return err
	}
	logrus.Infof("[WorkerPlane] Successfully ran KubeProxy container on host [%s]", host.Hostname)
	return nil
}

func doRunKubeProxy(host hosts.Host, masterHost hosts.Host, kubeproxyService Kubeproxy) error {
	imageCfg := &container.Config{
		Image: kubeproxyService.Image + ":" + kubeproxyService.Version,
		Cmd: []string{"/hyperkube",
			"proxy",
			"--v=2",
			"--healthz-bind-address=0.0.0.0",
			"--master=http://" + masterHost.IP + ":8080/"},
	}
	hostCfg := &container.HostConfig{
		NetworkMode:   "host",
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Privileged:    true,
	}
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, KubeproxyContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create KubeProxy container on host [%s]: %v", host.Hostname, err)
	}
	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start KubeProxy container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[WorkerPlane] Successfully started KubeProxy container: %s", resp.ID)
	return nil
}

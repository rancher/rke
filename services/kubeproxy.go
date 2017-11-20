package services

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

func runKubeproxy(host hosts.Host, kubeproxyService v1.KubeproxyService) error {
	imageCfg, hostCfg := buildKubeproxyConfig(host, kubeproxyService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, KubeproxyContainerName, host.AdvertisedHostname, WorkerRole)
}

func upgradeKubeproxy(host hosts.Host, kubeproxyService v1.KubeproxyService) error {
	logrus.Debugf("[upgrade/Kubeproxy] Checking for deployed version")
	containerInspect, err := docker.InspectContainer(host.DClient, host.AdvertisedHostname, KubeproxyContainerName)
	if err != nil {
		return err
	}
	if containerInspect.Config.Image == kubeproxyService.Image {
		logrus.Infof("[upgrade/Kubeproxy] Kubeproxy is already up to date")
		return nil
	}
	logrus.Debugf("[upgrade/Kubeproxy] Stopping old container")
	oldContainerName := "old-" + KubeproxyContainerName
	if err := docker.StopRenameContainer(host.DClient, host.AdvertisedHostname, KubeproxyContainerName, oldContainerName); err != nil {
		return err
	}
	// Container doesn't exist now!, lets deploy it!
	logrus.Debugf("[upgrade/Kubeproxy] Deploying new container")
	if err := runKubeproxy(host, kubeproxyService); err != nil {
		return err
	}
	logrus.Debugf("[upgrade/Kubeproxy] Removing old container")
	err = docker.RemoveContainer(host.DClient, host.AdvertisedHostname, oldContainerName)
	return err
}

func removeKubeproxy(host hosts.Host) error {
	return docker.DoRemoveContainer(host.DClient, KubeproxyContainerName, host.AdvertisedHostname)
}

func buildKubeproxyConfig(host hosts.Host, kubeproxyService v1.KubeproxyService) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeproxyService.Image,
		Cmd: []string{"/hyperkube",
			"proxy",
			"--v=2",
			"--healthz-bind-address=0.0.0.0",
			"--kubeconfig=" + pki.KubeProxyConfigPath,
		},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
		NetworkMode:   "host",
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Privileged:    true,
	}
	for arg, value := range kubeproxyService.ExtraArgs {
		cmd := fmt.Sprintf("--%s=%s", arg, value)
		imageCfg.Cmd = append(imageCfg.Cmd, cmd)
	}
	return imageCfg, hostCfg
}

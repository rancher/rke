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

func runScheduler(host hosts.Host, schedulerService v1.SchedulerService) error {
	imageCfg, hostCfg := buildSchedulerConfig(host, schedulerService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, SchedulerContainerName, host.AdvertisedHostname, ControlRole)
}

func upgradeScheduler(host hosts.Host, schedulerService v1.SchedulerService) error {
	logrus.Debugf("[upgrade/Scheduler] Checking for deployed version")
	containerInspect, err := docker.InspectContainer(host.DClient, host.AdvertisedHostname, SchedulerContainerName)
	if err != nil {
		return err
	}
	if containerInspect.Config.Image == schedulerService.Image {
		logrus.Infof("[upgrade/Scheduler] Scheduler is already up to date")
		return nil
	}
	logrus.Debugf("[upgrade/Scheduler] Stopping old container")
	oldContainerName := "old-" + SchedulerContainerName
	if err := docker.StopRenameContainer(host.DClient, host.AdvertisedHostname, SchedulerContainerName, oldContainerName); err != nil {
		return err
	}
	// Container doesn't exist now!, lets deploy it!
	logrus.Debugf("[upgrade/Scheduler] Deploying new container")
	if err := runScheduler(host, schedulerService); err != nil {
		return err
	}
	logrus.Debugf("[upgrade/Scheduler] Removing old container")
	err = docker.RemoveContainer(host.DClient, host.AdvertisedHostname, oldContainerName)
	return err
}

func removeScheduler(host hosts.Host) error {
	return docker.DoRemoveContainer(host.DClient, SchedulerContainerName, host.AdvertisedHostname)
}

func buildSchedulerConfig(host hosts.Host, schedulerService v1.SchedulerService) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: schedulerService.Image,
		Entrypoint: []string{"kube-scheduler",
			"--leader-elect=true",
			"--v=2",
			"--address=0.0.0.0",
			"--kubeconfig=" + pki.KubeSchedulerConfigPath,
		},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
		NetworkMode:   "host",
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}
	for arg, value := range schedulerService.ExtraArgs {
		cmd := fmt.Sprintf("--%s=%s", arg, value)
		imageCfg.Cmd = append(imageCfg.Cmd, cmd)
	}
	return imageCfg, hostCfg
}

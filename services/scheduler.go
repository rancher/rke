package services

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/hosts"
)

type Scheduler struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

func runScheduler(host hosts.Host, schedulerService Scheduler) error {
	isRunning, err := IsContainerRunning(host, SchedulerContainerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[ControlPlane] Scheduler is already running on host [%s]", host.Hostname)
		return nil
	}
	err = runSchedulerContainer(host, schedulerService)
	if err != nil {
		return err
	}
	return nil
}

func runSchedulerContainer(host hosts.Host, schedulerService Scheduler) error {
	logrus.Debugf("[ControlPlane] Pulling Scheduler Image on host [%s]", host.Hostname)
	err := PullImage(host, schedulerService.Image+":"+schedulerService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully pulled Scheduler image on host [%s]", host.Hostname)

	err = doRunScheduler(host, schedulerService)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully ran Scheduler container on host [%s]", host.Hostname)
	return nil
}

func doRunScheduler(host hosts.Host, schedulerService Scheduler) error {
	imageCfg := &container.Config{
		Image: schedulerService.Image + ":" + schedulerService.Version,
		Cmd: []string{"/hyperkube",
			"scheduler",
			"--v=2",
			"--address=0.0.0.0",
			"--master=http://" + host.IP + ":8080/"},
	}
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, SchedulerContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create Scheduler container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Scheduler container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[ControlPlane] Successfully started Scheduler container: %s", resp.ID)
	return nil
}

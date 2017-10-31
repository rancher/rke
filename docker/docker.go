package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/hosts"
)

func DoRunContainer(imageCfg *container.Config, hostCfg *container.HostConfig, containerName string, host *hosts.Host, plane string) error {
	isRunning, err := IsContainerRunning(host, containerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[%s] Container %s is already running on host [%s]", plane, containerName, host.Hostname)
		return nil
	}
	logrus.Debugf("[%s] Pulling Image on host [%s]", plane, host.Hostname)
	err = PullImage(host, imageCfg.Image)
	if err != nil {
		return err
	}
	logrus.Infof("[%s] Successfully pulled %s image on host [%s]", plane, containerName, host.Hostname)
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, containerName)
	if err != nil {
		return fmt.Errorf("Failed to create %s container on host [%s]: %v", containerName, host.Hostname, err)
	}
	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start %s container on host [%s]: %v", containerName, host.Hostname, err)
	}
	logrus.Debugf("[%s] Successfully started %s container: %s", plane, containerName, resp.ID)
	logrus.Infof("[%s] Successfully started %s container on host [%s]", plane, containerName, host.Hostname)
	return nil
}

func IsContainerRunning(host *hosts.Host, containerName string) (bool, error) {
	logrus.Debugf("Checking if container %s is running on host [%s]", containerName, host.Hostname)
	containers, err := host.DClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, fmt.Errorf("Can't get Docker containers for host [%s]: %v", host.Hostname, err)

	}
	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			return true, nil
		}
	}
	return false, nil
}

func PullImage(host *hosts.Host, containerImage string) error {
	out, err := host.DClient.ImagePull(context.Background(), containerImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Can't pull Docker image %s for host [%s]: %v", containerImage, host.Hostname, err)
	}
	defer out.Close()
	if logrus.GetLevel() == logrus.DebugLevel {
		io.Copy(os.Stdout, out)
	} else {
		io.Copy(ioutil.Discard, out)
	}

	return nil
}

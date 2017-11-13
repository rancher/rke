package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

func DoRunContainer(dClient *client.Client, imageCfg *container.Config, hostCfg *container.HostConfig, containerName string, hostname string, plane string) error {
	isRunning, err := IsContainerRunning(dClient, hostname, containerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[%s] Container %s is already running on host [%s]", plane, containerName, hostname)
		return nil
	}
	logrus.Debugf("[%s] Pulling Image on host [%s]", plane, hostname)
	err = PullImage(dClient, hostname, imageCfg.Image)
	if err != nil {
		return err
	}
	logrus.Infof("[%s] Successfully pulled %s image on host [%s]", plane, containerName, hostname)
	resp, err := dClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, containerName)
	if err != nil {
		return fmt.Errorf("Failed to create %s container on host [%s]: %v", containerName, hostname, err)
	}
	if err := dClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start %s container on host [%s]: %v", containerName, hostname, err)
	}
	logrus.Debugf("[%s] Successfully started %s container: %s", plane, containerName, resp.ID)
	logrus.Infof("[%s] Successfully started %s container on host [%s]", plane, containerName, hostname)
	return nil
}

func IsContainerRunning(dClient *client.Client, hostname string, containerName string) (bool, error) {
	logrus.Debugf("Checking if container %s is running on host [%s]", containerName, hostname)
	containers, err := dClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, fmt.Errorf("Can't get Docker containers for host [%s]: %v", hostname, err)

	}
	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			return true, nil
		}
	}
	return false, nil
}

func PullImage(dClient *client.Client, hostname string, containerImage string) error {
	out, err := dClient.ImagePull(context.Background(), containerImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Can't pull Docker image %s for host [%s]: %v", containerImage, hostname, err)
	}
	defer out.Close()
	if logrus.GetLevel() == logrus.DebugLevel {
		io.Copy(os.Stdout, out)
	} else {
		io.Copy(ioutil.Discard, out)
	}

	return nil
}

func RemoveContainer(dClient *client.Client, hostname string, containerName string) error {
	err := dClient.ContainerRemove(context.Background(), containerName, types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("Can't remove Docker container %s for host [%s]: %v", containerName, hostname, err)
	}
	return nil
}

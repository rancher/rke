package dind

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/util"
	"github.com/sirupsen/logrus"
)

const (
	DINDImage           = "docker:19.03.12-dind"
	DINDContainerPrefix = "rke-dind"
	DINDPlane           = "dind"
	DINDNetwork         = "dind-network"
	DINDSubnet          = "172.18.0.0/16"
)

func StartUpDindContainer(ctx context.Context, dindAddress, dindNetwork, dindStorageDriver, dindDNS, dindImportImagesList, dindImportImagesPath string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}
	// its recommended to use host's storage driver
	dockerInfo, err := cli.Info(ctx)
	if err != nil {
		return "", err
	}
	storageDriver := dindStorageDriver
	if len(storageDriver) == 0 {
		storageDriver = dockerInfo.Driver
	}

	// Get dind container name
	containerName := fmt.Sprintf("%s-%s", DINDContainerPrefix, dindAddress)
	_, err = cli.ContainerInspect(ctx, containerName)
	if err != nil {
		if !client.IsErrNotFound(err) {
			return "", err
		}
		if err := docker.UseLocalOrPull(ctx, cli, cli.DaemonHost(), DINDImage, DINDPlane, nil); err != nil {
			return "", err
		}
		binds := []string{
			fmt.Sprintf("/var/lib/kubelet-%s:/var/lib/kubelet:shared", containerName),
			"/etc/machine-id:/etc/machine-id:ro",
		}
		isLink, err := util.IsSymlink("/etc/resolv.conf")
		if err != nil {
			return "", err
		}
		if isLink {
			logrus.Infof("[%s] symlinked [/etc/resolv.conf] file detected. Using [%s] as DNS server.", DINDPlane, dindDNS)
		} else {
			binds = append(binds, "/etc/resolv.conf:/etc/resolv.conf")
		}
		// Export images to shared image path if configured
		if dindImportImagesList != "" {
			importImages := strings.Split(dindImportImagesList, ",")
			logrus.Infof("[%s] Found one or more custom images to use for testing in docker in docker: %v", DINDPlane, importImages)
			logrus.Infof("[%s] Using bind [%s] as shared image path for custom images", DINDPlane, dindImportImagesPath)
			// Add shared path to binds for dind containers
			binds = append(binds, fmt.Sprintf("%s/%s:%s", dindImportImagesPath, containerName, dindImportImagesPath))
			// Export each configured image
			for _, importImage := range importImages {
				logrus.Infof("[%s] Exporting image [%s]", DINDPlane, importImage)
				reader, err := cli.ImageSave(ctx, []string{importImage})
				if err != nil {
					return "", fmt.Errorf("Failed to save image [%s]: %v", importImage, err)
				}
				defer reader.Close()

				// Keep the filename valid (no breaking characters)
				imageFilename := strings.Replace(importImage, "/", "_", -1)
				imageFilename = strings.Replace(imageFilename, ":", "_", -1)
				tarFileName := fmt.Sprintf("dind-image-tar-%s.tar", imageFilename)
				// Create a temporary directory to extract the Docker image to.
				err = os.MkdirAll(filepath.Join(dindImportImagesPath, containerName), 0755)
				if err != nil {
					return "", fmt.Errorf("Failed to create directory [%s]", filepath.Join(dindImportImagesPath, containerName))
				}
				tarPath := filepath.Join(dindImportImagesPath, containerName, tarFileName)
				tarFile, err := os.Create(tarPath)
				if err != nil {
					return "", fmt.Errorf("Failed to create image file for image [%s]: %v", importImage, err)
				}
				defer tarFile.Close()

				// Copy the Docker image to the temporary file.
				_, err = io.Copy(tarFile, reader)
				if err != nil {
					return "", fmt.Errorf("Failed to copy reader to image file [%s] for image [%s]: %v", tarFileName, importImage, err)
				}
				logrus.Infof("[%s] Done exporting image [%s]", DINDPlane, importImage)
			}
			logrus.Infof("[%s] Done exporting all images", DINDPlane)
		}

		imageCfg := &container.Config{
			Image: DINDImage,
			Entrypoint: []string{
				"sh",
				"-c",
				"mount --make-shared / && " +
					"mount --make-shared /sys && " +
					"mount --make-shared /var/lib/docker && " +
					"dockerd-entrypoint.sh --storage-driver=" + storageDriver,
			},
			Hostname: dindAddress,
			Env:      []string{"DOCKER_TLS_CERTDIR="},
		}
		hostCfg := &container.HostConfig{
			Privileged: true,
			Binds:      binds,
			// this gets ignored if resolv.conf is bind mounted. So it's ok to have it anyway.
			DNS: []string{dindDNS},
			// Calico needs this
			Sysctls: map[string]string{
				"net.ipv4.conf.all.rp_filter": "1",
			},
		}
		resp, err := cli.ContainerCreate(ctx, imageCfg, hostCfg, nil, nil, containerName)
		if err != nil {
			return "", fmt.Errorf("Failed to create [%s] container on host [%s]: %v", containerName, cli.DaemonHost(), err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return "", fmt.Errorf("Failed to start [%s] container on host [%s]: %v", containerName, cli.DaemonHost(), err)
		}

		logrus.Infof("[%s] Successfully started [%s] container on host [%s]", DINDPlane, containerName, cli.DaemonHost())

		// Exec into containers to load images
		if dindImportImagesList != "" {
			logrus.Infof("[%s] Running docker load using docker exec in container [%s]", DINDPlane, containerName)
			cmd := exec.Command("docker", "exec", containerName, "sh", "-c", fmt.Sprintf("find %s -type f | xargs docker load -i", dindImportImagesPath))

			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("Failed to exec command in container [%s]: %v, output: %s", containerName, err, string(output))
			}
			logrus.Infof("[%s] Log output for docker exec in container [%s]:\n %s", DINDPlane, containerName, string(output))
		}
		dindContainer, err := cli.ContainerInspect(ctx, containerName)
		if err != nil {
			return "", fmt.Errorf("Failed to get the address of container [%s] on host [%s]: %v", containerName, cli.DaemonHost(), err)
		}
		dindIPAddress := dindContainer.NetworkSettings.IPAddress

		return dindIPAddress, nil
	}
	dindContainer, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("Failed to get the address of container [%s] on host [%s]: %v", containerName, cli.DaemonHost(), err)
	}
	dindIPAddress := dindContainer.NetworkSettings.IPAddress
	logrus.Infof("[%s] container [%s] is already running on host[%s]", DINDPlane, containerName, cli.DaemonHost())
	return dindIPAddress, nil
}

func RmoveDindContainer(ctx context.Context, dindAddress string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s-%s", DINDContainerPrefix, dindAddress)
	logrus.Infof("[%s] Removing dind container [%s] on host [%s]", DINDPlane, containerName, cli.DaemonHost())
	_, err = cli.ContainerInspect(ctx, containerName)
	if err != nil {
		if !client.IsErrNotFound(err) {
			return nil
		}
	}
	if err := cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true}); err != nil {
		if client.IsErrNotFound(err) {
			logrus.Debugf("[remove/%s] Container doesn't exist on host [%s]", containerName, cli.DaemonHost())
			return nil
		}
		return fmt.Errorf("Failed to remove dind container [%s] on host [%s]: %v", containerName, cli.DaemonHost(), err)
	}
	logrus.Infof("[%s] Successfully Removed dind container [%s] on host [%s]", DINDPlane, containerName, cli.DaemonHost())
	return nil
}

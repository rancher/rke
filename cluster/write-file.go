package cluster

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

// WriteFile - Writes file to Node.
func (c *Cluster) WriteFile(ctx context.Context, host *hosts.Host, role string, path string, mode string, content string) error {
	log.Infof(ctx, "[%s] Deploying file %s on node [%s]", role, path, host.Address)
	err := c.doWriteFile(ctx, host, role, path, mode, content)
	if err != nil {
		return fmt.Errorf("[%s] Failed to file %s on node [%s]: %v", role, path, host.Address, err)
	}
	logrus.Debugf("[%s] Successfully deployed %s on node [%s]", role, path, host.Address)

	return nil
}

func (c *Cluster) doWriteFile(ctx context.Context, host *hosts.Host, role string, path string, mode string, content string) error {
	isRunning, err := docker.IsContainerRunning(ctx, host.DClient, host.Address, "deploy-file", true)
	if err != nil {
		return err
	}
	if isRunning {
		err = docker.DoRemoveContainer(ctx, host.DClient, "write-file", host.Address)
		if err != nil {
			return err
		}
	}

	b64Content := base64.StdEncoding.EncodeToString([]byte(content))
	pathDir := filepath.Dir(path)
	imageCfg := &container.Config{
		Image: c.SystemImages.Alpine,
		Cmd: []string{
			"sh",
			"-c",
			fmt.Sprintf("cp -f %s %s-$(date -Iseconds -u); echo \"%s\" | base64 -d > %s && chmod %s %s", path, path, b64Content, path, mode, path),
		},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s:z", pathDir, pathDir),
		},
	}

	// Start container
	err = docker.DoRunContainer(ctx, host.DClient, imageCfg, hostCfg, "write-file", host.Address, role, c.PrivateRegistriesMap)
	if err != nil {
		return err
	}

	// Remove container
	err = docker.DoRemoveContainer(ctx, host.DClient, "write-file", host.Address)
	if err != nil {
		return err
	}

	return nil
}

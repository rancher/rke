package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

type Kubeproxy struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

func runKubeproxy(host hosts.Host, kubeproxyService Kubeproxy) error {
	imageCfg, hostCfg := buildKubeproxyConfig(host, kubeproxyService)
	err := docker.DoRunContainer(imageCfg, hostCfg, KubeproxyContainerName, &host, WorkerRole)
	if err != nil {
		return err
	}
	return nil
}

func buildKubeproxyConfig(host hosts.Host, kubeproxyService Kubeproxy) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeproxyService.Image + ":" + kubeproxyService.Version,
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

	return imageCfg, hostCfg
}

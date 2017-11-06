package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

func runKubeproxy(host hosts.Host, kubeproxyService Kubeproxy) error {
	imageCfg, hostCfg := buildKubeproxyConfig(host, kubeproxyService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, KubeproxyContainerName, host.Hostname, WorkerRole)
}

func buildKubeproxyConfig(host hosts.Host, kubeproxyService Kubeproxy) (*container.Config, *container.HostConfig) {
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

	return imageCfg, hostCfg
}

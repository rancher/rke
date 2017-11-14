package services

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
)

func runKubeproxy(host hosts.Host, kubeproxyService v1.KubeproxyService) error {
	imageCfg, hostCfg := buildKubeproxyConfig(host, kubeproxyService)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, KubeproxyContainerName, host.AdvertisedHostname, WorkerRole)
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

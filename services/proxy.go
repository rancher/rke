package services

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
)

const (
	NginxProxyImage   = "husseingalal/nginx-nodeporxy:dev"
	NginxProxyEnvName = "CP_HOSTS"
)

func RollingUpdateNginxProxy(cpHosts []hosts.Host, workerHosts []hosts.Host) error {
	nginxProxyEnv := buildProxyEnv(cpHosts)
	for _, host := range workerHosts {
		imageCfg, hostCfg := buildNginxProxyConfig(host, nginxProxyEnv)
		if err := docker.DoRollingUpdateContainer(host.DClient, imageCfg, hostCfg, NginxProxyContainerName, host.Address, WorkerRole); err != nil {
			return err
		}
	}
	return nil
}

func runNginxProxy(host hosts.Host, cpHosts []hosts.Host) error {
	nginxProxyEnv := buildProxyEnv(cpHosts)
	imageCfg, hostCfg := buildNginxProxyConfig(host, nginxProxyEnv)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, NginxProxyContainerName, host.Address, WorkerRole)
}

func removeNginxProxy(host hosts.Host) error {
	return docker.DoRemoveContainer(host.DClient, NginxProxyContainerName, host.Address)
}

func buildNginxProxyConfig(host hosts.Host, nginxProxyEnv string) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: NginxProxyImage,
		Env:   []string{fmt.Sprintf("%s=%s", NginxProxyEnvName, nginxProxyEnv)},
	}
	hostCfg := &container.HostConfig{
		NetworkMode:   "host",
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}

	return imageCfg, hostCfg
}

func buildProxyEnv(cpHosts []hosts.Host) string {
	proxyEnv := ""
	for i, cpHost := range cpHosts {
		proxyEnv += fmt.Sprintf("%s", cpHost.InternalAddress)
		if i < (len(cpHosts) - 1) {
			proxyEnv += ","
		}
	}
	return proxyEnv
}

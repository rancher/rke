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

func runNginxProxy(host hosts.Host, cpHosts []hosts.Host) error {
	nginxProxyEnv := buildProxyEnv(cpHosts)
	imageCfg, hostCfg := buildNginxProxyConfig(host, nginxProxyEnv)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, NginxProxyContainerName, host.AdvertisedHostname, WorkerRole)
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
		proxyEnv += fmt.Sprintf("%s", cpHost.AdvertiseAddress)
		if i < (len(cpHosts) - 1) {
			proxyEnv += ","
		}
	}
	return proxyEnv
}

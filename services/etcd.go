package services

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
)

type Etcd struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

func RunEtcdPlane(etcdHosts []hosts.Host, etcdService Etcd) error {
	logrus.Infof("[%s] Building up Etcd Plane..", ETCDRole)
	for _, host := range etcdHosts {
		imageCfg, hostCfg := buildEtcdConfig(host, etcdService)
		err := docker.DoRunContainer(imageCfg, hostCfg, EtcdContainerName, &host, ETCDRole)
		if err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully started Etcd Plane..", ETCDRole)
	return nil
}

func buildEtcdConfig(host hosts.Host, etcdService Etcd) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: etcdService.Image + ":" + etcdService.Version,
		Cmd: []string{"/usr/local/bin/etcd",
			"--name=etcd-" + host.Hostname,
			"--data-dir=/etcd-data",
			"--advertise-client-urls=http://" + host.ControlPlaneIP + ":2379,http://" + host.ControlPlaneIP + ":4001",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-advertise-peer-urls=http://" + host.ControlPlaneIP + ":2380",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--initial-cluster-token=etcd-cluster-1",
			"--initial-cluster=etcd-" + host.Hostname + "=http://" + host.ControlPlaneIP + ":2380"},
	}
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Binds: []string{
			"/var/lib/etcd:/etcd-data"},
		PortBindings: nat.PortMap{
			"2379/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "2379",
				},
			},
			"2380/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "2380",
				},
			},
		},
	}
	return imageCfg, hostCfg
}

func getEtcdConnString(hosts []hosts.Host) string {
	connString := ""
	for i, host := range hosts {
		connString += "http://" + host.ControlPlaneIP + ":2379"
		if i < (len(hosts) - 1) {
			connString += ","
		}
	}
	return connString
}

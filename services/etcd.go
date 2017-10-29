package services

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/hosts"
)

type Etcd struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
}

func RunEtcdPlane(etcdHosts []hosts.Host, etcdService Etcd) error {
	logrus.Infof("[Etcd] Building up Etcd Plane..")
	for _, host := range etcdHosts {
		isRunning, err := IsContainerRunning(host, EtcdContainerName)
		if err != nil {
			return err
		}
		if isRunning {
			logrus.Infof("[Etcd] Container is already running on host [%s]", host.Hostname)
			return nil
		}
		err = runEtcdContainer(host, etcdService)
		if err != nil {
			return err
		}
	}
	return nil
}

func runEtcdContainer(host hosts.Host, etcdService Etcd) error {
	logrus.Debugf("[Etcd] Pulling Image on host [%s]", host.Hostname)
	err := PullImage(host, etcdService.Image+":"+etcdService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[Etcd] Successfully pulled Etcd image on host [%s]", host.Hostname)
	err = doRunEtcd(host, etcdService)
	if err != nil {
		return err
	}
	logrus.Infof("[Etcd] Successfully ran Etcd container on host [%s]", host.Hostname)
	return nil
}

func doRunEtcd(host hosts.Host, etcdService Etcd) error {
	imageCfg := &container.Config{
		Image: etcdService.Image + ":" + etcdService.Version,
		Cmd: []string{"/usr/local/bin/etcd",
			"--name=etcd-" + host.Hostname,
			"--data-dir=/etcd-data",
			"--advertise-client-urls=http://" + host.IP + ":2379,http://" + host.IP + ":4001",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-advertise-peer-urls=http://" + host.IP + ":2380",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--initial-cluster-token=etcd-cluster-1",
			"--initial-cluster=etcd-" + host.Hostname + "=http://" + host.IP + ":2380"},
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
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, EtcdContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create Etcd container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Etcd container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[Etcd] Successfully started Etcd container: %s", resp.ID)
	return nil
}

func getEtcdConnString(hosts []hosts.Host) string {
	connString := ""
	for i, host := range hosts {
		connString += "http://" + host.IP + ":2379"
		if i < (len(hosts) - 1) {
			connString += ","
		}
	}
	return connString
}

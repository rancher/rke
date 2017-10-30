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

type KubeAPI struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
	ServiceClusterIPRange	string `yaml:"service_cluster_ip_range"`
}

func runKubeAPI(host hosts.Host, etcdHosts []hosts.Host, kubeAPIService KubeAPI) error {
	isRunning, err := IsContainerRunning(host, KubeAPIContainerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[ControlPlane] KubeAPI is already running on host [%s]", host.Hostname)
		return nil
	}
	etcdConnString := getEtcdConnString(etcdHosts)
	err = runKubeAPIContainer(host, kubeAPIService, etcdConnString)
	if err != nil {
		return err
	}
	return nil
}

func runKubeAPIContainer(host hosts.Host, kubeAPIService KubeAPI, etcdConnString string) error {
	logrus.Debugf("[ControlPlane] Pulling Kube API Image on host [%s]", host.Hostname)
	err := PullImage(host, kubeAPIService.Image+":"+kubeAPIService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully pulled Kube API image on host [%s]", host.Hostname)
	err = doRunKubeAPI(host, kubeAPIService, etcdConnString)
	if err != nil {
		return err
	}
	logrus.Infof("[ControlPlane] Successfully ran Kube API container on host [%s]", host.Hostname)
	return nil
}

func doRunKubeAPI(host hosts.Host, kubeAPIService KubeAPI, etcdConnString string) error {
	imageCfg := &container.Config{
		Image: kubeAPIService.Image + ":" + kubeAPIService.Version,
		Cmd: []string{"/hyperkube",
			"apiserver",
			"--insecure-bind-address=0.0.0.0",
			"--insecure-port=8080",
			"--cloud-provider=",
			"--allow_privileged=true",
			"--service-cluster-ip-range=" + kubeAPIService.ServiceClusterIPRange,
			"--admission-control=NamespaceLifecycle,LimitRanger,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota,DefaultTolerationSeconds",
			"--runtime-config=batch/v2alpha1",
			"--runtime-config=authentication.k8s.io/v1beta1=true",
			"--storage-backend=etcd3",
			"--etcd-servers=" + etcdConnString,
			"--advertise-address=" + host.IP},
	}
	hostCfg := &container.HostConfig{
		NetworkMode:   "host",
		RestartPolicy: container.RestartPolicy{Name: "always"},
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
	}
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, KubeAPIContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create Kube API container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Kube API container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[ControlPlane] Successfully started Kube API container: %s", resp.ID)
	return nil
}

package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

type KubeAPI struct {
	Version               string `yaml:"version"`
	Image                 string `yaml:"image"`
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range"`
}

func runKubeAPI(host hosts.Host, etcdHosts []hosts.Host, kubeAPIService KubeAPI) error {
	etcdConnString := getEtcdConnString(etcdHosts)
	imageCfg, hostCfg := buildKubeAPIConfig(host, kubeAPIService, etcdConnString)
	err := docker.DoRunContainer(imageCfg, hostCfg, KubeAPIContainerName, &host, ControlRole)
	if err != nil {
		return err
	}
	return nil
}

func buildKubeAPIConfig(host hosts.Host, kubeAPIService KubeAPI, etcdConnString string) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeAPIService.Image + ":" + kubeAPIService.Version,
		Cmd: []string{"/hyperkube",
			"apiserver",
			"--insecure-bind-address=0.0.0.0",
			"--insecure-port=8080",
			"--cloud-provider=",
			"--allow_privileged=true",
			"--service-cluster-ip-range=" + kubeAPIService.ServiceClusterIPRange,
			"--admission-control=ServiceAccount,NamespaceLifecycle,LimitRanger,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota,DefaultTolerationSeconds",
			"--runtime-config=batch/v2alpha1",
			"--runtime-config=authentication.k8s.io/v1beta1=true",
			"--storage-backend=etcd3",
			"--etcd-servers=" + etcdConnString,
			"--advertise-address=" + host.IP,
			"--client-ca-file=" + pki.CACertPath,
			"--tls-cert-file=" + pki.KubeAPICertPath,
			"--tls-private-key-file=" + pki.KubeAPIKeyPath,
			"--service-account-key-file=" + pki.KubeAPIKeyPath},
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
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
	return imageCfg, hostCfg
}

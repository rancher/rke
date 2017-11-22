package services

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

func runKubeAPI(host hosts.Host, etcdHosts []hosts.Host, kubeAPIService v1.KubeAPIService) error {
	etcdConnString := GetEtcdConnString(etcdHosts)
	imageCfg, hostCfg := buildKubeAPIConfig(host, kubeAPIService, etcdConnString)
	return docker.DoRunContainer(host.DClient, imageCfg, hostCfg, KubeAPIContainerName, host.AdvertisedHostname, ControlRole)
}

func upgradeKubeAPI(host hosts.Host, etcdHosts []hosts.Host, kubeAPIService v1.KubeAPIService) error {
	logrus.Debugf("[upgrade/KubeAPI] Checking for deployed version")
	containerInspect, err := docker.InspectContainer(host.DClient, host.AdvertisedHostname, KubeAPIContainerName)
	if err != nil {
		return err
	}
	if containerInspect.Config.Image == kubeAPIService.Image {
		logrus.Infof("[upgrade/KubeAPI] KubeAPI is already up to date")
		return nil
	}
	logrus.Debugf("[upgrade/KubeAPI] Stopping old container")
	oldContainerName := "old-" + KubeAPIContainerName
	if err := docker.StopRenameContainer(host.DClient, host.AdvertisedHostname, KubeAPIContainerName, oldContainerName); err != nil {
		return err
	}
	// Container doesn't exist now!, lets deploy it!
	logrus.Debugf("[upgrade/KubeAPI] Deploying new container")
	if err := runKubeAPI(host, etcdHosts, kubeAPIService); err != nil {
		return err
	}
	logrus.Debugf("[upgrade/KubeAPI] Removing old container")
	err = docker.RemoveContainer(host.DClient, host.AdvertisedHostname, oldContainerName)
	return err
}

func removeKubeAPI(host hosts.Host) error {
	return docker.DoRemoveContainer(host.DClient, KubeAPIContainerName, host.AdvertisedHostname)
}

func buildKubeAPIConfig(host hosts.Host, kubeAPIService v1.KubeAPIService, etcdConnString string) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeAPIService.Image,
		Entrypoint: []string{"kube-apiserver",
			"--insecure-bind-address=127.0.0.1",
			"--insecure-port=8080",
			"--secure-port=6443",
			"--cloud-provider=",
			"--allow_privileged=true",
			"--service-cluster-ip-range=" + kubeAPIService.ServiceClusterIPRange,
			"--admission-control=ServiceAccount,NamespaceLifecycle,LimitRanger,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota,DefaultTolerationSeconds",
			"--runtime-config=batch/v2alpha1",
			"--runtime-config=authentication.k8s.io/v1beta1=true",
			"--storage-backend=etcd3",
			"--etcd-servers=" + etcdConnString,
			"--advertise-address=" + host.AdvertiseAddress,
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

	for arg, value := range kubeAPIService.ExtraArgs {
		cmd := fmt.Sprintf("--%s=%s", arg, value)
		imageCfg.Cmd = append(imageCfg.Cmd, cmd)
	}
	return imageCfg, hostCfg
}

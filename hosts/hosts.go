package hosts

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type Host struct {
	v1.RKEConfigHost
	DClient *client.Client
}

const (
	ToCleanEtcdDir       = "/var/lib/etcd"
	ToCleanSSLDir        = "/etc/kubernetes/ssl"
	ToCleanCNIConf       = "/etc/cni"
	ToCleanCNIBin        = "/opt/cni"
	ToCleanCalicoRun     = "/var/run/calico"
	CleanerContainerName = "kube-cleaner"
	CleanerImage         = "alpine:latest"
)

func (h *Host) CleanUp() error {
	logrus.Infof("[down] Cleaning up host [%s]", h.AdvertisedHostname)
	toCleanDirs := []string{
		ToCleanEtcdDir,
		ToCleanSSLDir,
		ToCleanCNIConf,
		ToCleanCNIBin,
		ToCleanCalicoRun,
	}
	logrus.Infof("[down] Running cleaner container on host [%s]", h.AdvertisedHostname)
	imageCfg, hostCfg := buildCleanerConfig(h, toCleanDirs)
	if err := docker.DoRunContainer(h.DClient, imageCfg, hostCfg, CleanerContainerName, h.AdvertisedHostname, CleanerContainerName); err != nil {
		return err
	}

	if err := docker.WaitForContainer(h.DClient, CleanerContainerName); err != nil {
		return err
	}

	logrus.Infof("[down] Removing cleaner container on host [%s]", h.AdvertisedHostname)
	if err := docker.RemoveContainer(h.DClient, h.AdvertisedHostname, CleanerContainerName); err != nil {
		return err
	}
	logrus.Infof("[down] Successfully cleaned up host [%s]", h.AdvertisedHostname)
	return nil
}

func DeleteNode(toDeleteHost *Host, kubeClient *kubernetes.Clientset) error {
	logrus.Infof("[hosts] Cordoning host [%s]", toDeleteHost.AdvertisedHostname)
	err := k8s.CordonUncordon(kubeClient, toDeleteHost.AdvertisedHostname, true)
	if err != nil {
		return err
	}
	logrus.Infof("[hosts] Deleting host [%s] from the cluster", toDeleteHost.AdvertisedHostname)
	err = k8s.DeleteNode(kubeClient, toDeleteHost.AdvertisedHostname)
	if err != nil {
		return err
	}
	logrus.Infof("[hosts] Successfully deleted host [%s] from the cluster", toDeleteHost.AdvertisedHostname)
	return nil
}

func GetToDeleteHosts(currentHosts, configHosts []Host) []Host {
	toDeleteHosts := []Host{}
	for _, currentHost := range currentHosts {
		found := false
		for _, newHost := range configHosts {
			if currentHost.AdvertisedHostname == newHost.AdvertisedHostname {
				found = true
			}
		}
		if !found {
			toDeleteHosts = append(toDeleteHosts, currentHost)
		}
	}
	return toDeleteHosts
}

func IsHostListChanged(currentHosts, configHosts []Host) bool {
	changed := false
	for _, host := range currentHosts {
		found := false
		for _, configHost := range configHosts {
			if host.AdvertisedHostname == configHost.AdvertisedHostname {
				found = true
			}
		}
		if !found {
			return true
		}
	}
	for _, host := range configHosts {
		found := false
		for _, currentHost := range currentHosts {
			if host.AdvertisedHostname == currentHost.AdvertisedHostname {
				found = true
			}
		}
		if !found {
			return true
		}
	}
	return changed
}

func buildCleanerConfig(host *Host, toCleanDirs []string) (*container.Config, *container.HostConfig) {
	cmd := append([]string{"rm", "-rf"}, toCleanDirs...)
	imageCfg := &container.Config{
		Image: CleanerImage,
		Cmd:   cmd,
	}
	bindMounts := []string{}
	for _, vol := range toCleanDirs {
		bindMounts = append(bindMounts, fmt.Sprintf("%s:%s", vol, vol))
	}
	hostCfg := &container.HostConfig{
		Binds: bindMounts,
	}
	return imageCfg, hostCfg
}

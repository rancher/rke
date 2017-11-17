package hosts

import (
	"github.com/docker/docker/client"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type Host struct {
	v1.RKEConfigHost
	DClient *client.Client
}

func DeleteNode(toDeleteHost *Host, kubeClient *kubernetes.Clientset) error {
	logrus.Infof("[hosts] Cordoning host [%s]", toDeleteHost.AdvertisedHostname)
	err := k8s.CordonUncordon(kubeClient, toDeleteHost.AdvertisedHostname, true)
	if err != nil {
		return nil
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

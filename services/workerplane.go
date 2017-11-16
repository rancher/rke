package services

import (
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

func RunWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host, workerServices v1.RKEConfigServices) error {
	logrus.Infof("[%s] Building up Worker Plane..", WorkerRole)
	for _, host := range controlHosts {
		// only one master for now
		err := runKubelet(host, workerServices.Kubelet, true)
		if err != nil {
			return err
		}
		err = runKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	for _, host := range workerHosts {
		// run nginx proxy
		err := runNginxProxy(host, controlHosts)
		if err != nil {
			return err
		}
		// run kubelet
		err = runKubelet(host, workerServices.Kubelet, false)
		if err != nil {
			return err
		}
		// run kubeproxy
		err = runKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully started Worker Plane..", WorkerRole)
	return nil
}

func UpgradeWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host, workerServices v1.RKEConfigServices, localConfigPath string) error {
	logrus.Infof("[%s] Upgrading Worker Plane..", WorkerRole)
	k8sClient, err := k8s.NewClient(localConfigPath)
	if err != nil {
		return err
	}
	for _, host := range controlHosts {
		// cordone the node
		logrus.Debugf("[upgrade] Cordoning node: %s", host.AdvertisedHostname)
		if err = k8s.CordonUncordon(k8sClient, host.AdvertisedHostname, true); err != nil {
			return err
		}
		err = upgradeKubelet(host, workerServices.Kubelet, true)
		if err != nil {
			return err
		}
		err = upgradeKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}

		logrus.Debugf("[upgrade] Uncordoning node: %s", host.AdvertisedHostname)
		if err = k8s.CordonUncordon(k8sClient, host.AdvertisedHostname, false); err != nil {
			return err
		}
	}
	for _, host := range workerHosts {
		// cordone the node
		logrus.Debugf("[upgrade] Cordoning node: %s", host.AdvertisedHostname)
		if err = k8s.CordonUncordon(k8sClient, host.AdvertisedHostname, true); err != nil {
			return err
		}
		// upgrade kubelet
		err := upgradeKubelet(host, workerServices.Kubelet, false)
		if err != nil {
			return err
		}
		// upgrade kubeproxy
		err = upgradeKubeproxy(host, workerServices.Kubeproxy)
		if err != nil {
			return err
		}

		logrus.Debugf("[upgrade] Uncordoning node: %s", host.AdvertisedHostname)
		if err = k8s.CordonUncordon(k8sClient, host.AdvertisedHostname, false); err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully upgraded Worker Plane..", WorkerRole)
	return nil
}

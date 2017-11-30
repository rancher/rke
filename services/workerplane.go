package services

import (
	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

func RunWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host, workerServices v1.RKEConfigServices) error {
	logrus.Infof("[%s] Building up Worker Plane..", WorkerRole)
	for _, host := range controlHosts {
		// only one master for now
		if err := runKubelet(host, workerServices.Kubelet); err != nil {
			return err
		}
		if err := runKubeproxy(host, workerServices.Kubeproxy); err != nil {
			return err
		}
	}
	for _, host := range workerHosts {
		// run nginx proxy
		isControlPlaneHost := false
		for _, role := range host.Role {
			if role == ControlRole {
				isControlPlaneHost = true
				break
			}
		}
		if !isControlPlaneHost {
			if err := runNginxProxy(host, controlHosts); err != nil {
				return err
			}
		}
		// run kubelet
		if err := runKubelet(host, workerServices.Kubelet); err != nil {
			return err
		}
		// run kubeproxy
		if err := runKubeproxy(host, workerServices.Kubeproxy); err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully started Worker Plane..", WorkerRole)
	return nil
}

func RemoveWorkerPlane(controlHosts []hosts.Host, workerHosts []hosts.Host) error {
	logrus.Infof("[%s] Tearing down Worker Plane..", WorkerRole)
	for _, host := range controlHosts {
		if err := removeKubelet(host); err != nil {
			return err
		}
		if err := removeKubeproxy(host); err != nil {
			return err
		}
	}

	for _, host := range workerHosts {
		if err := removeKubelet(host); err != nil {
			return err
		}
		if err := removeKubeproxy(host); err != nil {
			return err
		}
		if err := removeNginxProxy(host); err != nil {
			return err
		}
	}
	logrus.Infof("[%s] Successfully teared down Worker Plane..", WorkerRole)
	return nil
}

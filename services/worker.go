package services

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/hosts"
)

func RunWorkerPlane(masterHosts []hosts.Host, workerHosts []hosts.Host, workerServices Services) error {
	logrus.Infof("[WorkerPlane] Building up Worker Plane..")
	for _, host := range masterHosts {
		// only one master for now
		err := runKubelet(host, masterHosts[0], workerServices.Kubelet, true)
		if err != nil {
			return err
		}
		err = runKubeproxy(host, masterHosts[0], workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	for _, host := range workerHosts {
		// run kubelet
		err := runKubelet(host, masterHosts[0], workerServices.Kubelet, false)
		if err != nil {
			return err
		}
		// run kubeproxy
		err = runKubeproxy(host, masterHosts[0], workerServices.Kubeproxy)
		if err != nil {
			return err
		}
	}
	return nil
}

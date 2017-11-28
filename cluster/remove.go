package cluster

import (
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/services"
)

func (c *Cluster) ClusterRemove() error {
	// Remove Worker Plane
	if err := services.RemoveWorkerPlane(c.ControlPlaneHosts, c.WorkerHosts); err != nil {
		return err
	}

	// Remove Contol Plane
	if err := services.RemoveControlPlane(c.ControlPlaneHosts); err != nil {
		return err
	}

	// Remove Etcd Plane
	if err := services.RemoveEtcdPlane(c.EtcdHosts); err != nil {
		return err
	}

	// Clean up all hosts
	return cleanUpHosts(c.ControlPlaneHosts, c.WorkerHosts, c.EtcdHosts)
}

func cleanUpHosts(cpHosts, workerHosts, etcdHosts []hosts.Host) error {
	allHosts := []hosts.Host{}
	allHosts = append(allHosts, cpHosts...)
	allHosts = append(allHosts, workerHosts...)
	allHosts = append(allHosts, etcdHosts...)

	for _, host := range allHosts {
		if err := host.CleanUp(); err != nil {
			return err
		}
	}
	return nil
}

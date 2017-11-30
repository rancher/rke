package cluster

import (
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
)

func (c *Cluster) ClusterRemove() error {
	// Remove Worker Plane
	if err := services.RemoveWorkerPlane(c.WorkerHosts, true); err != nil {
		return err
	}

	// Remove Contol Plane
	if err := services.RemoveControlPlane(c.ControlPlaneHosts, true); err != nil {
		return err
	}

	// Remove Etcd Plane
	if err := services.RemoveEtcdPlane(c.EtcdHosts); err != nil {
		return err
	}

	// Clean up all hosts
	if err := cleanUpHosts(c.ControlPlaneHosts, c.WorkerHosts, c.EtcdHosts); err != nil {
		return err
	}

	return pki.RemoveAdminConfig(c.LocalKubeConfigPath)
}

func cleanUpHosts(cpHosts, workerHosts, etcdHosts []*hosts.Host) error {
	allHosts := []*hosts.Host{}
	allHosts = append(allHosts, cpHosts...)
	allHosts = append(allHosts, workerHosts...)
	allHosts = append(allHosts, etcdHosts...)

	for _, host := range allHosts {
		if err := host.CleanUpAll(); err != nil {
			return err
		}
	}
	return nil
}

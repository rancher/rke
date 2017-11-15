package cluster

import (
	"fmt"

	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/sirupsen/logrus"
)

func (c *Cluster) ClusterUpgrade() error {
	// make sure all nodes are Ready
	logrus.Debugf("[upgrade] Checking node status")
	if err := checkK8sNodesState(); err != nil {
		return err
	}
	// upgrade Contol Plane
	logrus.Infof("[upgrade] Upgrading Control Plane Services")
	if err := services.UpgradeControlPlane(c.ControlPlaneHosts, c.EtcdHosts, c.Services); err != nil {
		return err
	}
	logrus.Infof("[upgrade] Control Plane Services updgraded successfully")

	// upgrade Worker Plane
	logrus.Infof("[upgrade] Upgrading Worker Plane Services")
	if err := services.UpgradeWorkerPlane(c.ControlPlaneHosts, c.WorkerHosts, c.Services); err != nil {
		return err
	}
	logrus.Infof("[upgrade] Worker Plane Services updgraded successfully")
	return nil
}

func checkK8sNodesState() error {
	k8sClient, err := k8s.NewClient(pki.KubeAdminConfigPath)
	if err != nil {
		return err
	}
	nodeList, err := k8s.GetNodeList(k8sClient)
	if err != nil {
		return err
	}
	for _, node := range nodeList.Items {
		ready := k8s.IsNodeReady(node)
		if !ready {
			return fmt.Errorf("[upgrade] Node: %s is NotReady", node.Name)
		}
	}
	logrus.Infof("[upgrade] All nodes are Ready")
	return nil
}

package cluster

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

const (
	ClusterCIDREnvName = "RKE_CLUSTER_CIDR"
)

func (c *Cluster) DeployNetworkPlugin() error {
	logrus.Infof("[network] Setting up network plugin: %s", c.NetworkPlugin)

	kubectlCmd := &KubectlCommand{
		Cmd: []string{"apply -f /network/" + c.NetworkPlugin + ".yaml"},
	}
	logrus.Infof("[network] Executing the deploy command..")
	err := c.RunKubectlCmd(kubectlCmd)
	if err != nil {
		return fmt.Errorf("Failed to run kubectl command: %v", err)
	}
	logrus.Infof("[network] Network plugin deployed successfully..")
	return nil
}

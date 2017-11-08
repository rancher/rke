package cluster

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/pki"
)

const (
	ClusterCIDREnvName = "RKE_CLUSTER_CIDR"
)

func (c *Cluster) DeployNetworkPlugin() error {
	logrus.Infof("[network] Setting up network plugin: %s", c.NetworkPlugin)
	deployerHost := c.ControlPlaneHosts[0]
	kubectlCmd := []string{"apply -f /network/" + c.NetworkPlugin + ".yaml"}
	env := []string{
		fmt.Sprintf("%s=%s", pki.KubeAdminConfigENVName, c.Certificates[pki.KubeAdminCommonName].Config),
		fmt.Sprintf("%s=%s", ClusterCIDREnvName, c.ClusterCIDR),
	}
	logrus.Infof("[network] Executing the deploy command..")
	err := k8s.RunKubectlCmd(deployerHost.DClient, deployerHost.Hostname, kubectlCmd, env)
	if err != nil {
		return fmt.Errorf("Failed to run kubectl command: %v", err)
	}
	logrus.Infof("[network] Network plugin deployed successfully..")
	return nil
}

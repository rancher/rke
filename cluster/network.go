package cluster

import (
	"github.com/rancher/rke/network"
	"github.com/sirupsen/logrus"
)

const (
	NetworkPluginResourceName = "rke-netwok-plugin"
)

func (c *Cluster) DeployNetworkPlugin() error {
	logrus.Infof("[network] Setting up network plugin: %s", c.Network.Plugin)

	pluginYaml := network.GetFlannelManifest(c.ClusterCIDR)

	if err := c.doAddonDeploy(pluginYaml, NetworkPluginResourceName); err != nil {
		return err
	}
	logrus.Infof("[network] Network plugin deployed successfully..")
	return nil
}

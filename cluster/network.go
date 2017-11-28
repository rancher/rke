package cluster

import (
	"fmt"

	"github.com/rancher/rke/network"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/sirupsen/logrus"
)

const (
	NetworkPluginResourceName = "rke-netwok-plugin"
	FlannelNetworkPlugin      = "flannel"
	CalicoNetworkPlugin       = "calico"
	CanalNetworkPlugin        = "canal"
)

func (c *Cluster) DeployNetworkPlugin() error {
	logrus.Infof("[network] Setting up network plugin: %s", c.Network.Plugin)
	switch c.Network.Plugin {
	case FlannelNetworkPlugin:
		return c.doFlannelDeploy()
	case CalicoNetworkPlugin:
		return c.doCalicoDeploy()
	case CanalNetworkPlugin:
		return c.doCanalDeploy()
	default:
		return fmt.Errorf("[network] Unsupported network plugin: %s", c.Network.Plugin)
	}
}

func (c *Cluster) doFlannelDeploy() error {
	pluginYaml := network.GetFlannelManifest(c.ClusterCIDR)
	return c.doAddonDeploy(pluginYaml, NetworkPluginResourceName)
}

func (c *Cluster) doCalicoDeploy() error {
	calicoConfig := make(map[string]string)
	calicoConfig["etcdEndpoints"] = services.GetEtcdConnString(c.EtcdHosts)
	calicoConfig["apiRoot"] = "https://127.0.0.1:6443"
	calicoConfig["clientCrt"] = pki.KubeNodeCertPath
	calicoConfig["clientKey"] = pki.KubeNodeKeyPath
	calicoConfig["clientCA"] = pki.CACertPath
	calicoConfig["kubeCfg"] = pki.KubeNodeConfigPath
	calicoConfig["clusterCIDR"] = c.ClusterCIDR
	pluginYaml := network.GetCalicoManifest(calicoConfig)
	return c.doAddonDeploy(pluginYaml, NetworkPluginResourceName)
}

func (c *Cluster) doCanalDeploy() error {
	canalConfig := make(map[string]string)
	canalConfig["clientCrt"] = pki.KubeNodeCertPath
	canalConfig["clientKey"] = pki.KubeNodeKeyPath
	canalConfig["clientCA"] = pki.CACertPath
	canalConfig["kubeCfg"] = pki.KubeNodeConfigPath
	canalConfig["clusterCIDR"] = c.ClusterCIDR
	pluginYaml := network.GetCanalManifest(canalConfig)
	return c.doAddonDeploy(pluginYaml, NetworkPluginResourceName)
}

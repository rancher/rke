package cluster

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/pki"
)

const (
	ClusterDnsServerIPEnvName = "RKE_DNS_SERVER"
	ClusterDomainEnvName      = "RKE_CLUSTER_DOMAIN"
)

func (c *Cluster) DeployK8sAddOns() error {
	if err := c.deployKubeDNS(); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) deployKubeDNS() error {
	logrus.Infof("[plugins] Setting up KubeDNS")
	deployerHost := c.ControlPlaneHosts[0]
	kubectlCmd := []string{"apply -f /addons/kubedns*.yaml"}

	env := []string{
		fmt.Sprintf("%s=%s", pki.KubeAdminConfigENVName, c.Certificates[pki.KubeAdminCommonName].Config),
		fmt.Sprintf("%s=%s", ClusterDnsServerIPEnvName, c.ClusterDnsServer),
		fmt.Sprintf("%s=%s", ClusterDomainEnvName, c.ClusterDomain),
	}

	logrus.Infof("[plugins] Executing the deploy command..")
	err := k8s.RunKubectlCmd(deployerHost.DClient, deployerHost.Hostname, kubectlCmd, env)
	if err != nil {
		return fmt.Errorf("Failed to run kubectl command: %v", err)
	}
	logrus.Infof("[plugins] kubeDNS deployed successfully..")
	return nil

}

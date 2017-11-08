package cluster

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/pki"
)

const (
	ClusterDNSServerIPEnvName = "RKE_DNS_SERVER"
	ClusterDomainEnvName      = "RKE_CLUSTER_DOMAIN"
)

func (c *Cluster) DeployK8sAddOns() error {
	err := c.deployKubeDNS()
	return err
}

func (c *Cluster) deployKubeDNS() error {
	logrus.Infof("[plugins] Setting up KubeDNS")
	deployerHost := c.ControlPlaneHosts[0]
	kubectlCmd := []string{"apply -f /addons/kubedns*.yaml"}

	env := []string{
		fmt.Sprintf("%s=%s", pki.KubeAdminConfigENVName, c.Certificates[pki.KubeAdminCommonName].Config),
		fmt.Sprintf("%s=%s", ClusterDNSServerIPEnvName, c.ClusterDNSServer),
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

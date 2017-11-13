package cluster

import (
	"fmt"

	"github.com/sirupsen/logrus"
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

	kubectlCmd := &KubectlCommand{
		Cmd: []string{"apply -f /addons/kubedns*.yaml"},
	}
	logrus.Infof("[plugins] Executing the deploy command..")
	err := c.RunKubectlCmd(kubectlCmd)
	if err != nil {
		return fmt.Errorf("Failed to run kubectl command: %v", err)
	}
	logrus.Infof("[plugins] kubeDNS deployed successfully..")
	return nil

}

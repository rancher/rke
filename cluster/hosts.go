package cluster

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
)

func (c *Cluster) TunnelHosts() error {
	for i := range c.EtcdHosts {
		err := c.EtcdHosts[i].TunnelUp()
		if err != nil {
			return fmt.Errorf("Failed to set up SSH tunneling for Etcd hosts: %v", err)
		}
	}
	for i := range c.ControlPlaneHosts {
		err := c.ControlPlaneHosts[i].TunnelUp()
		if err != nil {
			return fmt.Errorf("Failed to set up SSH tunneling for Control hosts: %v", err)
		}
	}
	for i := range c.WorkerHosts {
		err := c.WorkerHosts[i].TunnelUp()
		if err != nil {
			return fmt.Errorf("Failed to set up SSH tunneling for Worker hosts: %v", err)
		}
	}
	return nil
}

func (c *Cluster) InvertIndexHosts() error {
	c.EtcdHosts = make([]hosts.Host, 0)
	c.WorkerHosts = make([]hosts.Host, 0)
	c.ControlPlaneHosts = make([]hosts.Host, 0)
	for _, host := range c.Hosts {
		for _, role := range host.Role {
			logrus.Debugf("Host: " + host.Hostname + " has role: " + role)
			switch role {
			case services.ETCDRole:
				c.EtcdHosts = append(c.EtcdHosts, host)
			case services.ControlRole:
				c.ControlPlaneHosts = append(c.ControlPlaneHosts, host)
			case services.WorkerRole:
				c.WorkerHosts = append(c.WorkerHosts, host)
			default:
				return fmt.Errorf("Failed to recognize host [%s] role %s", host.Hostname, role)
			}
		}
	}
	return nil
}

func (c *Cluster) SetUpHosts(authType string) error {
	if authType == X509AuthenticationProvider {
		logrus.Infof("[certificates] Deploying kubernetes certificates to Cluster nodes")
		err := pki.DeployCertificatesOnMasters(c.ControlPlaneHosts, c.Certificates)
		if err != nil {
			return err
		}
		err = pki.DeployCertificatesOnWorkers(c.WorkerHosts, c.Certificates)
		if err != nil {
			return err
		}
		err = pki.DeployAdminConfig(c.Certificates[pki.KubeAdminCommonName].Config)
		if err != nil {
			return err
		}
		logrus.Infof("[certificates] Successfully deployed kubernetes certificates to Cluster nodes")
	}
	return nil
}

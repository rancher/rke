package cluster

import (
	"fmt"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/sirupsen/logrus"
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
			logrus.Debugf("Host: " + host.AdvertisedHostname + " has role: " + role)
			newHost := hosts.Host{
				RKEConfigHost: host,
			}
			switch role {
			case services.ETCDRole:
				c.EtcdHosts = append(c.EtcdHosts, newHost)
			case services.ControlRole:
				c.ControlPlaneHosts = append(c.ControlPlaneHosts, newHost)
			case services.WorkerRole:
				c.WorkerHosts = append(c.WorkerHosts, newHost)
			default:
				return fmt.Errorf("Failed to recognize host [%s] role %s", host.AdvertisedHostname, role)
			}
		}
	}
	return nil
}

func (c *Cluster) SetUpHosts() error {
	if c.Authentication.Strategy == X509AuthenticationProvider {
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

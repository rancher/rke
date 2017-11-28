package cluster

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	DefaultSSHKeyPath = "/.ssh/id_rsa"
)

func (c *Cluster) TunnelHosts() error {
	key, err := checkEncryptedKey(c.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("Failed to parse the private key: %v", err)
	}
	for i := range c.EtcdHosts {
		err := c.EtcdHosts[i].TunnelUp(key)
		if err != nil {
			return fmt.Errorf("Failed to set up SSH tunneling for Etcd hosts: %v", err)
		}
	}
	for i := range c.ControlPlaneHosts {
		err := c.ControlPlaneHosts[i].TunnelUp(key)
		if err != nil {
			return fmt.Errorf("Failed to set up SSH tunneling for Control hosts: %v", err)
		}
	}
	for i := range c.WorkerHosts {
		err := c.WorkerHosts[i].TunnelUp(key)
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
	for _, host := range c.Nodes {
		for _, role := range host.Role {
			logrus.Debugf("Host: " + host.Address + " has role: " + role)
			newHost := hosts.Host{
				RKEConfigNode: host,
			}
			switch role {
			case services.ETCDRole:
				c.EtcdHosts = append(c.EtcdHosts, newHost)
			case services.ControlRole:
				c.ControlPlaneHosts = append(c.ControlPlaneHosts, newHost)
			case services.WorkerRole:
				c.WorkerHosts = append(c.WorkerHosts, newHost)
			default:
				return fmt.Errorf("Failed to recognize host [%s] role %s", host.Address, role)
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
		err = pki.DeployAdminConfig(c.Certificates[pki.KubeAdminCommonName].Config, c.LocalKubeConfigPath)
		if err != nil {
			return err
		}
		logrus.Infof("[certificates] Successfully deployed kubernetes certificates to Cluster nodes")
	}
	return nil
}

func CheckEtcdHostsChanged(kubeCluster, currentCluster *Cluster) error {
	if currentCluster != nil {
		etcdChanged := hosts.IsHostListChanged(currentCluster.EtcdHosts, kubeCluster.EtcdHosts)
		if etcdChanged {
			return fmt.Errorf("Adding or removing Etcd nodes is not supported")
		}
	}
	return nil
}

func checkEncryptedKey(sshKeyPath string) (ssh.Signer, error) {
	logrus.Infof("[ssh] Checking private key")
	key, err := hosts.ParsePrivateKey(privateKeyPath(sshKeyPath))
	if err != nil {
		if strings.Contains(err.Error(), "decode encrypted private keys") {
			fmt.Printf("Passphrase for Private SSH Key: ")
			passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Printf("\n")
			if err != nil {
				return nil, err
			}
			key, err = hosts.ParsePrivateKeyWithPassPhrase(privateKeyPath(sshKeyPath), passphrase)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return key, nil
}

func privateKeyPath(sshKeyPath string) string {
	if len(sshKeyPath) == 0 {
		return os.Getenv("HOME") + DefaultSSHKeyPath
	}
	return sshKeyPath
}

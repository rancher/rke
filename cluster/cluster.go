package cluster

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
)

type Cluster struct {
	v1.RancherKubernetesEngineConfig `yaml:",inline"`
	ConfigPath                       string `yaml:"config_path"`
	LocalKubeConfigPath              string
	EtcdHosts                        []*hosts.Host
	WorkerHosts                      []*hosts.Host
	ControlPlaneHosts                []*hosts.Host
	KubeClient                       *kubernetes.Clientset
	KubernetesServiceIP              net.IP
	Certificates                     map[string]pki.CertificatePKI
	ClusterDomain                    string
	ClusterCIDR                      string
	ClusterDNSServer                 string
}

const (
	X509AuthenticationProvider   = "x509"
	DefaultClusterConfig         = "cluster.yml"
	DefaultServiceClusterIPRange = "10.233.0.0/18"
	DefaultClusterCIDR           = "10.233.64.0/18"
	DefaultClusterDNSService     = "10.233.0.3"
	DefaultClusterDomain         = "cluster.local"
	DefaultInfraContainerImage   = "gcr.io/google_containers/pause-amd64:3.0"
	DefaultAuthStrategy          = "x509"
	DefaultNetworkPlugin         = "flannel"
	DefaultClusterSSHKeyPath     = "~/.ssh/id_rsa"
	StateConfigMapName           = "cluster-state"
	UpdateStateTimeout           = 30
	GetStateTimeout              = 30
	KubernetesClientTimeOut      = 30
)

func (c *Cluster) DeployClusterPlanes() error {
	// Deploy Kubernetes Planes
	err := services.RunEtcdPlane(c.EtcdHosts, c.Services.Etcd)
	if err != nil {
		return fmt.Errorf("[etcd] Failed to bring up Etcd Plane: %v", err)
	}
	err = services.RunControlPlane(c.ControlPlaneHosts, c.EtcdHosts, c.Services)
	if err != nil {
		return fmt.Errorf("[controlPlane] Failed to bring up Control Plane: %v", err)
	}
	err = services.RunWorkerPlane(c.ControlPlaneHosts, c.WorkerHosts, c.Services)
	if err != nil {
		return fmt.Errorf("[workerPlane] Failed to bring up Worker Plane: %v", err)
	}
	return nil
}

func ParseConfig(clusterFile string) (*Cluster, error) {
	logrus.Debugf("Parsing cluster file [%v]", clusterFile)
	var err error
	c, err := parseClusterFile(clusterFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse the cluster file: %v", err)
	}
	err = c.InvertIndexHosts()
	if err != nil {
		return nil, fmt.Errorf("Failed to classify hosts from config file: %v", err)
	}

	err = c.ValidateCluster()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate cluster: %v", err)
	}

	c.KubernetesServiceIP, err = services.GetKubernetesServiceIP(c.Services.KubeAPI.ServiceClusterIPRange)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	c.ClusterDomain = c.Services.Kubelet.ClusterDomain
	c.ClusterCIDR = c.Services.KubeController.ClusterCIDR
	c.ClusterDNSServer = c.Services.Kubelet.ClusterDNSServer
	if len(c.ConfigPath) == 0 {
		c.ConfigPath = DefaultClusterConfig
	}
	c.LocalKubeConfigPath = GetLocalKubeConfig(c.ConfigPath)
	return c, nil
}

func parseClusterFile(clusterFile string) (*Cluster, error) {
	// parse hosts
	var kubeCluster Cluster
	err := yaml.Unmarshal([]byte(clusterFile), &kubeCluster)
	if err != nil {
		return nil, err
	}
	// Setting cluster Defaults
	kubeCluster.setClusterDefaults()

	return &kubeCluster, nil
}

func (c *Cluster) setClusterDefaults() {
	if len(c.SSHKeyPath) == 0 {
		c.SSHKeyPath = DefaultClusterSSHKeyPath
	}
	for i, host := range c.Nodes {
		if len(host.InternalAddress) == 0 {
			c.Nodes[i].InternalAddress = c.Nodes[i].Address
		}
		if len(host.HostnameOverride) == 0 {
			// This is a temporary modification
			c.Nodes[i].HostnameOverride = c.Nodes[i].Address
		}
		if len(host.SSHKeyPath) == 0 {
			c.Nodes[i].SSHKeyPath = c.SSHKeyPath
		}
	}
	if len(c.Services.KubeAPI.ServiceClusterIPRange) == 0 {
		c.Services.KubeAPI.ServiceClusterIPRange = DefaultServiceClusterIPRange
	}
	if len(c.Services.KubeController.ServiceClusterIPRange) == 0 {
		c.Services.KubeController.ServiceClusterIPRange = DefaultServiceClusterIPRange
	}
	if len(c.Services.KubeController.ClusterCIDR) == 0 {
		c.Services.KubeController.ClusterCIDR = DefaultClusterCIDR
	}
	if len(c.Services.Kubelet.ClusterDNSServer) == 0 {
		c.Services.Kubelet.ClusterDNSServer = DefaultClusterDNSService
	}
	if len(c.Services.Kubelet.ClusterDomain) == 0 {
		c.Services.Kubelet.ClusterDomain = DefaultClusterDomain
	}
	if len(c.Services.Kubelet.InfraContainerImage) == 0 {
		c.Services.Kubelet.InfraContainerImage = DefaultInfraContainerImage
	}
	if len(c.Authentication.Strategy) == 0 {
		c.Authentication.Strategy = DefaultAuthStrategy
	}
	if len(c.Network.Plugin) == 0 {
		c.Network.Plugin = DefaultNetworkPlugin
	}
}

func GetLocalKubeConfig(configPath string) string {
	baseDir := filepath.Dir(configPath)
	fileName := filepath.Base(configPath)
	baseDir += "/"
	return fmt.Sprintf("%s%s%s", baseDir, pki.KubeAdminConfigPrefix, fileName)
}

func rebuildLocalAdminConfig(kubeCluster *Cluster) error {
	logrus.Infof("[reconcile] Rebuilding and update local kube config")
	var workingConfig string
	currentKubeConfig := kubeCluster.Certificates[pki.KubeAdminCommonName]
	caCrt := kubeCluster.Certificates[pki.CACertName].Certificate
	for _, cpHost := range kubeCluster.ControlPlaneHosts {
		newConfig := pki.GetKubeConfigX509WithData(
			"https://"+cpHost.Address+":6443",
			pki.KubeAdminCommonName,
			string(cert.EncodeCertPEM(caCrt)),
			string(cert.EncodeCertPEM(currentKubeConfig.Certificate)),
			string(cert.EncodePrivateKeyPEM(currentKubeConfig.Key)))

		if err := pki.DeployAdminConfig(newConfig, kubeCluster.LocalKubeConfigPath); err != nil {
			return fmt.Errorf("Failed to redeploy local admin config with new host")
		}
		workingConfig = newConfig
		if _, err := GetK8sVersion(kubeCluster.LocalKubeConfigPath); err != nil {
			logrus.Infof("[reconcile] host [%s] is not active master on the cluster", cpHost.Address)
			continue
		} else {
			break
		}
	}
	currentKubeConfig.Config = workingConfig
	kubeCluster.Certificates[pki.KubeAdminCommonName] = currentKubeConfig
	return nil
}

func isLocalConfigWorking(localKubeConfigPath string) bool {
	if _, err := GetK8sVersion(localKubeConfigPath); err != nil {
		logrus.Infof("[reconcile] Local config is not vaild, rebuilding admin config")
		return false
	}
	return true
}

func getLocalConfigAddress(localConfigPath string) (string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", localConfigPath)
	if err != nil {
		return "", err
	}
	splittedAdress := strings.Split(config.Host, ":")
	address := splittedAdress[1]
	return address[2:], nil
}

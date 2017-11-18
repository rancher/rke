package cluster

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/cert"
)

type Cluster struct {
	v1.RancherKubernetesEngineConfig `yaml:",inline"`
	ConfigPath                       string `yaml:"config_path"`
	LocalKubeConfigPath              string
	EtcdHosts                        []hosts.Host
	WorkerHosts                      []hosts.Host
	ControlPlaneHosts                []hosts.Host
	KubeClient                       *kubernetes.Clientset
	KubernetesServiceIP              net.IP
	Certificates                     map[string]pki.CertificatePKI
	ClusterDomain                    string
	ClusterCIDR                      string
	ClusterDNSServer                 string
}

const (
	X509AuthenticationProvider = "x509"
	DefaultClusterConfig       = "cluster.yml"
	StateConfigMapName         = "cluster-state"
	UpdateStateTimeout         = 30
	GetStateTimeout            = 30
	KubernetesClientTimeOut    = 30
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
	for i, host := range kubeCluster.Hosts {
		if len(host.AdvertisedHostname) == 0 {
			return nil, fmt.Errorf("Hostname for host (%d) is not provided", i+1)
		} else if len(host.User) == 0 {
			return nil, fmt.Errorf("User for host (%d) is not provided", i+1)
		} else if len(host.Role) == 0 {
			return nil, fmt.Errorf("Role for host (%d) is not provided", i+1)

		} else if host.AdvertiseAddress == "" {
			// if control_plane_ip is not set,
			// default to the main IP
			kubeCluster.Hosts[i].AdvertiseAddress = host.IP
		}
		for _, role := range host.Role {
			if role != services.ETCDRole && role != services.ControlRole && role != services.WorkerRole {
				return nil, fmt.Errorf("Role [%s] for host (%d) is not recognized", role, i+1)
			}
		}
	}
	return &kubeCluster, nil
}

func GetLocalKubeConfig(configPath string) string {
	baseDir := filepath.Dir(configPath)
	fileName := filepath.Base(configPath)
	baseDir += "/"
	return fmt.Sprintf("%s%s%s", baseDir, pki.KubeAdminConfigPrefix, fileName)
}

func ReconcileCluster(kubeCluster, currentCluster *Cluster) error {
	logrus.Infof("[reconcile] Reconciling cluster state")
	if currentCluster == nil {
		logrus.Infof("[reconcile] This is newly generated cluster")
		return nil
	}
	if err := rebuildLocalAdminConfig(kubeCluster); err != nil {
		return err
	}
	kubeClient, err := k8s.NewClient(kubeCluster.LocalKubeConfigPath)
	if err != nil {
		return fmt.Errorf("Failed to initialize new kubernetes client: %v", err)
	}

	logrus.Infof("[reconcile] Check Control plane hosts to be deleted")
	cpToDelete := hosts.GetToDeleteHosts(currentCluster.ControlPlaneHosts, kubeCluster.ControlPlaneHosts)
	for _, toDeleteHost := range cpToDelete {
		hosts.DeleteNode(&toDeleteHost, kubeClient)
	}

	logrus.Infof("[reconcile] Check worker hosts to be deleted")
	wpToDelete := hosts.GetToDeleteHosts(currentCluster.WorkerHosts, kubeCluster.WorkerHosts)
	for _, toDeleteHost := range wpToDelete {
		hosts.DeleteNode(&toDeleteHost, kubeClient)
	}

	// Rolling update on change for nginx Proxy
	cpChanged := hosts.IsHostListChanged(currentCluster.ControlPlaneHosts, kubeCluster.ControlPlaneHosts)
	if cpChanged {
		logrus.Infof("[reconcile] Rolling update nginx hosts with new list of control plane hosts")
		err = services.RollingUpdateNginxProxy(kubeCluster.ControlPlaneHosts, kubeCluster.WorkerHosts)
		if err != nil {
			return fmt.Errorf("Failed to rolling update Nginx hosts with new control plane hosts")
		}
	}
	logrus.Infof("[reconcile] Reconciled cluster state successfully")
	return nil
}

func rebuildLocalAdminConfig(kubeCluster *Cluster) error {
	logrus.Infof("[reconcile] Rebuilding and update local kube config")
	currentKubeConfig := kubeCluster.Certificates[pki.KubeAdminCommonName]
	caCrt := kubeCluster.Certificates[pki.CACertName].Certificate
	newConfig := pki.GetKubeConfigX509WithData(
		"https://"+kubeCluster.ControlPlaneHosts[0].IP+":6443",
		pki.KubeAdminCommonName,
		string(cert.EncodeCertPEM(caCrt)),
		string(cert.EncodeCertPEM(currentKubeConfig.Certificate)),
		string(cert.EncodePrivateKeyPEM(currentKubeConfig.Key)))
	err := pki.DeployAdminConfig(newConfig, kubeCluster.LocalKubeConfigPath)
	if err != nil {
		return fmt.Errorf("Failed to redeploy local admin config with new host")
	}
	currentKubeConfig.Config = newConfig
	kubeCluster.Certificates[pki.KubeAdminCommonName] = currentKubeConfig
	return nil
}

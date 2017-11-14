package cluster

import (
	"fmt"
	"net"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	v1.RancherKubernetesEngineConfig `yaml:",inline"`
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

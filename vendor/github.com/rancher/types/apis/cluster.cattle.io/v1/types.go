package v1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterConditionType string

const (
	// ClusterConditionReady Cluster ready to serve API (healthy when true, unehalthy when false)
	ClusterConditionReady = "Ready"
	// ClusterConditionProvisioned Cluster is provisioned
	ClusterConditionProvisioned = "Provisioned"
	// ClusterConditionUpdating Cluster is being updating (upgrading, scaling up)
	ClusterConditionUpdating = "Updating"
	// ClusterConditionNoDiskPressure true when all cluster nodes have sufficient disk
	ClusterConditionNoDiskPressure = "NoDiskPressure"
	// ClusterConditionNoMemoryPressure true when all cluster nodes have sufficient memory
	ClusterConditionNoMemoryPressure = "NoMemoryPressure"
	// More conditions can be added if unredlying controllers request it
)

type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the the cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Spec ClusterSpec `json:"spec"`
	// Most recent observed status of the cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Status *ClusterStatus `json:"status"`
}

type ClusterSpec struct {
	GoogleKubernetesEngineConfig  *GoogleKubernetesEngineConfig  `json:"googleKubernetesEngineConfig,omitempty"`
	AzureKubernetesServiceConfig  *AzureKubernetesServiceConfig  `json:"azureKubernetesServiceConfig,omitempty"`
	RancherKubernetesEngineConfig *RancherKubernetesEngineConfig `json:"rancherKubernetesEngineConfig,omitempty"`
}

type ClusterStatus struct {
	//Conditions represent the latest available observations of an object's current state:
	//More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#typical-status-properties
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	//Component statuses will represent cluster's components (etcd/controller/scheduler) health
	// https://kubernetes.io/docs/api-reference/v1.8/#componentstatus-v1-core
	ComponentStatuses   []ClusterComponentStatus
	APIEndpoint         string          `json:"apiEndpoint,omitempty"`
	ServiceAccountToken string          `json:"serviceAccountToken,omitempty"`
	CACert              string          `json:"caCert,omitempty"`
	Capacity            v1.ResourceList `json:"capacity,omitempty"`
	Allocatable         v1.ResourceList `json:"allocatable,omitempty"`
	AppliedSpec         ClusterSpec     `json:"clusterSpec,omitempty"`
}

type ClusterComponentStatus struct {
	Name       string
	Conditions []v1.ComponentCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
}

type GoogleKubernetesEngineConfig struct {
	// ProjectID is the ID of your project to use when creating a cluster
	ProjectID string `json:"projectId,omitempty"`
	// The zone to launch the cluster
	Zone string `json:"zone,omitempty"`
	// The IP address range of the container pods
	ClusterIpv4Cidr string `json:"clusterIpv4Cidr,omitempty"`
	// An optional description of this cluster
	Description string `json:"description,omitempty"`
	// The number of nodes in this cluster
	NodeCount int64 `json:"nodeCount,omitempty"`
	// Size of the disk attached to each node
	DiskSizeGb int64 `json:"diskSizeGb,omitempty"`
	// The name of a Google Compute Engine
	MachineType string `json:"machineType,omitempty"`
	// Node kubernetes version
	NodeVersion string `json:"nodeVersion,omitempty"`
	// the master kubernetes version
	MasterVersion string `json:"masterVersion,omitempty"`
	// The map of Kubernetes labels (key/value pairs) to be applied
	// to each node.
	Labels map[string]string `json:"labels,omitempty"`
	// The path to the credential file(key.json)
	CredentialPath string `json:"credentialPath,omitempty"`
	// Enable alpha feature
	EnableAlphaFeature bool `json:"enableAlphaFeature,omitempty"`
}

type AzureKubernetesServiceConfig struct {
	//TBD
}

type RancherKubernetesEngineConfig struct {
	// Kubernetes nodes
	Hosts []RKEConfigHost `yaml:"hosts" json:"hosts,omitempty"`
	// Kubernetes components
	Services RKEConfigServices `yaml:"services" json:"services,omitempty"`
	// Network configuration used in the kubernetes cluster (flannel, calico)
	Network NetworkConfig `yaml:"network" json:"network,omitempty"`
	// Authentication configuration used in the cluster (default: x509)
	Authentication AuthConfig `yaml:"auth" json:"auth,omitempty"`
}

type RKEConfigHost struct {
	// SSH IP address of the host
	IP string `yaml:"ip" json:"ip,omitempty"`
	// Advertised address that will be used for components communication
	AdvertiseAddress string `yaml:"advertise_address" json:"advertiseAddress,omitempty"`
	// Host role in kubernetes cluster (controlplane, worker, or etcd)
	Role []string `yaml:"role" json:"role,omitempty"`
	// Hostname of the host
	AdvertisedHostname string `yaml:"advertised_hostname" json:"advertisedHostname,omitempty"`
	// SSH usesr that will be used by RKE
	User string `yaml:"user" json:"user,omitempty"`
	// Docker socket on the host that will be used in tunneling
	DockerSocket string `yaml:"docker_socket" json:"dockerSocket,omitempty"`
}

type RKEConfigServices struct {
	// Etcd Service
	Etcd ETCDService `yaml:"etcd" json:"etcd,omitempty"`
	// KubeAPI Service
	KubeAPI KubeAPIService `yaml:"kube-api" json:"kube-api,omitempty"`
	// KubeController Service
	KubeController KubeControllerService `yaml:"kube-controller" json:"kube-controller,omitempty"`
	// Scheduler Service
	Scheduler SchedulerService `yaml:"scheduler" json:"scheduler,omitempty"`
	// Kubelet Service
	Kubelet KubeletService `yaml:"kubelet" json:"kubelet,omitempty"`
	// KubeProxy Service
	Kubeproxy KubeproxyService `yaml:"kubeproxy" json:"kubeproxy,omitempty"`
}

type ETCDService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
}

type KubeAPIService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
	// Virtual IP range that will be used by Kubernetes services
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range" json:"serviceClusterIpRange,omitempty"`
}

type KubeControllerService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
	// CIDR Range for Pods in cluster
	ClusterCIDR string `yaml:"cluster_cidr" json:"clusterCidr,omitempty"`
	// Virtual IP range that will be used by Kubernetes services
	ServiceClusterIPRange string `yaml:"service_cluster_ip_range" json:"serviceClusterIpRange,omitempty"`
}

type KubeletService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
	// Domain of the cluster (default: "cluster.local")
	ClusterDomain string `yaml:"cluster_domain" json:"clusterDomain,omitempty"`
	// The image whose network/ipc namespaces containers in each pod will use
	InfraContainerImage string `yaml:"infra_container_image" json:"infraContainerImage,omitempty"`
	// Cluster DNS service ip
	ClusterDNSServer string `yaml:"cluster_dns_server" json:"clusterDnsServer,omitempty"`
}

type KubeproxyService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
}

type SchedulerService struct {
	// Base service properties
	BaseService `yaml:",inline" json:",inline"`
}

type BaseService struct {
	// Docker image of the service
	Image string `yaml:"image" json:"image,omitempty"`
	// Extra arguments that are added to the services
	ExtraArgs map[string]string `yaml:"extra_args" json:"extraArgs,omitempty"`
}

type NetworkConfig struct {
	// Network Plugin That will be used in kubernetes cluster
	Plugin string `yaml:"plugin" json:"plugin,omitempty"`
	// Plugin options to configure network properties
	Options map[string]string `yaml:"options" json:"options,omitempty"`
}

type AuthConfig struct {
	// Authentication strategy that will be used in kubernetes cluster
	Strategy string `yaml:"strategy" json:"strategy,omitempty"`
	// Authentication options
	Options map[string]string `yaml:"options" json:"options,omitempty"`
}

type ClusterNode struct {
	v1.Node
}

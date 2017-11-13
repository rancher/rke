package client

const (
	ClusterSpecType           = "clusterSpec"
	ClusterSpecFieldAKSConfig = "aksConfig"
	ClusterSpecFieldGKEConfig = "gkeConfig"
	ClusterSpecFieldRKEConfig = "rkeConfig"
)

type ClusterSpec struct {
	AKSConfig *AKSConfig `json:"aksConfig,omitempty"`
	GKEConfig *GKEConfig `json:"gkeConfig,omitempty"`
	RKEConfig *RKEConfig `json:"rkeConfig,omitempty"`
}

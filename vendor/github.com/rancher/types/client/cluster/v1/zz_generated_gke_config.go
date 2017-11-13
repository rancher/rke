package client

const (
	GKEConfigType                       = "gkeConfig"
	GKEConfigFieldClusterIpv4Cidr       = "clusterIpv4Cidr"
	GKEConfigFieldCredentialPath        = "credentialPath"
	GKEConfigFieldDescription           = "description"
	GKEConfigFieldDiskSizeGb            = "diskSizeGb"
	GKEConfigFieldEnableAlphaFeature    = "enableAlphaFeature"
	GKEConfigFieldInitialClusterVersion = "initialClusterVersion"
	GKEConfigFieldInitialNodeCount      = "initialNodeCount"
	GKEConfigFieldLabels                = "labels"
	GKEConfigFieldMachineType           = "machineType"
	GKEConfigFieldNodePoolID            = "nodePoolID"
	GKEConfigFieldProjectID             = "projectID"
	GKEConfigFieldUpdateConfig          = "updateConfig"
	GKEConfigFieldZone                  = "zone"
)

type GKEConfig struct {
	ClusterIpv4Cidr       string            `json:"clusterIpv4Cidr,omitempty"`
	CredentialPath        string            `json:"credentialPath,omitempty"`
	Description           string            `json:"description,omitempty"`
	DiskSizeGb            int64             `json:"diskSizeGb,omitempty"`
	EnableAlphaFeature    bool              `json:"enableAlphaFeature,omitempty"`
	InitialClusterVersion string            `json:"initialClusterVersion,omitempty"`
	InitialNodeCount      int64             `json:"initialNodeCount,omitempty"`
	Labels                map[string]string `json:"labels,omitempty"`
	MachineType           string            `json:"machineType,omitempty"`
	NodePoolID            string            `json:"nodePoolID,omitempty"`
	ProjectID             string            `json:"projectID,omitempty"`
	UpdateConfig          gkeUpdateConfig   `json:"updateConfig,omitempty"`
	Zone                  string            `json:"zone,omitempty"`
}

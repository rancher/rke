package client

const (
	gkeUpdateConfigType               = "gkeUpdateConfig"
	gkeUpdateConfigFieldMasterVersion = "masterVersion"
	gkeUpdateConfigFieldNodeCount     = "nodeCount"
	gkeUpdateConfigFieldNodeVersion   = "nodeVersion"
)

type gkeUpdateConfig struct {
	MasterVersion string `json:"masterVersion,omitempty"`
	NodeCount     int64  `json:"nodeCount,omitempty"`
	NodeVersion   string `json:"nodeVersion,omitempty"`
}

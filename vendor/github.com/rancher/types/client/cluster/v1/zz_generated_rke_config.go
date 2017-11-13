package client

const (
	RKEConfigType               = "rkeConfig"
	RKEConfigFieldAuthType      = "authType"
	RKEConfigFieldHosts         = "hosts"
	RKEConfigFieldNetworkPlugin = "networkPlugin"
	RKEConfigFieldServices      = "services"
)

type RKEConfig struct {
	AuthType      string            `json:"authType,omitempty"`
	Hosts         []RKEConfigHost   `json:"hosts,omitempty"`
	NetworkPlugin string            `json:"networkPlugin,omitempty"`
	Services      RKEConfigServices `json:"services,omitempty"`
}

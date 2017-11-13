package client

const (
	RKEConfigHostType                  = "rkeConfigHost"
	RKEConfigHostFieldAdvertiseAddress = "advertiseAddress"
	RKEConfigHostFieldDockerSocket     = "dockerSocket"
	RKEConfigHostFieldHostname         = "hostname"
	RKEConfigHostFieldIP               = "ip"
	RKEConfigHostFieldRole             = "role"
	RKEConfigHostFieldUser             = "user"
)

type RKEConfigHost struct {
	AdvertiseAddress string   `json:"advertiseAddress,omitempty"`
	DockerSocket     string   `json:"dockerSocket,omitempty"`
	Hostname         string   `json:"hostname,omitempty"`
	IP               string   `json:"ip,omitempty"`
	Role             []string `json:"role,omitempty"`
	User             string   `json:"user,omitempty"`
}

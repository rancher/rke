package client

const (
	DaemonEndpointType      = "daemonEndpoint"
	DaemonEndpointFieldPort = "port"
)

type DaemonEndpoint struct {
	Port int64 `json:"port,omitempty"`
}

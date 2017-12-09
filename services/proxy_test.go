package services

import (
	"fmt"
	"testing"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestNginxProxyImage            = "test/test:latest"
	TestNginxProxyConnectionString = "1.1.1.1,2.2.2.2"
)

func TestNginxProxyConfig(t *testing.T) {
	cpHosts := []*hosts.Host{
		&hosts.Host{
			RKEConfigNode: v3.RKEConfigNode{
				Address:          "1.1.1.1",
				InternalAddress:  "1.1.1.1",
				Role:             []string{"controlplane"},
				HostnameOverride: "cp1",
			},
			DClient: nil,
		},
		&hosts.Host{
			RKEConfigNode: v3.RKEConfigNode{
				Address:          "2.2.2.2",
				InternalAddress:  "2.2.2.2",
				Role:             []string{"controlplane"},
				HostnameOverride: "cp1",
			},
			DClient: nil,
		},
	}

	nginxProxyImage := TestNginxProxyImage
	nginxProxyEnv := buildProxyEnv(cpHosts)
	assertEqual(t, TestNginxProxyConnectionString, nginxProxyEnv,
		fmt.Sprintf("Failed to verify nginx connection string [%s]", TestNginxProxyConnectionString))

	imageCfg, hostCfg := buildNginxProxyConfig(nil, nginxProxyEnv, nginxProxyImage)
	// Test image and host config
	assertEqual(t, TestNginxProxyImage, imageCfg.Image,
		fmt.Sprintf("Failed to verify [%s] as Nginx Proxy Image", TestNginxProxyImage))
	assertEqual(t, true, hostCfg.NetworkMode.IsHost(),
		"Failed to verify that Nginx Proxy has host Network mode")
}

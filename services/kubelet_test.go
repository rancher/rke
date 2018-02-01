package services

import (
	"fmt"
	"testing"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestKubeletClusterDomain       = "cluster.local"
	TestKubeletClusterDNSServer    = "10.1.0.3"
	TestKubeletInfraContainerImage = "test/test:latest"
	TestKubeletImage               = "rancher/k8s:latest"
	TestKubeletVolumeBind          = "/etc/kubernetes:/etc/kubernetes"
	TestKubeletExtraArgs           = "--foo=bar"
	TestClusterDomainPrefix        = "--cluster-domain="
	TestClusterDNSServerPrefix     = "--cluster-dns="
	TestInfraContainerImagePrefix  = "--pod-infra-container-image="
	TestHostnameOverridePrefix     = "--hostname-override="
)

func TestKubeletConfig(t *testing.T) {

	host := &hosts.Host{
		RKEConfigNode: v3.RKEConfigNode{
			Address:          "1.1.1.1",
			InternalAddress:  "1.1.1.1",
			Role:             []string{"worker", "controlplane", "etcd"},
			HostnameOverride: "node1",
		},
		DClient: nil,
	}

	kubeletService := v3.KubeletService{}
	kubeletService.Image = TestKubeletImage
	kubeletService.ClusterDomain = TestKubeletClusterDomain
	kubeletService.ClusterDNSServer = TestKubeletClusterDNSServer
	kubeletService.InfraContainerImage = TestKubeletInfraContainerImage
	kubeletService.ExtraArgs = map[string]string{"foo": "bar"}

	imageCfg, hostCfg := buildKubeletConfig(host, kubeletService)
	// Test image and host config
	assertEqual(t, isStringInSlice(TestClusterDomainPrefix+TestKubeletClusterDomain, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet Command", TestClusterDomainPrefix+TestKubeletClusterDomain))
	assertEqual(t, isStringInSlice(TestClusterDNSServerPrefix+TestKubeletClusterDNSServer, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet Command", TestClusterDNSServerPrefix+TestKubeletClusterDNSServer))
	assertEqual(t, isStringInSlice(TestInfraContainerImagePrefix+TestKubeletInfraContainerImage, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet Command", TestInfraContainerImagePrefix+TestKubeletInfraContainerImage))
	assertEqual(t, TestKubeletImage, imageCfg.Image,
		fmt.Sprintf("Failed to verify [%s] as Kubelet Image", TestKubeletImage))
	assertEqual(t, isStringInSlice(TestHostnameOverridePrefix+host.HostnameOverride, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet Command", TestHostnameOverridePrefix+host.HostnameOverride))
	assertEqual(t, isStringInSlice(TestKubeletVolumeBind, hostCfg.Binds), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet Volume Binds", TestKubeletVolumeBind))
	assertEqual(t, isStringInSlice(TestKubeletExtraArgs, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Kubelet extra args", TestKubeletExtraArgs))
	assertEqual(t, true, hostCfg.Privileged,
		"Failed to verify that Kubelet is privileged")
	assertEqual(t, true, hostCfg.PidMode.IsHost(),
		"Failed to verify that Kubelet has host PID mode")
	assertEqual(t, true, hostCfg.NetworkMode.IsHost(),
		"Failed to verify that Kubelet has host Network mode")
}

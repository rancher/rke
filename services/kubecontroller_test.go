package services

import (
	"fmt"
	"testing"

	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestKubeControllerClusterCidr           = "10.0.0.0/16"
	TestKubeControllerServiceClusterIPRange = "10.1.0.0/16"
	TestKubeControllerImage                 = "rancher/k8s:latest"
	TestKubeControllerVolumeBind            = "/etc/kubernetes:/etc/kubernetes"
	TestKubeControllerExtraArgs             = "--foo=bar"
	TestClusterCidrPrefix                   = "--cluster-cidr="
	TestServiceIPRangePrefix                = "--service-cluster-ip-range="
)

func TestKubeControllerConfig(t *testing.T) {

	kubeControllerService := v3.KubeControllerService{}
	kubeControllerService.Image = TestKubeControllerImage
	kubeControllerService.ClusterCIDR = TestKubeControllerClusterCidr
	kubeControllerService.ServiceClusterIPRange = TestKubeControllerServiceClusterIPRange
	kubeControllerService.ExtraArgs = map[string]string{"foo": "bar"}

	imageCfg, hostCfg := buildKubeControllerConfig(kubeControllerService)
	// Test image and host config
	assertEqual(t, isStringInSlice(TestClusterCidrPrefix+TestKubeControllerClusterCidr, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in KubeController Command", TestClusterCidrPrefix+TestKubeControllerClusterCidr))
	assertEqual(t, isStringInSlice(TestServiceIPRangePrefix+TestKubeControllerServiceClusterIPRange, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in KubeController Command", TestServiceIPRangePrefix+TestKubeControllerServiceClusterIPRange))
	assertEqual(t, TestKubeControllerImage, imageCfg.Image,
		fmt.Sprintf("Failed to verify [%s] as KubeController Image", TestKubeControllerImage))
	assertEqual(t, isStringInSlice(TestKubeControllerVolumeBind, hostCfg.Binds), true,
		fmt.Sprintf("Failed to find [%s] in volume binds of KubeController", TestKubeControllerVolumeBind))
	assertEqual(t, isStringInSlice(TestKubeControllerExtraArgs, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in extra args of KubeController", TestKubeControllerExtraArgs))
	assertEqual(t, true, hostCfg.NetworkMode.IsHost(), "")
}

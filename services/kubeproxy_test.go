package services

import (
	"fmt"
	"testing"

	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestKubeproxyImage      = "rancher/k8s:latest"
	TestKubeproxyVolumeBind = "/etc/kubernetes:/etc/kubernetes"
	TestKubeproxyExtraArgs  = "--foo=bar"
)

func TestKubeproxyConfig(t *testing.T) {

	kubeproxyService := v3.KubeproxyService{}
	kubeproxyService.Image = TestKubeproxyImage
	kubeproxyService.ExtraArgs = map[string]string{"foo": "bar"}

	imageCfg, hostCfg := buildKubeproxyConfig(nil, kubeproxyService)
	// Test image and host config
	assertEqual(t, TestKubeproxyImage, imageCfg.Image,
		fmt.Sprintf("Failed to verify [%s] as KubeProxy Image", TestKubeproxyImage))
	assertEqual(t, isStringInSlice(TestKubeproxyVolumeBind, hostCfg.Binds), true,
		fmt.Sprintf("Failed to find [%s] in KubeProxy Volume Binds", TestKubeproxyVolumeBind))
	assertEqual(t, isStringInSlice(TestKubeproxyExtraArgs, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in KubeProxy extra args", TestKubeproxyExtraArgs))
	assertEqual(t, true, hostCfg.Privileged,
		"Failed to verify that KubeProxy is privileged")
	assertEqual(t, true, hostCfg.NetworkMode.IsHost(),
		"Failed to verify that KubeProxy has host Netowrk mode")
}

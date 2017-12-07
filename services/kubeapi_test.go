package services

import (
	"fmt"
	"testing"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestEtcdConnString      = "http://1.1.1.1:2379,http://2.2.2.2:2379"
	TestKubeAPIImage        = "rancher/k8s:latest"
	TestInsecureBindAddress = "--insecure-bind-address=127.0.0.1"
	TestKubeAPIVolumeBind   = "/etc/kubernetes:/etc/kubernetes"
	TestKubeAPIExtraArgs    = "--foo=bar"
)

func TestKubeAPIConfig(t *testing.T) {
	etcdHosts := []*hosts.Host{
		&hosts.Host{
			RKEConfigNode: v3.RKEConfigNode{
				Address:          "1.1.1.1",
				InternalAddress:  "1.1.1.1",
				Role:             []string{"etcd"},
				HostnameOverride: "etcd1",
			},
			DClient: nil,
		},
		&hosts.Host{
			RKEConfigNode: v3.RKEConfigNode{
				Address:          "2.2.2.2",
				InternalAddress:  "2.2.2.2",
				Role:             []string{"etcd"},
				HostnameOverride: "etcd2",
			},
			DClient: nil,
		},
	}

	cpHost := &hosts.Host{
		RKEConfigNode: v3.RKEConfigNode{
			Address:          "3.3.3.3",
			InternalAddress:  "3.3.3.3",
			Role:             []string{"controlplane"},
			HostnameOverride: "node1",
		},
		DClient: nil,
	}

	kubeAPIService := v3.KubeAPIService{}
	kubeAPIService.Image = TestKubeAPIImage
	kubeAPIService.ServiceClusterIPRange = "10.0.0.0/16"
	kubeAPIService.ExtraArgs = map[string]string{"foo": "bar"}
	// Test init cluster string
	etcdConnString := GetEtcdConnString(etcdHosts)
	assertEqual(t, etcdConnString, TestEtcdConnString, "")

	imageCfg, hostCfg := buildKubeAPIConfig(cpHost, kubeAPIService, etcdConnString)
	// Test image and host config
	assertEqual(t, isStringInSlice(TestInsecureBindAddress, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Entrypoint of KubeAPI", TestInsecureBindAddress))
	assertEqual(t, TestKubeAPIImage, imageCfg.Image,
		fmt.Sprintf("Failed to find correct image [%s] in KubeAPI Config", TestKubeAPIImage))
	assertEqual(t, isStringInSlice(TestKubeAPIVolumeBind, hostCfg.Binds), true,
		fmt.Sprintf("Failed to find [%s] in volume binds of KubeAPI", TestKubeAPIVolumeBind))
	assertEqual(t, isStringInSlice(TestKubeAPIExtraArgs, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in extra args of KubeAPI", TestKubeAPIExtraArgs))
}

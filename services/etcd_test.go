package services

import (
	"fmt"
	"testing"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestInitEtcdClusterString = "etcd-etcd1=https://1.1.1.1:2380,etcd-etcd2=https://2.2.2.2:2380"
	TestEtcdImage             = "etcd/etcdImage:latest"
	TestEtcdNamePrefix        = "--name=etcd-"
	TestEtcdVolumeBind        = "/var/lib/etcd:/etcd-data:z"
	TestEtcdExtraArgs         = "--foo=bar"
)

func TestEtcdConfig(t *testing.T) {
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

	etcdService := v3.ETCDService{}
	etcdService.Image = TestEtcdImage
	etcdService.ExtraArgs = map[string]string{"foo": "bar"}
	// Test init cluster string
	initCluster := getEtcdInitialCluster(etcdHosts)
	assertEqual(t, initCluster, TestInitEtcdClusterString, "")

	for _, host := range etcdHosts {
		nodeName := pki.GetEtcdCrtName(host.InternalAddress)
		imageCfg, hostCfg := buildEtcdConfig(host, etcdService, TestInitEtcdClusterString, nodeName)
		assertEqual(t, isStringInSlice(TestEtcdNamePrefix+host.HostnameOverride, imageCfg.Cmd), true,
			fmt.Sprintf("Failed to find [%s] in Etcd command", TestEtcdNamePrefix+host.HostnameOverride))
		assertEqual(t, TestEtcdImage, imageCfg.Image,
			fmt.Sprintf("Failed to verify [%s] as Etcd Image", TestEtcdImage))
		assertEqual(t, isStringInSlice(TestEtcdVolumeBind, hostCfg.Binds), true,
			fmt.Sprintf("Failed to find [%s] in volume binds of Etcd Service", TestEtcdVolumeBind))
		assertEqual(t, isStringInSlice(TestEtcdExtraArgs, imageCfg.Entrypoint), true,
			fmt.Sprintf("Failed to find [%s] in extra args of Etcd Service", TestEtcdExtraArgs))
	}
}

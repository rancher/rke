package cluster

import (
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"net"
	"os"
)

func GetLocalRKEConfig() *v3.RancherKubernetesEngineConfig {
	rkeLocalNode := GetLocalRKENodeConfig()
	rkeServices := v3.RKEConfigServices{
		Kubelet: v3.KubeletService{
			BaseService: v3.BaseService{
				Image:     DefaultK8sImage,
				ExtraArgs: map[string]string{"fail-swap-on": "false"},
			},
		},
	}
	return &v3.RancherKubernetesEngineConfig{
		Nodes:    []v3.RKEConfigNode{*rkeLocalNode},
		Services: rkeServices,
	}

}

func GetLocalRKENodeConfig() *v3.RKEConfigNode {
	localName, localAddress := localNodeNameResolver()
	logrus.Infof("Resolved local node name to %s [%s]", localName, localAddress)
	rkeLocalNode := &v3.RKEConfigNode{
		Address:          localAddress,
		HostnameOverride: localName,
		User:             LocalNodeUser,
		Role:             []string{services.ControlRole, services.WorkerRole, services.ETCDRole},
	}
	return rkeLocalNode
}

func localNodeNameResolver() (string, string) {
	name, err := os.Hostname()
	if err != nil {
		return LocalNodeHostname, LocalNodeAddress
	}
	addrs, err := net.LookupIP(name)
	if err != nil {
		return name, LocalNodeAddress
	}
	for _, a := range addrs {
		if ipv4 := a.To4(); ipv4 != nil {
			return name, ipv4.String()
		}
	}
	return name, LocalNodeAddress
}

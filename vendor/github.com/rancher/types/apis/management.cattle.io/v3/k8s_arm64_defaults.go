package v3

import (
	"fmt"
	"strings"
)

const (
	DefaultArm64K8s = "v1.11.6-rancher1-1"
)

var (
	Arm64K8sVersionsCurrent = []string{
		"v1.11.6-rancher1-1",
		"v1.12.5-rancher1-1",
		"v1.13.1-rancher1-1",
	}

	// Arm64K8sVersionToRKESystemImages is dynamically populated on init() with the latest versions
	Arm64K8sVersionToRKESystemImages map[string]RKESystemImages

	AllArm64K8sVersions = map[string]RKESystemImages{
		"v1.11.6-rancher1-1": {
			Etcd:                      m("jianghang8421/coreos-etcd-arm64:v3.2.18"),
			Kubernetes:                m("jianghang8421/hyperkube:v1.11.6-rancher1"),
			Alpine:                    m("jianghang8421/rke-tools:v0.1.18"),
			NginxProxy:                m("jianghang8421/rke-tools:v0.1.18"),
			CertDownloader:            m("jianghang8421/rke-tools:v0.1.18"),
			KubernetesServicesSidecar: m("jianghang8421/rke-tools:v0.1.18"),
			KubeDNS:                   m("jianghang8421/k8s-dns-kube-dns-arm64:1.14.10"),
			DNSmasq:                   m("jianghang8421/k8s-dns-dnsmasq-nanny-arm64:1.14.10"),
			KubeDNSSidecar:            m("jianghang8421/k8s-dns-sidecar-arm64:1.14.10"),
			KubeDNSAutoscaler:         m("jianghang8421/cluster-proportional-autoscaler-arm64:1.0.0"),
			Flannel:                   m("jianghang8421/coreos-flannel-arm64:v0.10.0"),
			FlannelCNI:                m("jianghang8421/coreos-flannel-cni-arm64:v0.3.0"),
			CalicoNode:                m("quay.io/calico/node:v3.1.3"),
			CalicoCNI:                 m("quay.io/calico/cni:v3.1.3"),
			CalicoCtl:                 m("quay.io/calico/ctl:v2.0.0"),
			CanalNode:                 m("quay.io/calico/node:v3.1.3"),
			CanalCNI:                  m("quay.io/calico/cni:v3.1.3"),
			CanalFlannel:              m("quay.io/coreos/flannel:v0.10.0"),
			WeaveNode:                 m("weaveworks/weave-kube:2.1.2"),
			WeaveCNI:                  m("weaveworks/weave-npc:2.1.2"),
			PodInfraContainer:         m("jianghang8421/pause-arm64:3.1"),
			Ingress:                   m("jianghang8421/nginx-ingress-controller:0.21.0-rancher1"),
			IngressBackend:            m("jianghang8421/nginx-ingress-controller-defaultbackend-arm64:1.4"),
			MetricsServer:             m("jianghang8421/metrics-server-arm64:v0.2.1"),
		},
		"v1.12.5-rancher1-1": {
			Etcd:                      m("jianghang8421/coreos-etcd-arm64:v3.2.24"),
			Kubernetes:                m("jianghang8421/hyperkube:v1.12.5-rancher1"),
			Alpine:                    m("jianghang8421/rke-tools:v0.1.23"),
			NginxProxy:                m("jianghang8421/rke-tools:v0.1.23"),
			CertDownloader:            m("jianghang8421/rke-tools:v0.1.23"),
			KubernetesServicesSidecar: m("jianghang8421/rke-tools:v0.1.23"),
			KubeDNS:                   m("jianghang8421/k8s-dns-kube-dns-arm64:1.14.13"),
			DNSmasq:                   m("jianghang8421/k8s-dns-dnsmasq-nanny-arm64:1.14.13"),
			KubeDNSSidecar:            m("jianghang8421/k8s-dns-sidecar-arm64:1.14.13"),
			KubeDNSAutoscaler:         m("jianghang8421/cluster-proportional-autoscaler-arm64:1.0.0"),
			Flannel:                   m("jianghang8421/coreos-flannel-arm64:v0.10.0"),
			FlannelCNI:                m("jianghang8421/coreos-flannel-cni-arm64:v0.3.0"),
			CalicoNode:                m("quay.io/calico/node:v3.1.3"),
			CalicoCNI:                 m("quay.io/calico/cni:v3.1.3"),
			CalicoCtl:                 m("quay.io/calico/ctl:v2.0.0"),
			CanalNode:                 m("quay.io/calico/node:v3.1.3"),
			CanalCNI:                  m("quay.io/calico/cni:v3.1.3"),
			CanalFlannel:              m("quay.io/coreos/flannel:v0.10.0"),
			WeaveNode:                 m("weaveworks/weave-kube:2.5.0"),
			WeaveCNI:                  m("weaveworks/weave-npc:2.5.0"),
			PodInfraContainer:         m("jianghang8421/pause-arm64:3.1"),
			Ingress:                   m("jianghang8421/nginx-ingress-controller:0.21.0-rancher1"),
			IngressBackend:            m("jianghang8421/nginx-ingress-controller-defaultbackend-arm64:1.4"),
			MetricsServer:             m("jianghang8421/metrics-server-arm64:v0.3.1"),
			CoreDNS:                   m("coredns/coredns:1.2.2"),
			CoreDNSAutoscaler:         m("jianghang8421/cluster-proportional-autoscaler-arm64:1.0.0"),
		},
		"v1.13.1-rancher1-1": {
			Etcd:                      m("jianghang8421/coreos-etcd-arm64:v3.2.24"),
			Kubernetes:                m("jianghang8421/hyperkube:v1.13.1-rancher1"),
			Alpine:                    m("jianghang8421/rke-tools:v0.1.23"),
			NginxProxy:                m("jianghang8421/rke-tools:v0.1.23"),
			CertDownloader:            m("jianghang8421/rke-tools:v0.1.23"),
			KubernetesServicesSidecar: m("jianghang8421/rke-tools:v0.1.23"),
			KubeDNS:                   m("jianghang8421/k8s-dns-kube-dns-arm64:1.15.0"),
			DNSmasq:                   m("jianghang8421/k8s-dns-dnsmasq-nanny-arm64:1.15.0"),
			KubeDNSSidecar:            m("jianghang8421/k8s-dns-sidecar-arm64:1.15.0"),
			KubeDNSAutoscaler:         m("jianghang8421/cluster-proportional-autoscaler-arm64:1.0.0"),
			Flannel:                   m("jianghang8421/coreos-flannel-arm64:v0.10.0"),
			FlannelCNI:                m("jianghang8421/coreos-flannel-cni-arm64:v0.3.0"),
			CalicoNode:                m("quay.io/calico/node:v3.4.0"),
			CalicoCNI:                 m("quay.io/calico/cni:v3.4.0"),
			CalicoCtl:                 m("quay.io/calico/ctl:v2.0.0"),
			CanalNode:                 m("quay.io/calico/node:v3.4.0"),
			CanalCNI:                  m("quay.io/calico/cni:v3.4.0"),
			CanalFlannel:              m("quay.io/coreos/flannel:v0.10.0"),
			WeaveNode:                 m("weaveworks/weave-kube:2.5.0"),
			WeaveCNI:                  m("weaveworks/weave-npc:2.5.0"),
			PodInfraContainer:         m("jianghang8421/pause-arm64:3.1"),
			Ingress:                   m("jianghang8421/nginx-ingress-controller:0.21.0-rancher1"),
			IngressBackend:            m("jianghang8421/nginx-ingress-controller-defaultbackend-arm64:1.4"),
			MetricsServer:             m("jianghang8421/metrics-server-arm64:v0.3.1"),
			CoreDNS:                   m("coredns/coredns:1.2.6"),
			CoreDNSAutoscaler:         m("jianghang8421/cluster-proportional-autoscaler-arm64:1.0.0"),
		},
	}
)

func initArm64() {
	if Arm64K8sVersionToRKESystemImages != nil {
		panic("Do not initialize or add values to Arm64K8sVersionToRKESystemImages")
	}

	Arm64K8sVersionToRKESystemImages = map[string]RKESystemImages{}

	for version, images := range AllArm64K8sVersions {
		longName := "jianghang8421/hyperkube:" + version
		if !strings.HasPrefix(longName, images.Kubernetes) {
			panic(fmt.Sprintf("For K8s version %s, the Kubernetes image tag should be a substring of %s, currently it is %s", version, version, images.Kubernetes))
		}
	}

	for _, latest := range Arm64K8sVersionsCurrent {
		images, ok := AllArm64K8sVersions[latest]
		if !ok {
			panic("K8s version " + " is not found in AllArm64K8sVersions map")
		}
		Arm64K8sVersionToRKESystemImages[latest] = images
	}

	if _, ok := Arm64K8sVersionToRKESystemImages[DefaultArm64K8s]; !ok {
		panic("Default K8s version " + DefaultArm64K8s + " is not found in k8sVersionsCurrent list")
	}
}

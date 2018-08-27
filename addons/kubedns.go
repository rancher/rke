package addons

import "github.com/rancher/rke/templates"

import "fmt"

const (
	KubeDNSImage           = "KubeDNSImage"
	DNSMasqImage           = "DNSMasqImage"
	KubeDNSSidecarImage    = "KubednsSidecarImage"
	KubeDNSAutoScalerImage = "KubeDNSAutoScalerImage"
	KubeDNSServer          = "ClusterDNSServer"
	KubeDNSClusterDomain   = "ClusterDomain"
	MetricsServerImage     = "MetricsServerImage"
	RBAC                   = "RBAC"
	MetricsServerOptions   = "MetricsServerOptions"
)

func GetKubeDNSManifest(kubeDNSConfig map[string]string) (string, error) {

	fmt.Println(">>>>> >>>>> /root/go/src/github.com/rancher/rke/addons/kubedns.go GetKubeDNSManifest")

	return templates.CompileTemplateFromMap(templates.KubeDNSTemplate, kubeDNSConfig)
}

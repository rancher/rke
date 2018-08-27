package addons

import "github.com/rancher/rke/templates"

import "fmt"

const (
	CoreDNSImage           = "CoreDNSImage"
        CoreDNSAutoScalerImage = "CoreDNSAutoScalerImage"
        CoreDNSServer          = "ClusterDNSServer"
        CoreDNSClusterDomain   = "ClusterDomain"
)

func GetCoreDNSManifest(coreDNSConfig map[string]string) (string, error) {

        fmt.Println(">>>>> >>>>> /root/go/src/github.com/rancher/rke/addons/coredns.go GetCoreDNSManifest")

	return templates.CompileTemplateFromMap(templates.CoreDNSTemplate, coreDNSConfig)
}

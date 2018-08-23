package addons

import "github.com/rancher/rke/templates"

const (
	CoreDNSImage           = "CoreDNSImage"
)

func GetCoreDNSManifest(coreDNSConfig map[string]string) (string, error) {

	return templates.CompileTemplateFromMap(templates.CoreDNSTemplate, coreDNSConfig)
}

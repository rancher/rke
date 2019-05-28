package addons

import (
	rkeData "github.com/rancher/kontainer-driver-metadata/rke/templates"
	"github.com/rancher/rke/templates"
)

func GetMetricsServerManifest(MetricsServerConfig interface{}) (string, error) {

	return templates.CompileTemplateFromMap(templates.GetDefaultVersionedTemplate(rkeData.MetricsServer), MetricsServerConfig)
}

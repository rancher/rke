package templates

import (
	"bytes"
	"text/template"

	"github.com/rancher/rke/metadata"

	"github.com/rancher/rke/util"
)

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func GetVersionedTemplates(templateName string, k8sVersion string) string {

	versionedTemplate := metadata.K8sVersionToTemplates[templateName]
	if t, ok := versionedTemplate[util.GetTagMajorVersion(k8sVersion)]; ok {
		return t
	}
	return versionedTemplate["default"]
}

func GetDefaultVersionedTemplate(templateName string) string {
	return metadata.K8sVersionToTemplates[templateName]["default"]
}

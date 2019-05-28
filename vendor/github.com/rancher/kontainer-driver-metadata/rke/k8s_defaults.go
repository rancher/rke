package rke

import (
	"fmt"
	"strings"

	"github.com/rancher/kontainer-driver-metadata/rke/templates"

	"github.com/rancher/types/apis/management.cattle.io/v3"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	"github.com/rancher/types/image"
)

var (
	m = image.Mirror

	// K8sVersionToRKESystemImages is dynamically populated on init() with the latest versions
	K8sVersionToRKESystemImages map[string]v3.RKESystemImages
	K8sVersionWindowsSystemImages map[string]v3.WindowsSystemImages

	// K8sVersionServiceOptions - service options per k8s version
	K8sVersionServiceOptions map[string]v3.KubernetesServicesOptions
	K8sVersionWindowsServiceOptions map[string]v3.KubernetesServicesOptions

	// K8sVersionToRKEVersions - min/max RKE versions per k8s version
	K8sVersionToRKEVersions map[string]v3.RKEVersions

	// Default k8s version per rke version
	RKEDefaultK8sVersions map[string]string

	// Addon Templates per K8s version / "default" where nothing changes for k8s version
	K8sVersionedTemplates map[string]map[string]string

	// ToolsSystemImages default images for alert, pipeline, logging, globaldns
	ToolsSystemImages = struct {
		AlertSystemImages    v3.AlertSystemImages
		PipelineSystemImages projectv3.PipelineSystemImages
		LoggingSystemImages  v3.LoggingSystemImages
		AuthSystemImages     v3.AuthSystemImages
	}{
		AlertSystemImages: v3.AlertSystemImages{
			AlertManager:       m("prom/alertmanager:v0.15.2"),
			AlertManagerHelper: m("rancher/alertmanager-helper:v0.0.2"),
		},
		PipelineSystemImages: projectv3.PipelineSystemImages{
			Jenkins:       m("rancher/pipeline-jenkins-server:v0.1.0"),
			JenkinsJnlp:   m("jenkins/jnlp-slave:3.10-1-alpine"),
			AlpineGit:     m("rancher/pipeline-tools:v0.1.9"),
			PluginsDocker: m("plugins/docker:17.12"),
			Minio:         m("minio/minio:RELEASE.2018-05-25T19-49-13Z"),
			Registry:      m("registry:2"),
			RegistryProxy: m("rancher/pipeline-tools:v0.1.9"),
			KubeApply:     m("rancher/pipeline-tools:v0.1.9"),
		},
		LoggingSystemImages: v3.LoggingSystemImages{
			Fluentd:                       m("rancher/fluentd:v0.1.11"),
			FluentdHelper:                 m("rancher/fluentd-helper:v0.1.2"),
			LogAggregatorFlexVolumeDriver: m("rancher/log-aggregator:v0.1.4"),
		},
		AuthSystemImages: v3.AuthSystemImages{
			KubeAPIAuth: m("rancher/kube-api-auth:v0.1.3"),
		},
	}

	AllK8sVersions map[string]v3.RKESystemImages
)

func InitRKE() {
	if K8sVersionToRKESystemImages != nil {
		panic("Do not initialize or add values to K8sVersionToRKESystemImages")
	}

	K8sVersionToRKESystemImages = loadK8sRKESystemImages()

	for version, images := range K8sVersionToRKESystemImages {
		longName := "rancher/hyperkube:" + version
		if !strings.HasPrefix(longName, images.Kubernetes) {
			panic(fmt.Sprintf("For K8s version %s, the Kubernetes image tag should be a substring of %s, currently it is %s", version, version, images.Kubernetes))
		}
	}

	K8sVersionServiceOptions = loadK8sVersionServiceOptions()
	K8sVersionToRKEVersions = loadK8sRKEVersions()
	RKEDefaultK8sVersions = loadRKEDefaultK8sVersions()
	K8sVersionedTemplates = templates.LoadK8sVersionedTemplates()

	for _, defaultK8s := range RKEDefaultK8sVersions {
		if _, ok := K8sVersionToRKESystemImages[defaultK8s]; !ok {
			panic(fmt.Sprintf("Default K8s version %v is not found in the K8sVersionToRKESystemImages", defaultK8s))
		}
	}

	// init Windows versions
	K8sVersionWindowsSystemImages = loadK8sVersionWindowsSystemimages()
	K8sVersionWindowsServiceOptions = loadK8sVersionWindowsServiceOptions()

}

package metadata

import (
	"context"
	rkeDriverMetadata "github.com/rancher/kontainer-driver-metadata/rke"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"strings"
)

var (
	RKEVersion                  string
	DefaultK8sVersion           string
	K8sVersionToTemplates       map[string]map[string]string
	K8sVersionToRKESystemImages map[string]v3.RKESystemImages
	K8sVersionToServiceOptions  map[string]v3.KubernetesServicesOptions
	K8sVersionsCurrent          []string
	K8sBadVersions              = map[string]bool{}
)

func InitMetadata(ctx context.Context) error {
	rkeDriverMetadata.InitRKE()
	initK8sRKESystemImages()
	initAddonTemplates()
	initServiceOptions()
	return nil
}

func initAddonTemplates() {
	K8sVersionToTemplates = rkeDriverMetadata.K8sVersionedTemplates
}

func initServiceOptions() {
	K8sVersionToServiceOptions = rkeDriverMetadata.K8sVersionServiceOptions
}

func initK8sRKESystemImages() {
	K8sVersionToRKESystemImages = map[string]v3.RKESystemImages{}
	if defaultK8s, ok := rkeDriverMetadata.RKEDefaultK8sVersions[RKEVersion]; ok {
		DefaultK8sVersion = defaultK8s
	}
	maxVersionForMajorK8sVersion := map[string]string{}
	for k8sVersion, systemImages := range rkeDriverMetadata.K8sVersionToRKESystemImages {
		if rkeVersionInfo, ok := rkeDriverMetadata.K8sVersionToRKEVersions[k8sVersion]; ok {
			greaterThanMax := rkeVersionInfo.MaxRKEVersion != "" && RKEVersion > rkeVersionInfo.MaxRKEVersion
			lowerThanMin := rkeVersionInfo.MinRKEVersion != "" && RKEVersion < rkeVersionInfo.MinRKEVersion
			if greaterThanMax || lowerThanMin {
				K8sBadVersions[k8sVersion] = true
				continue
			}
		}
		K8sVersionToRKESystemImages[k8sVersion] = systemImages
		majorVersion := getTagMajorVersion(k8sVersion)
		if curr, ok := maxVersionForMajorK8sVersion[majorVersion]; !ok || k8sVersion > curr {
			maxVersionForMajorK8sVersion[majorVersion] = k8sVersion
		}
	}
	for _, k8sVersion := range maxVersionForMajorK8sVersion {
		K8sVersionsCurrent = append(K8sVersionsCurrent, k8sVersion)
	}
}

func getTagMajorVersion(tag string) string {
	splitTag := strings.Split(tag, ".")
	if len(splitTag) < 2 {
		return ""
	}
	return strings.Join(splitTag[:2], ".")
}

package image

import "strings"

var Mirrors = map[string]string{}

func Mirror(image string) string {
	orig := image
	if strings.HasPrefix(image, "weaveworks") || strings.HasPrefix(image, "noiro") {
		return image
	}

	image = strings.Replace(image, "gcr.io/google_containers/", "rancher/mirrored-", 1)
	image = strings.Replace(image, "quay.io/coreos/", "rancher/mirrored-coreos-", 1)
	image = strings.Replace(image, "quay.io/calico/", "rancher/mirrored-calico-", 1)
	image = strings.Replace(image, "plugins/docker", "rancher/mirrored-plugins-docker", 1)
	image = strings.Replace(image, "k8s.gcr.io/defaultbackend", "rancher/mirrored-nginx-ingress-controller-defaultbackend", 1)
	image = strings.Replace(image, "k8s.gcr.io/k8s-dns-node-cache", "rancher/mirrored-k8s-dns-node-cache", 1)
	image = strings.Replace(image, "kibana", "rancher/mirrored-kibana", 1)
	image = strings.Replace(image, "jenkins/", "rancher/mirrored-jenkins-", 1)
	image = strings.Replace(image, "alpine/git", "rancher/alpine-git", 1)
	image = strings.Replace(image, "prom/", "rancher/mirrored-prom-", 1)
	image = strings.Replace(image, "quay.io/pires/", "rancher/mirrored-", 1)
	image = strings.Replace(image, "coredns/", "rancher/mirrored-coredns-", 1)
	image = strings.Replace(image, "minio/", "rancher/mirrored-minio-", 1)

	Mirrors[image] = orig
	return image
}

package pki

func getKubeConfigX509(kubernetesURL string, componentName string, caPath string, crtPath string, keyPath string) string {
	return `apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    certificate-authority: ` + caPath + `
    server: "` + kubernetesURL + `"
  name: "local"
contexts:
- context:
    cluster: "local"
    user: "` + componentName + `"
  name: "Default"
current-context: "Default"
users:
- name: "` + componentName + `"
  user:
    client-certificate: ` + crtPath + `
    client-key: ` + keyPath + ``
}

package templates

// Secrets encryption template. Secret is 32byte base64 encoded string.
var (
	EncryptionConfigurationTemplate = map[string]string{
		"v1.13": `apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: {{ .Secret }}
    - identity: {}`,
		"v1.12": `apiVersion: v1
kind: EncryptionConfig
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: {{ .Secret }}
    - identity: {}`,
		"v1.11": `apiVersion: v1
kind: EncryptionConfig
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: {{ .Secret }}
    - identity: {}`,
	}
)

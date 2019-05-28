package rke

import "github.com/rancher/types/apis/management.cattle.io/v3"

const (
	tlsCipherSuites        = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
	enableAdmissionPlugins = "NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
)

func loadK8sVersionServiceOptions() map[string]v3.KubernetesServicesOptions {
	return map[string]v3.KubernetesServicesOptions{
		"v1.14": {
			KubeAPI: map[string]string{
				"tls-cipher-suites":        tlsCipherSuites,
				"enable-admission-plugins": enableAdmissionPlugins,
			},
			Kubelet: map[string]string{
				"tls-cipher-suites": tlsCipherSuites,
			},
		},
		"v1.13": {
			KubeAPI: map[string]string{
				"tls-cipher-suites":        tlsCipherSuites,
				"enable-admission-plugins": enableAdmissionPlugins,
				"repair-malformed-updates": "false",
			},
			Kubelet: map[string]string{
				"tls-cipher-suites": tlsCipherSuites,
			},
		},
		"v1.12": {
			KubeAPI: map[string]string{
				"tls-cipher-suites":        tlsCipherSuites,
				"enable-admission-plugins": enableAdmissionPlugins,
				"repair-malformed-updates": "false",
			},
			Kubelet: map[string]string{
				"tls-cipher-suites": tlsCipherSuites,
			},
		},
		"v1.11": {
			KubeAPI: map[string]string{
				"tls-cipher-suites":        tlsCipherSuites,
				"enable-admission-plugins": enableAdmissionPlugins,
				"repair-malformed-updates": "false",
			},
			Kubelet: map[string]string{
				"tls-cipher-suites": tlsCipherSuites,
				"cadvisor-port":     "0",
			},
		},
		"v1.10": {
			KubeAPI: map[string]string{
				"tls-cipher-suites":        tlsCipherSuites,
				"endpoint-reconciler-type": "lease",
				"enable-admission-plugins": enableAdmissionPlugins,
				"repair-malformed-updates": "false",
			},
			Kubelet: map[string]string{
				"tls-cipher-suites": tlsCipherSuites,
				"cadvisor-port":     "0",
			},
		},
		"v1.9": {
			KubeAPI: map[string]string{
				"endpoint-reconciler-type": "lease",
				"admission-control":        "ServiceAccount,NamespaceLifecycle,LimitRanger,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota,DefaultTolerationSeconds",
				"repair-malformed-updates": "false",
			},
			Kubelet: map[string]string{
				"cadvisor-port": "0",
			},
		},
	}
}

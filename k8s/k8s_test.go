package k8s

import (
	"testing"

	check "gopkg.in/check.v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	KubeConfigPath = "/etc/kubernetes/ssl/kubeconfig"
	ConfigMapName  = "testconfigmap"
	ConfigYaml     = `---
foo: bar
test: test123`
	SecretFile = `secret123`
	SecretName = "secret"
)

type KubernetesOperationsTestSuite struct {
	kubeClient *kubernetes.Clientset
}

func Test(t *testing.T) {
	check.TestingT(t)
}

var _ = check.Suite(&KubernetesOperationsTestSuite{})

func (k *KubernetesOperationsTestSuite) SetUpSuite(c *check.C) {
	var err error
	k.kubeClient, err = NewClient(KubeConfigPath)
	meta := metav1.ObjectMeta{Name: metav1.NamespaceSystem}
	ns := &v1.Namespace{
		ObjectMeta: meta,
	}
	if _, err = k.kubeClient.CoreV1().Namespaces().Create(ns); err != nil {
		c.Fatalf("Failed to set up test suite: %v", err)
	}

}

func (k *KubernetesOperationsTestSuite) TestSaveConfig(c *check.C) {
	var err error
	if err != nil {
		c.Fatalf("Failed to initialize kubernetes client")
	}

	// Make sure that config yaml file can be stored as a config map
	err = UpdateConfigMap(k.kubeClient, []byte(ConfigYaml), ConfigMapName)
	if err != nil {
		c.Fatalf("Failed to store config map %s: %v", ConfigMapName, err)
	}

	cfgMap, err := GetConfigMap(k.kubeClient, ConfigMapName)
	if err != nil {
		c.Fatalf("Failed to fetch config map %s: %v", ConfigMapName, err)
	}

	if cfgMap.Data[ConfigMapName] != ConfigYaml {
		c.Fatalf("Failed to verify the config map %s: %v", ConfigMapName, err)
	}
}

func (k *KubernetesOperationsTestSuite) TestSaveSecret(c *check.C) {
	var err error
	if err != nil {
		c.Fatalf("Failed to initialize kubernetes client")
	}

	err = UpdateSecret(k.kubeClient, SecretName, []byte(SecretFile), SecretName)
	if err != nil {
		c.Fatalf("Failed to store secret %s: %v", SecretName, err)
	}

	secret, err := GetSecret(k.kubeClient, SecretName)
	if err != nil {
		c.Fatalf("Failed to fetch secret %s: %v", SecretName, err)
	}

	if string(secret.Data[SecretName]) != SecretFile {
		c.Fatalf("Failed to verify the secret %s: %v", SecretName, err)
	}
}

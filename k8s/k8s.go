package k8s

import (
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(kubeConfigPath string) (*kubernetes.Clientset, error) {
	// use the current admin kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	K8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return K8sClientSet, nil
}

func UpdateConfigMap(k8sClient *kubernetes.Clientset, configYaml []byte, configMapName string) error {
	cfgMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			configMapName: string(configYaml),
		},
	}

	if _, err := k8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(cfgMap); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		if _, err := k8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(cfgMap); err != nil {
			return err
		}
	}
	return nil
}

func GetConfigMap(k8sClient *kubernetes.Clientset, configMapName string) (*v1.ConfigMap, error) {
	return k8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(configMapName, metav1.GetOptions{})
}

func UpdateSecret(k8sClient *kubernetes.Clientset, fieldName string, secretData []byte, secretName string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			fieldName: secretData,
		},
	}
	if _, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Create(secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		// update secret if its already exist
		oldSecret, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newData := oldSecret.Data
		newData[fieldName] = secretData
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: metav1.NamespaceSystem,
			},
			Data: newData,
		}
		if _, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Update(secret); err != nil {
			return err
		}
	}
	return nil
}

func GetSecret(k8sClient *kubernetes.Clientset, secretName string) (*v1.Secret, error) {
	return k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
}

func DeleteNode(k8sClient *kubernetes.Clientset, nodeName string) error {
	return k8sClient.CoreV1().Nodes().Delete(nodeName, &metav1.DeleteOptions{})
}

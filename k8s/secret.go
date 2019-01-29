package k8s

import (
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetSecret(k8sClient *kubernetes.Clientset, secretName string) (*v1.Secret, error) {
	return k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
}

// GetAllSecrets - Iterate through namespaces and return a slice of secrets.
func GetAllSecrets(k8sClient *kubernetes.Clientset) ([]v1.Secret, error) {
	secrets := []v1.Secret{}
	namespaces, err := k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		s, err := k8sClient.CoreV1().Secrets(ns.ObjectMeta.Name).List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, s.Items...)
	}

	return secrets, nil
}

func UpdateSecret(k8sClient *kubernetes.Clientset, secretDataMap map[string][]byte, secretName string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: secretDataMap,
	}
	if _, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Create(secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		// update secret if its already exist
		if _, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Update(secret); err != nil {
			return err
		}
	}
	return nil
}

// UpdateSecretObject - Update a secret with v1.Secret object.
func UpdateSecretObject(k8sClient *kubernetes.Clientset, secret *v1.Secret) error {
	// Clear version so we can submit as update.
	secret.SetResourceVersion("")
	_, err := k8sClient.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(secret)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		// update secret if its already exist
		_, err := k8sClient.CoreV1().Secrets(secret.ObjectMeta.Namespace).Update(secret)
		if err != nil {
			return err
		}
	}
	return nil
}

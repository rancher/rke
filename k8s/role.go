package k8s

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func UpdateRoleBindingFromYaml(k8sClient *kubernetes.Clientset, roleBindingYaml, namespace string) error {
	roleBinding := rbacv1.RoleBinding{}
	if err := DecodeYamlResource(&roleBinding, roleBindingYaml); err != nil {
		return err
	}
	roleBinding.Namespace = namespace
	return retryTo(updateRoleBinding, k8sClient, roleBinding, DefaultRetries, DefaultSleepSeconds)
}

func updateRoleBinding(k8sClient *kubernetes.Clientset, rb interface{}) error {
	roleBinding := rb.(rbacv1.RoleBinding)
	if _, err := k8sClient.RbacV1().RoleBindings(roleBinding.Namespace).Create(context.TODO(), &roleBinding, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		if _, err := k8sClient.RbacV1().RoleBindings(roleBinding.Namespace).Update(context.TODO(), &roleBinding, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func UpdateRoleFromYaml(k8sClient *kubernetes.Clientset, roleYaml, namespace string) error {
	role := rbacv1.Role{}
	if err := DecodeYamlResource(&role, roleYaml); err != nil {
		return err
	}
	role.Namespace = namespace
	return retryTo(updateRole, k8sClient, role, DefaultRetries, DefaultSleepSeconds)
}

func updateRole(k8sClient *kubernetes.Clientset, r interface{}) error {
	role := r.(rbacv1.Role)
	if _, err := k8sClient.RbacV1().Roles(role.Namespace).Create(context.TODO(), &role, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		if _, err := k8sClient.RbacV1().Roles(role.Namespace).Update(context.TODO(), &role, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

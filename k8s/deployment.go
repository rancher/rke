package k8s

import (
	"fmt"
	"time"

	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDeployment(k8sClient *kubernetes.Clientset, deploymentName, namespace string) (*v1beta1.Deployment, error) {
	return k8sClient.AppsV1beta1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
}

func UpdateDeployment(k8sClient *kubernetes.Clientset, deploymentName, namespace string, deploymentObj *v1beta1.Deployment) error {
	deployment, err := k8sClient.AppsV1beta1().Deployments(namespace).Update(deploymentObj)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 5)

	return retryToWithTimeout(ensureDeploymentAvailable, k8sClient, deployment, DefaultTimeout)
}

func ensureDeploymentAvailable(k8sClient *kubernetes.Clientset, d interface{}) error {
	dObj := d.(*v1beta1.Deployment)
	deployment, err := GetDeployment(k8sClient, dObj.Name, dObj.Namespace)
	if err != nil {
		return fmt.Errorf("Failed to get deployment status: %v", err)
	}
	if deployment.Status.AvailableReplicas == 0 {
		return fmt.Errorf("No available replicas for deployment [%s]", deployment.Name)
	}
	return nil
}

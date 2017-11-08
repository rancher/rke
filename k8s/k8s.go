package k8s

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rancher/rke/docker"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubectlImage    = "melsayed/kubectl:latest"
	KubctlContainer = "kubectl"
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
	if _, err := k8sClient.ConfigMaps(metav1.NamespaceSystem).Create(cfgMap); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		if _, err := k8sClient.ConfigMaps(metav1.NamespaceSystem).Update(cfgMap); err != nil {
			return err
		}
	}
	return nil
}

func GetConfigMap(k8sClient *kubernetes.Clientset, configMapName string) (*v1.ConfigMap, error) {
	return k8sClient.ConfigMaps(metav1.NamespaceSystem).Get(configMapName, metav1.GetOptions{})
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
	if _, err := k8sClient.Secrets(metav1.NamespaceSystem).Create(secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		// update secret if its already exist
		oldSecret, err := k8sClient.Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
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
		if _, err := k8sClient.Secrets(metav1.NamespaceSystem).Update(secret); err != nil {
			return err
		}
	}
	return nil
}

func GetSecret(k8sClient *kubernetes.Clientset, secretName string) (*v1.Secret, error) {
	return k8sClient.Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
}

func DeleteNode(k8sClient *kubernetes.Clientset, nodeName string) error {
	return k8sClient.Nodes().Delete(nodeName, &metav1.DeleteOptions{})
}

func RunKubectlCmd(dClient *client.Client, hostname string, cmd []string, withEnv []string) error {

	logrus.Debugf("[kubectl] Using host [%s] for deployment", hostname)
	logrus.Debugf("[kubectl] Pulling kubectl image..")

	if err := docker.PullImage(dClient, hostname, KubectlImage); err != nil {
		return err
	}
	env := []string{}
	if withEnv != nil {
		env = append(env, withEnv...)
	}
	imageCfg := &container.Config{
		Image: KubectlImage,
		Env:   env,
		Cmd:   cmd,
	}
	logrus.Debugf("[kubectl] Creating kubectl container..")
	resp, err := dClient.ContainerCreate(context.Background(), imageCfg, nil, nil, KubctlContainer)
	if err != nil {
		return fmt.Errorf("Failed to create kubectl container on host [%s]: %v", hostname, err)
	}
	logrus.Debugf("[kubectl] Container %s created..", resp.ID)
	if err := dClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start kubectl container on host [%s]: %v", hostname, err)
	}
	logrus.Debugf("[kubectl] running command: %s", cmd)
	statusCh, errCh := dClient.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("Failed to execute kubectl container on host [%s]: %v", hostname, err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("kubectl command failed on host [%s]: exit status %v", hostname, status.StatusCode)
		}
	}
	if err := dClient.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("Failed to remove kubectl container on host[%s]: %v", hostname, err)
	}
	return nil
}

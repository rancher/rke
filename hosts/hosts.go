package hosts

import (
	"github.com/docker/docker/client"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/io.cattle.cluster/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type Host struct {
	v1.RKEConfigHost
	DClient *client.Client
}

func ReconcileWorkers(currentWorkers []Host, newWorkers []Host, kubeClient *kubernetes.Clientset) error {
	for _, currentWorker := range currentWorkers {
		found := false
		for _, newWorker := range newWorkers {
			if currentWorker.Hostname == newWorker.Hostname {
				found = true
			}
		}
		if !found {
			if err := deleteWorkerNode(&currentWorker, kubeClient); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteWorkerNode(workerNode *Host, kubeClient *kubernetes.Clientset) error {
	logrus.Infof("[hosts] Deleting host [%s] from the cluster", workerNode.Hostname)
	err := k8s.DeleteNode(kubeClient, workerNode.Hostname)
	if err != nil {
		return err
	}
	logrus.Infof("[hosts] Successfully deleted host [%s] from the cluster", workerNode.Hostname)
	return nil
}

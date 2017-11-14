package hosts

import (
	"github.com/docker/docker/client"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
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
			if currentWorker.AdvertisedHostname == newWorker.AdvertisedHostname {
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
	logrus.Infof("[hosts] Deleting host [%s] from the cluster", workerNode.AdvertisedHostname)
	err := k8s.DeleteNode(kubeClient, workerNode.AdvertisedHostname)
	if err != nil {
		return err
	}
	logrus.Infof("[hosts] Successfully deleted host [%s] from the cluster", workerNode.AdvertisedHostname)
	return nil
}

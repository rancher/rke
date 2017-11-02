package hosts

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/rancher/rke/k8s"
	"k8s.io/client-go/kubernetes"
)

type Host struct {
	IP               string   `yaml:"ip"`
	AdvertiseAddress string   `yaml:"advertise_address"`
	Role             []string `yaml:"role"`
	Hostname         string   `yaml:"hostname"`
	User             string   `yaml:"user"`
	DockerSocket     string   `yaml:"docker_socket"`
	DClient          *client.Client
}

func ReconcileWorkers(currentWorkers []Host, newWorkers []Host, kClient *kubernetes.Clientset) error {
	for _, currentWorker := range currentWorkers {
		found := false
		for _, newWorker := range newWorkers {
			if currentWorker.Hostname == newWorker.Hostname {
				found = true
			}
		}
		if !found {
			if err := deleteWorkerNode(&currentWorker, kClient); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteWorkerNode(workerNode *Host, kClient *kubernetes.Clientset) error {
	logrus.Infof("[hosts] Deleting host [%s] from the cluster", workerNode.Hostname)
	err := k8s.DeleteNode(kClient, workerNode.Hostname)
	if err != nil {
		return err
	}
	logrus.Infof("[hosts] Successfully deleted host [%s] from the cluster", workerNode.Hostname)
	return nil
}

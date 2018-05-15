package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/log"
)

const (
	RancherDeploymentName = "cattle"
	RancherNameSpace      = "cattle-system"
	RancherUpdatedLabel   = "io.rancher.updated"
)

func UpdateRancherVersion(ctx context.Context, localConfigPath, rancherServerTag string) error {
	log.Infof(ctx, "Upgrading rancher 2.x deployment to tag [%s]", rancherServerTag)

	// simple validation
	if strings.Contains(rancherServerTag, "/") ||
		strings.Contains(rancherServerTag, ":") {
		return fmt.Errorf("Rancher tag validation failed. Please use docker image tag, e.g.: v2.0.0")
	}
	// not going to use a k8s dialer here.. this is a CLI command
	k8sClient, err := k8s.NewClient(localConfigPath, nil)
	if err != nil {
		return fmt.Errorf("Failed to create Kubernetes Client: %v", err)
	}
	rancherDeployment, err := k8s.GetDeployment(k8sClient, RancherDeploymentName, RancherNameSpace)
	if err != nil {
		return err
	}

	baseImage := strings.Split(rancherDeployment.Spec.Template.Spec.Containers[0].Image, ":")[0]
	// setting the new image in the deployment spec
	rancherDeployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", baseImage, rancherServerTag)

	// k8s doesn't like to use fixed tags like 'stable' and 'latest'. If these are used, the deployment wouldn't get updated.
	// To work around this, we add a timestamp label to the pod. All this does is change the pod template hash and forces the upgrade even with the same image tag
	rancherDeployment.Spec.Template.Labels[RancherUpdatedLabel] = fmt.Sprintf("%d", rancherDeployment.Generation+1)

	if err := k8s.UpdateDeployment(k8sClient, RancherDeploymentName, RancherNameSpace, rancherDeployment); err != nil {
		return fmt.Errorf("Failed to upgrade Rancher 2.x deployment: %v", err)
	}

	log.Infof(ctx, "Upgrading rancher 2.x deployment to tag [%s] completed successfully", rancherServerTag)
	return nil
}

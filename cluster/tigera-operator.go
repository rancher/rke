package cluster

import (
	"context"

	operator "github.com/tigera/operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
)

func tigeraClient(config *rest.Config) (*rest.RESTClient, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: operator.GroupVersion.Group, Version: operator.GroupVersion.Version}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.UnversionedRESTClientFor(&crdConfig)
}

// check if the Tigera Operator is used by Calico CNI
func (c *Cluster) checkTigeraOperator(ctx context.Context) (bool, error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.LocalKubeConfigPath)
	if err != nil {
		return false, err
	}

	client, err := tigeraClient(config)
	if err != nil {
		return false, err
	}

	installation := operator.InstallationList{}
	err = client.
		Get().
		Resource("installations").
		Do(ctx).
		Into(&installation)

	//If no Installation is found, assume no Operator is used
	if err != nil || len(installation.Items) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

package cluster

import (
	"testing"

	"github.com/rancher/rke/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateNetworkOptions(t *testing.T) {
	t.Run("weave with k8s v1.30.0 or greater", func(tt *testing.T) {
		cluster := &Cluster{
			RancherKubernetesEngineConfig: types.RancherKubernetesEngineConfig{
				Version: "v1.30.0-rancher1",
				Network: types.NetworkConfig{
					Plugin: WeaveNetworkPlugin,
				},
			},
		}

		err := validateNetworkOptions(cluster)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "weave CNI support is removed for k8s version >=1.30.0")

		cluster.Version = "v1.30.1-rancher1"
		err = validateNetworkOptions(cluster)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "weave CNI support is removed for k8s version >=1.30.0")
	})

	t.Run("weave with k8s version less than v1.30.0", func(tt *testing.T) {
		cluster := &Cluster{
			RancherKubernetesEngineConfig: types.RancherKubernetesEngineConfig{
				Version: "v1.29.0-rancher1",
				Network: types.NetworkConfig{
					Plugin: WeaveNetworkPlugin,
				},
			},
		}

		err := validateNetworkOptions(cluster)
		assert.Nil(t, err)

		cluster.Version = "v1.28.5-rancher1"
		err = validateNetworkOptions(cluster)
		assert.Nil(t, err)
	})

}

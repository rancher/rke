package cluster

import (
	"context"
	"testing"

	"github.com/rancher/rke/metadata"
	"github.com/rancher/rke/types"
	v3 "github.com/rancher/rke/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRestoreImage(t *testing.T) {
	ctx := context.Background()

	metadata.InitMetadata(ctx)

	cluster := &Cluster{
		RancherKubernetesEngineConfig: v3.RancherKubernetesEngineConfig{
			SystemImages: types.RKESystemImages{
				Etcd:   "rancher/mirrored-coreos-etcd:v3.5.7",
				Alpine: "rancher/rke-tools:v0.1.90",
			},
		},
	}

	expectedRestoreImage := cluster.getBackupImage()
	restoreImage := cluster.getRestoreImage()

	assert.NotEmpty(t, restoreImage, "")
	assert.Equal(t, expectedRestoreImage, restoreImage,
		"expected restoreImage is different when etcd image version is v3.5.7")

	cluster.SystemImages.Etcd = "rancher/mirrored-coreos-etcd:v3.5.8"

	expectedRestoreImage = cluster.getBackupImage()
	restoreImage = cluster.getRestoreImage()

	assert.NotEmpty(t, restoreImage, "")
	assert.Equal(t, expectedRestoreImage, restoreImage,
		"expected restoreImage is different when etcd image version is greater than v3.5.7")

	cluster.SystemImages.Etcd = "rancher/mirrored-coreos-etcd:v3.5.6"

	expectedRestoreImage = cluster.SystemImages.Etcd
	restoreImage = cluster.getRestoreImage()

	assert.NotEmpty(t, restoreImage, "")
	assert.Equal(t, expectedRestoreImage, restoreImage,
		"expected restoreImage is different when etcd image version is less than v3.5.7")

	// test for custom image
	cluster.SystemImages.Etcd = "custom/mirrored-coreos-etcd:v3.5.7"

	expectedRestoreImage = cluster.SystemImages.Etcd
	restoreImage = cluster.getRestoreImage()

	assert.NotEmpty(t, restoreImage, "")
	assert.Equal(t, expectedRestoreImage, restoreImage,
		"expected restoreImage is different when custom etcd image is used")
}

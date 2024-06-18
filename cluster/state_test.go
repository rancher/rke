package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/rancher/rke/pki"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/kubernetes/fake"
)

func setup(t *testing.T, withConfigMap bool) (context.Context, FullState, kubernetes.Interface) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	fullState := FullState{
		CurrentState: State{
			RancherKubernetesEngineConfig: GetLocalRKEConfig(),
			CertificatesBundle: map[string]pki.CertificatePKI{
				"test": {
					CertificatePEM: "fake cert",
					KeyPEM:         "fake key",
				},
			},
		},
	}

	if withConfigMap {
		fullStateBytes, err := json.Marshal(fullState)
		assert.NoError(t, err)

		_, err = client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: FullStateConfigMapName,
			},
			Data: map[string]string{
				FullStateConfigMapName: string(fullStateBytes),
			},
		}, metav1.CreateOptions{})
		assert.NoError(t, err)
	}

	return ctx, fullState, client
}

func checkSecretMatches(t *testing.T, ctx context.Context, client kubernetes.Interface, expected FullState) {
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(ctx, FullStateSecretName, metav1.GetOptions{})
	assert.NoError(t, err)
	fullStateFromSecret := FullState{}
	err = json.Unmarshal(secret.Data[FullStateConfigMapName], &fullStateFromSecret)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(fullStateFromSecret, expected))
}

func checkConfigMapDeleted(t *testing.T, ctx context.Context, client kubernetes.Interface) {
	_, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ctx, FullStateConfigMapName, metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))
}

func TestSaveFullStateToK8s_Nil(t *testing.T) {
	err := SaveFullStateToK8s(context.Background(), &fake.Clientset{}, nil)
	assert.True(t, errors.Is(err, ErrFullStateIsNil))
}

// Tests the scenario where the cluster stores no existing state. In this case, a new full state secret should be
// created and the old configmap should be deleted.
func TestSaveAndGetFullStateFromK8s_ClusterWithoutSecretOrCM(t *testing.T) {
	// Set up a fake cluster without a secret or configmap.
	ctx, fullState, client := setup(t, false)

	// We should not be able to fetch and load the state from the secret or configmap.
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	fetchedFullState, err := GetFullStateFromK8s(ctx, client)
	assert.True(t, apierrors.IsNotFound(err))

	// Create the secret and delete the configmap.
	err = SaveFullStateToK8s(ctx, client, &fullState)
	assert.NoError(t, err)

	// There should be a secret containing the full state.
	checkSecretMatches(t, ctx, client, fullState)

	// There should be no configmap.
	checkConfigMapDeleted(t, ctx, client)

	// We should be able to fetch and load the state from the secret.
	fetchedFullState, err = GetFullStateFromK8s(ctx, client)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(*fetchedFullState, fullState))
}

// Tests the scenario where the cluster already stores a full state secret but no configmap. In this case, the secret
// should be updated and there should still be no configmap.
func TestSaveAndGetFullStateFromK8s_ClusterWithSecretAndNoCM(t *testing.T) {
	// Set up a fake cluster without a secret or configmap.
	ctx, fullState, client := setup(t, false)

	// Add the secret to the cluster.
	err := SaveFullStateToK8s(ctx, client, &fullState)
	assert.NoError(t, err)

	// There should be a secret containing the full state.
	checkSecretMatches(t, ctx, client, fullState)

	// Change the state.
	for k, v := range fullState.CurrentState.CertificatesBundle {
		v.CertificatePEM = "fake PEM"
		fullState.CurrentState.CertificatesBundle[k] = v
	}

	// Saving again should update the existing secret.
	err = SaveFullStateToK8s(ctx, client, &fullState)
	assert.NoError(t, err)

	// There should be a secret containing the updated full state.
	checkSecretMatches(t, ctx, client, fullState)

	// There should be no configmap.
	checkConfigMapDeleted(t, ctx, client)

	// We should be able to fetch and load the state from the secret.
	fullStateFromK8s, err := GetFullStateFromK8s(ctx, client)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(*fullStateFromK8s, fullState))
}

// Tests the scenario where the cluster already stores existing state in a configmap and there is no secret. In this
// case, a new full state secret should be created and the configmap should be deleted.
func TestSaveAndGetFullStateFromK8s_OldClusterWithCM(t *testing.T) {
	// Create a fake cluster without a secret but with a configmap.
	ctx, fullState, client := setup(t, true)

	// Make sure we can fall back to the configmap when we fetch and load full cluster state given that the secret does
	// not yet exist.
	fullStateFromK8s, err := GetFullStateFromK8s(ctx, client)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(*fullStateFromK8s, fullState))

	// Saving should create a new secret.
	err = SaveFullStateToK8s(ctx, client, &fullState)
	assert.NoError(t, err)

	// There should be a secret containing the full state.
	checkSecretMatches(t, ctx, client, fullState)

	// The configmap should have been deleted.
	checkConfigMapDeleted(t, ctx, client)

	// We should be able to fetch and load the state from the secret.
	fullStateFromK8s, err = GetFullStateFromK8s(ctx, client)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(*fullStateFromK8s, fullState))
}

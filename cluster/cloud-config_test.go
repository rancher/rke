package cluster

import (
	"encoding/json"
	"testing"

	types "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenstackConfigGenerate(t *testing.T) {
	provider := &types.CloudProvider{
		Name: OpenstackProvider,
		CloudConfig: map[string]string{
			"auth-url":    "https://localhost:5000/v3",
			"username":    "admin",
			"password":    "admin",
			"tenant-name": "admin",
			"domain-name": "Default",
		},
	}

	expected := `[Global]
auth-url=https://localhost:5000/v3
domain-name=Default
password=admin
tenant-name=admin
username=admin
`

	cfg, err := GenerateCloudConfig(provider)
	require.NoError(t, err, "failed to generate cloud config")
	assert.EqualValues(t, expected, cfg, "expected openstack cloud config in a specific order")
}

func TestGenericConfigGenerate(t *testing.T) {
	provider := &types.CloudProvider{
		Name: "random",
		CloudConfig: map[string]string{
			"any":   "value",
			"other": "value",
		},
	}

	expected := `{

"any": "value",

"other": "value"
}`

	cfg, err := GenerateCloudConfig(provider)
	require.NoError(t, err, "failed to generate cloud config")

	// json marshals keys in same order
	assert.EqualValues(t, expected, cfg, "expected generic cloud config in a specific order")

	var obj map[string]interface{}
	json.Unmarshal([]byte(cfg), &obj)
	assert.Equal(t, "value", obj["any"])
	assert.Equal(t, "value", obj["other"])
	assert.Len(t, obj, 2)
}

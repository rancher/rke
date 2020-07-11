package docker

import (
	"testing"

	v3 "github.com/rancher/rke/types"
	"github.com/stretchr/testify/assert"
)

const (
	basicRepoUname = "basicUser"
	basicRepoPass  = "basicPass"
	basicImage     = "repo.com/rancher/rke-tools:v1"
	repoUname      = "user"
	repoPass       = "pass"
	image          = "repo.com/foo/bar/rancher/rke-tools:v1"
)

func TestPrivateRegistry(t *testing.T) {
	privateRegistries := map[string]v3.PrivateRegistry{}
	pr1 := v3.PrivateRegistry{
		URL:      "repo.com",
		User:     basicRepoUname,
		Password: basicRepoPass,
	}
	a1, err := getRegistryAuth(pr1)
	assert.Nil(t, err)
	privateRegistries[pr1.URL] = pr1

	pr2 := v3.PrivateRegistry{
		URL:      "repo.com/foo/bar",
		User:     repoUname,
		Password: repoPass,
	}
	a2, err := getRegistryAuth(pr2)
	assert.Nil(t, err)
	privateRegistries[pr2.URL] = pr2

	a, _, err := GetImageRegistryConfig(basicImage, privateRegistries)
	assert.Nil(t, err)
	assert.Equal(t, a, a1)

	a, _, err = GetImageRegistryConfig(image, privateRegistries)
	assert.Nil(t, err)
	assert.Equal(t, a, a2)

}

func TestGetKubeletDockerConfig(t *testing.T) {
	e := "{\"auths\":{\"https://registry.example.com\":{\"auth\":\"dXNlcjE6cGFzc3d+cmQ=\"}}}"
	c, err := GetKubeletDockerConfig(map[string]v3.PrivateRegistry{
		"https://registry.example.com": v3.PrivateRegistry{
			User:     "user1",
			Password: "passw~rd",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, c, e)
}

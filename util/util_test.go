package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveZFromBinds(t *testing.T) {
	binds := []string{"/etc/kubernetes:/etc/kubernetes:z", "/var/log/kube-audit:/var/log/kube-audit:rw,z", "/var/lib/test:/var/lib/test:Z,ro", "/usr/local/lib/test:/usr/local/lib/test:ro,z,noexec", "/etc/normalz:/etc/normalz"}
	expectedBinds := []string{"/etc/kubernetes:/etc/kubernetes", "/var/log/kube-audit:/var/log/kube-audit:rw", "/var/lib/test:/var/lib/test:ro", "/usr/local/lib/test:/usr/local/lib/test:ro,noexec", "/etc/normalz:/etc/normalz"}

	removedBinds := RemoveZFromBinds(binds)
	assert.ElementsMatch(t, expectedBinds, removedBinds)

	emptyBinds, expectedEmptyBinds := []string{}, []string{}
	removedEmptyBinds := RemoveZFromBinds(emptyBinds)
	assert.ElementsMatch(t, expectedEmptyBinds, removedEmptyBinds)
}

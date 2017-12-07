package services

import (
	"fmt"
	"testing"
)

const (
	TestServiceIP                      = "10.233.0.1"
	TestIncorrectClusterServiceIPRange = "#!453.23423.dsf.23"
	TestClusterServiceIPRange          = "10.233.0.0/18"
)

func TestKubernetesServiceIP(t *testing.T) {
	kubernetesServiceIP, err := GetKubernetesServiceIP(TestClusterServiceIPRange)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, kubernetesServiceIP.String(), TestServiceIP,
		fmt.Sprintf("Failed to get correct kubernetes service IP [%s] for range [%s]", kubernetesServiceIP.String(), TestClusterServiceIPRange))
}

func TestIncorrectKubernetesServiceIP(t *testing.T) {
	_, err := GetKubernetesServiceIP(TestIncorrectClusterServiceIPRange)
	if err == nil {
		t.Fatalf("Failed to catch error when parsing incorrect cluster service ip range")
	}
}

func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

package pki

import (
	"crypto/x509"
	"fmt"
	"net"
	"testing"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	FakeClusterDomain       = "cluster.test"
	FakeKubernetesServiceIP = "10.0.0.1"
)

func TestPKI(t *testing.T) {
	cpHosts := []*hosts.Host{
		&hosts.Host{
			RKEConfigNode: v3.RKEConfigNode{
				Address:          "1.1.1.1",
				InternalAddress:  "192.168.1.5",
				Role:             []string{"controlplane"},
				HostnameOverride: "server1",
			},
			DClient: nil,
		},
	}
	certificateMap, err := StartCertificatesGeneration(cpHosts, cpHosts, FakeClusterDomain, "", net.ParseIP(FakeKubernetesServiceIP))
	if err != nil {
		t.Fatalf("Failed To generate certificates: %v", err)
	}
	assertEqual(t, certificateMap[CACertName].Certificate.IsCA, true, "")
	roots := x509.NewCertPool()
	roots.AddCert(certificateMap[CACertName].Certificate)

	certificatesToVerify := []string{
		KubeAPICertName,
		KubeNodeName,
		KubeProxyName,
		KubeControllerName,
		KubeSchedulerName,
		KubeAdminCommonName,
	}
	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	for _, cert := range certificatesToVerify {
		if _, err := certificateMap[cert].Certificate.Verify(opts); err != nil {
			t.Fatalf("Failed to verify certificate %s: %v", cert, err)
		}
	}
	// Test DNS ALT names
	kubeAPIDNSNames := []string{
		"localhost",
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc." + FakeClusterDomain,
	}
	for _, testDNS := range kubeAPIDNSNames {
		assertEqual(
			t,
			isStringInSlice(
				testDNS,
				certificateMap[KubeAPICertName].Certificate.DNSNames),
			true,
			fmt.Sprintf("DNS %s is not found in ALT names of Kube API certificate", testDNS))
	}

	// Test ALT IPs
	kubeAPIAltIPs := []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP(cpHosts[0].InternalAddress),
		net.ParseIP(cpHosts[0].Address),
		net.ParseIP(FakeKubernetesServiceIP),
	}

	for _, testIP := range kubeAPIAltIPs {
		found := false
		for _, altIP := range certificateMap[KubeAPICertName].Certificate.IPAddresses {
			if testIP.Equal(altIP) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("IP Address %v is not found in ALT Ips of kube API", testIP)
		}
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

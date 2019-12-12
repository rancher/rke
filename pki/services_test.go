package pki

import (
	"context"
	"github.com/rancher/rke/hosts"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestDeleteUnusedCerts(t *testing.T) {
	tests := []struct {
		ctx             context.Context
		name            string
		certs           map[string]CertificatePKI
		certName        string
		hosts           []*hosts.Host
		expectLeftCerts map[string]CertificatePKI
	}{
		{
			ctx:  context.Background(),
			name: "Keep valid etcd certs",
			certs: map[string]CertificatePKI{
				"kube-etcd-172-17-0-3":    CertificatePKI{},
				"kube-etcd-172-17-0-4":    CertificatePKI{},
				"kube-node":               CertificatePKI{},
				"kube-kubelet-172-17-0-4": CertificatePKI{},
				"kube-apiserver":          CertificatePKI{},
				"kube-proxy":              CertificatePKI{},
			},
			certName: EtcdCertName,
			hosts: []*hosts.Host{
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.3",
				}},
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.4",
				}},
			},
			expectLeftCerts: map[string]CertificatePKI{
				"kube-etcd-172-17-0-3":    CertificatePKI{},
				"kube-etcd-172-17-0-4":    CertificatePKI{},
				"kube-node":               CertificatePKI{},
				"kube-kubelet-172-17-0-4": CertificatePKI{},
				"kube-apiserver":          CertificatePKI{},
				"kube-proxy":              CertificatePKI{},
			},
		},
		{
			ctx:  context.Background(),
			name: "Keep valid kubelet certs",
			certs: map[string]CertificatePKI{
				"kube-kubelet-172-17-0-5": CertificatePKI{},
				"kube-kubelet-172-17-0-6": CertificatePKI{},
				"kube-node":               CertificatePKI{},
				"kube-apiserver":          CertificatePKI{},
				"kube-proxy":              CertificatePKI{},
				"kube-etcd-172-17-0-6":    CertificatePKI{},
			},
			certName: KubeletCertName,
			hosts: []*hosts.Host{
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.5",
				}},
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.6",
				}},
			},
			expectLeftCerts: map[string]CertificatePKI{
				"kube-kubelet-172-17-0-5": CertificatePKI{},
				"kube-kubelet-172-17-0-6": CertificatePKI{},
				"kube-node":               CertificatePKI{},
				"kube-apiserver":          CertificatePKI{},
				"kube-proxy":              CertificatePKI{},
				"kube-etcd-172-17-0-6":    CertificatePKI{},
			},
		},
		{
			ctx:  context.Background(),
			name: "Remove unused etcd certs",
			certs: map[string]CertificatePKI{
				"kube-etcd-172-17-0-11":    CertificatePKI{},
				"kube-etcd-172-17-0-10":    CertificatePKI{},
				"kube-kubelet-172-17-0-11": CertificatePKI{},
				"kube-node":                CertificatePKI{},
				"kube-apiserver":           CertificatePKI{},
				"kube-proxy":               CertificatePKI{},
			},
			certName: EtcdCertName,
			hosts: []*hosts.Host{
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.11",
				}},
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.12",
				}},
			},
			expectLeftCerts: map[string]CertificatePKI{
				"kube-etcd-172-17-0-11":    CertificatePKI{},
				"kube-kubelet-172-17-0-11": CertificatePKI{},
				"kube-node":                CertificatePKI{},
				"kube-apiserver":           CertificatePKI{},
				"kube-proxy":               CertificatePKI{},
			},
		},
		{
			ctx:  context.Background(),
			name: "Remove unused kubelet certs",
			certs: map[string]CertificatePKI{
				"kube-kubelet-172-17-0-11": CertificatePKI{},
				"kube-kubelet-172-17-0-10": CertificatePKI{},
				"kube-etcd-172-17-0-10":    CertificatePKI{},
				"kube-node":                CertificatePKI{},
				"kube-apiserver":           CertificatePKI{},
				"kube-proxy":               CertificatePKI{},
			},
			certName: KubeletCertName,
			hosts: []*hosts.Host{
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.11",
				}},
				{RKEConfigNode: v3.RKEConfigNode{
					Address: "172.17.0.12",
				}},
			},
			expectLeftCerts: map[string]CertificatePKI{
				"kube-kubelet-172-17-0-11": CertificatePKI{},
				"kube-etcd-172-17-0-10":    CertificatePKI{},
				"kube-node":                CertificatePKI{},
				"kube-apiserver":           CertificatePKI{},
				"kube-proxy":               CertificatePKI{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteUnusedCerts(tt.ctx, tt.certs, tt.certName, tt.hosts)
			assert.Equal(t, true, reflect.DeepEqual(tt.certs, tt.expectLeftCerts))
		})
	}
}

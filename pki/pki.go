package pki

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/hosts"
	"github.com/urfave/cli"
	"k8s.io/client-go/util/cert"
)

type CertificatePKI struct {
	certificate *x509.Certificate
	key         *rsa.PrivateKey
	config      string
}

// StartCertificatesGeneration ...
func StartCertificatesGeneration(ctx *cli.Context, cpHosts []hosts.Host, workerHosts []hosts.Host, clusterDomain string, KubernetesServiceIP net.IP) error {
	forceDeploy := ctx.Bool("force-crts")
	logrus.Infof("[certificates] Generating kubernetes certificates")
	certs, err := generateCerts(cpHosts, clusterDomain, KubernetesServiceIP)
	if err != nil {
		return err
	}
	err = deployCertificatesOnMasters(cpHosts, certs, forceDeploy)
	if err != nil {
		return err
	}
	err = deployCertificatesOnWorkers(workerHosts, certs, forceDeploy)
	if err != nil {
		return err
	}
	return nil
}

func generateCerts(cpHosts []hosts.Host, clusterDomain string, KubernetesServiceIP net.IP) (map[string]CertificatePKI, error) {
	certs := make(map[string]CertificatePKI)
	// generate CA certificate and key
	logrus.Infof("[certificates] Generating CA kubernetes certificates")
	caCrt, caKey, err := generateCACertAndKey()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] CA Certificate: %s", string(cert.EncodeCertPEM(caCrt)))
	certs[CACertName] = CertificatePKI{
		certificate: caCrt,
		key:         caKey,
	}

	// generate API certificate and key
	logrus.Infof("[certificates] Generating Kubernetes API server certificates")
	kubeAPIAltNames := getAltNames(cpHosts, clusterDomain, KubernetesServiceIP)
	kubeAPICrt, kubeAPIKey, err := generateKubeAPICertAndKey(caCrt, caKey, kubeAPIAltNames)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] Kube API Certificate: %s", string(cert.EncodeCertPEM(kubeAPICrt)))
	certs[KubeAPICertName] = CertificatePKI{
		certificate: kubeAPICrt,
		key:         kubeAPIKey,
	}

	// generate Kube controller-manager certificate and key
	logrus.Infof("[certificates] Generating Kube Controller certificates")
	kubeControllerCrt, kubeControllerKey, err := generateClientCertAndKey(caCrt, caKey, KubeControllerCommonName, []string{})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] Kube Controller Certificate: %s", string(cert.EncodeCertPEM(kubeControllerCrt)))
	certs[KubeControllerName] = CertificatePKI{
		certificate: kubeControllerCrt,
		key:         kubeControllerKey,
		config:      getKubeConfigX509("https://"+cpHosts[0].IP+":6443", KubeControllerName, CACertPath, KubeControllerCertPath, KubeControllerKeyPath),
	}

	// generate Kube scheduler certificate and key
	logrus.Infof("[certificates] Generating Kube Scheduler certificates")
	kubeSchedulerCrt, kubeSchedulerKey, err := generateClientCertAndKey(caCrt, caKey, KubeSchedulerCommonName, []string{})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] Kube Scheduler Certificate: %s", string(cert.EncodeCertPEM(kubeSchedulerCrt)))
	certs[KubeSchedulerName] = CertificatePKI{
		certificate: kubeSchedulerCrt,
		key:         kubeSchedulerKey,
		config:      getKubeConfigX509("https://"+cpHosts[0].IP+":6443", KubeSchedulerName, CACertPath, KubeSchedulerCertPath, KubeSchedulerKeyPath),
	}

	// generate Kube Proxy certificate and key
	logrus.Infof("[certificates] Generating Kube Proxy certificates")
	kubeProxyCrt, kubeProxyKey, err := generateClientCertAndKey(caCrt, caKey, KubeProxyCommonName, []string{})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] Kube Proxy Certificate: %s", string(cert.EncodeCertPEM(kubeProxyCrt)))
	certs[KubeProxyName] = CertificatePKI{
		certificate: kubeProxyCrt,
		key:         kubeProxyKey,
		config:      getKubeConfigX509("https://"+cpHosts[0].IP+":6443", KubeProxyName, CACertPath, KubeProxyCertPath, KubeProxyKeyPath),
	}

	logrus.Infof("[certificates] Generating Node certificate")
	nodeCrt, nodeKey, err := generateClientCertAndKey(caCrt, caKey, KubeNodeCommonName, []string{KubeNodeOrganizationName})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[certificates] Node Certificate: %s", string(cert.EncodeCertPEM(kubeProxyCrt)))
	certs[KubeNodeName] = CertificatePKI{
		certificate: nodeCrt,
		key:         nodeKey,
		config:      getKubeConfigX509("https://"+cpHosts[0].IP+":6443", KubeNodeName, CACertPath, KubeNodeCertPath, KubeNodeKeyPath),
	}
	return certs, nil
}

func generateClientCertAndKey(caCrt *x509.Certificate, caKey *rsa.PrivateKey, commonName string, orgs []string) (*x509.Certificate, *rsa.PrivateKey, error) {
	rootKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate private key for %s certificate: %v", commonName, err)
	}
	caConfig := cert.Config{
		CommonName:   commonName,
		Organization: orgs,
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCert, err := cert.NewSignedCert(caConfig, rootKey, caCrt, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate %s certificate: %v", commonName, err)
	}

	return clientCert, rootKey, nil
}

func generateKubeAPICertAndKey(caCrt *x509.Certificate, caKey *rsa.PrivateKey, altNames *cert.AltNames) (*x509.Certificate, *rsa.PrivateKey, error) {
	rootKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate private key for kube-apiserver certificate: %v", err)
	}
	caConfig := cert.Config{
		CommonName: KubeAPICertName,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		AltNames:   *altNames,
	}
	kubeCACert, err := cert.NewSignedCert(caConfig, rootKey, caCrt, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate kube-apiserver certificate: %v", err)
	}

	return kubeCACert, rootKey, nil
}

func generateCACertAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	rootKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate private key for CA certificate: %v", err)
	}
	caConfig := cert.Config{
		CommonName: CACertName,
	}
	kubeCACert, err := cert.NewSelfSignedCACert(caConfig, rootKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate CA certificate: %v", err)
	}

	return kubeCACert, rootKey, nil
}

func getAltNames(cpHosts []hosts.Host, clusterDomain string, KubernetesServiceIP net.IP) *cert.AltNames {
	ips := []net.IP{}
	dnsNames := []string{}
	for _, host := range cpHosts {
		ips = append(ips, net.ParseIP(host.IP))
		dnsNames = append(dnsNames, host.Hostname)
	}
	ips = append(ips, net.ParseIP("127.0.0.1"))
	ips = append(ips, KubernetesServiceIP)
	dnsNames = append(dnsNames, []string{
		"localhost",
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc." + clusterDomain,
	}...)
	return &cert.AltNames{
		IPs:      ips,
		DNSNames: dnsNames,
	}
}

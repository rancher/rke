package pki

import (
	"context"
	"crypto/x509"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki/cert"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
)

func GenerateKubeAPICertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate API certificate and key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	kubernetesServiceIP, err := GetKubernetesServiceIP(rkeConfig.Services.KubeAPI.ServiceClusterIPRange)
	if err != nil {
		return fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	clusterDomain := rkeConfig.Services.Kubelet.ClusterDomain
	cpHosts := hosts.NodesToHosts(rkeConfig.Nodes, controlRole)
	kubeAPIAltNames := GetAltNames(cpHosts, clusterDomain, kubernetesServiceIP, rkeConfig.Authentication.SANs)
	kubeAPICert := certs[KubeAPICertName].Certificate
	if kubeAPICert != nil &&
		reflect.DeepEqual(kubeAPIAltNames.DNSNames, kubeAPICert.DNSNames) &&
		DeepEqualIPsAltNames(kubeAPIAltNames.IPs, kubeAPICert.IPAddresses) && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Kubernetes API server certificates")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeAPICertName].Key
	}
	kubeAPICrt, kubeAPIKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, KubeAPICertName, kubeAPIAltNames, serviceKey, nil)
	if err != nil {
		return err
	}
	kubeAPIChain := []*x509.Certificate{kubeAPICrt}
	kubeAPIChain = append(kubeAPIChain, caChain...)
	certs[KubeAPICertName] = ToCertObject(KubeAPICertName, "", "", kubeAPIChain, kubeAPIKey, nil)
	// handle service account tokens in old clusters
	apiCert := certs[KubeAPICertName]
	if certs[ServiceAccountTokenKeyName].Key == nil {
		logrus.Info("[certificates] Generating Service account token key")
		certs[ServiceAccountTokenKeyName] = ToCertObject(ServiceAccountTokenKeyName, ServiceAccountTokenKeyName, "", apiCert.Chain, apiCert.Key, nil)
	}
	return nil
}

func GenerateKubeAPICSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate API csr and key
	kubernetesServiceIP, err := GetKubernetesServiceIP(rkeConfig.Services.KubeAPI.ServiceClusterIPRange)
	if err != nil {
		return fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	clusterDomain := rkeConfig.Services.Kubelet.ClusterDomain
	cpHosts := hosts.NodesToHosts(rkeConfig.Nodes, controlRole)
	kubeAPIAltNames := GetAltNames(cpHosts, clusterDomain, kubernetesServiceIP, rkeConfig.Authentication.SANs)
	kubeAPIChain := certs[KubeAPICertName].Chain
	oldKubeAPICSR := certs[KubeAPICertName].CSR
	if oldKubeAPICSR != nil &&
		reflect.DeepEqual(kubeAPIAltNames.DNSNames, oldKubeAPICSR.DNSNames) &&
		DeepEqualIPsAltNames(kubeAPIAltNames.IPs, oldKubeAPICSR.IPAddresses) {
		return nil
	}
	logrus.Info("[certificates] Generating Kubernetes API server csr")
	kubeAPICSR, kubeAPIKey, err := GenerateCertSigningRequestAndKey(true, KubeAPICertName, kubeAPIAltNames, certs[KubeAPICertName].Key, nil)
	if err != nil {
		return err
	}
	certs[KubeAPICertName] = ToCertObject(KubeAPICertName, "", "", kubeAPIChain, kubeAPIKey, kubeAPICSR)
	return nil
}

func GenerateKubeControllerCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate Kube controller-manager certificate and key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	if certs[KubeControllerCertName].Certificate != nil && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Controller certificates")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeControllerCertName].Key
	}
	kubeControllerCrt, kubeControllerKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeControllerCertName), nil, serviceKey, nil)
	if err != nil {
		return err
	}
	kubeControllerChain := []*x509.Certificate{kubeControllerCrt}
	kubeControllerChain = append(kubeControllerChain, caChain...)
	certs[KubeControllerCertName] = ToCertObject(KubeControllerCertName, "", "", kubeControllerChain, kubeControllerKey, nil)
	return nil
}

func GenerateKubeControllerCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate Kube controller-manager csr and key
	kubeControllerChain := certs[KubeControllerCertName].Chain
	kubeControllerCSRPEM := certs[KubeControllerCertName].CSRPEM
	if kubeControllerCSRPEM != "" {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Controller csr")
	kubeControllerCSR, kubeControllerKey, err := GenerateCertSigningRequestAndKey(false, getDefaultCN(KubeControllerCertName), nil, certs[KubeControllerCertName].Key, nil)
	if err != nil {
		return err
	}
	certs[KubeControllerCertName] = ToCertObject(KubeControllerCertName, "", "", kubeControllerChain, kubeControllerKey, kubeControllerCSR)
	return nil
}

func GenerateKubeSchedulerCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate Kube scheduler certificate and key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	if certs[KubeSchedulerCertName].Certificate != nil && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Scheduler certificates")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeSchedulerCertName].Key
	}
	kubeSchedulerCrt, kubeSchedulerKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeSchedulerCertName), nil, serviceKey, nil)
	if err != nil {
		return err
	}
	kubeSchedulerChain := []*x509.Certificate{kubeSchedulerCrt}
	kubeSchedulerChain = append(kubeSchedulerChain, caChain...)
	certs[KubeSchedulerCertName] = ToCertObject(KubeSchedulerCertName, "", "", kubeSchedulerChain, kubeSchedulerKey, nil)
	return nil
}

func GenerateKubeSchedulerCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate Kube scheduler csr and key
	kubeSchedulerChain := certs[KubeSchedulerCertName].Chain
	kubeSchedulerCSRPEM := certs[KubeSchedulerCertName].CSRPEM
	if kubeSchedulerCSRPEM != "" {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Scheduler csr")
	kubeSchedulerCSR, kubeSchedulerKey, err := GenerateCertSigningRequestAndKey(false, getDefaultCN(KubeSchedulerCertName), nil, certs[KubeSchedulerCertName].Key, nil)
	if err != nil {
		return err
	}
	certs[KubeSchedulerCertName] = ToCertObject(KubeSchedulerCertName, "", "", kubeSchedulerChain, kubeSchedulerKey, kubeSchedulerCSR)
	return nil
}

func GenerateKubeProxyCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate Kube Proxy certificate and key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	if certs[KubeProxyCertName].Certificate != nil && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Proxy certificates")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeProxyCertName].Key
	}
	kubeProxyCrt, kubeProxyKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeProxyCertName), nil, serviceKey, nil)
	if err != nil {
		return err
	}
	kubeProxyChain := []*x509.Certificate{kubeProxyCrt}
	kubeProxyChain = append(kubeProxyChain, caChain...)
	certs[KubeProxyCertName] = ToCertObject(KubeProxyCertName, "", "", kubeProxyChain, kubeProxyKey, nil)
	return nil
}

func GenerateKubeProxyCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate Kube Proxy csr and key
	kubeProxyChain := certs[KubeProxyCertName].Chain
	kubeProxyCSRPEM := certs[KubeProxyCertName].CSRPEM
	if kubeProxyCSRPEM != "" {
		return nil
	}
	logrus.Info("[certificates] Generating Kube Proxy csr")
	kubeProxyCSR, kubeProxyKey, err := GenerateCertSigningRequestAndKey(false, getDefaultCN(KubeProxyCertName), nil, certs[KubeProxyCertName].Key, nil)
	if err != nil {
		return err
	}
	certs[KubeProxyCertName] = ToCertObject(KubeProxyCertName, "", "", kubeProxyChain, kubeProxyKey, kubeProxyCSR)
	return nil
}

func GenerateKubeNodeCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate kubelet certificate
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	if certs[KubeNodeCertName].Certificate != nil && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Node certificate")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeProxyCertName].Key
	}
	nodeCrt, nodeKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, KubeNodeCommonName, nil, serviceKey, []string{KubeNodeOrganizationName})
	if err != nil {
		return err
	}
	nodeChain := []*x509.Certificate{nodeCrt}
	nodeChain = append(nodeChain, caChain...)
	certs[KubeNodeCertName] = ToCertObject(KubeNodeCertName, KubeNodeCommonName, KubeNodeOrganizationName, nodeChain, nodeKey, nil)
	return nil
}

func GenerateKubeNodeCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate kubelet csr and key
	nodeChain := certs[KubeNodeCertName].Chain
	nodeCSRPEM := certs[KubeNodeCertName].CSRPEM
	if nodeCSRPEM != "" {
		return nil
	}
	logrus.Info("[certificates] Generating Node csr and key")
	nodeCSR, nodeKey, err := GenerateCertSigningRequestAndKey(false, KubeNodeCommonName, nil, certs[KubeNodeCertName].Key, []string{KubeNodeOrganizationName})
	if err != nil {
		return err
	}
	certs[KubeNodeCertName] = ToCertObject(KubeNodeCertName, KubeNodeCommonName, KubeNodeOrganizationName, nodeChain, nodeKey, nodeCSR)
	return nil
}

func GenerateKubeAdminCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate Admin certificate and key
	logrus.Info("[certificates] Generating admin certificates and kubeconfig")
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	cpHosts := hosts.NodesToHosts(rkeConfig.Nodes, controlRole)
	if len(configPath) == 0 {
		configPath = ClusterConfig
	}
	localKubeConfigPath := GetLocalKubeConfig(configPath, configDir)
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[KubeAdminCertName].Key
	}
	kubeAdminCrt, kubeAdminKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, KubeAdminCertName, nil, serviceKey, []string{KubeAdminOrganizationName})
	if err != nil {
		return err
	}
	kubeAdminChain := []*x509.Certificate{kubeAdminCrt}
	kubeAdminChain = append(kubeAdminChain, caChain...)
	kubeAdminCertObj := ToCertObject(KubeAdminCertName, KubeAdminCertName, KubeAdminOrganizationName, kubeAdminChain, kubeAdminKey, nil)
	if len(cpHosts) > 0 {
		kubeAdminConfig := GetKubeConfigX509WithData(
			"https://"+cpHosts[0].Address+":6443",
			rkeConfig.ClusterName,
			KubeAdminCertName,
			string(cert.EncodeCertPEM(caCrt)),
			string(cert.EncodeCertPEM(kubeAdminCrt)),
			string(cert.EncodePrivateKeyPEM(kubeAdminKey)))
		kubeAdminCertObj.Config = kubeAdminConfig
		kubeAdminCertObj.ConfigPath = localKubeConfigPath
	} else {
		kubeAdminCertObj.Config = ""
	}
	certs[KubeAdminCertName] = kubeAdminCertObj
	return nil
}

func GenerateKubeAdminCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	// generate Admin certificate and key
	kubeAdminChain := certs[KubeAdminCertName].Chain
	kubeAdminCSRPEM := certs[KubeAdminCertName].CSRPEM
	if kubeAdminCSRPEM != "" {
		return nil
	}
	kubeAdminCSR, kubeAdminKey, err := GenerateCertSigningRequestAndKey(false, KubeAdminCertName, nil, certs[KubeAdminCertName].Key, []string{KubeAdminOrganizationName})
	if err != nil {
		return err
	}
	logrus.Info("[certificates] Generating admin csr and kubeconfig")
	kubeAdminCertObj := ToCertObject(KubeAdminCertName, KubeAdminCertName, KubeAdminOrganizationName, kubeAdminChain, kubeAdminKey, kubeAdminCSR)
	certs[KubeAdminCertName] = kubeAdminCertObj
	return nil
}

func GenerateAPIProxyClientCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	//generate API server proxy client key and certs
	caChain := certs[RequestHeaderCACertName].Chain
	caCrt := certs[RequestHeaderCACertName].Certificate
	caKey := certs[RequestHeaderCACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("Request Header CA Certificate or Key is empty")
	}
	if certs[APIProxyClientCertName].Certificate != nil && !rotate {
		return nil
	}
	logrus.Info("[certificates] Generating Kubernetes API server proxy client certificates")
	var serviceKey cert.PrivateKey
	if !rotate {
		serviceKey = certs[APIProxyClientCertName].Key
	}
	apiserverProxyClientCrt, apiserverProxyClientKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, APIProxyClientCertName, nil, serviceKey, nil)
	if err != nil {
		return err
	}
	apiserverProxyClientChain := []*x509.Certificate{apiserverProxyClientCrt}
	apiserverProxyClientChain = append(apiserverProxyClientChain, caChain...)
	certs[APIProxyClientCertName] = ToCertObject(APIProxyClientCertName, "", "", apiserverProxyClientChain, apiserverProxyClientKey, nil)
	return nil
}

func GenerateAPIProxyClientCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	//generate API server proxy client key and certs
	apiserverProxyClientChain := certs[APIProxyClientCertName].Chain
	apiserverProxyClientCSRPEM := certs[APIProxyClientCertName].CSRPEM
	if apiserverProxyClientCSRPEM != "" {
		return nil
	}
	logrus.Info("[certificates] Generating Kubernetes API server proxy client csr")
	apiserverProxyClientCSR, apiserverProxyClientKey, err := GenerateCertSigningRequestAndKey(true, APIProxyClientCertName, nil, certs[APIProxyClientCertName].Key, nil)
	if err != nil {
		return err
	}
	certs[APIProxyClientCertName] = ToCertObject(APIProxyClientCertName, "", "", apiserverProxyClientChain, apiserverProxyClientKey, apiserverProxyClientCSR)
	return nil
}

func GenerateExternalEtcdCertificates(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	clientCert, err := cert.ParseCertsPEM([]byte(rkeConfig.Services.Etcd.Cert))
	if err != nil {
		return err
	}
	clientKey, err := cert.ParsePrivateKeyPEM([]byte(rkeConfig.Services.Etcd.Key))
	if err != nil {
		return err
	}
	certs[EtcdClientCertName] = ToCertObject(EtcdClientCertName, "", "", clientCert, clientKey.(cert.PrivateKey), nil)

	caCert, err := cert.ParseCertsPEM([]byte(rkeConfig.Services.Etcd.CACert))
	if err != nil {
		return err
	}
	certs[EtcdClientCACertName] = ToCertObject(EtcdClientCACertName, "", "", caCert, nil, nil)
	return nil
}

func GenerateEtcdCertificates(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	kubernetesServiceIP, err := GetKubernetesServiceIP(rkeConfig.Services.KubeAPI.ServiceClusterIPRange)
	if err != nil {
		return fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	clusterDomain := rkeConfig.Services.Kubelet.ClusterDomain
	etcdHosts := hosts.NodesToHosts(rkeConfig.Nodes, etcdRole)
	etcdAltNames := GetAltNames(etcdHosts, clusterDomain, kubernetesServiceIP, []string{})
	var (
		dnsNames = make([]string, len(etcdAltNames.DNSNames))
		ips      = []string{}
	)
	copy(dnsNames, etcdAltNames.DNSNames)
	sort.Strings(dnsNames)
	for _, ip := range etcdAltNames.IPs {
		ips = append(ips, ip.String())
	}
	sort.Strings(ips)
	for _, host := range etcdHosts {
		etcdName := GetCrtNameForHost(host, EtcdCertName)
		if _, ok := certs[etcdName]; ok && certs[etcdName].CertificatePEM != "" && !rotate {
			cert := certs[etcdName].Certificate
			if cert != nil && len(dnsNames) == len(cert.DNSNames) && len(ips) == len(cert.IPAddresses) {
				var (
					certDNSNames = make([]string, len(cert.DNSNames))
					certIPs      = []string{}
				)
				copy(certDNSNames, cert.DNSNames)
				sort.Strings(certDNSNames)
				for _, ip := range cert.IPAddresses {
					certIPs = append(certIPs, ip.String())
				}
				sort.Strings(certIPs)

				if reflect.DeepEqual(dnsNames, certDNSNames) && reflect.DeepEqual(ips, certIPs) {
					continue
				}
			}
		}
		var serviceKey cert.PrivateKey
		if !rotate {
			serviceKey = certs[etcdName].Key
		}
		logrus.Infof("[certificates] Generating %s certificate and key", etcdName)
		etcdCrt, etcdKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, EtcdCertName, etcdAltNames, serviceKey, nil)
		if err != nil {
			return err
		}
		etcdChain := []*x509.Certificate{etcdCrt}
		etcdChain = append(etcdChain, caChain...)
		certs[etcdName] = ToCertObject(etcdName, "", "", etcdChain, etcdKey, nil)
	}
	deleteUnusedCerts(ctx, certs, EtcdCertName, etcdHosts)
	return nil
}

func GenerateEtcdCSRs(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	kubernetesServiceIP, err := GetKubernetesServiceIP(rkeConfig.Services.KubeAPI.ServiceClusterIPRange)
	if err != nil {
		return fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	clusterDomain := rkeConfig.Services.Kubelet.ClusterDomain
	etcdHosts := hosts.NodesToHosts(rkeConfig.Nodes, etcdRole)
	etcdAltNames := GetAltNames(etcdHosts, clusterDomain, kubernetesServiceIP, []string{})
	for _, host := range etcdHosts {
		etcdName := GetCrtNameForHost(host, EtcdCertName)
		etcdChain := certs[etcdName].Chain
		etcdCsr := certs[etcdName].CSR
		if etcdCsr != nil {
			if reflect.DeepEqual(etcdAltNames.DNSNames, etcdCsr.DNSNames) &&
				DeepEqualIPsAltNames(etcdAltNames.IPs, etcdCsr.IPAddresses) {
				continue
			}
		}
		logrus.Infof("[certificates] Generating etcd-%s csr and key", host.InternalAddress)
		etcdCSR, etcdKey, err := GenerateCertSigningRequestAndKey(true, EtcdCertName, etcdAltNames, certs[etcdName].Key, nil)
		if err != nil {
			return err
		}
		certs[etcdName] = ToCertObject(etcdName, "", "", etcdChain, etcdKey, etcdCSR)
	}
	return nil
}

func GenerateServiceTokenKey(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate service account token key
	privateAPIKey := certs[ServiceAccountTokenKeyName].Key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	if certs[ServiceAccountTokenKeyName].Certificate != nil {
		return nil
	}
	// handle rotation on old clusters
	if certs[ServiceAccountTokenKeyName].Key == nil {
		privateAPIKey = certs[KubeAPICertName].Key
	}
	tokenCrt, tokenKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, ServiceAccountTokenKeyName, nil, privateAPIKey, nil)
	if err != nil {
		return fmt.Errorf("Failed to generate private key for service account token: %v", err)
	}
	tokenChain := []*x509.Certificate{tokenCrt}
	tokenChain = append(tokenChain, caChain...)
	certs[ServiceAccountTokenKeyName] = ToCertObject(ServiceAccountTokenKeyName, ServiceAccountTokenKeyName, "", tokenChain, tokenKey, nil)
	return nil
}

func GenerateRKECACerts(ctx context.Context, certs map[string]CertificatePKI, configPath, configDir string) error {
	if err := GenerateRKEMasterCACert(ctx, certs, configPath, configDir); err != nil {
		return err
	}
	return GenerateRKERequestHeaderCACert(ctx, certs, configPath, configDir)
}

func GenerateRKEMasterCACert(ctx context.Context, certs map[string]CertificatePKI, configPath, configDir string) error {
	// generate kubernetes CA certificate and key
	logrus.Info("[certificates] Generating CA kubernetes certificates")

	caCrt, caKey, err := GenerateCACertAndKey(CACertName, nil)
	if err != nil {
		return err
	}
	caChain := []*x509.Certificate{caCrt}
	certs[CACertName] = ToCertObject(CACertName, "", "", caChain, caKey, nil)
	return nil
}

func GenerateRKERequestHeaderCACert(ctx context.Context, certs map[string]CertificatePKI, configPath, configDir string) error {
	// generate request header client CA certificate and key
	logrus.Info("[certificates] Generating Kubernetes API server aggregation layer requestheader client CA certificates")
	requestHeaderCACrt, requestHeaderCAKey, err := GenerateCACertAndKey(RequestHeaderCACertName, nil)
	if err != nil {
		return err
	}
	requestHeaderCAChain := []*x509.Certificate{requestHeaderCACrt}
	certs[RequestHeaderCACertName] = ToCertObject(RequestHeaderCACertName, "", "", requestHeaderCAChain, requestHeaderCAKey, nil)
	return nil
}

func GenerateKubeletCertificate(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	// generate kubelet certificate and key
	caChain := certs[CACertName].Chain
	caCrt := certs[CACertName].Certificate
	caKey := certs[CACertName].Key
	if caCrt == nil || caKey == nil {
		return fmt.Errorf("CA Certificate or Key is empty")
	}
	log.Debugf(ctx, "[certificates] Generating Kubernetes Kubelet certificates")
	allHosts := hosts.NodesToHosts(rkeConfig.Nodes, "")
	for _, host := range allHosts {
		kubeletName := GetCrtNameForHost(host, KubeletCertName)
		kubeletCert := certs[kubeletName].Certificate
		if kubeletCert != nil && !rotate {
			continue
		}
		kubeletAltNames := GetIPHostAltnamesForHost(host)
		if kubeletCert != nil &&
			reflect.DeepEqual(kubeletAltNames.DNSNames, kubeletCert.DNSNames) &&
			DeepEqualIPsAltNames(kubeletAltNames.IPs, kubeletCert.IPAddresses) && !rotate {
			continue
		}
		var serviceKey cert.PrivateKey
		if !rotate {
			serviceKey = certs[kubeletName].Key
		}
		log.Debugf(ctx, "[certificates] Generating %s certificate and key", kubeletName)
		kubeletCrt, kubeletKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, kubeletName, kubeletAltNames, serviceKey, nil)
		if err != nil {
			return err
		}
		kubeletChain := []*x509.Certificate{kubeletCrt}
		kubeletChain = append(kubeletChain, caChain...)
		certs[kubeletName] = ToCertObject(kubeletName, "", "", kubeletChain, kubeletKey, nil)
	}
	deleteUnusedCerts(ctx, certs, KubeletCertName, allHosts)
	return nil
}

func GenerateKubeletCSR(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	allHosts := hosts.NodesToHosts(rkeConfig.Nodes, "")
	for _, host := range allHosts {
		kubeletName := GetCrtNameForHost(host, KubeletCertName)
		kubeletChain := certs[kubeletName].Chain
		oldKubeletCSR := certs[kubeletName].CSR
		kubeletAltNames := GetIPHostAltnamesForHost(host)
		if oldKubeletCSR != nil &&
			reflect.DeepEqual(kubeletAltNames.DNSNames, oldKubeletCSR.DNSNames) &&
			DeepEqualIPsAltNames(kubeletAltNames.IPs, oldKubeletCSR.IPAddresses) {
			continue
		}
		logrus.Infof("[certificates] Generating %s Kubernetes Kubelet csr", kubeletName)
		kubeletCSR, kubeletKey, err := GenerateCertSigningRequestAndKey(true, kubeletName, kubeletAltNames, certs[kubeletName].Key, nil)
		if err != nil {
			return err
		}
		certs[kubeletName] = ToCertObject(kubeletName, "", "", kubeletChain, kubeletKey, kubeletCSR)
	}
	return nil
}

func GenerateRKEServicesCerts(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string, rotate bool) error {
	RKECerts := []GenFunc{
		GenerateKubeAPICertificate,
		GenerateServiceTokenKey,
		GenerateKubeControllerCertificate,
		GenerateKubeSchedulerCertificate,
		GenerateKubeProxyCertificate,
		GenerateKubeNodeCertificate,
		GenerateKubeAdminCertificate,
		GenerateAPIProxyClientCertificate,
		GenerateEtcdCertificates,
	}
	if IsKubeletGenerateServingCertificateEnabledinConfig(&rkeConfig) {
		RKECerts = append(RKECerts, GenerateKubeletCertificate)
	} else {
		//Clean up kubelet certs when GenerateServingCertificate is disabled
		logrus.Info("[certificates] GenerateServingCertificate is disabled, checking if there are unused kubelet certificates")
		for k := range certs {
			if strings.HasPrefix(k, KubeletCertName) {
				logrus.Infof("[certificates] Deleting unused kubelet certificate: %s", k)
				delete(certs, k)
			}
		}
	}
	for _, gen := range RKECerts {
		if err := gen(ctx, certs, rkeConfig, configPath, configDir, rotate); err != nil {
			return err
		}
	}
	if len(rkeConfig.Services.Etcd.ExternalURLs) > 0 {
		return GenerateExternalEtcdCertificates(ctx, certs, rkeConfig, configPath, configDir, false)
	}
	return nil
}

func GenerateRKEServicesCSRs(ctx context.Context, certs map[string]CertificatePKI, rkeConfig v3.RancherKubernetesEngineConfig) error {
	RKECerts := []CSRFunc{
		GenerateKubeAPICSR,
		GenerateKubeControllerCSR,
		GenerateKubeSchedulerCSR,
		GenerateKubeProxyCSR,
		GenerateKubeNodeCSR,
		GenerateKubeAdminCSR,
		GenerateAPIProxyClientCSR,
		GenerateEtcdCSRs,
	}
	if IsKubeletGenerateServingCertificateEnabledinConfig(&rkeConfig) {
		RKECerts = append(RKECerts, GenerateKubeletCSR)
	}
	for _, csr := range RKECerts {
		if err := csr(ctx, certs, rkeConfig); err != nil {
			return err
		}
	}
	return nil
}

func deleteUnusedCerts(ctx context.Context, certs map[string]CertificatePKI, certName string, hostList []*hosts.Host) {
	hostAddresses := hosts.GetInternalAddressForHosts(hostList)
	logrus.Tracef("Checking and deleting unused certificates with prefix [%s] for the following [%d] node(s): %s", certName, len(hostAddresses), strings.Join(hostAddresses, ","))
	unusedCerts := make(map[string]bool)
	for k := range certs {
		if strings.HasPrefix(k, certName) {
			unusedCerts[k] = true
		}
	}
	for _, host := range hostList {
		Name := GetCrtNameForHost(host, certName)
		delete(unusedCerts, Name)
	}
	for k := range unusedCerts {
		logrus.Infof("[certificates] Deleting unused certificate: %s", k)
		delete(certs, k)
	}
}

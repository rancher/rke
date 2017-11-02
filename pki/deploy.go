package pki

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"k8s.io/client-go/util/cert"
)

func convertCrtToENV(name string, certificate *x509.Certificate) string {
	encodedCrt := cert.EncodeCertPEM(certificate)
	return fmt.Sprintf("%s=%s", name, string(encodedCrt))
}

func convertKeyToENV(name string, key *rsa.PrivateKey) string {
	encodedKey := cert.EncodePrivateKeyPEM(key)
	return fmt.Sprintf("%s=%s", name, string(encodedKey))
}

func convertConfigToENV(name string, config string) string {
	return fmt.Sprintf("%s=%s", name, config)
}

func DeployCertificatesOnMasters(cpHosts []hosts.Host, crtMap map[string]CertificatePKI) error {
	env := []string{
		convertCrtToENV(CACertENVName, crtMap[CACertName].Certificate),
		convertKeyToENV(CAKeyENVName, crtMap[CACertName].Key),
		convertCrtToENV(KubeAPICertENVName, crtMap[KubeAPICertName].Certificate),
		convertKeyToENV(KubeAPIKeyENVName, crtMap[KubeAPICertName].Key),
		convertCrtToENV(KubeControllerCertENVName, crtMap[KubeControllerName].Certificate),
		convertKeyToENV(KubeControllerKeyENVName, crtMap[KubeControllerName].Key),
		convertConfigToENV(KubeControllerConfigENVName, crtMap[KubeControllerName].Config),
		convertCrtToENV(KubeSchedulerCertENVName, crtMap[KubeSchedulerName].Certificate),
		convertKeyToENV(KubeSchedulerKeyENVName, crtMap[KubeSchedulerName].Key),
		convertConfigToENV(KubeSchedulerConfigENVName, crtMap[KubeSchedulerName].Config),
		convertCrtToENV(KubeProxyCertENVName, crtMap[KubeProxyName].Certificate),
		convertKeyToENV(KubeProxyKeyENVName, crtMap[KubeProxyName].Key),
		convertConfigToENV(KubeProxyConfigENVName, crtMap[KubeProxyName].Config),
		convertCrtToENV(KubeNodeCertENVName, crtMap[KubeNodeName].Certificate),
		convertKeyToENV(KubeNodeKeyENVName, crtMap[KubeNodeName].Key),
		convertConfigToENV(KubeNodeConfigENVName, crtMap[KubeNodeName].Config),
	}
	for i := range cpHosts {
		err := doRunDeployer(&cpHosts[i], env)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeployCertificatesOnWorkers(workerHosts []hosts.Host, crtMap map[string]CertificatePKI) error {
	env := []string{
		convertCrtToENV(CACertENVName, crtMap[CACertName].Certificate),
		convertCrtToENV(KubeProxyCertENVName, crtMap[KubeProxyName].Certificate),
		convertKeyToENV(KubeProxyKeyENVName, crtMap[KubeProxyName].Key),
		convertConfigToENV(KubeProxyConfigENVName, crtMap[KubeProxyName].Config),
		convertCrtToENV(KubeNodeCertENVName, crtMap[KubeNodeName].Certificate),
		convertKeyToENV(KubeNodeKeyENVName, crtMap[KubeNodeName].Key),
		convertConfigToENV(KubeNodeConfigENVName, crtMap[KubeNodeName].Config),
	}
	for i := range workerHosts {
		err := doRunDeployer(&workerHosts[i], env)
		if err != nil {
			return err
		}
	}
	return nil
}

func doRunDeployer(host *hosts.Host, containerEnv []string) error {
	logrus.Debugf("[certificates] Pulling Certificate downloader Image on host [%s]", host.Hostname)
	err := docker.PullImage(host.DClient, host.Hostname, CrtDownloaderImage)
	if err != nil {
		return err
	}
	imageCfg := &container.Config{
		Image: CrtDownloaderImage,
		Env:   containerEnv,
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
		},
		Privileged:    true,
		RestartPolicy: container.RestartPolicy{Name: "never"},
	}
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, CrtDownloaderContainer)
	if err != nil {
		return fmt.Errorf("Failed to create Certificates deployer container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Certificates deployer container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[certificates] Successfully started Certificate deployer container: %s", resp.ID)
	for {
		isDeployerRunning, err := docker.IsContainerRunning(host.DClient, host.Hostname, CrtDownloaderContainer)
		if err != nil {
			return err
		}
		if isDeployerRunning {
			time.Sleep(5 * time.Second)
			continue
		}
		if err := host.DClient.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
			return fmt.Errorf("Failed to delete Certificates deployer container on host[%s]: %v", host.Hostname, err)
		}
		return nil
	}
}

func DeployAdminConfig(kubeConfig string) error {
	logrus.Debugf("Deploying admin Kubeconfig locally: %s", kubeConfig)
	err := ioutil.WriteFile(KubeAdminConfigPath, []byte(kubeConfig), 0644)
	if err != nil {
		return fmt.Errorf("Failed to create local admin kubeconfig file: %v", err)
	}
	return nil
}

package pki

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
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
	return name + "=" + string(encodedCrt)
}

func convertKeyToENV(name string, key *rsa.PrivateKey) string {
	encodedKey := cert.EncodePrivateKeyPEM(key)
	return name + "=" + string(encodedKey)
}

func convertConfigToENV(name string, config string) string {
	return name + "=" + config
}

func deployCertificatesOnMasters(cpHosts []hosts.Host, crtMap map[string]CertificatePKI, forceDeploy bool) error {
	forceDeployEnv := "FORCE_DEPLOY=false"
	if forceDeploy {
		forceDeployEnv = "FORCE_DEPLOY=true"
	}
	env := []string{
		forceDeployEnv,
		convertCrtToENV(CACertENVName, crtMap[CACertName].certificate),
		convertKeyToENV(CAKeyENVName, crtMap[CACertName].key),
		convertCrtToENV(KubeAPICertENVName, crtMap[KubeAPICertName].certificate),
		convertKeyToENV(KubeAPIKeyENVName, crtMap[KubeAPICertName].key),
		convertCrtToENV(KubeControllerCertENVName, crtMap[KubeControllerName].certificate),
		convertKeyToENV(KubeControllerKeyENVName, crtMap[KubeControllerName].key),
		convertConfigToENV(KubeControllerConfigENVName, crtMap[KubeControllerName].config),
		convertCrtToENV(KubeSchedulerCertENVName, crtMap[KubeSchedulerName].certificate),
		convertKeyToENV(KubeSchedulerKeyENVName, crtMap[KubeSchedulerName].key),
		convertConfigToENV(KubeSchedulerConfigENVName, crtMap[KubeSchedulerName].config),
		convertCrtToENV(KubeProxyCertENVName, crtMap[KubeProxyName].certificate),
		convertKeyToENV(KubeProxyKeyENVName, crtMap[KubeProxyName].key),
		convertConfigToENV(KubeProxyConfigENVName, crtMap[KubeProxyName].config),
		convertCrtToENV(KubeNodeCertENVName, crtMap[KubeNodeName].certificate),
		convertKeyToENV(KubeNodeKeyENVName, crtMap[KubeNodeName].key),
		convertConfigToENV(KubeNodeConfigENVName, crtMap[KubeNodeName].config),
	}
	for _, host := range cpHosts {
		err := doRunDeployer(&host, env)
		if err != nil {
			return err
		}
	}
	return nil
}

func deployCertificatesOnWorkers(workerHosts []hosts.Host, crtMap map[string]CertificatePKI, forceDeploy bool) error {
	forceDeployEnv := "FORCE_DEPLOY=false"
	if forceDeploy {
		forceDeployEnv = "FORCE_DEPLOY=true"
	}
	env := []string{
		forceDeployEnv,
		convertCrtToENV(CACertENVName, crtMap[CACertName].certificate),
		convertCrtToENV(KubeProxyCertENVName, crtMap[KubeProxyName].certificate),
		convertKeyToENV(KubeProxyKeyENVName, crtMap[KubeProxyName].key),
		convertConfigToENV(KubeProxyConfigENVName, crtMap[KubeProxyName].config),
		convertCrtToENV(KubeNodeCertENVName, crtMap[KubeNodeName].certificate),
		convertKeyToENV(KubeNodeKeyENVName, crtMap[KubeNodeName].key),
		convertConfigToENV(KubeNodeConfigENVName, crtMap[KubeNodeName].config),
	}
	for _, host := range workerHosts {
		err := doRunDeployer(&host, env)
		if err != nil {
			return err
		}
	}
	return nil
}

func doRunDeployer(host *hosts.Host, containerEnv []string) error {
	logrus.Debugf("[certificates] Pulling Certificate downloader Image on host [%s]", host.Hostname)
	err := docker.PullImage(host, CrtDownloaderImage)
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
		isDeployerRunning, err := docker.IsContainerRunning(host, CrtDownloaderContainer)
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

func deployAdminConfig(kubeConfig string, forceDeploy bool) error {
	logrus.Debugf("Deploying admin Kubeconfig locally: %s", kubeConfig)
	if _, err := os.Stat(KubeAdminConfigPath); os.IsNotExist(err) || forceDeploy {
		err := ioutil.WriteFile(KubeAdminConfigPath, []byte(kubeConfig), 0644)
		if err != nil {
			return fmt.Errorf("Failed to create local admin kubeconfig file: %v", err)
		}
	}
	return nil
}

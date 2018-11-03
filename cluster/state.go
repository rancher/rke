package cluster

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/util"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/cert"
)

const (
	stateFileExt = ".rkestate"
)

type RKEFullState struct {
	DesiredState RKEState `json:"desiredState,omitempty"`
	CurrentState RKEState `json:"currentState,omitempty"`
}

type RKEState struct {
	RancherKubernetesEngineConfig *v3.RancherKubernetesEngineConfig `json:"rkeConfig,omitempty"`
	CertificatesBundle            map[string]v3.CertificatePKI      `json:"certificatesBundle,omitempty"`
}

// UpdateClusterState will update the cluster's current state
func (c *Cluster) UpdateClusterState(ctx context.Context, fullState *RKEFullState) error {
	fullState.CurrentState.RancherKubernetesEngineConfig = c.RancherKubernetesEngineConfig.DeepCopy()
	fullState.CurrentState.CertificatesBundle = TransformCertsToV3Certs(c.Certificates)
	return fullState.WriteStateFile(ctx, c.StateFilePath)
}

func (c *Cluster) SaveClusterState(ctx context.Context, rkeConfig *v3.RancherKubernetesEngineConfig) error {
	if len(c.ControlPlaneHosts) > 0 {
		// Reinitialize kubernetes Client
		var err error
		c.KubeClient, err = k8s.NewClient(c.LocalKubeConfigPath, c.K8sWrapTransport)
		if err != nil {
			return fmt.Errorf("Failed to re-initialize Kubernetes Client: %v", err)
		}
		err = saveClusterCerts(ctx, c.KubeClient, c.Certificates)
		if err != nil {
			return fmt.Errorf("[certificates] Failed to Save Kubernetes certificates: %v", err)
		}
		err = saveStateToKubernetes(ctx, c.KubeClient, c.LocalKubeConfigPath, rkeConfig)
		if err != nil {
			return fmt.Errorf("[state] Failed to save configuration state to k8s: %v", err)
		}
	}
	// save state to the cluster nodes as a backup
	uniqueHosts := hosts.GetUniqueHostList(c.EtcdHosts, c.ControlPlaneHosts, c.WorkerHosts)
	if err := saveStateToNodes(ctx, uniqueHosts, rkeConfig, c.SystemImages.Alpine, c.PrivateRegistriesMap); err != nil {
		return fmt.Errorf("[state] Failed to save configuration state to nodes: %v", err)
	}
	return nil
}

func TransformV3CertsToCerts(in map[string]v3.CertificatePKI) map[string]pki.CertificatePKI {
	out := map[string]pki.CertificatePKI{}
	for k, v := range in {
		certs, _ := cert.ParseCertsPEM([]byte(v.Certificate))
		key, _ := cert.ParsePrivateKeyPEM([]byte(v.Key))
		o := pki.CertificatePKI{
			ConfigEnvName: v.ConfigEnvName,
			Name:          v.Name,
			Config:        v.Config,
			CommonName:    v.CommonName,
			OUName:        v.OUName,
			EnvName:       v.EnvName,
			Path:          v.Path,
			KeyEnvName:    v.KeyEnvName,
			KeyPath:       v.KeyPath,
			ConfigPath:    v.ConfigPath,
			Certificate:   certs[0],
			Key:           key.(*rsa.PrivateKey),
		}
		out[k] = o
	}
	return out
}

func TransformCertsToV3Certs(in map[string]pki.CertificatePKI) map[string]v3.CertificatePKI {
	out := map[string]v3.CertificatePKI{}
	for k, v := range in {
		certificate := string(cert.EncodeCertPEM(v.Certificate))
		key := string(cert.EncodePrivateKeyPEM(v.Key))
		o := v3.CertificatePKI{
			Name:          v.Name,
			Config:        v.Config,
			Certificate:   certificate,
			Key:           key,
			EnvName:       v.EnvName,
			KeyEnvName:    v.KeyEnvName,
			ConfigEnvName: v.ConfigEnvName,
			Path:          v.Path,
			KeyPath:       v.KeyPath,
			ConfigPath:    v.ConfigPath,
			CommonName:    v.CommonName,
			OUName:        v.OUName,
		}
		out[k] = o
	}
	return out
}
func (c *Cluster) GetClusterState(ctx context.Context, fullState *RKEFullState, configDir string) (*Cluster, error) {
	var err error
	if fullState.CurrentState.RancherKubernetesEngineConfig == nil {
		return nil, nil
	}

	currentCluster, err := InitClusterObject(ctx, fullState.CurrentState.RancherKubernetesEngineConfig, c.ConfigPath, configDir)
	if err != nil {
		return nil, err
	}

	currentCluster.Certificates = TransformV3CertsToCerts(fullState.CurrentState.CertificatesBundle)

	currentCluster.SetupDialers(ctx, c.DockerDialerFactory, c.LocalConnDialerFactory, c.K8sWrapTransport)

	activeEtcdHosts := currentCluster.EtcdHosts
	for _, inactiveHost := range c.InactiveHosts {
		activeEtcdHosts = removeFromHosts(inactiveHost, activeEtcdHosts)
	}
	// make sure I have all the etcd certs, We need handle dialer failure for etcd nodes https://github.com/rancher/rancher/issues/12898
	for _, host := range activeEtcdHosts {
		certName := pki.GetEtcdCrtName(host.InternalAddress)
		if (currentCluster.Certificates[certName] == pki.CertificatePKI{}) {
			if currentCluster.Certificates, err = pki.RegenerateEtcdCertificate(ctx,
				currentCluster.Certificates,
				host,
				activeEtcdHosts,
				currentCluster.ClusterDomain,
				currentCluster.KubernetesServiceIP); err != nil {
				return nil, err
			}
		}
	}
	currentCluster.Certificates, err = regenerateAPICertificate(c, currentCluster.Certificates)
	if err != nil {
		return nil, fmt.Errorf("Failed to regenerate KubeAPI certificate %v", err)
	}

	currentCluster.setClusterDefaults(ctx)

	return currentCluster, nil
}

func saveStateToKubernetes(ctx context.Context, kubeClient *kubernetes.Clientset, kubeConfigPath string, rkeConfig *v3.RancherKubernetesEngineConfig) error {
	log.Infof(ctx, "[state] Saving cluster state to Kubernetes")
	clusterFile, err := yaml.Marshal(*rkeConfig)
	if err != nil {
		return err
	}
	timeout := make(chan bool, 1)
	go func() {
		for {
			_, err := k8s.UpdateConfigMap(kubeClient, clusterFile, StateConfigMapName)
			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}
			log.Infof(ctx, "[state] Successfully Saved cluster state to Kubernetes ConfigMap: %s", StateConfigMapName)
			timeout <- true
			break
		}
	}()
	select {
	case <-timeout:
		return nil
	case <-time.After(time.Second * UpdateStateTimeout):
		return fmt.Errorf("[state] Timeout waiting for kubernetes to be ready")
	}
}

func saveStateToNodes(ctx context.Context, uniqueHosts []*hosts.Host, clusterState *v3.RancherKubernetesEngineConfig, alpineImage string, prsMap map[string]v3.PrivateRegistry) error {
	log.Infof(ctx, "[state] Saving cluster state to cluster nodes")
	clusterFile, err := yaml.Marshal(*clusterState)
	if err != nil {
		return err
	}
	var errgrp errgroup.Group

	hostsQueue := util.GetObjectQueue(uniqueHosts)
	for w := 0; w < WorkerThreads; w++ {
		errgrp.Go(func() error {
			var errList []error
			for host := range hostsQueue {
				err := pki.DeployStateOnPlaneHost(ctx, host.(*hosts.Host), alpineImage, prsMap, string(clusterFile))
				if err != nil {
					errList = append(errList, err)
				}
			}
			return util.ErrList(errList)
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}
	return nil
}

func getStateFromKubernetes(ctx context.Context, kubeClient *kubernetes.Clientset, kubeConfigPath string) (*Cluster, error) {
	log.Infof(ctx, "[state] Fetching cluster state from Kubernetes")
	var cfgMap *v1.ConfigMap
	var currentCluster Cluster
	var err error
	timeout := make(chan bool, 1)
	go func() {
		for {
			cfgMap, err = k8s.GetConfigMap(kubeClient, StateConfigMapName)
			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}
			log.Infof(ctx, "[state] Successfully Fetched cluster state to Kubernetes ConfigMap: %s", StateConfigMapName)
			timeout <- true
			break
		}
	}()
	select {
	case <-timeout:
		clusterData := cfgMap.Data[StateConfigMapName]
		err := yaml.Unmarshal([]byte(clusterData), &currentCluster)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal cluster data")
		}
		return &currentCluster, nil
	case <-time.After(time.Second * GetStateTimeout):
		log.Infof(ctx, "Timed out waiting for kubernetes cluster to get state")
		return nil, fmt.Errorf("Timeout waiting for kubernetes cluster to get state")
	}
}

func getStateFromNodes(ctx context.Context, uniqueHosts []*hosts.Host, alpineImage string, prsMap map[string]v3.PrivateRegistry) *Cluster {
	log.Infof(ctx, "[state] Fetching cluster state from Nodes")
	var currentCluster Cluster
	var clusterFile string
	var err error

	for _, host := range uniqueHosts {
		filePath := path.Join(host.PrefixPath, pki.TempCertPath, pki.ClusterStateFile)
		clusterFile, err = pki.FetchFileFromHost(ctx, filePath, alpineImage, host, prsMap, pki.StateDeployerContainerName, "state")
		if err == nil {
			break
		}
	}
	if len(clusterFile) == 0 {
		return nil
	}
	err = yaml.Unmarshal([]byte(clusterFile), &currentCluster)
	if err != nil {
		logrus.Debugf("[state] Failed to unmarshal the cluster file fetched from nodes: %v", err)
		return nil
	}
	log.Infof(ctx, "[state] Successfully fetched cluster state from Nodes")
	return &currentCluster

}

func GetK8sVersion(localConfigPath string, k8sWrapTransport k8s.WrapTransport) (string, error) {
	logrus.Debugf("[version] Using %s to connect to Kubernetes cluster..", localConfigPath)
	k8sClient, err := k8s.NewClient(localConfigPath, k8sWrapTransport)
	if err != nil {
		return "", fmt.Errorf("Failed to create Kubernetes Client: %v", err)
	}
	discoveryClient := k8sClient.DiscoveryClient
	logrus.Debugf("[version] Getting Kubernetes server version..")
	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("Failed to get Kubernetes server version: %v", err)
	}
	return fmt.Sprintf("%#v", *serverVersion), nil
}

func RebuildState(ctx context.Context, rkeConfig *v3.RancherKubernetesEngineConfig, oldState *RKEFullState, configPath, configDir string) (*RKEFullState, error) {
	newState := &RKEFullState{
		DesiredState: RKEState{
			RancherKubernetesEngineConfig: rkeConfig.DeepCopy(),
		},
	}

	// Rebuilding the certificates of the desired state
	if oldState.DesiredState.CertificatesBundle == nil {
		// Get the certificate Bundle
		certBundle, err := pki.GenerateRKECerts(ctx, *rkeConfig, "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to generate certificate bundle: %v", err)
		}
		// Convert rke certs to v3.certs
		newState.DesiredState.CertificatesBundle = TransformCertsToV3Certs(certBundle)
	} else {
		// Regenerating etcd certificates for any new etcd nodes
		pkiCertBundle := TransformV3CertsToCerts(oldState.DesiredState.CertificatesBundle)
		if err := pki.GenerateEtcdCertificates(ctx, pkiCertBundle, *rkeConfig, "", ""); err != nil {
			return nil, err
		}
		// Regenerating kubeapi certificates for any new kubeapi nodes
		if err := pki.GenerateKubeAPICertificate(ctx, pkiCertBundle, *rkeConfig, "", ""); err != nil {
			return nil, err
		}
		// Regenerating kubeadmin certificates/config
		if err := pki.GenerateKubeAdminCertificate(ctx, pkiCertBundle, *rkeConfig, configPath, configDir); err != nil {
			return nil, err
		}
		newState.DesiredState.CertificatesBundle = TransformCertsToV3Certs(pkiCertBundle)
	}
	newState.CurrentState = oldState.CurrentState
	return newState, nil
}

func (s *RKEFullState) WriteStateFile(ctx context.Context, statePath string) error {
	stateFile, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to Marshal state object: %v", err)
	}
	logrus.Debugf("Writing state file: %s", stateFile)
	if err := ioutil.WriteFile(statePath, stateFile, 0640); err != nil {
		return fmt.Errorf("Failed to write state file: %v", err)
	}
	log.Infof(ctx, "Successfully Deployed state file at [%s]", statePath)
	return nil
}

func GetStateFilePath(configPath, configDir string) string {
	if configPath == "" {
		configPath = pki.ClusterConfig
	}
	baseDir := filepath.Dir(configPath)
	if len(configDir) > 0 {
		baseDir = filepath.Dir(configDir)
	}
	fileName := filepath.Base(configPath)
	baseDir += "/"
	fullPath := fmt.Sprintf("%s%s", baseDir, fileName)
	trimmedName := strings.TrimSuffix(fullPath, filepath.Ext(fullPath))
	return trimmedName + stateFileExt
}

func ReadStateFile(ctx context.Context, statePath string) (*RKEFullState, error) {
	rkeFullState := &RKEFullState{}
	fp, err := filepath.Abs(statePath)
	if err != nil {
		return rkeFullState, fmt.Errorf("failed to lookup current directory name: %v", err)
	}
	file, err := os.Open(fp)
	if err != nil {
		return rkeFullState, fmt.Errorf("Can not find RKE state file: %v", err)
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return rkeFullState, fmt.Errorf("failed to read state file: %v", err)
	}
	if err := json.Unmarshal(buf, rkeFullState); err != nil {
		return rkeFullState, fmt.Errorf("failed to unmarshal the state file: %v", err)
	}
	return rkeFullState, nil
}

func RemoveStateFile(ctx context.Context, statePath string) {
	log.Infof(ctx, "Removing state file: %s", statePath)
	if err := os.Remove(statePath); err != nil {
		logrus.Warningf("Failed to remove state file: %v", err)
		return
	}
	log.Infof(ctx, "State file removed successfully")
}

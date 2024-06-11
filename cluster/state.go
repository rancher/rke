package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/transport"
)

const (
	stateFileExt = ".rkestate"
	certDirExt   = "_certs"
)

var (
	ErrFullStateIsNil = errors.New("fullState argument cannot be nil")
)

type FullState struct {
	DesiredState State `json:"desiredState,omitempty"`
	CurrentState State `json:"currentState,omitempty"`
}

type State struct {
	RancherKubernetesEngineConfig *v3.RancherKubernetesEngineConfig `json:"rkeConfig,omitempty"`
	CertificatesBundle            map[string]pki.CertificatePKI     `json:"certificatesBundle,omitempty"`
	EncryptionConfig              string                            `json:"encryptionConfig,omitempty"`
}

func (c *Cluster) UpdateClusterCurrentState(ctx context.Context, fullState *FullState) error {
	fullState.CurrentState.RancherKubernetesEngineConfig = c.RancherKubernetesEngineConfig.DeepCopy()
	fullState.CurrentState.CertificatesBundle = c.Certificates
	fullState.CurrentState.EncryptionConfig = c.EncryptionConfig.EncryptionProviderFile
	return fullState.WriteStateFile(ctx, c.StateFilePath)
}

func (c *Cluster) GetClusterState(ctx context.Context, fullState *FullState) (*Cluster, error) {
	var err error
	if fullState.CurrentState.RancherKubernetesEngineConfig == nil {
		return nil, nil
	}

	// resetup external flags
	flags := GetExternalFlags(false, false, false, false, c.ConfigDir, c.ConfigPath)
	currentCluster, err := InitClusterObject(ctx, fullState.CurrentState.RancherKubernetesEngineConfig, flags, fullState.CurrentState.EncryptionConfig)
	if err != nil {
		return nil, err
	}
	currentCluster.Certificates = fullState.CurrentState.CertificatesBundle
	currentCluster.EncryptionConfig.EncryptionProviderFile = fullState.CurrentState.EncryptionConfig
	// resetup dialers
	dialerOptions := hosts.GetDialerOptions(c.DockerDialerFactory, c.LocalConnDialerFactory, c.K8sWrapTransport)
	if err := currentCluster.SetupDialers(ctx, dialerOptions); err != nil {
		return nil, err
	}
	return currentCluster, nil
}

func (c *Cluster) GetStateFileFromConfigMap(ctx context.Context) (string, error) {
	kubeletImage := c.Services.Kubelet.Image
	for _, host := range c.ControlPlaneHosts {
		stateFile, err := services.RunGetStateFileFromConfigMap(ctx, host, c.PrivateRegistriesMap, kubeletImage, c.Version)
		if err != nil || stateFile == "" {
			logrus.Infof("Could not get ConfigMap with cluster state from host [%s]", host.Address)
			continue
		}
		return stateFile, nil
	}
	return "", fmt.Errorf("[state] Unable to get ConfigMap with cluster state from any Control Plane host")
}

// SaveFullStateToK8s saves the full cluster state to a k8s secret. If any errors that occur on attempts to update
// the secret will be retired up until some limit.
func SaveFullStateToK8s(ctx context.Context, k8sClient kubernetes.Interface, fullState *FullState) error {
	log.Infof(ctx, "[state] Saving full cluster state to Kubernetes")

	if fullState == nil {
		return ErrFullStateIsNil
	}

	secrets := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem)
	configMaps := k8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem)
	stateBytes, err := json.Marshal(fullState)
	if err != nil {
		return fmt.Errorf("[state] error marshalling full state to JSON: %w", err)
	}

	// Back off for 1s between attempts.
	backoff := wait.Backoff{
		Duration: time.Second,
		Steps:    int(UpdateStateTimeout.Seconds()),
	}

	// Try to create or update the secret and delete the old configmap in k8s, if it still exists.
	saveState := func(ctx context.Context) (bool, error) {
		// Check if the secret already exists.
		existingSecret, err := secrets.Get(ctx, FullStateSecretName, metav1.GetOptions{})
		if err == nil {
			// The secret already exists, update it.
			existingSecretCopy := existingSecret.DeepCopy()
			existingSecretCopy.Data[FullStateSecretName] = stateBytes
			if _, err := secrets.Update(ctx, existingSecretCopy, metav1.UpdateOptions{}); err != nil {
				return false, fmt.Errorf("[state] error updating secret: %w", err)
			}
		} else if apierrors.IsNotFound(err) {
			// The secret does not exist, create it.
			_, err := secrets.Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FullStateSecretName,
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string][]byte{
					FullStateSecretName: stateBytes,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return false, fmt.Errorf("[state] error creating secret: %w", err)
			}
		} else {
			return false, fmt.Errorf("[state] error getting secret: %w", err)
		}

		// Delete the old configmap.
		err = configMaps.Delete(ctx, FullStateConfigMapName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("[state] error deleting configmap: %w", err)
		}

		return true, nil
	}

	// Retry until success or backoff.Steps has been reached ctx is cancelled.
	if err = wait.ExponentialBackoffWithContext(ctx, backoff, saveState); err != nil {
		return fmt.Errorf("[state] error updating secret: %w", err)
	}

	return nil
}

// GetFullStateFromK8s fetches the full cluster state from the k8s cluster.
// In earlier versions of RKE, the full cluster state was stored in a configmap, but it has since been moved
// to a secret. This function tries fetching it from the secret first and will fall back on the configmap if the secret
// doesn't exist.
func GetFullStateFromK8s(ctx context.Context, k8sClient kubernetes.Interface) (*FullState, error) {
	// Back off for 1s between attempts.
	backoff := wait.Backoff{
		Duration: time.Second,
		Steps:    int(GetStateTimeout.Seconds()),
	}

	// Try to fetch secret or configmap in k8s.
	var fullState FullState
	getState := func(ctx context.Context) (bool, error) {
		fullStateBytes, err := getFullStateBytesFromSecret(ctx, k8sClient, FullStateSecretName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logrus.Debug("full-state secret not found, falling back to configmap")

				fullStateBytes, err = getFullStateBytesFromConfigMap(ctx, k8sClient, FullStateConfigMapName)
				if err != nil {
					return false, fmt.Errorf("[state] error getting full state from configmap: %w", err)
				}
			} else {
				return false, fmt.Errorf("[state] error getting full state from secret: %w", err)
			}
		}

		if err := json.Unmarshal(fullStateBytes, &fullState); err != nil {
			return false, fmt.Errorf("[state] error unmarshalling full state from JSON: %w", err)
		}

		return true, nil
	}

	// Retry until success or backoff.Steps has been reached or ctx is cancelled.
	err := wait.ExponentialBackoffWithContext(ctx, backoff, getState)
	return &fullState, err
}

// getFullStateBytesFromConfigMap fetches the full state from the configmap with the given name in the kube-system namespace.
func getFullStateBytesFromConfigMap(ctx context.Context, k8sClient kubernetes.Interface, name string) ([]byte, error) {
	confMap, err := k8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("[state] error getting configmap %s: %w", name, err)
	}

	data, ok := confMap.Data[name]
	if !ok {
		return nil, fmt.Errorf("[state] expected configmap %s to have field %s, but none was found", name, name)
	}

	return []byte(data), nil
}

// getFullStateBytesFromSecret fetches the full state from the secret with the given name in the kube-system namespace.
func getFullStateBytesFromSecret(ctx context.Context, k8sClient kubernetes.Interface, name string) ([]byte, error) {
	secret, err := k8sClient.CoreV1().Secrets(metav1.NamespaceSystem).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("[state] error getting secret %s: %w", name, err)
	}

	data, ok := secret.Data[name]
	if !ok {
		return nil, fmt.Errorf("[state] expected secret %s to have field %s, but none was found", name, name)
	}

	return data, nil
}

func GetStateFromKubernetes(ctx context.Context, kubeCluster *Cluster) (*Cluster, error) {
	log.Infof(ctx, "[state] Fetching cluster state from Kubernetes")
	k8sClient, err := k8s.NewClient(kubeCluster.LocalKubeConfigPath, kubeCluster.K8sWrapTransport)
	if err != nil {
		return nil, fmt.Errorf("[state] Failed to create Kubernetes Client: %v", err)
	}
	var cfgMap *v1.ConfigMap
	var currentCluster Cluster
	timeout := make(chan bool, 1)
	go func() {
		for {
			cfgMap, err = k8s.GetConfigMap(k8sClient, StateConfigMapName)
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
			return nil, fmt.Errorf("[state] Failed to unmarshal cluster data")
		}
		return &currentCluster, nil
	case <-time.After(GetStateTimeout):
		log.Infof(ctx, "Timed out waiting for kubernetes cluster to get state")
		return nil, fmt.Errorf("[state] Timeout waiting for kubernetes cluster to get state")
	}
}

func GetK8sVersion(localConfigPath string, k8sWrapTransport transport.WrapperFunc) (string, error) {
	logrus.Debugf("[version] Using %s to connect to Kubernetes cluster..", localConfigPath)
	k8sClient, err := k8s.NewClient(localConfigPath, k8sWrapTransport)
	if err != nil {
		return "", fmt.Errorf("[state] Failed to create Kubernetes Client: %v", err)
	}
	discoveryClient := k8sClient.DiscoveryClient
	logrus.Debugf("[version] Getting Kubernetes server version..")
	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("[state] Failed to get Kubernetes server version: %v", err)
	}
	return fmt.Sprintf("%#v", *serverVersion), nil
}

func RebuildState(ctx context.Context, kubeCluster *Cluster, oldState *FullState, flags ExternalFlags) (*FullState, error) {
	rkeConfig := &kubeCluster.RancherKubernetesEngineConfig
	newState := &FullState{
		DesiredState: State{
			RancherKubernetesEngineConfig: rkeConfig.DeepCopy(),
		},
	}

	if flags.CustomCerts {
		certBundle, err := pki.ReadCertsAndKeysFromDir(flags.CertificateDir)
		if err != nil {
			return nil, fmt.Errorf("[state] Failed to read certificates from dir [%s]: %v", flags.CertificateDir, err)
		}
		// make sure all custom certs are included
		if err := pki.ValidateBundleContent(rkeConfig, certBundle, flags.ClusterFilePath, flags.ConfigDir); err != nil {
			return nil, fmt.Errorf("[state] Failed to validates certificates from dir [%s]: %v", flags.CertificateDir, err)
		}
		newState.DesiredState.CertificatesBundle = certBundle
		newState.CurrentState = oldState.CurrentState

		err = updateEncryptionConfig(kubeCluster, oldState, newState)
		if err != nil {
			return nil, err
		}
		return newState, nil
	}

	// Rebuilding the certificates of the desired state
	if oldState.DesiredState.CertificatesBundle == nil { // this is a fresh cluster
		if err := buildFreshState(ctx, kubeCluster, newState); err != nil {
			return nil, err
		}
	} else { // This is an existing cluster with an old DesiredState
		if err := rebuildExistingState(ctx, kubeCluster, oldState, newState, flags); err != nil {
			return nil, err
		}
	}
	newState.CurrentState = oldState.CurrentState
	return newState, nil
}

func (s *FullState) WriteStateFile(ctx context.Context, statePath string) error {
	stateFile, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("[state] Failed to Marshal state object: %v", err)
	}
	logrus.Tracef("Writing state file: %s", stateFile)
	if err := os.WriteFile(statePath, stateFile, 0600); err != nil {
		return fmt.Errorf("[state] Failed to write state file: %v", err)
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

func GetCertificateDirPath(configPath, configDir string) string {
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
	return trimmedName + certDirExt
}

func StringToFullState(ctx context.Context, stateFileContent string) (*FullState, error) {
	rkeFullState := &FullState{}
	logrus.Tracef("stateFileContent: %s", stateFileContent)
	if err := json.Unmarshal([]byte(stateFileContent), rkeFullState); err != nil {
		return rkeFullState, err
	}
	rkeFullState.DesiredState.CertificatesBundle = pki.TransformPEMToObject(rkeFullState.DesiredState.CertificatesBundle)
	rkeFullState.CurrentState.CertificatesBundle = pki.TransformPEMToObject(rkeFullState.CurrentState.CertificatesBundle)
	logrus.Tracef("rkeFullState: %+v", rkeFullState)

	return rkeFullState, nil
}

func ReadStateFile(ctx context.Context, statePath string) (*FullState, error) {
	rkeFullState := &FullState{}
	fp, err := filepath.Abs(statePath)
	if err != nil {
		return rkeFullState, fmt.Errorf("[state] failed to lookup current directory name: %v", err)
	}
	file, err := os.Open(fp)
	if err != nil {
		return rkeFullState, fmt.Errorf("[state] Can not find RKE state file: %v", err)
	}
	defer file.Close()
	buf, err := io.ReadAll(file)
	if err != nil {
		return rkeFullState, fmt.Errorf("[state] failed to read state file: %v", err)
	}
	if err := json.Unmarshal(buf, rkeFullState); err != nil {
		return rkeFullState, fmt.Errorf("[state] failed to unmarshal the state file: %v", err)
	}
	rkeFullState.DesiredState.CertificatesBundle = pki.TransformPEMToObject(rkeFullState.DesiredState.CertificatesBundle)
	rkeFullState.CurrentState.CertificatesBundle = pki.TransformPEMToObject(rkeFullState.CurrentState.CertificatesBundle)
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

func GetStateFromNodes(ctx context.Context, kubeCluster *Cluster) *Cluster {
	var currentCluster Cluster
	var clusterFile string
	var err error

	uniqueHosts := hosts.GetUniqueHostList(kubeCluster.EtcdHosts, kubeCluster.ControlPlaneHosts, kubeCluster.WorkerHosts)
	for _, host := range uniqueHosts {
		filePath := path.Join(pki.TempCertPath, pki.ClusterStateFile)
		clusterFile, err = pki.FetchFileFromHost(ctx, filePath, kubeCluster.SystemImages.Alpine, host, kubeCluster.PrivateRegistriesMap, pki.StateDeployerContainerName, "state", kubeCluster.Version)
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

func buildFreshState(ctx context.Context, kubeCluster *Cluster, newState *FullState) error {
	rkeConfig := &kubeCluster.RancherKubernetesEngineConfig
	// Get the certificate Bundle
	certBundle, err := pki.GenerateRKECerts(ctx, *rkeConfig, "", "")
	if err != nil {
		return fmt.Errorf("[state] Failed to generate certificate bundle: %v", err)
	}
	newState.DesiredState.CertificatesBundle = certBundle
	if isEncryptionEnabled(rkeConfig) {
		if newState.DesiredState.EncryptionConfig, err = kubeCluster.getEncryptionProviderFile(); err != nil {
			return err
		}
	}
	return nil
}

func rebuildExistingState(ctx context.Context, kubeCluster *Cluster, oldState, newState *FullState, flags ExternalFlags) error {
	rkeConfig := &kubeCluster.RancherKubernetesEngineConfig
	pkiCertBundle := oldState.DesiredState.CertificatesBundle
	// check for legacy clusters prior to requestheaderca
	if pkiCertBundle[pki.RequestHeaderCACertName].Certificate == nil {
		if err := pki.GenerateRKERequestHeaderCACert(ctx, pkiCertBundle, flags.ClusterFilePath, flags.ConfigDir); err != nil {
			return err
		}
	}
	if err := pki.GenerateRKEServicesCerts(ctx, pkiCertBundle, *rkeConfig, flags.ClusterFilePath, flags.ConfigDir, false); err != nil {
		return err
	}
	newState.DesiredState.CertificatesBundle = pkiCertBundle
	err := updateEncryptionConfig(kubeCluster, oldState, newState)
	return err
}

func updateEncryptionConfig(kubeCluster *Cluster, oldState *FullState, newState *FullState) error {
	if isEncryptionEnabled(&kubeCluster.RancherKubernetesEngineConfig) {
		if oldState.DesiredState.EncryptionConfig != "" {
			newState.DesiredState.EncryptionConfig = oldState.DesiredState.EncryptionConfig
		} else {
			var err error
			if newState.DesiredState.EncryptionConfig, err = kubeCluster.getEncryptionProviderFile(); err != nil {
				return err
			}
		}
	}
	return nil
}

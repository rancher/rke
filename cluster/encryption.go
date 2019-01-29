package cluster

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/rancher/rke/k8s"
	"github.com/rancher/rke/services"
	"github.com/rancher/rke/templates"
	"github.com/sirupsen/logrus"
)

// DeploySecretsEncryptionConfig -  Deploy encryption.yaml to ControlPlane hosts.
func (c *Cluster) DeploySecretsEncryptionConfig(ctx context.Context, currentCluster *Cluster) error {
	if currentCluster == nil {
		currentCluster = &Cluster{}
	}

	if c.Services.KubeAPI.SecretsEncryptionConfig == currentCluster.Services.KubeAPI.SecretsEncryptionConfig {
		logrus.Debug("[controlplane] Secrets encryption config matches state. No changes needed.")
		return nil
	}

	for _, host := range c.ControlPlaneHosts {
		err := c.WriteFile(ctx, host, "controlplane", "/etc/kubernetes/encryption.yaml", "0600", c.Services.KubeAPI.SecretsEncryptionConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReconcileSecretsEncryptionConfig - Restart controlplane and replace all secrets in the cluster when the encryption config changes.
func (c *Cluster) ReconcileSecretsEncryptionConfig(ctx context.Context, currentCluster *Cluster) error {
	if currentCluster != nil {
		if c.Services.KubeAPI.SecretsEncryptionConfig == currentCluster.Services.KubeAPI.SecretsEncryptionConfig {
			logrus.Debug("[reconcile] Secrets encryption config matches. No changes needed.")
			return nil
		}

		logrus.Info("[reconcile] New secrets encryption config. Updating existing secrets to apply encryption.")

		err := services.RestartControlPlane(ctx, c.ControlPlaneHosts)
		if err != nil {
			return err
		}

		kubeClient, err := k8s.NewClient(c.LocalKubeConfigPath, c.K8sWrapTransport)
		if err != nil {
			return fmt.Errorf("Failed to initialize new kubernetes client: %v", err)
		}

		secrets, err := k8s.GetAllSecrets(kubeClient)
		if err != nil {
			return err
		}

		for _, secret := range secrets {
			err = k8s.UpdateSecretObject(kubeClient, &secret)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Need some logic to keep users from shooting themselves in the foot.
// If you rotate keys without the old key in the list you will lock yourself out of the existing secrets.
// Ways people will break this:
// - Add a config after a cluster is up with a generated config
// - Change a config without putting the old key in the list before re-create of secrets.

// SetSecretsEncryptionConfig - Set a default encryption config and secret if not provided.
func (c *Cluster) SetSecretsEncryptionConfig(currentCluster *Cluster) error {
	if currentCluster == nil {
		currentCluster = &Cluster{}
	}
	// cluster.yaml not defined or populated
	if len(c.Services.KubeAPI.SecretsEncryptionConfig) == 0 {
		// Generate from template if not populated in current config.
		if len(currentCluster.Services.KubeAPI.SecretsEncryptionConfig) == 0 {
			clusterMajorVersion := getTagMajorVersion(c.Version)
			encSecretKey := make([]byte, 32)
			_, err := rand.Read(encSecretKey)
			if err != nil {
				return err
			}
			encConfig := &bytes.Buffer{}
			secret := struct {
				Secret string
			}{
				Secret: base64.StdEncoding.EncodeToString(encSecretKey),
			}

			_, ok := templates.EncryptionConfigurationTemplate[clusterMajorVersion]
			if !ok {
				return fmt.Errorf("Unsupported Kubernetes version, no encryption config template available: %s", c.Version)
			}
			encTemplate := template.New("encryption.yaml")
			encTemplate, err = encTemplate.Parse(templates.EncryptionConfigurationTemplate[clusterMajorVersion])
			if err != nil {
				return err
			}
			err = encTemplate.Execute(encConfig, secret)
			if err != nil {
				return err
			}

			c.Services.KubeAPI.SecretsEncryptionConfig = encConfig.String()
		} else {
			// set previous cluster config (c is populated with cluster.yml, so currently empty)
			c.Services.KubeAPI.SecretsEncryptionConfig = currentCluster.Services.KubeAPI.SecretsEncryptionConfig
		}
		// cluster.yml is defined.
	} else {
		// Should be a new cluster - just let it pass
		if len(currentCluster.Services.KubeAPI.SecretsEncryptionConfig) == 0 {
			return nil
		}
		// cluster.yml equals current config
		if c.Services.KubeAPI.SecretsEncryptionConfig == currentCluster.Services.KubeAPI.SecretsEncryptionConfig {
			return nil
		}
		// cluster.yml is different than current config
		// Are you sure you know what you are doing?
		if !c.Force {
			reader := bufio.NewReader(os.Stdin)
			logrus.Warn("A new secrets encryption configuration has been detected!")
			logrus.Warn("Improper configuration may result in kube-apiserver failing to start and permanently losing access to stored secrets.")
			logrus.Warn("For details on using or changing a custom encryption configuration see:")
			logrus.Warn("https://rancher.com/docs/rke/v0.1.x/en/config-options/services/kube-api")
			fmt.Print("Are you sure you want to continue [y/n]: ")
			input, err := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if err != nil {
				return err
			}
			if input != "y" && input != "Y" {
				return fmt.Errorf("Aborted")
			}
		}
	}
	return nil
}

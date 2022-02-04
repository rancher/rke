package cmd

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const s3Endpoint = "s3.amazonaws.com"

func EtcdCommand() cli.Command {
	snapshotFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "Specify snapshot name",
		},
		cli.StringFlag{
			Name:   "config",
			Usage:  "Specify an alternate cluster YAML file",
			Value:  pki.ClusterConfig,
			EnvVar: "RKE_CONFIG",
		},
		cli.BoolFlag{
			Name:  "s3",
			Usage: "Enabled backup to s3",
		},
		cli.StringFlag{
			Name:  "s3-endpoint",
			Usage: "Specify s3 endpoint url",
			Value: s3Endpoint,
		},
		cli.StringFlag{
			Name:  "s3-endpoint-ca",
			Usage: "Specify a custom CA cert to connect to S3 endpoint",
		},
		cli.StringFlag{
			Name:  "access-key",
			Usage: "Specify s3 accessKey",
		},
		cli.StringFlag{
			Name:  "secret-key",
			Usage: "Specify s3 secretKey",
		},
		cli.StringFlag{
			Name:  "bucket-name",
			Usage: "Specify s3 bucket name",
		},
		cli.StringFlag{
			Name:  "region",
			Usage: "Specify the s3 bucket location (optional)",
		},
		cli.StringFlag{
			Name:  "folder",
			Usage: "Specify s3 folder name",
		},
	}

	snapshotSaveFlags := append(snapshotFlags, commonFlags...)

	snapshotRestoreFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "cert-dir",
			Usage: "Specify a certificate dir path",
		},
		cli.BoolFlag{
			Name:  "custom-certs",
			Usage: "Use custom certificates from a cert dir",
		},
		cli.BoolFlag{
			Name:  "use-local-state",
			Usage: "Use local state file (do not check or use snapshot archive for state file)",
		},
	}
	snapshotRestoreFlags = append(append(snapshotFlags, snapshotRestoreFlags...), commonFlags...)

	return cli.Command{
		Name:  "etcd",
		Usage: "etcd snapshot save/restore operations in k8s cluster",
		Subcommands: []cli.Command{
			{
				Name:   "snapshot-save",
				Usage:  "Take snapshot on all etcd hosts",
				Flags:  snapshotSaveFlags,
				Action: SnapshotSaveEtcdHostsFromCli,
			},
			{
				Name:   "snapshot-restore",
				Usage:  "Restore existing snapshot",
				Flags:  snapshotRestoreFlags,
				Action: RestoreEtcdSnapshotFromCli,
			},
		},
	}
}

func SnapshotSaveEtcdHosts(
	ctx context.Context,
	rkeConfig *v3.RancherKubernetesEngineConfig,
	dialersOptions hosts.DialersOptions,
	flags cluster.ExternalFlags, snapshotName string) error {

	log.Infof(ctx, "Starting saving snapshot on etcd hosts")

	stateFilePath := cluster.GetStateFilePath(flags.ClusterFilePath, flags.ConfigDir)

	kubeCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, "")
	if err != nil {
		return err
	}
	if err := kubeCluster.SetupDialers(ctx, dialersOptions); err != nil {
		return err
	}

	if err := kubeCluster.TunnelHosts(ctx, flags); err != nil {
		return err
	}

	if err := kubeCluster.DeployStateFile(ctx, stateFilePath, snapshotName); err != nil {
		return err
	}

	if err := kubeCluster.SnapshotEtcd(ctx, snapshotName); err != nil {
		return err
	}

	log.Infof(ctx, "Finished saving/uploading snapshot [%s] on all etcd hosts", snapshotName)
	return nil
}

func RestoreEtcdSnapshot(
	ctx context.Context,
	rkeConfig *v3.RancherKubernetesEngineConfig,
	dialersOptions hosts.DialersOptions,
	flags cluster.ExternalFlags,
	data map[string]interface{},
	snapshotName string) (string, string, string, string, map[string]pki.CertificatePKI, error) {
	var APIURL, caCrt, clientCert, clientKey string

	rkeFullState := &cluster.FullState{}
	stateFileRetrieved := false
	// Local state file
	stateFilePath := cluster.GetStateFilePath(flags.ClusterFilePath, flags.ConfigDir)

	if !flags.UseLocalState {
		log.Infof(ctx, "Checking if state file is included in snapshot file for [%s]", snapshotName)
		// Creating temp cluster to check if snapshot archive contains state file and retrieve it
		tempCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, "")
		if err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
		if err := tempCluster.SetupDialers(ctx, dialersOptions); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
		if err := tempCluster.TunnelHosts(ctx, flags); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}

		// Extract state file from snapshot
		stateFile, err := tempCluster.GetStateFileFromSnapshot(ctx, snapshotName)
		// If state file is not in snapshot (or can't be retrieved), fallback to local state file
		if err != nil {
			logrus.Infof("Could not extract state file from snapshot [%s] on any host, falling back to local state file: %v", snapshotName, err)
			rkeFullState, _ = cluster.ReadStateFile(ctx, stateFilePath)
		} else {
			// Parse extracted state file to FullState struct
			rkeFullState, err = cluster.StringToFullState(ctx, stateFile)
			if err != nil {
				logrus.Errorf("Error when converting state file contents to rkeFullState: %v", err)
				return APIURL, caCrt, clientCert, clientKey, nil, err
			}
			logrus.Infof("State file is successfully extracted from snapshot [%s]", snapshotName)
			stateFileRetrieved = true
		}
	} else {
		var err error
		log.Infof(ctx, "Not checking if state file is included in snapshot file for [%s], using local state file [%s]", snapshotName, stateFilePath)
		rkeFullState, err = cluster.ReadStateFile(ctx, stateFilePath)
		if err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}
	log.Infof(ctx, "Restoring etcd snapshot %s", snapshotName)

	kubeCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, rkeFullState.DesiredState.EncryptionConfig)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err := validateCerts(rkeFullState.DesiredState); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	// If we can't retrieve state file from snapshot, and we don't have local, we need to check for legacy cluster
	if !stateFileRetrieved || flags.UseLocalState {
		if err := checkLegacyCluster(ctx, kubeCluster, rkeFullState, flags); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}

	rkeFullState.CurrentState = cluster.State{}
	if err := rkeFullState.WriteStateFile(ctx, stateFilePath); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.SetupDialers(ctx, dialersOptions); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.TunnelHosts(ctx, flags); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	// if we fail after cleanup, we can't find the certs to do the download, we need to redeploy them
	if err := kubeCluster.DeployRestoreCerts(ctx, rkeFullState.DesiredState.CertificatesBundle); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	// first download and check
	if err := kubeCluster.PrepareBackup(ctx, snapshotName); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	log.Infof(ctx, "Cleaning old kubernetes cluster")
	if err := kubeCluster.CleanupNodes(ctx); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.RestoreEtcdSnapshot(ctx, snapshotName); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err := ClusterInit(ctx, rkeConfig, dialersOptions, flags); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	APIURL, caCrt, clientCert, clientKey, certs, err := ClusterUp(ctx, dialersOptions, flags, data)
	if err != nil {
		if !strings.Contains(err.Error(), "Provisioning incomplete") {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
		log.Warnf(ctx, err.Error())
	}

	if err := cluster.RestartClusterPods(ctx, kubeCluster); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if err := kubeCluster.RemoveOldNodes(ctx); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	log.Infof(ctx, "Finished restoring snapshot [%s] on all etcd hosts", snapshotName)
	return APIURL, caCrt, clientCert, clientKey, certs, err
}

func validateCerts(state cluster.State) error {
	var failedErrs error

	if state.RancherKubernetesEngineConfig == nil {
		// possibly already started a restore
		return nil
	}
	for name, certPKI := range state.CertificatesBundle {
		if name == pki.ServiceAccountTokenKeyName || name == pki.RequestHeaderCACertName || name == pki.KubeAdminCertName {
			continue
		}

		cert := certPKI.Certificate
		if cert == nil {
			if failedErrs == nil {
				failedErrs = fmt.Errorf("Certificate [%s] is nil", certPKI.Name)
			} else {
				failedErrs = errors.Wrap(failedErrs, fmt.Sprintf("Certificate [%s] is nil", certPKI.Name))
			}
			continue
		}

		certPool := x509.NewCertPool()
		certPool.AddCert(cert)
		if _, err := cert.Verify(x509.VerifyOptions{Roots: certPool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); err != nil {
			if failedErrs == nil {
				failedErrs = fmt.Errorf("Certificate [%s] failed verification: %v", certPKI.Name, err)
			} else {
				failedErrs = errors.Wrap(failedErrs, fmt.Sprintf("Certificate [%s] failed verification: %v", certPKI.Name, err))
			}
		}
	}
	if failedErrs != nil {
		return errors.Wrap(failedErrs, "[etcd] Failed to restore etcd snapshot: invalid certs")
	}
	return nil
}

func SnapshotSaveEtcdHostsFromCli(ctx *cli.Context) error {
	logrus.Infof("Running RKE version: %v", ctx.App.Version)
	clusterFile, filePath, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster file: %v", err)
	}

	rkeConfig, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return fmt.Errorf("failed to parse cluster file: %v", err)
	}

	rkeConfig, err = setOptionsFromCLI(ctx, rkeConfig)
	if err != nil {
		return err
	}
	// Check snapshot name
	etcdSnapshotName := ctx.String("name")
	if etcdSnapshotName == "" {
		etcdSnapshotName = fmt.Sprintf("rke_etcd_snapshot_%s", time.Now().Format(time.RFC3339))
		logrus.Warnf("Name of the snapshot is not specified, using [%s]", etcdSnapshotName)
	}
	// setting up the flags
	flags := cluster.GetExternalFlags(false, false, false, false, "", filePath)

	return SnapshotSaveEtcdHosts(context.Background(), rkeConfig, hosts.DialersOptions{}, flags, etcdSnapshotName)
}

func RestoreEtcdSnapshotFromCli(ctx *cli.Context) error {
	logrus.Infof("Running RKE version: %v", ctx.App.Version)
	clusterFile, filePath, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster file: %v", err)
	}

	rkeConfig, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return fmt.Errorf("failed to parse cluster file: %v", err)
	}

	rkeConfig, err = setOptionsFromCLI(ctx, rkeConfig)
	if err != nil {
		return err
	}
	etcdSnapshotName := ctx.String("name")
	if etcdSnapshotName == "" {
		return fmt.Errorf("you must specify the snapshot name to restore")
	}
	// Warn user if etcdSnapshotName contains extension (should just be snapshotname, not the filename)
	if strings.HasSuffix(etcdSnapshotName, ".zip") {
		logrus.Warnf("The snapshot name [%s] ends with the file extension (.zip) which is not needed, the snapshot name should be provided without the extension", etcdSnapshotName)
	}
	// setting up the flags
	// flag to use local state file
	useLocalState := ctx.Bool("use-local-state")

	flags := cluster.GetExternalFlags(false, false, false, useLocalState, "", filePath)
	// Custom certificates and certificate dir flags
	flags.CertificateDir = ctx.String("cert-dir")
	flags.CustomCerts = ctx.Bool("custom-certs")

	_, _, _, _, _, err = RestoreEtcdSnapshot(context.Background(), rkeConfig, hosts.DialersOptions{}, flags, map[string]interface{}{}, etcdSnapshotName)
	return err
}

func SnapshotRemoveFromEtcdHosts(
	ctx context.Context,
	rkeConfig *v3.RancherKubernetesEngineConfig,
	dialersOptions hosts.DialersOptions,
	flags cluster.ExternalFlags, snapshotName string) error {

	log.Infof(ctx, "Starting snapshot remove on etcd hosts")
	kubeCluster, err := cluster.InitClusterObject(ctx, rkeConfig, flags, "")
	if err != nil {
		return err
	}
	if err := kubeCluster.SetupDialers(ctx, dialersOptions); err != nil {
		return err
	}

	if err := kubeCluster.TunnelHosts(ctx, flags); err != nil {
		return err
	}

	if err := kubeCluster.RemoveEtcdSnapshot(ctx, snapshotName); err != nil {
		return err
	}

	log.Infof(ctx, "Finished removing snapshot [%s] from all etcd hosts", snapshotName)
	return nil
}

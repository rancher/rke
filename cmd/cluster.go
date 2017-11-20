package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/util/cert"
)

func ClusterCommand() cli.Command {
	clusterUpDownUpgradeFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "cluster-file",
			Usage:  "Specify an alternate cluster YAML file",
			Value:  cluster.DefaultClusterConfig,
			EnvVar: "CLUSTER_FILE",
		},
	}
	clusterForceRemove := []cli.Flag{
		cli.BoolFlag{
			Name:  "force",
			Usage: "Force removal of the cluster",
		},
	}
	return cli.Command{
		Name:      "cluster",
		ShortName: "cluster",
		Usage:     "Operations on the cluster",
		Flags:     clusterUpDownUpgradeFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "up",
				Usage:  "Bring the cluster up",
				Action: clusterUpFromCli,
				Flags:  clusterUpDownUpgradeFlags,
			},
			cli.Command{
				Name:   "down",
				Usage:  "Teardown the cluster and clean cluster nodes",
				Action: clusterDownFromCli,
				Flags:  append(clusterUpDownUpgradeFlags, clusterForceRemove...),
			},
			cli.Command{
				Name:   "version",
				Usage:  "Show Cluster Kubernetes version",
				Action: getClusterVersion,
				Flags:  clusterUpDownUpgradeFlags,
			},
			cli.Command{
				Name:   "upgrade",
				Usage:  "Upgrade Cluster Kubernetes version",
				Action: clusterUpgradeFromCli,
				Flags:  clusterUpDownUpgradeFlags,
			},
		},
	}
}

func ClusterUp(clusterFile string) (string, string, string, string, error) {
	logrus.Infof("Building Kubernetes cluster")
	var APIURL, caCrt, clientCert, clientKey string
	kubeCluster, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.TunnelHosts()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	currentCluster, err := kubeCluster.GetClusterState()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = cluster.SetUpAuthentication(kubeCluster, currentCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.SetUpHosts()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.DeployClusterPlanes()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = cluster.ReconcileCluster(kubeCluster, currentCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.SaveClusterState(clusterFile)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.DeployNetworkPlugin()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.DeployK8sAddOns()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.DeployUserAddOns()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	APIURL = fmt.Sprintf("https://" + kubeCluster.ControlPlaneHosts[0].IP + ":6443")
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Key))

	logrus.Infof("Finished building Kubernetes cluster successfully")
	return APIURL, caCrt, clientCert, clientKey, nil
}

func ClusterUpgrade(clusterFile string) (string, string, string, string, error) {
	logrus.Infof("Upgrading Kubernetes cluster")
	var APIURL, caCrt, clientCert, clientKey string
	kubeCluster, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	logrus.Debugf("Getting current cluster")
	currentCluster, err := kubeCluster.GetClusterState()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}
	logrus.Debugf("Setting up upgrade tunnels")
	/*
		kubeCluster is the cluster.yaml definition. It should have updated configuration
		currentCluster is the current state fetched from kubernetes
		we add currentCluster certs to kubeCluster, kubeCluster would have the latest configuration from cluster.yaml and the certs to connect to k8s and apply the upgrade
	*/
	kubeCluster.Certificates = currentCluster.Certificates
	err = kubeCluster.TunnelHosts()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}
	logrus.Debugf("Starting cluster upgrade")
	err = kubeCluster.ClusterUpgrade()
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	err = kubeCluster.SaveClusterState(clusterFile)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, err
	}

	logrus.Infof("Cluster upgraded successfully")

	APIURL = fmt.Sprintf("https://" + kubeCluster.ControlPlaneHosts[0].IP + ":6443")
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCommonName].Key))
	return APIURL, caCrt, clientCert, clientKey, nil
}

func ClusterDown(clusterFile string) error {
	logrus.Infof("Tearing down Kubernetes cluster")
	kubeCluster, err := cluster.ParseConfig(clusterFile)
	if err != nil {
		return err
	}

	err = kubeCluster.TunnelHosts()
	if err != nil {
		return err
	}

	logrus.Debugf("Starting Cluster removal")
	err = kubeCluster.ClusterDown()
	if err != nil {
		return err
	}

	logrus.Infof("Cluster removed successfully")
	return nil
}

func clusterUpFromCli(ctx *cli.Context) error {
	clusterFile, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to resolve cluster file: %v", err)
	}
	_, _, _, _, err = ClusterUp(clusterFile)
	return err
}

func clusterUpgradeFromCli(ctx *cli.Context) error {
	clusterFile, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to resolve cluster file: %v", err)
	}
	_, _, _, _, err = ClusterUpgrade(clusterFile)
	return err
}

func clusterDownFromCli(ctx *cli.Context) error {
	force := ctx.Bool("force")
	if !force {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure you want to remove Kubernetes cluster [y/n]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if input != "y" && input != "Y" {
			return nil
		}
	}
	clusterFile, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to resolve cluster file: %v", err)
	}
	return ClusterDown(clusterFile)
}

func resolveClusterFile(ctx *cli.Context) (string, error) {
	clusterFile := ctx.String("cluster-file")
	fp, err := filepath.Abs(clusterFile)
	if err != nil {
		return "", fmt.Errorf("failed to lookup current directory name: %v", err)
	}
	file, err := os.Open(fp)
	if err != nil {
		return "", fmt.Errorf("Can not find cluster configuration file: %v", err)
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	clusterFileBuff := string(buf)

	/*
		This is a hacky way to add config path to cluster object without messing with
		ClusterUp function and to avoid conflict with calls from kontainer-engine, basically
		i add config path (cluster.yml by default) to a field into the config buffer
		to be parsed later and added as ConfigPath field into cluster object.
	*/
	clusterFileBuff = fmt.Sprintf("%s\nconfig_path: %s\n", clusterFileBuff, clusterFile)
	return clusterFileBuff, nil
}

func getClusterVersion(ctx *cli.Context) error {
	localKubeConfig := cluster.GetLocalKubeConfig(ctx.String("cluster-file"))
	serverVersion, err := cluster.GetK8sVersion(localKubeConfig)
	if err != nil {
		return err
	}
	fmt.Printf("Server Version: %s\n", serverVersion)
	return nil
}

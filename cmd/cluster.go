package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/urfave/cli"
)

func ClusterCommand() cli.Command {
	clusterUpFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "cluster-file",
			Usage:  "Specify an alternate cluster YAML file (default: cluster.yml)",
			Value:  "cluster.yml",
			EnvVar: "CLUSTER_FILE",
		},
		cli.BoolFlag{
			Name:  "force-crts",
			Usage: "Force rotating the Kubernetes components certificates",
		},
	}
	return cli.Command{
		Name:      "cluster",
		ShortName: "cluster",
		Usage:     "Operations on the cluster",
		Flags:     clusterUpFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "up",
				Usage:  "Bring the cluster up",
				Action: clusterUp,
				Flags:  clusterUpFlags,
			},
		},
	}
}

func clusterUp(ctx *cli.Context) error {
	logrus.Infof("Building up Kubernetes cluster")
	clusterFile, err := resolveClusterFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to bring cluster up: %v", err)
	}
	logrus.Debugf("Parsing cluster file [%v]", clusterFile)
	servicesLookup, k8shosts, err := parseClusterFile(clusterFile)
	if err != nil {
		return fmt.Errorf("Failed to parse the cluster file: %v", err)
	}
	for i := range k8shosts {
		// Set up socket tunneling
		k8shosts[i].TunnelUp(ctx)
		defer k8shosts[i].DClient.Close()
		if err != nil {
			return err
		}
	}
	etcdHosts, cpHosts, workerHosts := hosts.DivideHosts(k8shosts)
	KubernetesServiceIP, err := services.GetKubernetesServiceIp(servicesLookup.Services.KubeAPI.ServiceClusterIPRange)
	clusterDomain := servicesLookup.Services.Kubelet.ClusterDomain
	if err != nil {
		return err
	}
	err = pki.StartCertificatesGeneration(ctx, cpHosts, workerHosts, clusterDomain, KubernetesServiceIP)
	if err != nil {
		return fmt.Errorf("[Certificates] Failed to generate Kubernetes certificates: %v", err)
	}
	err = services.RunEtcdPlane(etcdHosts, servicesLookup.Services.Etcd)
	if err != nil {
		return fmt.Errorf("[Etcd] Failed to bring up Etcd Plane: %v", err)
	}
	err = services.RunControlPlane(cpHosts, etcdHosts, servicesLookup.Services)
	if err != nil {
		return fmt.Errorf("[ControlPlane] Failed to bring up Control Plane: %v", err)
	}
	err = services.RunWorkerPlane(cpHosts, workerHosts, servicesLookup.Services)
	if err != nil {
		return fmt.Errorf("[WorkerPlane] Failed to bring up Worker Plane: %v", err)
	}
	return nil
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
	clusterFile = string(buf)

	return clusterFile, nil
}

func parseClusterFile(clusterFile string) (*services.Container, []hosts.Host, error) {
	// parse hosts
	k8shosts := hosts.Hosts{}
	err := yaml.Unmarshal([]byte(clusterFile), &k8shosts)
	if err != nil {
		return nil, nil, err
	}
	for i, host := range k8shosts.Hosts {
		if len(host.Hostname) == 0 {
			return nil, nil, fmt.Errorf("Hostname for host (%d) is not provided", i+1)
		} else if len(host.User) == 0 {
			return nil, nil, fmt.Errorf("User for host (%d) is not provided", i+1)
		} else if len(host.Role) == 0 {
			return nil, nil, fmt.Errorf("Role for host (%d) is not provided", i+1)

		} else if host.AdvertiseAddress == "" {
			// if control_plane_ip is not set,
			// default to the main IP
			k8shosts.Hosts[i].AdvertiseAddress = host.IP
		}
		for _, role := range host.Role {
			if role != services.ETCDRole && role != services.ControlRole && role != services.WorkerRole {
				return nil, nil, fmt.Errorf("Role [%s] for host (%d) is not recognized", role, i+1)
			}
		}
	}
	// parse services
	var servicesContainer services.Container
	err = yaml.Unmarshal([]byte(clusterFile), &servicesContainer)
	if err != nil {
		return nil, nil, err
	}
	return &servicesContainer, k8shosts.Hosts, nil
}

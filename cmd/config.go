package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/cluster.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

func ConfigCommand() cli.Command {
	return cli.Command{
		Name:      "config",
		ShortName: "config",
		Usage:     "Setup cluster configuration",
		Action:    clusterConfig,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name,n",
				Usage: "Name of the configuration file",
				Value: cluster.DefaultClusterConfig,
			},
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Generate Empty configuration file",
			},
			cli.BoolFlag{
				Name:  "print,p",
				Usage: "Print configuration",
			},
		},
	}
}

func getConfig(reader *bufio.Reader, text, def string) (string, error) {
	for {
		fmt.Printf("%s [%s]: ", text, def)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input != "" {
			return input, nil
		}
		return def, nil
	}
}

func writeConfig(cluster *v1.RancherKubernetesEngineConfig, configFile string, print bool) error {
	yamlConfig, err := yaml.Marshal(*cluster)
	if err != nil {
		return err
	}
	logrus.Debugf("Deploying cluster configuration file: %s", configFile)

	if print {
		fmt.Printf("Configuration File: \n%s", string(yamlConfig))
		return nil
	}
	return ioutil.WriteFile(configFile, yamlConfig, 0640)

}

func clusterConfig(ctx *cli.Context) error {
	configFile := ctx.String("name")
	print := ctx.Bool("print")
	cluster := v1.RancherKubernetesEngineConfig{}

	// Get cluster config from user
	reader := bufio.NewReader(os.Stdin)

	// Generate empty configuration file
	if ctx.Bool("empty") {
		cluster.Hosts = make([]v1.RKEConfigHost, 1)
		return writeConfig(&cluster, configFile, print)
	}

	// Get number of hosts
	numberOfHostsString, err := getConfig(reader, "Number of Hosts", "3")
	if err != nil {
		return err
	}
	numberOfHostsInt, err := strconv.Atoi(numberOfHostsString)
	if err != nil {
		return err
	}

	// Get Hosts config
	cluster.Hosts = make([]v1.RKEConfigHost, 0)
	for i := 0; i < numberOfHostsInt; i++ {
		hostCfg, err := getHostConfig(reader, i)
		if err != nil {
			return err
		}
		cluster.Hosts = append(cluster.Hosts, *hostCfg)
	}

	// Get Network config
	networkConfig, err := getNetworkConfig(reader)
	if err != nil {
		return err
	}
	cluster.Network = *networkConfig

	// Get Authentication Config
	authConfig, err := getAuthConfig(reader)
	if err != nil {
		return err
	}
	cluster.Authentication = *authConfig

	// Get Services Config
	serviceConfig, err := getServiceConfig(reader)
	if err != nil {
		return err
	}
	cluster.Services = *serviceConfig

	return writeConfig(&cluster, configFile, print)
}

func getHostConfig(reader *bufio.Reader, index int) (*v1.RKEConfigHost, error) {
	host := v1.RKEConfigHost{}
	advertisedHostname, err := getConfig(reader, fmt.Sprintf("Hostname of host (%d)", index+1), "")
	if err != nil {
		return nil, err
	}
	host.AdvertisedHostname = advertisedHostname

	sshIP, err := getConfig(reader, fmt.Sprintf("SSH IP of host (%s)", advertisedHostname), "")
	if err != nil {
		return nil, err
	}
	host.IP = sshIP

	advertisedIP, err := getConfig(reader, fmt.Sprintf("Advertised IP of host (%s)", advertisedHostname), "")
	if err != nil {
		return nil, err
	}
	host.AdvertiseAddress = advertisedIP

	isControlHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) a control host (y/n)?", advertisedHostname), "y")
	if err != nil {
		return nil, err
	}
	if isControlHost == "y" || isControlHost == "Y" {
		host.Role = append(host.Role, services.ControlRole)
	}

	isWorkerHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) a worker host (y/n)?", advertisedHostname), "n")
	if err != nil {
		return nil, err
	}
	if isWorkerHost == "y" || isWorkerHost == "Y" {
		host.Role = append(host.Role, services.WorkerRole)
	}

	isEtcdHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) an Etcd host (y/n)?", advertisedHostname), "n")
	if err != nil {
		return nil, err
	}
	if isEtcdHost == "y" || isEtcdHost == "Y" {
		host.Role = append(host.Role, services.ETCDRole)
	}

	sshUser, err := getConfig(reader, fmt.Sprintf("SSH User of host (%s)", advertisedHostname), "ubuntu")
	if err != nil {
		return nil, err
	}
	host.User = sshUser

	dockerSocketPath, err := getConfig(reader, fmt.Sprintf("Docker socket path on host (%s)", advertisedHostname), "/var/run/docker.sock")
	if err != nil {
		return nil, err
	}
	host.DockerSocket = dockerSocketPath
	return &host, nil
}

func getServiceConfig(reader *bufio.Reader) (*v1.RKEConfigServices, error) {
	servicesConfig := v1.RKEConfigServices{}
	servicesConfig.Etcd = v1.ETCDService{}
	servicesConfig.KubeAPI = v1.KubeAPIService{}
	servicesConfig.KubeController = v1.KubeControllerService{}
	servicesConfig.Scheduler = v1.SchedulerService{}
	servicesConfig.Kubelet = v1.KubeletService{}
	servicesConfig.Kubeproxy = v1.KubeproxyService{}

	etcdImage, err := getConfig(reader, "Etcd Docker Image", "quay.io/coreos/etcd:latest")
	if err != nil {
		return nil, err
	}
	servicesConfig.Etcd.Image = etcdImage

	kubeImage, err := getConfig(reader, "Kubernetes Docker image", "quay.io/coreos/hyperkube:v1.7.5_coreos.0")
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeAPI.Image = kubeImage
	servicesConfig.KubeController.Image = kubeImage
	servicesConfig.Scheduler.Image = kubeImage
	servicesConfig.Kubelet.Image = kubeImage
	servicesConfig.Kubeproxy.Image = kubeImage

	clusterDomain, err := getConfig(reader, "Cluster domain", "cluster.local")
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.ClusterDomain = clusterDomain

	serviceClusterIPRange, err := getConfig(reader, "Service Cluster IP Range", "10.233.0.0/18")
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeAPI.ServiceClusterIPRange = serviceClusterIPRange
	servicesConfig.KubeController.ServiceClusterIPRange = serviceClusterIPRange

	clusterNetworkCidr, err := getConfig(reader, "Cluster Network CIDR", "10.233.64.0/18")
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeController.ClusterCIDR = clusterNetworkCidr

	clusterDNSServiceIP, err := getConfig(reader, "Cluster DNS Service IP", "10.233.0.3")
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.ClusterDNSServer = clusterDNSServiceIP

	infraPodImage, err := getConfig(reader, "Infra Container image", "gcr.io/google_containers/pause-amd64:3.0")
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.InfraContainerImage = infraPodImage
	return &servicesConfig, nil
}

func getAuthConfig(reader *bufio.Reader) (*v1.AuthConfig, error) {
	authConfig := v1.AuthConfig{}

	authType, err := getConfig(reader, "Authentication Strategy", "x509")
	if err != nil {
		return nil, err
	}
	authConfig.Strategy = authType
	return &authConfig, nil
}

func getNetworkConfig(reader *bufio.Reader) (*v1.NetworkConfig, error) {
	networkConfig := v1.NetworkConfig{}

	networkPlugin, err := getConfig(reader, "Network Plugin Type", "flannel")
	if err != nil {
		return nil, err
	}
	networkConfig.Plugin = networkPlugin
	return &networkConfig, nil
}

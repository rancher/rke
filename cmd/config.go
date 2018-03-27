package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/services"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

const (
	comments = `# If you intened to deploy Kubernetes in an air-gapped environment,
# please consult the documentation on how to configure custom RKE images.`
)

func ConfigCommand() cli.Command {
	return cli.Command{
		Name:   "config",
		Usage:  "Setup cluster configuration",
		Action: clusterConfig,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name,n",
				Usage: "Name of the configuration file",
				Value: pki.ClusterConfig,
			},
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Generate Empty configuration file",
			},
			cli.BoolFlag{
				Name:  "print,p",
				Usage: "Print configuration",
			},
			cli.BoolFlag{
				Name:  "from-machine,m",
				Usage: "Get node configuration from docker-machine",
			},
		},
	}
}

func getConfig(reader *bufio.Reader, text, def string) (string, error) {
	for {
		if def == "" {
			fmt.Printf("[+] %s [%s]: ", text, "none")
		} else {
			fmt.Printf("[+] %s [%s]: ", text, def)
		}
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

func writeConfig(cluster *v3.RancherKubernetesEngineConfig, configFile string, print bool) error {
	yamlConfig, err := yaml.Marshal(*cluster)
	if err != nil {
		return err
	}
	logrus.Debugf("Deploying cluster configuration file: %s", configFile)

	configString := fmt.Sprintf("%s\n%s", comments, string(yamlConfig))
	if print {
		fmt.Printf("Configuration File: \n%s", configString)
		return nil
	}
	return ioutil.WriteFile(configFile, []byte(configString), 0640)
}

func clusterConfig(ctx *cli.Context) error {
	configFile := ctx.String("name")
	print := ctx.Bool("print")
	cluster := v3.RancherKubernetesEngineConfig{}

	// Get cluster config from user
	reader := bufio.NewReader(os.Stdin)

	// Generate empty configuration file
	if ctx.Bool("empty") {
		cluster.Nodes = make([]v3.RKEConfigNode, 1)
		return writeConfig(&cluster, configFile, print)
	}

	sshKeyPath, err := getConfig(reader, "Cluster Level SSH Private Key Path", "~/.ssh/id_rsa")
	if err != nil {
		return err
	}
	cluster.SSHKeyPath = sshKeyPath

	cluster.Nodes = make([]v3.RKEConfigNode, 0)
	if ctx.Bool("from-machine") {
		nodes, err := configureFromMachine(reader)

		if err != nil {
			return err
		}
		cluster.Nodes = append(cluster.Nodes, nodes...)
	} else {

		// Get number of hosts
		numberOfHostsString, err := getConfig(reader, "Number of Hosts", "1")
		if err != nil {
			return err
		}
		numberOfHostsInt, err := strconv.Atoi(numberOfHostsString)
		if err != nil {
			return err
		}

		// Get Hosts config
		for i := 0; i < numberOfHostsInt; i++ {
			hostCfg, err := getHostConfig(reader, i, cluster.SSHKeyPath)
			if err != nil {
				return err
			}
			cluster.Nodes = append(cluster.Nodes, *hostCfg)
		}
	}
	// Get Network config
	networkConfig, err := getNetworkConfig(reader)
	if err != nil {
		return err
	}
	cluster.Network = *networkConfig

	// Get Authentication Config
	authnConfig, err := getAuthnConfig(reader)
	if err != nil {
		return err
	}
	cluster.Authentication = *authnConfig

	// Get Authorization config
	authzConfig, err := getAuthzConfig(reader)
	if err != nil {
		return err
	}
	cluster.Authorization = *authzConfig

	// Get Services Config
	serviceConfig, err := getServiceConfig(reader)
	if err != nil {
		return err
	}
	cluster.Services = *serviceConfig

	return writeConfig(&cluster, configFile, print)
}

func getHostConfig(reader *bufio.Reader, index int, clusterSSHKeyPath string) (*v3.RKEConfigNode, error) {
	host := v3.RKEConfigNode{}

	address, err := getConfig(reader, fmt.Sprintf("SSH Address of host (%d)", index+1), "")
	if err != nil {
		return nil, err
	}
	host.Address = address

	port, err := getConfig(reader, fmt.Sprintf("SSH Port of host (%d)", index+1), cluster.DefaultSSHPort)
	if err != nil {
		return nil, err
	}
	host.Port = port

	sshKeyPath, err := getConfig(reader, fmt.Sprintf("SSH Private Key Path of host (%s)", address), "")
	if err != nil {
		return nil, err
	}
	if len(sshKeyPath) == 0 {
		fmt.Printf("[-] You have entered empty SSH key path, trying fetch from SSH key parameter\n")
		sshKey, err := getConfig(reader, fmt.Sprintf("SSH Private Key of host (%s)", address), "")
		if err != nil {
			return nil, err
		}
		if len(sshKey) == 0 {
			fmt.Printf("[-] You have entered empty SSH key, defaulting to cluster level SSH key: %s\n", clusterSSHKeyPath)
			host.SSHKeyPath = clusterSSHKeyPath
		} else {
			host.SSHKey = sshKey
		}
	} else {
		host.SSHKeyPath = sshKeyPath
	}

	sshUser, err := getConfig(reader, fmt.Sprintf("SSH User of host (%s)", address), "ubuntu")
	if err != nil {
		return nil, err
	}
	host.User = sshUser

	isControlHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) a control host (y/n)?", address), "y")
	if err != nil {
		return nil, err
	}
	if isControlHost == "y" || isControlHost == "Y" {
		host.Role = append(host.Role, services.ControlRole)
	}

	isWorkerHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) a worker host (y/n)?", address), "n")
	if err != nil {
		return nil, err
	}
	if isWorkerHost == "y" || isWorkerHost == "Y" {
		host.Role = append(host.Role, services.WorkerRole)
	}

	isEtcdHost, err := getConfig(reader, fmt.Sprintf("Is host (%s) an Etcd host (y/n)?", address), "n")
	if err != nil {
		return nil, err
	}
	if isEtcdHost == "y" || isEtcdHost == "Y" {
		host.Role = append(host.Role, services.ETCDRole)
	}

	hostnameOverride, err := getConfig(reader, fmt.Sprintf("Override Hostname of host (%s)", address), "")
	if err != nil {
		return nil, err
	}
	host.HostnameOverride = hostnameOverride

	internalAddress, err := getConfig(reader, fmt.Sprintf("Internal IP of host (%s)", address), "")
	if err != nil {
		return nil, err
	}
	host.InternalAddress = internalAddress

	dockerSocketPath, err := getConfig(reader, fmt.Sprintf("Docker socket path on host (%s)", address), cluster.DefaultDockerSockPath)
	if err != nil {
		return nil, err
	}
	host.DockerSocket = dockerSocketPath
	return &host, nil
}

func getServiceConfig(reader *bufio.Reader) (*v3.RKEConfigServices, error) {
	servicesConfig := v3.RKEConfigServices{}
	servicesConfig.Etcd = v3.ETCDService{}
	servicesConfig.KubeAPI = v3.KubeAPIService{}
	servicesConfig.KubeController = v3.KubeControllerService{}
	servicesConfig.Scheduler = v3.SchedulerService{}
	servicesConfig.Kubelet = v3.KubeletService{}
	servicesConfig.Kubeproxy = v3.KubeproxyService{}

	imageDefaults := v3.K8sVersionToRKESystemImages[cluster.DefaultK8sVersion]

	etcdImage, err := getConfig(reader, "Etcd Docker Image", imageDefaults.Etcd)
	if err != nil {
		return nil, err
	}
	servicesConfig.Etcd.Image = etcdImage

	kubeImage, err := getConfig(reader, "Kubernetes Docker image", imageDefaults.Kubernetes)
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeAPI.Image = kubeImage
	servicesConfig.KubeController.Image = kubeImage
	servicesConfig.Scheduler.Image = kubeImage
	servicesConfig.Kubelet.Image = kubeImage
	servicesConfig.Kubeproxy.Image = kubeImage

	clusterDomain, err := getConfig(reader, "Cluster domain", cluster.DefaultClusterDomain)
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.ClusterDomain = clusterDomain

	serviceClusterIPRange, err := getConfig(reader, "Service Cluster IP Range", cluster.DefaultServiceClusterIPRange)
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeAPI.ServiceClusterIPRange = serviceClusterIPRange
	servicesConfig.KubeController.ServiceClusterIPRange = serviceClusterIPRange

	podSecurityPolicy, err := getConfig(reader, "Enable PodSecurityPolicy", "n")
	if err != nil {
		return nil, err
	}
	if podSecurityPolicy == "y" || podSecurityPolicy == "Y" {
		servicesConfig.KubeAPI.PodSecurityPolicy = true
	} else {
		servicesConfig.KubeAPI.PodSecurityPolicy = false
	}

	clusterNetworkCidr, err := getConfig(reader, "Cluster Network CIDR", cluster.DefaultClusterCIDR)
	if err != nil {
		return nil, err
	}
	servicesConfig.KubeController.ClusterCIDR = clusterNetworkCidr

	clusterDNSServiceIP, err := getConfig(reader, "Cluster DNS Service IP", cluster.DefaultClusterDNSService)
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.ClusterDNSServer = clusterDNSServiceIP

	infraPodImage, err := getConfig(reader, "Infra Container image", imageDefaults.PodInfraContainer)
	if err != nil {
		return nil, err
	}
	servicesConfig.Kubelet.InfraContainerImage = infraPodImage
	return &servicesConfig, nil
}

func getAuthnConfig(reader *bufio.Reader) (*v3.AuthnConfig, error) {
	authnConfig := v3.AuthnConfig{}

	authnType, err := getConfig(reader, "Authentication Strategy", cluster.DefaultAuthStrategy)
	if err != nil {
		return nil, err
	}
	authnConfig.Strategy = authnType
	return &authnConfig, nil
}

func getAuthzConfig(reader *bufio.Reader) (*v3.AuthzConfig, error) {
	authzConfig := v3.AuthzConfig{}
	authzMode, err := getConfig(reader, "Authorization Mode (rbac, none)", cluster.DefaultAuthorizationMode)
	if err != nil {
		return nil, err
	}
	authzConfig.Mode = authzMode
	return &authzConfig, nil
}

func getNetworkConfig(reader *bufio.Reader) (*v3.NetworkConfig, error) {
	networkConfig := v3.NetworkConfig{}

	networkPlugin, err := getConfig(reader, "Network Plugin Type (flannel, calico, weave, canal)", cluster.DefaultNetworkPlugin)
	if err != nil {
		return nil, err
	}
	networkConfig.Plugin = networkPlugin
	return &networkConfig, nil
}

type NodeMachineConfig struct {
	ConfigVersion int `json:"ConfigVersion,omitempty"`
	Driver        struct {
		IPAddress  string `json:"IPAddress,omitempty"`
		SSHKeyPair string `json:"SSHKeyPair,omitempty"`
		SSHKeyPath string `json:"SSHKeyPath,omitempty"`
		SSHPort    int    `json:"SSHPort,omitempty"`
		SSHUser    string `json:"SSHUser,omitempty"`
	} `json:"Driver,omitempty"`
	HostOptions struct {
		EngineOptions struct {
			Labels []string `json:"Labels,omitempty"`
		} `json:"EngineOptions,omitempty"`
	}
	Name string `json:"Name,omitempty"`
}

func configureFromMachine(reader *bufio.Reader) ([]v3.RKEConfigNode, error) {

	dockerMachineStore := os.Getenv("MACHINE_STORAGE_PATH")

	if dockerMachineStore == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		dockerMachineStore = fmt.Sprintf("%s/.docker/machine/machines", usr.HomeDir)
	}

	// Get the docker-machine store path
	machineStorePath, err := getConfig(reader, "Docker Machine storage path", dockerMachineStore)

	if err != nil {
		return nil, err
	}

	machines, err := getMachines(machineStorePath)

	if err != nil {
		return nil, err
	}

	selectedNodes, err := getConfig(reader, "Which nodes would you like to use", strings.Join(machines, ","))
	if err != nil {
		return nil, err
	}
	nodes := make([]v3.RKEConfigNode, 0)
	for _, m := range strings.Split(selectedNodes, ",") {
		node, err := getMachineConfig(machineStorePath, m)

		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func getMachines(storePath string) ([]string, error) {
	var machines []string
	dirs, err := ioutil.ReadDir(storePath)

	if err != nil {
		return nil, err
	}

	for _, d := range dirs {
		machines = append(machines, d.Name())
	}

	return machines, nil
}

func getMachineConfig(storePath, machine string) (v3.RKEConfigNode, error) {
	nodeConfig := v3.RKEConfigNode{}
	configPath := fmt.Sprintf("%s/%s/config.json", storePath, machine)
	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nodeConfig, err
	}

	config := &NodeMachineConfig{}

	if err := json.Unmarshal(configFile, config); err != nil {
		return nodeConfig, err
	}

	if config.Driver.IPAddress == "" {
		return nodeConfig, errors.New(fmt.Sprintf("IPAddress not defined in config.json for %s. "+
			"Please ensure that the driver is setting these values correctly", machine))
	}

	nodeConfig.Role = getRolesFromLabels(config)
	nodeConfig.SSHKeyPath = config.Driver.SSHKeyPath
	nodeConfig.User = config.Driver.SSHUser
	nodeConfig.Address = config.Driver.IPAddress
	nodeConfig.Port = strconv.Itoa(config.Driver.SSHPort)

	return nodeConfig, nil
}

func getRolesFromLabels(conf *NodeMachineConfig) []string {
	var nodeRoles []string
	labels := conf.HostOptions.EngineOptions.Labels

	if len(labels) == 0 {
		return append(nodeRoles, "worker", "controlplane", "etcd")
	}

	// Parse the labels and determine roles
	for _, l := range labels {
		if strings.ContainsAny(l, "worker controlplane etcd") {
			label := strings.Split(l, "=")
			ok, err := strconv.ParseBool(label[1])

			if err != nil {
				logrus.Infof("[config] Could not parse bool from label %s", label[0])
				continue
			}

			if ok {
				nodeRoles = append(nodeRoles, label[0])
			}

		}
	}

	if len(nodeRoles) == 0 {
		nodeRoles = append(nodeRoles, "worker", "controlplane", "etcd")
	}

	return nodeRoles

}

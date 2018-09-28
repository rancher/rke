package dockermachine

import (
	"bufio"
	"encoding/json"
	"fmt"
	ncp "github.com/rancher/rke/nodeconfigproviders"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os/user"
	"strconv"
	"strings"
)

type Provider struct {
	dockerMachineStorePath string
}

type dockerMachineConfig struct {
	ConfigVersion int `json:"ConfigVersion,omitempty"`
	Driver        struct {
		IPAddress  string `json:"IPAddress"`
		SSHKeyPair string `json:"SSHKeyPair"`
		SSHKeyPath string `json:"SSHKeyPath"`
		SSHPort    int    `json:"SSHPort"`
		SSHUser    string `json:"SSHUser"`
	} `json:"Driver,omitempty"`
	HostOptions struct {
		EngineOptions struct {
			Labels []string `json:"Labels,omitempty"`
		} `json:"EngineOptions,omitempty"`
	}
	Name string `json:"Name,omitempty"`
}

func init() {
	ncp.RegisterNodeConfigProvider("docker-machine", &Provider{})
}

func (d *Provider) Init() error {
	if err := d.setMachinePath(); err != nil {
		return err
	}

	return nil
}

func (d *Provider) GetNodesFromConfig(reader *bufio.Reader) ([]string, error) {
	var err error

	// Get the docker-machine store path if it needs to read from a new directory
	d.dockerMachineStorePath, err = ncp.Prompt(reader, "Docker Machine storage path", d.dockerMachineStorePath)

	if err != nil {
		return nil, err
	}

	machines, err := d.getMachines()

	if err != nil {
		return nil, err
	}

	deployNodes := ""
	// make sure at least a node is passed
	for deployNodes == "" {
		deployNodes, err = ncp.Prompt(reader, "Which nodes would you like to use (cannot be none ctrl+c to exit)", strings.Join(machines, ","))
	}

	if err != nil {
		return nil, err
	}

	return strings.Split(deployNodes, ","), nil
}
func (d *Provider) ReadNodeConfigurations(nodes []string) ([]v3.RKEConfigNode, error) {
	var nodeConfigs []v3.RKEConfigNode

	for _, n := range nodes {
		node, err := readNodeConfig(n, d.dockerMachineStorePath)

		if err != nil {
			return nil, err
		}

		nodeConfigs = append(nodeConfigs, node)

	}

	return nodeConfigs, nil
}

func (d *Provider) getMachines() ([]string, error) {
	var machines []string
	dirs, err := ioutil.ReadDir(d.dockerMachineStorePath)

	if err != nil {
		return nil, err
	}

	for _, d := range dirs {
		machines = append(machines, d.Name())
	}

	return machines, nil
}

func readNodeConfig(node, storePath string) (v3.RKEConfigNode, error) {
	nodeConfig := v3.RKEConfigNode{}
	config := &dockerMachineConfig{}
	configPath := fmt.Sprintf("%s/%s/config.json", storePath, node)

	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nodeConfig, err
	}

	if err := json.Unmarshal(configFile, config); err != nil {
		return nodeConfig, err
	}

	if config.Driver.IPAddress == "" {
		logrus.Infof("IPAddress is not defined in the config.json for the machine %s", node)
	}

	sshPort := strconv.Itoa(config.Driver.SSHPort)

	if sshPort == "" {
		sshPort = "22"
	}

	nodeConfig.Address = config.Driver.IPAddress
	nodeConfig.SSHKeyPath = config.Driver.SSHKeyPath
	nodeConfig.User = config.Driver.SSHUser
	nodeConfig.Port = sshPort
	nodeConfig.Role = getRolesFromLabels(config)

	return nodeConfig, nil
}

func getRolesFromLabels(conf *dockerMachineConfig) []string {
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

func (d *Provider) setMachinePath() error {
	// Set the default Docker Machine Storage Path
	if d.dockerMachineStorePath == "" {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		d.dockerMachineStorePath = fmt.Sprintf("%s/.docker/machine/machines", usr.HomeDir)
	} else {
		// Make sure that the machines directory is present in the path provided
		if !strings.Contains(d.dockerMachineStorePath, "machines") {
			d.dockerMachineStorePath = fmt.Sprintf("%s/machines", d.dockerMachineStorePath)
		}
	}

	return nil
}

package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
)

var dockerMachineStorePath = os.Getenv("MACHINE_STORAGE_PATH")

type DockerMachineProvider struct{}

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

func (d *DockerMachineProvider) ListNodes(reader *bufio.Reader) ([]string, error) {

	var err error

	// Set the default Docker Machine Storage Path
	if dockerMachineStorePath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		dockerMachineStorePath = fmt.Sprintf("%s/.docker/machine/machines", usr.HomeDir)
	} else {
		if !strings.Contains(dockerMachineStorePath, "machines") {
			dockerMachineStorePath = fmt.Sprintf("%s/machines", dockerMachineStorePath)
		}
	}

	// Get the docker-machine store path if it needs to read from a new directory
	dockerMachineStorePath, err = prompt(reader, "Docker Machine storage path", dockerMachineStorePath)

	if err != nil {
		return nil, err
	}

	machines, err := getMachines()

	if err != nil {
		return nil, err
	}

	deployNodes, err := prompt(reader, "Which nodes would you like to use", strings.Join(machines, ","))

	if err != nil {
		return nil, err
	}

	return strings.Split(deployNodes, ","), nil
}

func getMachines() ([]string, error) {
	var machines []string
	dirs, err := ioutil.ReadDir(dockerMachineStorePath)

	if err != nil {
		return nil, err
	}

	for _, d := range dirs {
		machines = append(machines, d.Name())
	}

	return machines, nil
}

func (d *DockerMachineProvider) GetNodesConfig(nodes []string) ([]v3.RKEConfigNode, error) {
	var nodeConfigs []v3.RKEConfigNode

	for _, n := range nodes {
		node, err := readNodeConfig(n, dockerMachineStorePath)

		if err != nil {
			return nil, err
		}

		nodeConfigs = append(nodeConfigs, node)

	}

	return nodeConfigs, nil
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
		return nodeConfig, fmt.Errorf("IPAddress not defined in config.json for %s. "+
			"Please ensure that the driver is setting these values correctly", node)
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

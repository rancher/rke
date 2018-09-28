package nodeconfigproviders

import (
	"bufio"
	"fmt"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"strings"
)

type NodeConfigProvider interface {
	Init() error
	GetNodesFromConfig(reader *bufio.Reader) ([]string, error)
	ReadNodeConfigurations(deployNodes []string) ([]v3.RKEConfigNode, error)
}

var (
	nodeProviders = make(map[string]NodeConfigProvider)
)

func RegisterNodeConfigProvider(name string, provider NodeConfigProvider) {
	if _, exists := nodeProviders[name]; exists {
		logrus.Fatalf("Node configuration provider %s exists", name)
	}

	nodeProviders[name] = provider
}

func GetNodeProvider(name string) (NodeConfigProvider, error) {
	if provider, ok := nodeProviders[name]; ok {
		if err := provider.Init(); err != nil {
			return nil, err
		}
		return provider, nil
	}

	return nil, fmt.Errorf("no such provider %s", name)
}

func Prompt(reader *bufio.Reader, text, def string) (string, error) {
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

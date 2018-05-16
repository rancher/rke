package providers

import (
	"bufio"
	"fmt"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"strings"
	"sync"
)

type NodeProvider interface {
	ListNodes(reader *bufio.Reader) ([]string, error)
	GetNodesConfig(deployNodes []string) ([]v3.RKEConfigNode, error)
}

var (
	nodeProviders = map[string]NodeProvider{
		"docker-machine": new(DockerMachineProvider),
	}
	nodeProvidersMu = sync.RWMutex{}
)

func SetNodeProvider(name string, provider NodeProvider) {
	nodeProvidersMu.Lock()
	nodeProviders[name] = provider
	nodeProvidersMu.Unlock()
}

func GetNodeProvider(name string) (NodeProvider, bool) {
	nodeProvidersMu.RLock()
	defer nodeProvidersMu.RUnlock()
	provider, ok := nodeProviders[name]
	return provider, ok
}

func ListProviders() string {
	var providers []string
	for k := range nodeProviders {
		providers = append(providers, k)
	}

	return strings.Join(providers, ",")
}

func prompt(reader *bufio.Reader, text, def string) (string, error) {
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

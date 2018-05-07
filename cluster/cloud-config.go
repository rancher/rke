package cluster

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	types "github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	OpenstackProvider = "openstack"
)

// GenerateCloudConfig takes the provider cloud config values and generates
// a textual config format for the provider
func GenerateCloudConfig(provider *types.CloudProvider) (string, error) {
	switch provider.Name {
	case OpenstackProvider:
		return openstackCloudConfig(provider)
	default:
		return genericCloudConfig(provider)
	}
}

// openstackCloudConfig generates a conformant openstack cloud config in INI
// All config key/values are expected to be part of the [Global] section
func openstackCloudConfig(provider *types.CloudProvider) (string, error) {
	var (
		keys, lines []string
		cfg         = provider.CloudConfig
	)
	for k := range cfg {
		keys = append(keys, k)
	}
	// keep this sorted for reproducible configs
	sort.Strings(keys)

	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", k, cfg[k]))
	}
	return fmt.Sprintf(`[Global]
%s
`, strings.Join(lines, "\n")), nil
}

// genericCloudConfig generates generic cloud config in json format
func genericCloudConfig(provider *types.CloudProvider) (string, error) {
	cfgMap := parseConfigValues(provider.CloudConfig)
	jsonString, err := json.MarshalIndent(cfgMap, "", "\n")
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func parseConfigValues(cfg map[string]string) map[string]interface{} {
	cfgMap := make(map[string]interface{})
	for key, value := range cfg {
		tmpBool, err := strconv.ParseBool(value)
		if err == nil {
			cfgMap[key] = tmpBool
			continue
		}
		tmpInt, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			cfgMap[key] = tmpInt
			continue
		}
		tmpFloat, err := strconv.ParseFloat(value, 64)
		if err == nil {
			cfgMap[key] = tmpFloat
			continue
		}
		cfgMap[key] = value
	}
	return cfgMap
}

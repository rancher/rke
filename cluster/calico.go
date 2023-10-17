package cluster

import (
	"encoding/json"
	"fmt"

	calico "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func getSubMap(parent map[interface{}]interface{}, key string) interface{} {
	for k := range parent {
		switch k := k.(type) {
		case string:
			if k == key {
				return parent[k]
			}
		default:
			{
				return nil
			}
		}
	}
	return nil
}

func resolveFelixConfiguration(clusterFile string, felixConfigOut *calico.FelixConfigurationSpec) (string, error) {
	var clusterConfig map[string]interface{}

	err := yaml.Unmarshal([]byte(clusterFile), &clusterConfig)
	if err != nil {
		return clusterFile, fmt.Errorf("error unmarshalling clusterfile: %v", err)
	}

	network, ok := clusterConfig["network"].(map[interface{}]interface{})
	if network == nil || !ok {
		return clusterFile, nil
	}

	calicoNetworkProvider, ok := getSubMap(network, "calicoNetworkProvider").(map[interface{}]interface{})
	if calicoNetworkProvider == nil || !ok {
		return clusterFile, nil
	}

	felixConfiguration, ok := getSubMap(calicoNetworkProvider, "felixConfiguration").(map[interface{}]interface{})
	if felixConfiguration == nil || !ok {
		return clusterFile, nil
	}
	if ok && felixConfiguration != nil {
		delete(calicoNetworkProvider, "felixConfiguration")
		newClusterFile, err := yaml.Marshal(clusterConfig)
		if err != nil {
			return clusterFile, fmt.Errorf("error marshalling clusterfile: %v", err)
		}
		err = parseFelixConfiguration(felixConfiguration, felixConfigOut)
		return string(newClusterFile), err
	}

	return clusterFile, nil
}

func convertMap(in map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k := range in {
		switch t := k.(type) {
		case string:
			out[k.(string)] = in[k]
		default:
			{
				logrus.Errorf("wrong type found for %s: %s", k, t)
				return nil
			}
		}
	}
	return out
}

func parseFelixConfiguration(felixConfig map[interface{}]interface{}, felixConfigOut *calico.FelixConfigurationSpec) error {
	data, err := json.Marshal(convertMap(felixConfig))
	if err != nil {
		return fmt.Errorf("error marshalling FelixConfiguration: %v", err)
	}

	// calico.FelixConfiguration struct has json tags defined, using JSON Unmarshal instead of runtime serializer
	err = json.Unmarshal(data, felixConfigOut)
	if err != nil {
		return fmt.Errorf("error unmarshalling FelixConfiguration: %v", err)
	}

	return nil
}

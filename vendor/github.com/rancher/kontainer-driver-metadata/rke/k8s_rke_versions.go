package rke

import "github.com/rancher/types/apis/management.cattle.io/v3"

func loadRKEDefaultK8sVersions() map[string]string {
	return map[string]string {
		"0.2.3": "v1.13.5-rancher1-3",
	}
}

func loadK8sRKEVersions() map[string]v3.RKEVersions {
	return map[string]v3.RKEVersions{
		"v1.8.10-rancher1-1": {
			MaxRKEVersion: "0.2.2",
		},
		"v1.8.11-rancher1" : {
			MaxRKEVersion: "0.2.2",
		},
		"v1.8.11-rancher2-1" :{
			MaxRKEVersion: "0.2.2",
		},
		"v1.9.5-rancher1-1" :{
			MaxRKEVersion: "0.2.2",
		},
		"v1.9.7-rancher1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.9.7-rancher2-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.9.7-rancher2-2":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.0-rancher1-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.1-rancher1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.1-rancher2-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.3-rancher2-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.5-rancher1-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.5-rancher1-2":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.11-rancher1-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.10.12-rancher1-1":{
			MaxRKEVersion: "0.2.2",
		},
		"v1.11.1-rancher1-1":{

		},
		"v1.11.2-rancher1-1":{

		},
		"v1.11.2-rancher1-2":{

		},
		"v1.11.3-rancher1-1":{

		},
		"v1.11.5-rancher1-1":{

		},
		"v1.11.8-rancher1-1":{

		},
		"v1.11.6-rancher1-1":{

		},
		"v1.11.9-rancher1-1":{

		},
		"v1.11.9-rancher1-2":{

		},
		"v1.12.0-rancher1-1":{

		},
		"v1.12.1-rancher1-1":{

		},
		"v1.12.3-rancher1-1":{

		},
		"v1.12.4-rancher1-1":{

		},
		"v1.12.5-rancher1-1":{

		},
		"v1.12.5-rancher1-2":{

		},
		"v1.12.6-rancher1-1":{

		},
		"v1.12.6-rancher1-2":{

		},
		"v1.12.7-rancher1-1":{

		},
		"v1.12.7-rancher1-2":{

		},
		"v1.12.7-rancher1-3":{

		},
		"v1.13.1-rancher1-1":{

		},
		"v1.13.1-rancher1-2":{

		},
		"v1.13.4-rancher1-1":{

		},
		"v1.13.4-rancher1-2":{

		},
		"v1.13.5-rancher1-2":{

		},
		"v1.13.5-rancher1-1":{

		},
		"v1.13.5-rancher1-3":{

		},
		"v1.14.1-rancher1-1":{

		},
		"v1.14.1-rancher1-2":{

		},
	}
}

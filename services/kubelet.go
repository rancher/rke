package services

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
)

type Kubelet struct {
	Version             string `yaml:"version"`
	Image               string `yaml:"image"`
	ClusterDomain       string `yaml:"cluster_domain"`
	InfraContainerImage string `yaml:"infra_container_image"`
}

func runKubelet(host hosts.Host, kubeletService Kubelet, isMaster bool) error {
	imageCfg, hostCfg := buildKubeletConfig(host, kubeletService, isMaster)
	err := docker.DoRunContainer(imageCfg, hostCfg, KubeletContainerName, &host, WorkerRole)
	if err != nil {
		return err
	}
	return nil
}

func buildKubeletConfig(host hosts.Host, kubeletService Kubelet, isMaster bool) (*container.Config, *container.HostConfig) {
	imageCfg := &container.Config{
		Image: kubeletService.Image + ":" + kubeletService.Version,
		Cmd: []string{"/hyperkube",
			"kubelet",
			"--v=2",
			"--address=0.0.0.0",
			"--cluster-domain=" + kubeletService.ClusterDomain,
			"--hostname-override=" + host.Hostname,
			"--pod-infra-container-image=" + kubeletService.InfraContainerImage,
			"--cgroup-driver=cgroupfs",
			"--cgroups-per-qos=True",
			"--enforce-node-allocatable=",
			"--cluster-dns=10.233.0.3",
			"--network-plugin=cni",
			"--cni-conf-dir=/etc/cni/net.d",
			"--cni-bin-dir=/opt/cni/bin",
			"--resolv-conf=/etc/resolv.conf",
			"--allow-privileged=true",
			"--cloud-provider=",
			"--kubeconfig=" + pki.KubeNodeConfigPath,
			"--require-kubeconfig=True",
		},
	}
	if isMaster {
		imageCfg.Cmd = append(imageCfg.Cmd, "--register-with-taints=node-role.kubernetes.io/master=:NoSchedule")
		imageCfg.Cmd = append(imageCfg.Cmd, "--node-labels=node-role.kubernetes.io/master=true")
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
			"/etc/kubernetes:/etc/kubernetes",
			"/etc/cni:/etc/cni:ro",
			"/opt/cni:/opt/cni:ro",
			"/etc/resolv.conf:/etc/resolv.conf",
			"/sys:/sys:ro",
			"/var/lib/docker:/var/lib/docker:rw",
			"/var/lib/kubelet:/var/lib/kubelet:shared",
			"/var/run:/var/run:rw",
			"/run:/run",
			"/dev:/host/dev"},
		NetworkMode:   "host",
		PidMode:       "host",
		Privileged:    true,
		RestartPolicy: container.RestartPolicy{Name: "always"},
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
	}
	return imageCfg, hostCfg
}

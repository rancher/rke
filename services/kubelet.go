package services

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/rancher/rke/hosts"
)

type Kubelet struct {
	Version string `yaml:"version"`
	Image   string `yaml:"image"`
	ClusterDomain	string `yaml:"cluster_domain"`
	InfraContainerImage string `yaml:"infra_container_image"`
}

func runKubelet(host hosts.Host, masterHost hosts.Host, kubeletService Kubelet, isMaster bool) error {
	isRunning, err := IsContainerRunning(host, KubeletContainerName)
	if err != nil {
		return err
	}
	if isRunning {
		logrus.Infof("[WorkerPlane] Kubelet is already running on host [%s]", host.Hostname)
		return nil
	}
	err = runKubeletContainer(host, masterHost, kubeletService, isMaster)
	if err != nil {
		return err
	}
	return nil
}

func runKubeletContainer(host hosts.Host, masterHost hosts.Host, kubeletService Kubelet, isMaster bool) error {
	logrus.Debugf("[WorkerPlane] Pulling Kubelet Image on host [%s]", host.Hostname)
	err := PullImage(host, kubeletService.Image+":"+kubeletService.Version)
	if err != nil {
		return err
	}
	logrus.Infof("[WorkerPlane] Successfully pulled Kubelet image on host [%s]", host.Hostname)

	err = doRunKubelet(host, masterHost, kubeletService, isMaster)
	if err != nil {
		return err
	}
	logrus.Infof("[WorkerPlane] Successfully ran Kubelet container on host [%s]", host.Hostname)
	return nil
}

func doRunKubelet(host hosts.Host, masterHost hosts.Host, kubeletService Kubelet, isMaster bool) error {
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
			"--api-servers=http://" + masterHost.IP + ":8080/",
		},
	}
	if isMaster {
		imageCfg.Cmd = append(imageCfg.Cmd, "--register-with-taints=node-role.kubernetes.io/master=:NoSchedule")
		imageCfg.Cmd = append(imageCfg.Cmd, "--node-labels=node-role.kubernetes.io/master=true")
	}
	hostCfg := &container.HostConfig{
		Binds: []string{
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
	resp, err := host.DClient.ContainerCreate(context.Background(), imageCfg, hostCfg, nil, KubeletContainerName)
	if err != nil {
		return fmt.Errorf("Failed to create Kubelet container on host [%s]: %v", host.Hostname, err)
	}

	if err := host.DClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start Kubelet container on host [%s]: %v", host.Hostname, err)
	}
	logrus.Debugf("[WorkerPlane] Successfully started Kubelet container: %s", resp.ID)
	return nil
}

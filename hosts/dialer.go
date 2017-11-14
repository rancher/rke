package hosts

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type dialer struct {
	host *Host
}

const (
	DockerAPIVersion = "1.24"
)

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	sshAddr := d.host.IP + ":22"
	// Build SSH client configuration
	cfg, err := makeSSHConfig(d.host.User)
	if err != nil {
		logrus.Fatalf("Error configuring SSH: %v", err)
	}
	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", sshAddr, cfg)
	if err != nil {
		logrus.Fatalf("Error establishing SSH connection: %v", err)
	}
	if len(d.host.DockerSocket) == 0 {
		d.host.DockerSocket = "/var/run/docker.sock"
	}
	remote, err := conn.Dial("unix", d.host.DockerSocket)
	if err != nil {
		logrus.Fatalf("Error connecting to Docker socket on host [%s]: %v", d.host.AdvertisedHostname, err)
	}
	return remote, err
}

func (h *Host) TunnelUp() error {
	logrus.Infof("[ssh] Start tunnel for host [%s]", h.AdvertisedHostname)

	dialer := &dialer{
		host: h,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	// set Docker client
	var err error
	logrus.Debugf("Connecting to Docker API for host [%s]", h.AdvertisedHostname)
	h.DClient, err = client.NewClient("unix:///var/run/docker.sock", DockerAPIVersion, httpClient, nil)
	if err != nil {
		return fmt.Errorf("Can't connect to Docker for host [%s]: %v", h.AdvertisedHostname, err)
	}
	return nil
}

func privateKeyPath() string {
	return os.Getenv("HOME") + "/.ssh/id_rsa"
}

// Get private key for ssh authentication
func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

func makeSSHConfig(user string) (*ssh.ClientConfig, error) {
	key, err := parsePrivateKey(privateKeyPath())
	if err != nil {
		return nil, err
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &config, nil
}

package hosts

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type dialer struct {
	host   *Host
	signer ssh.Signer
}

const (
	DockerAPIVersion = "1.24"
)

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	sshAddr := d.host.Address + ":22"
	// Build SSH client configuration
	cfg, err := makeSSHConfig(d.host.User, d.signer)
	if err != nil {
		return nil, fmt.Errorf("Error configuring SSH: %v", err)
	}
	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", sshAddr, cfg)
	if err != nil {
		return nil, fmt.Errorf("Error establishing SSH connection: %v", err)
	}
	if len(d.host.DockerSocket) == 0 {
		d.host.DockerSocket = "/var/run/docker.sock"
	}
	remote, err := conn.Dial("unix", d.host.DockerSocket)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to Docker socket on host [%s]: %v", d.host.Address, err)
	}
	return remote, err
}

func (h *Host) TunnelUp(signer ssh.Signer) error {
	logrus.Infof("[ssh] Start tunnel for host [%s]", h.Address)

	dialer := &dialer{
		host:   h,
		signer: signer,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	// set Docker client
	var err error
	logrus.Debugf("Connecting to Docker API for host [%s]", h.Address)
	h.DClient, err = client.NewClient("unix:///var/run/docker.sock", DockerAPIVersion, httpClient, nil)
	if err != nil {
		return fmt.Errorf("Can't connect to Docker for host [%s]: %v", h.Address, err)
	}
	return nil
}

func ParsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

func ParsePrivateKeyWithPassPhrase(keyPath string, passphrase []byte) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKeyWithPassphrase(buff, passphrase)
}

func makeSSHConfig(user string, signer ssh.Signer) (*ssh.ClientConfig, error) {
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &config, nil
}

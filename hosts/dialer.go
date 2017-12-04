package hosts

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
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

func (h *Host) TunnelUp() error {
	logrus.Infof("[ssh] Start tunnel for host [%s]", h.Address)
	key, err := checkEncryptedKey(h.SSHKey, h.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("Failed to parse the private key: %v", err)
	}
	dialer := &dialer{
		host:   h,
		signer: key,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	// set Docker client
	logrus.Debugf("Connecting to Docker API for host [%s]", h.Address)
	h.DClient, err = client.NewClient("unix:///var/run/docker.sock", DockerAPIVersion, httpClient, nil)
	if err != nil {
		return fmt.Errorf("Can't connect to Docker for host [%s]: %v", h.Address, err)
	}
	return nil
}

func parsePrivateKey(keyBuff string) (ssh.Signer, error) {
	return ssh.ParsePrivateKey([]byte(keyBuff))
}

func parsePrivateKeyWithPassPhrase(keyBuff string, passphrase []byte) (ssh.Signer, error) {
	return ssh.ParsePrivateKeyWithPassphrase([]byte(keyBuff), passphrase)
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

func checkEncryptedKey(sshKey, sshKeyPath string) (ssh.Signer, error) {
	logrus.Debugf("[ssh] Checking private key")
	var err error
	var key ssh.Signer
	if len(sshKey) > 0 {
		key, err = parsePrivateKey(sshKey)
	} else {
		key, err = parsePrivateKey(privateKeyPath(sshKeyPath))
	}
	if err == nil {
		return key, nil
	}

	// parse encrypted key
	if strings.Contains(err.Error(), "decode encrypted private keys") {
		fmt.Printf("Passphrase for Private SSH Key: ")
		passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Printf("\n")
		if err != nil {
			return nil, err
		}
		if len(sshKey) > 0 {
			key, err = parsePrivateKeyWithPassPhrase(sshKey, passphrase)
		} else {
			key, err = parsePrivateKeyWithPassPhrase(privateKeyPath(sshKeyPath), passphrase)
		}
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

func privateKeyPath(sshKeyPath string) string {
	if sshKeyPath[:2] == "~/" {
		sshKeyPath = filepath.Join(os.Getenv("HOME"), sshKeyPath[2:])
	}
	buff, _ := ioutil.ReadFile(sshKeyPath)
	return string(buff)
}

package startup

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	clusteryml = `kubernetes_version: %K8SVERSION%
addon_job_timeout: 75
nodes:
- address: 10.10.10.100
  internal_address: 10.10.10.100
  role: [etcd, controlplane, worker]
  user: vagrant
  ssh_key_path: "~/.ssh/id_rsa"
- address: 10.10.10.101
  internal_address: 10.10.10.101
  role: [worker]
  user: vagrant
  ssh_key_path: "~/.ssh/id_rsa"
`
)

// TODO: Add remove cluster

// TODO: How to run tests in parallel for the different versions (how to solve the IPs)

type NodeError struct {
	Node string
	Cmd  string
	Err  error
}

func (ne *NodeError) Error() string {
	return fmt.Sprintf("failed creating cluster: %s: %v", ne.Cmd, ne.Err)
}

func (ne *NodeError) Unwrap() error {
	return ne.Err
}

func newNodeError(cmd, node string, err error) *NodeError {
	return &NodeError{
		Cmd:  cmd,
		Node: node,
		Err:  err,
	}
}

// CreateCluster creates a cluster with the given number of server and agent nodes
func CreateCluster(nodeOS string, serverCount, agentCount int) ([]string, []string, error) {

	logrus.Info("MANU - Creating cluster")
	serverNodeNames, agentNodeNames, err := vagrantUp(nodeOS, serverCount, agentCount)

	logrus.Info("MANU - vagrantUp done")

	// Fetch the supported versions
	versions, err := RunCommand("../../bin/rke --quiet config --all --list-version | sort -V")
	if err != nil {
		return nil, nil, err
	}
	logrus.Infof("MANU - These are the versions: %s", versions)

	// Pick one of the versions
	logrus.Infof("Picking version: %s", strings.Split(versions, "\n")[0])
	version := strings.Split(versions, "\n")[0]
	// Create the config
	clusterJSON := strings.ReplaceAll(clusteryml, "%K8SVERSION%", version)
	err = writeFile("cluster.yml", clusterJSON)
	if err != nil {
		return nil, nil, err
	}

	logrus.Info("MANU - Config file written")
	// Create the cluster using rke
	cmd := fmt.Sprintf("../../bin/rke --debug up --config cluster.yml")
	err = RunCommandStreamingOutput(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run command: %s: %v", cmd, err)
	}

	logrus.Info("rke command executed")

	return serverNodeNames, agentNodeNames, err
}

func writeFile(name string, content string) error {
	os.MkdirAll(filepath.Dir(name), 0755)
	err := os.WriteFile(name, []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
}

// vagrantUp brings up the nodes in parallel
func vagrantUp(nodeOS string, serverCount, agentCount int) ([]string, []string, error) {

	serverNodeNames, agentNodeNames, nodeEnvs := genNodeEnvs(nodeOS, serverCount, agentCount)

	var testOptions string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "E2E_") {
			testOptions += " " + env
		}
	}
	// Bring up the first server node
	cmd := fmt.Sprintf(`%s %s vagrant up %s &> vagrant.log`, nodeEnvs, testOptions, serverNodeNames[0])

	fmt.Println(cmd)
	if _, err := RunCommand(cmd); err != nil {
		return nil, nil, newNodeError(cmd, serverNodeNames[0], err)
	}
	// Bring up the rest of the nodes in parallel
	errg, _ := errgroup.WithContext(context.Background())
	for _, node := range append(serverNodeNames[1:], agentNodeNames...) {
		cmd := fmt.Sprintf(`%s %s vagrant up %s &>> vagrant.log`, nodeEnvs, testOptions, node)
		errg.Go(func() error {
			if _, err := RunCommand(cmd); err != nil {
				return newNodeError(cmd, node, err)
			}
			return nil
		})
		// We must wait a bit between provisioning nodes to avoid too many learners attempting to join the cluster
		if strings.Contains(node, "agent") {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(30 * time.Second)
		}
	}
	if err := errg.Wait(); err != nil {
		return nil, nil, err
	}

	return serverNodeNames, agentNodeNames, nil
}

// genNodeEnvs generates the node and testing environment variables for vagrant up
func genNodeEnvs(nodeOS string, serverCount, agentCount int) ([]string, []string, string) {
	serverNodeNames := make([]string, serverCount)
	for i := 0; i < serverCount; i++ {
		serverNodeNames[i] = "server-" + strconv.Itoa(i)
	}
	agentNodeNames := make([]string, agentCount)
	for i := 0; i < agentCount; i++ {
		agentNodeNames[i] = "agent-" + strconv.Itoa(i)
	}

	nodeRoles := strings.Join(serverNodeNames, " ") + " " + strings.Join(agentNodeNames, " ")
	nodeRoles = strings.TrimSpace(nodeRoles)

	nodeBoxes := strings.Repeat(nodeOS+" ", serverCount+agentCount)
	nodeBoxes = strings.TrimSpace(nodeBoxes)

	nodeEnvs := fmt.Sprintf(`E2E_NODE_ROLES="%s" E2E_NODE_BOXES="%s"`, nodeRoles, nodeBoxes)

	return serverNodeNames, agentNodeNames, nodeEnvs
}

func RunCommand(cmd string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	if kc, ok := os.LookupEnv("E2E_KUBECONFIG"); ok {
		c.Env = append(os.Environ(), "KUBECONFIG="+kc)
	}

	out, err := c.CombinedOutput()
	return string(out), err
}

func RunCommandStreamingOutput(cmd string) error {
	logrus.Info("MANU - Starting RunCommandStreamingOutput")
	c := exec.Command("bash", "-c", cmd)
	if kc, ok := os.LookupEnv("E2E_KUBECONFIG"); ok {
		c.Env = append(os.Environ(), "KUBECONFIG="+kc)
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	err = c.Start()
	if err != nil {
		return err
	}

	go copyOutput(stdout)
	go copyOutput(stderr)

	err = c.Wait()
	if err != nil {
		return err
	}

	return err
}

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logrus.Infof(scanner.Text())
	}
}

// GetVagrantLog returns the logs of on vagrant commands that initialize the nodes and provision K3s on each node.
// It also attempts to fetch the systemctl logs of K3s on nodes where the k3s.service failed.
func GetVagrantLog(cErr error) string {
	var nodeErr *NodeError
	nodeJournal := ""
	if errors.As(cErr, &nodeErr) {
		nodeJournal, _ = GetJournalLogs(nodeErr.Node)
		nodeJournal = "\nNode Journal Logs:\n" + nodeJournal
	}

	log, err := os.Open("vagrant.log")
	if err != nil {
		return err.Error()
	}
	bytes, err := io.ReadAll(log)
	if err != nil {
		return err.Error()
	}
	return string(bytes) + nodeJournal
}

func GetJournalLogs(node string) (string, error) {
	cmd := "journalctl -u k3s* --no-pager"
	return RunCmdOnNode(cmd, node)
}

// RunCmdOnNode executes a command from within the given node as sudo
func RunCmdOnNode(cmd string, nodename string) (string, error) {
	injectEnv := ""
	if _, ok := os.LookupEnv("E2E_GOCOVER"); ok && strings.HasPrefix(cmd, "k3s") {
		injectEnv = "GOCOVERDIR=/tmp/k3scov "
	}
	runcmd := "vagrant ssh " + nodename + " -c \"sudo " + injectEnv + cmd + "\""
	out, err := RunCommand(runcmd)
	if err != nil {
		return out, fmt.Errorf("failed to run command: %s on node %s: %s, %v", cmd, nodename, out, err)
	}
	return out, nil
}

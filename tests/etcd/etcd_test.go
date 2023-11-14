package etcd

import (
	"flag"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/rke/tests"
)

// Valid nodeOS: generic/ubuntu2004, opensuse/Leap-15.3.x86_64
var nodeOS = flag.String("nodeOS", "generic/ubuntu2004", "VM operating system")
var serverCount = flag.Int("serverCount", 1, "number of server nodes")
var agentCount = flag.Int("agentCount", 1, "number of agent nodes")
var ci = flag.Bool("ci", false, "running on CI")

func Test_E2EDualStack(t *testing.T) {
	flag.Parse()
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	RunSpecs(t, "Startup Test Suite", suiteConfig, reporterConfig)
}

var (
	serverNodeNames []string
	agentNodeNames  []string
)

//var _ = ReportAfterEach(e2e.GenReport)

//TODO: Fetch Kubeconfig

var _ = Describe("Verify RKE starts correctly", Ordered, func() {

	It("Starts up with no issues", func() {
		var err error
		serverNodeNames, agentNodeNames, err = tests.CreateCluster(*nodeOS, *serverCount, *agentCount)
		Expect(err).NotTo(HaveOccurred(), tests.GetVagrantLog(err))
		fmt.Println("CLUSTER CONFIG")
		fmt.Println("OS:", *nodeOS)
		fmt.Println("Server Nodes:", serverNodeNames)
		fmt.Println("Agent Nodes:", agentNodeNames)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Create an etcd snapshot", func() {
		err := tests.CreateSnapshot()
		Expect(err).NotTo(HaveOccurred())
	})

	It("Verify etcd snapshot was created", func() {
		output, err := tests.RunCmdOnNode("ls -l /opt/rke/etcd-snapshots/test_etcd_snapshot.zip", serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(output).Should(ContainSubstring("test_etcd_snapshot.zip"))
	})

	It("Deploy a dummy workload", func() {
		_, err := tests.DeployWorkload("test-daemonset.yaml")
		Expect(err).NotTo(HaveOccurred())
		cmd := "kubectl get daemonset -n default --kubeconfig=kube_config_cluster.yml"
		output, err := tests.RunCommand(cmd)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).Should(ContainSubstring("test-daemonset"))
	})

	It("Restore etcd snapshot", func() {
		err := tests.RestoreSnapshot()
		Expect(err).NotTo(HaveOccurred())
	})

	It("Verify etcd snapshot was restored", func() {
		cmd := "kubectl get daemonset -n default --kubeconfig=kube_config_cluster.yml"
		output, err := tests.RunCommand(cmd)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).ShouldNot(ContainSubstring("test-daemonset"))
	})

	It("Destroys with no issues", func() {
		Expect(tests.RemoveCluster()).To(Succeed())
	})
})

var failed bool
var _ = AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var _ = AfterSuite(func() {
	if failed && !*ci {
		fmt.Println("FAILED!")
	} else {
		Expect(tests.VagrantDestroy()).To(Succeed())
	}
})

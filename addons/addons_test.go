package addons

import (
	"bytes"
	"fmt"
	"testing"

	"k8s.io/api/batch/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	AddonSuffix    = "-deploy-job"
	FakeAddonName  = "example-addon"
	FakeNodeName   = "node1"
	FakeAddonImage = "example/example:latest"
)

func TestJobManifest(t *testing.T) {
	jobYaml, err := GetAddonsExecuteJob(FakeAddonName, FakeNodeName, FakeAddonImage)
	if err != nil {
		t.Fatalf("Failed to get addon execute job: %v", err)
	}
	job := v1.Job{}
	decoder := yamlutil.NewYAMLToJSONDecoder(bytes.NewReader([]byte(jobYaml)))
	err = decoder.Decode(&job)
	if err != nil {
		t.Fatalf("Failed To decode Job yaml: %v", err)
	}
	assertEqual(t, job.Name, FakeAddonName+AddonSuffix,
		fmt.Sprintf("Failed to verify job name [%s]", FakeAddonName+AddonSuffix))
	assertEqual(t, job.Spec.Template.Spec.NodeName, FakeNodeName,
		fmt.Sprintf("Failed to verify node name [%s] in the job", FakeNodeName))
	assertEqual(t, job.Spec.Template.Spec.Containers[0].Image, FakeAddonImage,
		fmt.Sprintf("Failed to verify container image [%s] in the job", FakeAddonImage))
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

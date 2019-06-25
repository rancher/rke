package hosts

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rancher/types/apis/management.cattle.io/v3"
	"k8s.io/api/core/v1"
)

var testTaint = v1.Taint{
	Key:    "key",
	Effect: v1.TaintEffectNoSchedule,
}

var testCases = []struct {
	toAdd       []string
	toDel       []string
	currentHost *Host
	specHost    *Host
}{
	{
		toAdd: []string{},
		toDel: []string{},
		currentHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{},
		},
		specHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{},
		},
	},
	{
		toAdd: []string{},
		toDel: []string{GetTaintString(testTaint)},
		currentHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{
					testTaint,
				},
			},
		},
		specHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{},
			},
		},
	},
	{
		toAdd: []string{GetTaintString(testTaint)},
		toDel: []string{},
		currentHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{},
			},
		},
		specHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{
					testTaint,
				},
			},
		},
	},
	{
		toAdd: []string{},
		toDel: []string{},
		currentHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{
					testTaint,
				},
			},
		},
		specHost: &Host{
			RKEConfigNode: v3.RKEConfigNode{
				Taints: []v1.Taint{
					testTaint,
				},
			},
		},
	},
}

func Test_HostDiffTaints(t *testing.T) {
	for _, testCase := range testCases {
		toAdd, toDel := GetHostDiffTaints(testCase.currentHost, testCase.specHost)
		assertEqual(t, toAdd, testCase.toAdd, fmt.Sprintf("taints to add are not as expected, test case: %v", testCase))
		assertEqual(t, toDel, testCase.toDel, fmt.Sprintf("taints to delete are not as expected, test case: %v", testCase))
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if reflect.DeepEqual(a, b) {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

package services

import (
	"fmt"
	"testing"

	"github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	TestSchedulerImage      = "rancher/k8s:latest"
	TestSchedulerVolumeBind = "/etc/kubernetes:/etc/kubernetes"
	TestSchedulerExtraArgs  = "--foo=bar"
)

func TestSchedulerConfig(t *testing.T) {

	schedulerService := v3.SchedulerService{}
	schedulerService.Image = TestSchedulerImage
	schedulerService.ExtraArgs = map[string]string{"foo": "bar"}

	imageCfg, hostCfg := buildSchedulerConfig(nil, schedulerService)
	// Test image and host config
	assertEqual(t, TestSchedulerImage, imageCfg.Image,
		fmt.Sprintf("Failed to verify [%s] as Scheduler Image", TestSchedulerImage))
	assertEqual(t, isStringInSlice(TestSchedulerVolumeBind, hostCfg.Binds), true,
		fmt.Sprintf("Failed to find [%s] in Scheduler Volume Binds", TestSchedulerVolumeBind))
	assertEqual(t, isStringInSlice(TestSchedulerExtraArgs, imageCfg.Entrypoint), true,
		fmt.Sprintf("Failed to find [%s] in Scheduler extra args", TestSchedulerExtraArgs))
	assertEqual(t, true, hostCfg.NetworkMode.IsHost(),
		"Failed to verify that Scheduler has host Network mode")
}

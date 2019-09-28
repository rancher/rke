package cluster

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rancher/rke/docker"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/util"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	TimeCheckContainer = "rke-time-checker"
)

func (c *Cluster) CheckClusterTime(ctx context.Context, currentCluster *Cluster) error {
	allHosts := hosts.GetUniqueHostList(c.EtcdHosts, c.ControlPlaneHosts, c.WorkerHosts)
	var errgrp errgroup.Group
	hostsQueue := util.GetObjectQueue(allHosts)
	for w := 0; w < WorkerThreads; w++ {
		errgrp.Go(func() error {
			var errList []error
			for host := range hostsQueue {
				err := c.checkTimeDiffLocalVsNode(ctx, host.(*hosts.Host), c.SystemImages.Alpine, c.PrivateRegistriesMap)
				if err != nil {
					errList = append(errList, err)
				}
			}
			return util.ErrList(errList)
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) checkTimeDiffLocalVsNode(ctx context.Context, host *hosts.Host, image string, prsMap map[string]v3.PrivateRegistry) error {
	imageCfg := &container.Config{
		Image: image,
		Cmd: []string{
			"date",
			"+%s",
		},
	}
	hostCfg := &container.HostConfig{
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
	}
	if err := docker.DoRemoveContainer(ctx, host.DClient, TimeCheckContainer, host.Address); err != nil {
		return err
	}
	// Closest point to actual running the container
	startEpoch := util.GetEpochTime()
	logrus.Debugf("[time] Time on local host: [%d] [%s]", startEpoch, time.Unix(startEpoch, 0).UTC())
	if err := docker.DoRunOnetimeContainer(ctx, host.DClient, imageCfg, hostCfg, TimeCheckContainer, host.Address, "time", prsMap); err != nil {
		return err
	}
	endEpoch := util.GetEpochTime()

	containerStderrLog, containerStdoutLog, logsErr := docker.GetContainerLogsStdoutStderr(ctx, host.DClient, TimeCheckContainer, "all", true)
	if logsErr != nil {
		log.Warnf(ctx, "[time] Failed to get time check logs: %v, logsErr: %v", containerStderrLog, logsErr)
	}
	logrus.Debugf("[time] containerStdoutLog [%s] on host: %s", strings.TrimSuffix(containerStdoutLog, "\n"), host.Address)

	if err := docker.RemoveContainer(ctx, host.DClient, host.Address, TimeCheckContainer); err != nil {
		return err
	}
	strippedContainerStdoutLog := strings.TrimSuffix(containerStdoutLog, "\n")
	remoteEpoch, _ := strconv.Atoi(strippedContainerStdoutLog)
	remoteEpochInt64 := int64(remoteEpoch)
	logrus.Debugf("[time] Time on remote host [%s]: [%d] [%s]", host.Address, remoteEpochInt64, time.Unix(remoteEpochInt64, 0).UTC())
	hostname, _ := os.Hostname()
	// If local time is later than remote, certificate will not be valid yet
	if startEpoch > remoteEpochInt64 {
		return fmt.Errorf("[time] time [%s] on host [%s] is earlier than time [%s] on host [%s] used for provisioning. Please correct time on the host(s) or use time synchronization software", time.Unix(remoteEpochInt64, 0).UTC(), host.Address, time.Unix(startEpoch, 0).UTC(), hostname)
	}
	// If remote time is later than local time, it won't break directly but is still an issue
	if remoteEpochInt64 > endEpoch {
		logrus.Warningf("[time] Time [%s] on host [%s] is later than time [%s] on host [%s] used for provisioning", time.Unix(remoteEpochInt64, 0).UTC(), host.Address, time.Unix(endEpoch, 0).UTC(), hostname)
	}
	return nil
}

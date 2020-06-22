package shepherd_unittest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type Platform struct {
	// Contains the response string for a given type of a request
	DockerAppMetrics     string
	DockerClusterMetrics string
	// Cloudlet-level test data
	CloudletMetrics  string
	VmAppInstMetrics string
	// TODO - add Prometheus/nginx strings here EDGECLOUD-1252
}

func (s *Platform) GetType() string {
	return "unit test"
}

func (s *Platform) Init(ctx context.Context, pc *platform.PlatformConfig) error {
	return nil
}

func (s *Platform) GetClusterIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, error) {
	return "localhost", nil
}

func (s *Platform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return &UTClient{pf: s}, nil
}

// Query local system for the resource usage
func (s *Platform) GetPlatformStats(ctx context.Context) (shepherd_common.CloudletMetrics, error) {
	metrics := shepherd_common.CloudletMetrics{}
	if err := json.Unmarshal([]byte(s.CloudletMetrics), &metrics); err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to marshal unit test metrics", "stats", s.CloudletMetrics, "err", err.Error())
		return metrics, err
	}
	return metrics, nil
}

func (s *Platform) GetVmStats(ctx context.Context, key *edgeproto.AppInstKey) (shepherd_common.AppMetrics, error) {
	metrics := shepherd_common.AppMetrics{}
	if err := json.Unmarshal([]byte(s.VmAppInstMetrics), &metrics); err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to marshal unit test metrics", "stats", s.VmAppInstMetrics, "err", err.Error())
		return metrics, err
	}
	ts, _ := types.TimestampProto(time.Now())
	metrics.CpuTS, metrics.MemTS, metrics.DiskTS, metrics.NetSentTS, metrics.NetRecvTS = ts, ts, ts, ts, ts
	return metrics, nil
}

func (s *Platform) GetMetricsCollectInterval() time.Duration {
	return 0
}

// UTClient hijacks a set of commands and returns predetermined output
// For all other commands it just calls pc.LocalClient equivalents
type UTClient struct {
	pc.LocalClient
	pf *Platform
}

func (s *UTClient) Output(command string) (string, error) {
	out, err := s.getUTData(command)
	if err != nil {
		return s.LocalClient.Output(command)
	}
	return out, nil
}

func (s *UTClient) OutputWithTimeout(command string, timeout time.Duration) (string, error) {
	out, err := s.getUTData(command)
	if err != nil {
		return s.LocalClient.OutputWithTimeout(command, timeout)
	}
	return out, nil
}

func (s *UTClient) getUTData(command string) (string, error) {
	// docker stats unit test
	if strings.Contains(command, "docker stats ") {
		// take the json with line breaks and compact it, as that's what the command expects
		return s.pf.DockerAppMetrics, nil
	} else if strings.Contains(command, shepherd_common.ResTrackerCmd) {
		return s.pf.DockerClusterMetrics, nil
	}
	// nginx-stats and envoy-stats unit test
	// "docker exec containername curl http://url"
	if strings.Contains(command, "docker exec") && strings.Contains(command, "curl") {
		split := strings.SplitN(command, " ", 4)
		if len(split) == 4 {
			return s.LocalClient.Output(split[3])
		}
	}
	// "docker exec containername echo text"
	if strings.Contains(command, "docker exec") && strings.Contains(command, "echo") {
		split := strings.SplitN(command, " ", 4)
		if len(split) == 4 {
			return s.LocalClient.Output(split[3])
		}
	}
	return "", fmt.Errorf("No UT Data found")
}

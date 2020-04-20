package shepherd_edgebox

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type Platform struct {
	pf           dind.Platform
	SharedClient ssh.Client
}

func (s *Platform) GetType() string {
	return "edgebox"
}

func (s *Platform) Init(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName, vaultAddr string, vars map[string]string) error {
	s.SharedClient, _ = s.pf.GetNodePlatformClient(ctx, nil)
	return nil
}

func (s *Platform) GetClusterIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, error) {
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		return clusterInst.AllocatedIp, nil
	}
	return "localhost", nil
}

func (s *Platform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return s.SharedClient, nil
}

func (s *Platform) GetMetricsCollectInterval() time.Duration {
	return 0
}

func (s *Platform) GetPlatformStats(ctx context.Context) (shepherd_common.CloudletMetrics, error) {
	cloudletMetric := shepherd_common.CloudletMetrics{}
	cloudletMetric.ComputeTS, _ = types.TimestampProto(time.Now())
	cloudletMetric.NetworkTS = cloudletMetric.ComputeTS

	cpu, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return cloudletMetric, err
	}
	cloudletMetric.VCpuMax = uint64(cpu)
	// for local we are using all VCpus
	cloudletMetric.VCpuUsed = uint64(cpu)
	m, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return cloudletMetric, err
	}
	// Limits for Mem is in MBs
	cloudletMetric.MemUsed = uint64(m.Used >> 20)
	cloudletMetric.MemMax = uint64(m.Total >> 20)
	d, err := disk.Usage("/")
	if err != nil {
		return cloudletMetric, err
	}
	// Limits for Disk is in GBs
	cloudletMetric.DiskMax = d.Total >> 30
	cloudletMetric.DiskUsed = d.Used >> 30

	n, err := net.IOCounters(false)
	if err != nil {
		return cloudletMetric, err
	}
	cloudletMetric.NetRecv = n[0].BytesRecv >> 10
	cloudletMetric.NetSent = n[0].BytesSent >> 10
	return cloudletMetric, nil
}

func (s *Platform) GetVmStats(ctx context.Context, key *edgeproto.AppInstKey) (shepherd_common.AppMetrics, error) {
	return shepherd_common.AppMetrics{}, fmt.Errorf("VM on DIND is unsupported")
}

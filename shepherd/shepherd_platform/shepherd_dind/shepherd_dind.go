package shepherd_dind

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type Platform struct {
	pf           dind.Platform
	SharedClient pc.PlatformClient
}

func (s *Platform) GetType() string {
	return "dind"
}

func (s *Platform) Init(ctx context.Context, key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	s.SharedClient, _ = s.pf.GetPlatformClient(ctx, nil)
	return nil
}

func (s *Platform) GetClusterIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, error) {
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		return clusterInst.AllocatedIp, nil
	}
	return "localhost", nil
}

func (s *Platform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.SharedClient, nil
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
	cloudletMetric.MemUsed = uint64(m.Used)
	cloudletMetric.MemMax = uint64(m.Total)
	d, err := disk.Usage("/")
	if err != nil {
		return cloudletMetric, err
	}
	cloudletMetric.DiskMax = d.Total
	cloudletMetric.DiskUsed = d.Used

	n, err := net.IOCounters(false)
	if err != nil {
		return cloudletMetric, err
	}
	cloudletMetric.NetRecv = n[0].BytesRecv
	cloudletMetric.NetSent = n[0].BytesSent
	return cloudletMetric, nil
}

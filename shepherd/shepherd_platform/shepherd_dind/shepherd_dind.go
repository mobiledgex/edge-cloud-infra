package shepherd_dind

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	ctx          context.Context
	pf           dind.Platform
	SharedClient pc.PlatformClient
}

func (s *Platform) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Platform) GetType() string {
	return "dind"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	s.SharedClient, _ = s.pf.GetPlatformClient(nil)
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		return clusterInst.AllocatedIp, nil
	}
	return "localhost", nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.SharedClient, nil
}

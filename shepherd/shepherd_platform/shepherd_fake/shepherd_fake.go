package shepherd_fake

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	ctx context.Context
}

func (s *Platform) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Platform) GetType() string {
	return "fake"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	return "localhost", nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

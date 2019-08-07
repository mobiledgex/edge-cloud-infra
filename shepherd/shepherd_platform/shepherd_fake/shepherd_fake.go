package shepherd_fake

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "fakecloudlet"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	//start the fake prom server for e2e tests
	if l, err := SetupFakeProm(); err != nil {
		return "", err
	} else {
		go RunFakeProm(l)
	}
	addr := "127.0.0.1"
	return addr, nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

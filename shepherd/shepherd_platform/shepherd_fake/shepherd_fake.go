package shepherd_fake

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	fakePromOwner bool
}

func (s *Platform) GetType() string {
	return "fakecloudlet"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	s.fakePromOwner = false
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	if s.fakePromOwner == false {
		//start the fake prom server for e2e tests
		if l, err := SetupFakeProm(); err != nil {
			return "", err
		} else {
			s.fakePromOwner = true
			RunFakeProm(l)
		}
	}
	return "127.0.0.1", nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

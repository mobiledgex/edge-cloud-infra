package dind

import (
	//"errors"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "dind"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	//dont need to do anything here?
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		return clusterInst.AllocatedIp, nil
	}
	return "localhost", nil
}

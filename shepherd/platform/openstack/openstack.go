package openstack

import (
	//"errors"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "openstack"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	//somehow connect through to rootlb, can assume its already been setup by crm's openstack platform init
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	return mexos.GetMasterIP(clusterInst, "")
}

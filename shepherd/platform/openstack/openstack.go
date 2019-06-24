package openstack

import (
	//"errors"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	rootLbName   string
	SharedClient pc.PlatformClient
	pf           openstack.Platform
}

func (s *Platform) GetType() string {
	return "openstack"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	//get the platform client so we can ssh in to make curl commands to the prometheus apps
	var err error
	if err = mexos.InitOpenstackProps(); err != nil {
		return err
	}
	//need to have a separate one for dedicated rootlbs, see openstack.go line 111,
	s.rootLbName = cloudcommon.GetRootLBFQDN(key)
	s.SharedClient, err = s.pf.GetPlatformClientRootLB(s.rootLbName)
	if err != nil {
		return err
	}
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	return mexos.GetMasterIP(clusterInst)
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	if clusterInst != nil && clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLb := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey)
		pc, err := s.pf.GetPlatformClientRootLB(rootLb)
		return pc, err //this is not ok, pc will go away when the func returns
	} else {
		return s.SharedClient, nil
	}
}

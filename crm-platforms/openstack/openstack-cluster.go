package openstack

import (
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *Platform) CreateCluster(clusterInst *edgeproto.ClusterInst, flavor *edgeproto.ClusterFlavor) error {
	lbName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IpAccessDedicated {
		lbName = mexos.GetDedicatedRootLBNameForCluster(clusterInst, s.cloudletKey)
	}
	return mexos.CreateCluster(lbName, clusterInst, flavor)
}

func (s *Platform) DeleteCluster(clusterInst *edgeproto.ClusterInst) error {
	lbName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IpAccessDedicated {
		lbName = mexos.GetDedicatedRootLBNameForCluster(clusterInst, s.cloudletKey)
	}
	return mexos.DeleteCluster(lbName, clusterInst)
}

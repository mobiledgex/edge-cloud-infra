package openstack

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	lbName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
	}
	return mexos.UpdateCluster(s.ctx, lbName, clusterInst, updateCallback)
}

func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	lbName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
	}
	//find the flavor and check the disk size
	for _, flavor := range s.flavorList {
		if flavor.Name == clusterInst.NodeFlavor && flavor.Disk < MINIMUM_DISK_SIZE {
			return fmt.Errorf("Insufficient disk size, please specify a flavor with at least %dgb", MINIMUM_DISK_SIZE)
		}
	}

	//adjust the timeout just a bit to give some buffer for the API exchange and also sleep loops
	timeout -= time.Minute

	return mexos.CreateCluster(s.ctx, lbName, clusterInst, updateCallback, timeout)
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	lbName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
	}
	return mexos.DeleteCluster(s.ctx, lbName, clusterInst)
}

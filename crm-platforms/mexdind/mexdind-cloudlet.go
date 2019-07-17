package mexdind

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create cloudlet for mexdind")
	s.generic.CreateCloudlet(cloudlet, flavor, updateCallback)
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete cloudlet for mexdind")
	s.generic.DeleteCloudlet(cloudlet)
	return nil
}

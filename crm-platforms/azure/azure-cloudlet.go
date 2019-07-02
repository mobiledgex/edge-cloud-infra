package azure

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreatePlatform(pf *edgeproto.Platform, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create platform for azure")
	return nil
}

func (s *Platform) DeletePlatform(pf *edgeproto.Platform) error {
	log.DebugLog(log.DebugLevelMexos, "delete platform for azure")
	return nil
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pf *edgeproto.Platform, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create cloudlet for azure")
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet, pf *edgeproto.Platform) error {
	log.DebugLog(log.DebugLevelMexos, "delete cloudlet for azure")
	return nil
}

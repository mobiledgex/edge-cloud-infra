package mexdind

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create cloudlet for mexdind")
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete cloudlet for mexdind")
	return nil
}

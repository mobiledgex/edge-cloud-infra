package fakeinfra

import (
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
	fake.Platform
}

func (s *Platform) GetType() string {
	return "fakeinfra"
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create fake cloudlet", "key", cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(cloudlet, pfConfig)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fake cloudlet create failed", "err", err)
		return err
	}
	return intprocess.StartShepherdService(cloudlet, pfConfig)
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete fake Cloudlet", "key", cloudlet.Key)
	err := cloudcommon.StopCRMService(cloudlet)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fake cloudlet delete failed", "err", err)
		return err
	}
	return intprocess.StopShepherdService(cloudlet)
}

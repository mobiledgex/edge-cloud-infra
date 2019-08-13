package fakeinfra

import (
	"fmt"
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "fakeinfra"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "running in fakeinfra cloudlet mode")
	updateCallback(edgeproto.UpdateTask, "Done intializing fakeinfra platform")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	info.OsMaxRam = 500
	info.OsMaxVcores = 50
	info.OsMaxVolGb = 5000
	info.Flavors = []*edgeproto.FlavorInfo{
		&edgeproto.FlavorInfo{
			Name:  "flavor1",
			Vcpus: uint64(10),
			Ram:   uint64(101024),
			Disk:  uint64(500),
		},
	}
	return nil
}

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster not supported for fakeinfra cloudlets")
}
func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	updateCallback(edgeproto.UpdateTask, "First Create Task")
	updateCallback(edgeproto.UpdateTask, "Second Create Task")
	log.DebugLog(log.DebugLevelMexos, "fakeinfra ClusterInst ready")
	return nil
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	log.DebugLog(log.DebugLevelMexos, "fakeinfra ClusterInst deleted")
	return nil
}

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Creating App Inst")
	log.DebugLog(log.DebugLevelMexos, "fakeinfra AppInst ready")
	return nil
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "fakeinfra AppInst deleted")
	return nil
}

func (s *Platform) UpdateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "fakeinfra appInst updated")
	return nil
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return &edgeproto.AppInstRuntime{}, nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return req.Command, nil
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create fakeinfra cloudlet", "key", cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(cloudlet, pfConfig)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fakeinfra cloudlet create failed", "err", err)
		return err
	}
	return intprocess.StartShepherdService(cloudlet, pfConfig)
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete fakeinfra Cloudlet", "key", cloudlet.Key)
	err := cloudcommon.StopCRMService(cloudlet)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fakeinfra cloudlet delete failed", "err", err)
		return err
	}
	return intprocess.StopShepherdService(cloudlet)
}

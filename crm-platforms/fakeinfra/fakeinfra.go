package fakeinfra

import (
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
	generic fake.Platform
}

func (s *Platform) GetType() string {
	return "fakeinfra"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig) error {
	log.DebugLog(log.DebugLevelMexos, "running in fakeinfra cloudlet mode")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return s.generic.GatherCloudletInfo(info)
}

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.generic.UpdateClusterInst(clusterInst, updateCallback)
}
func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	return s.generic.CreateClusterInst(clusterInst, updateCallback, timeout)
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	return s.generic.DeleteClusterInst(clusterInst)
}

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.generic.CreateAppInst(clusterInst, app, appInst, flavor, updateCallback)
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	return s.generic.DeleteAppInst(clusterInst, app, appInst)
}

func (s *Platform) UpdateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.generic.UpdateAppInst(clusterInst, app, appInst, updateCallback)
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return s.generic.GetAppInstRuntime(clusterInst, app, appInst)
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.generic.GetPlatformClient(clusterInst)
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return s.generic.GetContainerCommand(clusterInst, app, appInst, req)
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	err := s.generic.CreateCloudlet(cloudlet, pfConfig, flavor, updateCallback)
	if err != nil {
		return err
	}
	return intprocess.StartShepherdService(cloudlet, pfConfig)
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	err := s.generic.DeleteCloudlet(cloudlet)
	if err != nil {
		return err
	}
	return intprocess.StopShepherdService(cloudlet)
}

package fakeinfra

import (
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	fake fake.Platform
}

func (s *Platform) GetType() string {
	return "fakeinfra"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.fake.Init(platformConfig, updateCallback)
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return s.fake.GatherCloudletInfo(info)
}

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.fake.UpdateClusterInst(clusterInst, updateCallback)
}
func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	return s.fake.CreateClusterInst(clusterInst, updateCallback, timeout)
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	return s.fake.DeleteClusterInst(clusterInst)
}

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.fake.CreateAppInst(clusterInst, app, appInst, flavor, updateCallback)
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	return s.fake.DeleteAppInst(clusterInst, app, appInst)
}

func (s *Platform) UpdateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.fake.UpdateAppInst(clusterInst, app, appInst, updateCallback)
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return s.fake.GetAppInstRuntime(clusterInst, app, appInst)
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.fake.GetPlatformClient(clusterInst)
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return s.GetContainerCommand(clusterInst, app, appInst, req)
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	if err := s.fake.CreateCloudlet(cloudlet, pfConfig, flavor, updateCallback); err != nil {
		return err
	}
	return intprocess.StartShepherdService(cloudlet, pfConfig)
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	if err := s.fake.DeleteCloudlet(cloudlet); err != nil {
		return err
	}
	return intprocess.StopShepherdService(cloudlet)
}

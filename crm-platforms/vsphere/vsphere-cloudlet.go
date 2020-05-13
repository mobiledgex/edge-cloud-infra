package vsphere

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var flavorLock sync.Mutex
var clusterLock sync.Mutex
var appLock sync.Mutex

var flavors []*edgeproto.FlavorInfo

func (v *VSpherePlatform) VerifyApiEndpoint(ctx context.Context, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	// Verify if Openstack API Endpoint is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if VCenter API Endpoint is reachable")
	host, portstr, err := v.GetVCenterAddress()
	if err != nil {
		return err
	}
	_, err = client.Output(
		fmt.Sprintf(
			"nc %s %s -w 5", host, portstr,
		),
	)
	if err != nil {
		return fmt.Errorf("unable to reach Vcenter Address: %s", host)
	}
	return nil
}

func (o *VSpherePlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("SaveCloudletAccessVars not implemented for vsphere")
}

func (v *VSpherePlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	// we don't currently have the ability to download and setup the template, but we will verify it is there
	imgPath := v.vmProperties.GetCloudletOSImage()
	_, err := v.GetServerDetail(ctx, imgPath)
	if err != nil {
		return "", fmt.Errorf("Vsphere base image template not present: %s", imgPath)
	}
	return imgPath, nil
}

func (v *VSpherePlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	flavorLock.Lock()
	defer flavorLock.Unlock()
	// we just send the controller back the same list of flavors it gave us, because VSphere has no flavor concept
	return flavors, nil
}

func (v *VSpherePlatform) SyncControllerFlavors(ctx context.Context, controllerData *platform.ControllerData) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerFlavors")
	flavorLock.Lock()
	defer flavorLock.Unlock()
	flavorkeys := make(map[edgeproto.FlavorKey]context.Context)
	controllerData.FlavorCache.GetAllKeys(ctx, flavorkeys)
	for k := range flavorkeys {
		log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerFlavors found flavor", "key", k)
		var flav edgeproto.Flavor
		if controllerData.FlavorCache.Get(&k, &flav) {
			var flavInfo edgeproto.FlavorInfo
			flavInfo.Name = flav.Key.Name
			flavInfo.Disk = flav.Disk
			flavInfo.Ram = flav.Ram
			flavInfo.Vcpus = flav.Vcpus
			flavors = append(flavors, &flavInfo)
		} else {
			return fmt.Errorf("fail to fetch flavor %s", k)
		}
	}
	return nil
}

/*
func (v *VSpherePlatform) SyncControllerClusters(ctx context.Context, controllerData *platform.ControllerData) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerClusters")

	clusterKeys := make(map[edgeproto.ClusterInstKey]context.Context)
	controllerData.ClusterInstCache.GetAllKeys(ctx, clusterKeys)
	for k := range clusterKeys {
		var clus edgeproto.ClusterInst
		if controllerData.ClusterCache.Get(&k, &clus) {
            v.SyncCluster(ctx, &clus)
		} else {
			return fmt.Errorf("fail to fetch cluster %s", k)
		}
	}
	return nil
}*/

func (v *VSpherePlatform) SyncControllerData(ctx context.Context, controllerData *platform.ControllerData, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "XXXXXXX SyncControllerData")
	err := v.SyncControllerFlavors(ctx, controllerData)
	if err != nil {
		return err
	}
	return nil
}

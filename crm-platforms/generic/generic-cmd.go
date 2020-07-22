package generic

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *GenericPlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	// we just send the controller back the same list of flavors it gave us,
	// because currently generic platform has no flavors
	// Make sure each flavor is at least a minimum size to run the platform
	// FIXME: Copy of Vsphere code
	var flavors []*edgeproto.FlavorInfo
	if s.caches == nil {
		log.WarnLog("flavor cache is nil")
		return nil, fmt.Errorf("Flavor cache is nil")
	}
	flavorkeys := make(map[edgeproto.FlavorKey]struct{})
	s.caches.FlavorCache.GetAllKeys(ctx, func(k *edgeproto.FlavorKey, modRev int64) {
		flavorkeys[*k] = struct{}{}
	})
	for k := range flavorkeys {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList found flavor", "key", k)
		var flav edgeproto.Flavor
		if s.caches.FlavorCache.Get(&k, &flav) {
			var flavInfo edgeproto.FlavorInfo
			flavInfo.Name = flav.Key.Name
			if flav.Ram >= vmlayer.MINIMUM_RAM_SIZE {
				flavInfo.Ram = flav.Ram
			} else {
				flavInfo.Ram = vmlayer.MINIMUM_RAM_SIZE
			}
			if flav.Vcpus >= vmlayer.MINIMUM_VCPUS {
				flavInfo.Vcpus = flav.Vcpus
			} else {
				flavInfo.Vcpus = vmlayer.MINIMUM_VCPUS
			}
			if flav.Disk >= vmlayer.MINIMUM_DISK_SIZE {
				flavInfo.Disk = flav.Disk
			} else {
				flavInfo.Disk = vmlayer.MINIMUM_DISK_SIZE
			}
			flavors = append(flavors, &flavInfo)
		} else {
			return nil, fmt.Errorf("fail to fetch flavor %s", k)
		}
	}

	// add the default platform flavor as well
	var rlbFlav edgeproto.Flavor
	err := s.VMProperties.GetCloudletSharedRootLBFlavor(&rlbFlav)
	if err != nil {
		return nil, err
	}
	rootlbFlavorInfo := edgeproto.FlavorInfo{
		Name:  "mex-rootlb-flavor",
		Vcpus: rlbFlav.Vcpus,
		Ram:   rlbFlav.Ram,
		Disk:  rlbFlav.Disk,
	}
	flavors = append(flavors, &rootlbFlavorInfo)
	return flavors, nil
}

func (s *GenericPlatform) SetPowerState(ctx context.Context, serverName, serverAction string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetPowerState not supported")
	return nil
}

func (s *GenericPlatform) GetCloudletImageSuffix(ctx context.Context) string {
	return ".qcow2"
}

func (s *GenericPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddCloudletImageIfNotPresent not supported")
	imgPath := vmlayer.GetCloudletVMImagePath(imgPathPrefix, imgVersion, s.GetCloudletImageSuffix(ctx))

	// Fetch platform base image name
	pfImageName, err := cloudcommon.GetFileName(imgPath)
	if err != nil {
		return "", err
	}
	return pfImageName, nil
}

func (s *GenericPlatform) DeleteImage(ctx context.Context, folder, imageName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteImage not supported")
	return nil
}

func (s *GenericPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer not supported")
	return nil
}

func (s *GenericPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName string, portName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer not supported")
	return nil
}

func (s *GenericPlatform) GetNetworkList(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

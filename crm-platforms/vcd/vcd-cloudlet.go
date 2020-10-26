package vcd

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Cloudlet related operations

// The CreateImageFromUrl in vsphere could be placed into common-utils, if we separate the
// fetch, from v.ImportImage, allowing platform to do the needful.
// Then have a look at that AddCloudImageIfNotPresent for potential refactor. xxx uses GetServerDetail is that standard?

// PI
// Is this only for appInst images? Or our cloudlet template too?
func (v *VcdPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "PI AddCloudletImageIfNotPresent  TBI ", "imgPathPrefix", imgPathPrefix, "ImgVersion", imgVersion)
	// how about just returning our ubuntu18.04 image here? For now
	fmt.Printf("AddCloudletImageIfNotPresent-i-TBI\n")

	//	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, imageUrl, md5Sum)

	return "", nil
}

// PI   Security calls this to save what it gets from vault?
func (v *VcdPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	// ENOTIMP
	return nil
}

// This appears to only deal with non-eixstant flavors in vmware world
func (v *VcdPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo ")
	// get the flavor list
	// whatelse do we need here?
	var err error
	info.Flavors, err = v.GetFlavorList(ctx)
	if err != nil {
		fmt.Printf("\n\nGatherCloudlentInfo-E-GetFlavorList err: %s\n", err.Error())
		return err
	}
	return nil
}

// PI  why is this needed

func (v *VcdPlatform) GetCloudletImageSuffix(ctx context.Context) string {
	fmt.Printf("GetCloudletImageSuffix TBI\n")
	// needed? Follow convention
	return "-vcd.qcow2"
}

//
// XXX Its the result of the heat stack apply in openstack, not supported by vmpool, and TBI as the OVF file for the cloudlet in vsphere...
// Not sure we'll have a single OVF file for our entire cloudlet? (with muliple clusterInsts?)
func (v *VcdPlatform) GetCloudletManifest(ctx context.Context, name, cloudletImagePath string, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest name: %s imagePath? %s ", name, cloudletImagePath)
	return "", nil

}

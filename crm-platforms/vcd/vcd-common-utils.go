package vcd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"path/filepath"
)

// This functions could be shared across non-openstack platforms, and were basicly hijacked from vsphere
// While in VcdPlatform currently, if approved, move into vm-common-mumble package
func (v *VcdPlatform) GetFlavor(ctx context.Context, flavorName string) (*edgeproto.FlavorInfo, error) {
	flavs, err := v.GetFlavorList(ctx)
	if err != nil {
		return nil, err
	}
	for _, f := range flavs {
		if f.Name == flavorName {
			return f, nil
		}
	}
	return nil, fmt.Errorf("no flavor found named: %s", flavorName)
}

func (v *VcdPlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList")
	// we just send the controller back the same list of flavors it gave us, because VSphere has no flavor concept.
	// Make sure each flavor is at least a minimum size to run the platform

	var flavors []*edgeproto.FlavorInfo
	if v.caches == nil {
		log.WarnLog("flavor cache is nil")
		return nil, fmt.Errorf("Flavor cache is nil")
	}
	flavorkeys := make(map[edgeproto.FlavorKey]struct{})
	v.caches.FlavorCache.GetAllKeys(ctx, func(k *edgeproto.FlavorKey, modRev int64) {

		flavorkeys[*k] = struct{}{}
	})

	for k := range flavorkeys {
		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList found flavor", "key", k)
		}
		var flav edgeproto.Flavor
		if v.caches.FlavorCache.Get(&k, &flav) {
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
	err := v.vmProperties.GetCloudletSharedRootLBFlavor(&rlbFlav)
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

// Here, we're implementing AddCloudletImageIfNotPresent
// We'll lift CreateImageFromUrl, but leave out the return v.ImportImage
// Really, it could have the platform passed in, and we could return p.ImportImage
// but not yet, just let the caller do the ImportImage
//
//CreateImageFromUrl downloads image from URL and then imports to the datastore
//func (v *VSpherePlatform) CreateImageFromUrl(ctx context.Context, imageName, imageUrl, md5Sum string) error {

var qcowConvertTimeout = 5 * time.Minute

func (v *VcdPlatform) CreateImageFromUrl(ctx context.Context, imageName, imageUrl, md5Sum string) (string, error) {

	// dne	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, imageUrl, md5Sum)
	filePath := ""
	defer func() {
		// Stale file might be present if download fails/succeeds, deleting it
		if delerr := infracommon.DeleteFile(filePath); delerr != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "filePath", filePath)
		}
	}()

	vmdkFile, err := v.ConvertQcowToVmdk(ctx, filePath, vmlayer.MINIMUM_DISK_SIZE)
	if err != nil {
		return "", err
	}
	return vmdkFile, nil
	// return v.ImportImage(ctx, imageName, vmdkFile)
}

func (v *VcdPlatform) ConvertQcowToVmdk(ctx context.Context, sourceFile string, size uint64) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ConvertQcowToVmdk", "sourceFile", sourceFile, "size", size)
	destFile := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
	destFile = destFile + ".vmdk"

	convertChan := make(chan string, 1)
	var convertErr string
	go func() {
		//resize to the correct size
		sizeInGB := fmt.Sprintf("%dG", size)
		log.SpanLog(ctx, log.DebugLevelInfra, "Resizing to", "size", sizeInGB)
		out, err := sh.Command("qemu-img", "resize", sourceFile, "--shrink", sizeInGB).CombinedOutput()

		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "qemu-img resize failed", "out", string(out), "err", err)
			convertChan <- fmt.Sprintf("qemu-img resize failed: %s %v", out, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "doing qemu-img convert", "destFile", destFile)
		out, err = sh.Command("qemu-img", "convert", "-O", "vmdk", "-o", "subformat=streamOptimized", sourceFile, destFile).CombinedOutput()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "qemu-img convert failed", "out", string(out), "err", err)
			convertChan <- fmt.Sprintf("qemu-img convert failed: %s %v", out, err)
		} else {
			convertChan <- ""

		}
	}()
	select {
	case convertErr = <-convertChan:
	case <-time.After(qcowConvertTimeout):
		return "", fmt.Errorf("ConvertQcowToVmdk timed out")
	}
	if convertErr != "" {
		return "", errors.New(convertErr)
	}
	return destFile, nil
}

package vcd

import (
	"context"
	"fmt"
	"path/filepath"

	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// appinst related functionality
// TBI
func (v *VcdPlatform) AddAppImageIfNotPresent(ctx context.Context, imageInfo *infracommon.ImageInfo, app *edgeproto.App, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddAppImageIfNotPresent", "app.ImagePath", app.ImagePath, "imageInfo", imageInfo, "flavor", flavor)

	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	f, err := v.GetFlavor(ctx, flavor)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Downloading VM Image")
	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.PlatformConfig.AccessApi, imageInfo.LocalImageName, app.ImagePath, imageInfo.Md5sum)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "downloaded file", "filePath", filePath)

	vmdkFile := filePath
	if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		updateCallback(edgeproto.UpdateTask, "Converting Image to VMDK")
		vmdkFile, err = v.ConvertQcowToVmdk(ctx, filePath, f.Disk)
		if err != nil {
			return err
		}
	}
	ovfFile := strings.TrimSuffix(vmdkFile, filepath.Ext(vmdkFile))
	ovfFile = ovfFile + ".ovf"

	err = v.UploadOvaFile(ctx, ovfFile, imageInfo.LocalImageName, "VM App OVF", vcdClient)
	return err
	/*
		defer func() {
			// Stale file might be present if download fails/succeeds, deleting it
			if delerr := infracommon.DeleteFile(filePath); delerr != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "filePath", filePath)
			}
			if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
				if delerr := infracommon.DeleteFile(vmdkFile); delerr != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "vmdkFile", vmdkFile)
				}
			}
		}()*/
}

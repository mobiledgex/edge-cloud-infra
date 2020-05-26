package vsphere

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (v *VSpherePlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddAppImageIfNotPresent", "app.ImagePath", app.ImagePath)

	imageName, err := cloudcommon.GetFileName(app.ImagePath)
	if err != nil {
		return err
	}
	_, md5Sum, err := infracommon.GetUrlInfo(ctx, v.vmProperties.CommonPf.VaultConfig, app.ImagePath)

	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, app.ImagePath, md5Sum)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "downloaded file", "filePath", filePath)

	vmdkFile := filePath
	if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		vmdkFile, err = v.ConvertQcowToVmdk(ctx, filePath)
		if err != nil {
			return err
		}
	}

	defer func() {
		// Stale file might be present if download fails/succeeds, deleting it
		if delerr := infracommon.DeleteFile(filePath); delerr != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "filePath", filePath)
		}
		if delerr := infracommon.DeleteFile(vmdkFile); delerr != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "vmdkFile", vmdkFile)
		}
	}()
	return v.ImportImage(ctx, vmdkFile)
}

func (v *VSpherePlatform) ConvertQcowToVmdk(ctx context.Context, sourceFile string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ConvertQcowToVmdk", "sourceFile", sourceFile)
	time.Sleep(time.Second * 60)
	destFile := strings.ReplaceAll(sourceFile, ".qcow2", "")
	destFile = destFile + ".vmdk"
	out, err := sh.Command("qemu-img", "convert", "-O", "vmdk", "-o", "subformat=streamOptimized", sourceFile, destFile).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "qemu-img convert failed", "out", string(out), "err", err)
		return "", fmt.Errorf("qemu-img convert failed: %s %v", out, err)
	}
	return destFile, nil
}

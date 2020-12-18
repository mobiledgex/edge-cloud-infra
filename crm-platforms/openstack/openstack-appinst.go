package openstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *OpenstackPlatform) GetConsoleUrl(ctx context.Context, serverName string) (string, error) {
	consoleUrl, err := o.OSGetConsoleUrl(ctx, serverName)
	if err != nil {
		return "", err
	}
	return consoleUrl.Url, nil
}

func (o *OpenstackPlatform) AddAppImageIfNotPresent(ctx context.Context, localImageName string, app *edgeproto.App, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddAppImageIfNotPresent", "localImageName", localImageName, "imagePath", app.ImagePath)

	sourceImageTime, md5Sum, err := infracommon.GetUrlInfo(ctx, o.VMProperties.CommonPf.PlatformConfig.AccessApi, app.ImagePath)
	imageDetail, err := o.GetImageDetail(ctx, localImageName)
	createImage := false
	if err != nil {
		if strings.Contains(err.Error(), ResourceNotFound) {
			// Add image to Glance
			log.SpanLog(ctx, log.DebugLevelInfra, "image is not present in glance, add image")
			createImage = true
		} else {
			return err
		}
	} else {
		if imageDetail.Status != "active" {
			return fmt.Errorf("image in store %s is not active", localImageName)
		}
		if imageDetail.Checksum != md5Sum {
			if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW && imageDetail.DiskFormat == vmlayer.ImageFormatVmdk {
				log.SpanLog(ctx, log.DebugLevelInfra, "image was imported as vmdk, checksum match not possible")
			} else {
				return fmt.Errorf("mismatch in md5sum for image in glance: %s", localImageName)
			}
		}
		glanceImageTime, err := time.Parse(time.RFC3339, imageDetail.UpdatedAt)
		if err != nil {
			return err
		}
		if !sourceImageTime.IsZero() {
			if sourceImageTime.Sub(glanceImageTime) > 0 {
				// Update the image in Glance
				updateCallback(edgeproto.UpdateTask, "Image in store is outdated, deleting old image")
				err = o.DeleteImage(ctx, "", localImageName)
				if err != nil {
					return err
				}
				createImage = true
			}
		}
	}
	if createImage {
		updateCallback(edgeproto.UpdateTask, "Creating VM Image from URL")
		err = o.CreateImageFromUrl(ctx, localImageName, app.ImagePath, md5Sum)
		if err != nil {
			return err
		}
	}
	return nil
}

package openstack

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *OpenstackPlatform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.vmPlatform.CreateAppInst(ctx, clusterInst, app, appInst, appFlavor, privacyPolicy, updateCallback)
}

func (o *OpenstackPlatform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	return o.vmPlatform.DeleteAppInst(ctx, clusterInst, app, appInst)
}

func (o *OpenstackPlatform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.vmPlatform.UpdateAppInst(ctx, clusterInst, app, appInst, updateCallback)
}

func (o *OpenstackPlatform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return o.vmPlatform.GetAppInstRuntime(ctx, clusterInst, app, appInst)
}

func (o *OpenstackPlatform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return o.vmPlatform.GetContainerCommand(ctx, clusterInst, app, appInst, req)
}

func (o *OpenstackPlatform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeVM:
		objName := cloudcommon.GetAppFQN(&app.Key)
		consoleUrl, err := o.OSGetConsoleUrl(ctx, objName)
		if err != nil {
			return "", err
		}
		return consoleUrl.Url, nil
	default:
		return "", fmt.Errorf("unsupported deployment type %s", deployment)
	}

}

// SetPowerState is mostly common code that could eventually move
func (o *OpenstackPlatform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	PowerState := appInst.PowerState
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeVM:
		serverName := cloudcommon.GetAppFQN(&app.Key)
		fqdn := appInst.Uri

		log.SpanLog(ctx, log.DebugLevelInfra, "setting server state", "serverName", serverName, "fqdn", fqdn, "PowerState", PowerState)

		updateCallback(edgeproto.UpdateTask, "Verifying AppInst state")
		serverDetail, err := o.GetServerDetail(ctx, serverName)
		if err != nil {
			return err
		}

		serverAction := ""
		switch PowerState {
		case edgeproto.PowerState_POWER_ON_REQUESTED:
			if serverDetail.Status == "ACTIVE" {
				return fmt.Errorf("server %s is already active", serverName)
			}
			serverAction = "start"
		case edgeproto.PowerState_POWER_OFF_REQUESTED:
			if serverDetail.Status == "SHUTOFF" {
				return fmt.Errorf("server %s is already stopped", serverName)
			}
			serverAction = "stop"
		case edgeproto.PowerState_REBOOT_REQUESTED:
			serverAction = "reboot"
			if serverDetail.Status != "ACTIVE" {
				return fmt.Errorf("server %s is not active", serverName)
			}
		default:
			return fmt.Errorf("unsupported server power action: %s", PowerState)
		}

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
		oldServerIP, err := o.GetIPFromServerName(ctx, o.vmPlatform.GetCloudletExternalNetwork(), serverName)
		if err != nil || oldServerIP.ExternalAddr == "" {
			return fmt.Errorf("unable to fetch external ip for %s, err %v", serverName, err)
		}

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Performing action %s on %s", serverAction, serverName))
		err = o.OSSetPowerState(ctx, serverName, serverAction)
		if err != nil {
			return err
		}

		if PowerState == edgeproto.PowerState_POWER_ON_REQUESTED || PowerState == edgeproto.PowerState_REBOOT_REQUESTED {
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Waiting for server %s to become active", serverName))
			serverDetail, err := o.GetActiveServerDetails(ctx, serverName)
			if err != nil {
				return err
			}
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
			newServerIP, err := o.GetIPFromServerName(ctx, o.vmPlatform.GetCloudletExternalNetwork(), serverName)
			if err != nil || newServerIP.ExternalAddr == "" {
				return fmt.Errorf("unable to fetch external ip for %s, addr %s, err %v", serverName, serverDetail.Addresses, err)
			}
			if oldServerIP.ExternalAddr != newServerIP.ExternalAddr {
				// IP changed, update DNS entry
				updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Updating DNS entry as IP changed for %s", serverName))
				log.SpanLog(ctx, log.DebugLevelInfra, "updating DNS entry", "serverName", serverName, "fqdn", fqdn, "ip", newServerIP)
				err = o.vmPlatform.CommonPf.ActivateFQDNA(ctx, fqdn, newServerIP.ExternalAddr)
				if err != nil {
					return fmt.Errorf("unable to update fqdn for %s, addr %s, err %v", serverName, newServerIP.ExternalAddr, err)
				}
			}
		}
		updateCallback(edgeproto.UpdateTask, "Performed power control action successfully")
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}

	return nil
}

func (o *OpenstackPlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error {
	imageName, err := cloudcommon.GetFileName(app.ImagePath)
	if err != nil {
		return err
	}
	sourceImageTime, md5Sum, err := infracommon.GetUrlInfo(ctx, o.vmPlatform.CommonPf.VaultConfig, app.ImagePath)
	imageDetail, err := o.GetImageDetail(ctx, imageName)
	createImage := false
	if err != nil {
		if strings.Contains(err.Error(), "Could not find resource") {
			// Add image to Glance
			log.SpanLog(ctx, log.DebugLevelInfra, "image is not present in glance, add image")
			createImage = true
		} else {
			return err
		}
	} else {
		if imageDetail.Status != "active" {
			return fmt.Errorf("image in store %s is not active", imageName)
		}
		if imageDetail.Checksum != md5Sum {
			if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW && imageDetail.DiskFormat == infracommon.ImageFormatVmdk {
				log.SpanLog(ctx, log.DebugLevelInfra, "image was imported as vmdk, checksum match not possible")
			} else {
				return fmt.Errorf("mismatch in md5sum for image in glance: %s", imageName)
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
				err = o.DeleteImage(ctx, imageName)
				if err != nil {
					return err
				}
				createImage = true
			}
		}
	}
	if createImage {
		updateCallback(edgeproto.UpdateTask, "Creating VM Image from URL")
		err = o.CreateImageFromUrl(ctx, imageName, app.ImagePath, md5Sum)
		if err != nil {
			return err
		}
	}
	return nil
}

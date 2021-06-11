package vsphere

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

var clusterLock sync.Mutex
var appLock sync.Mutex

const govcLocation = "https://github.com/vmware/govmomi/tree/master/govc"

func (v *VSpherePlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("SaveCloudletAccessVars not implemented for vsphere")
}

func (v *VSpherePlatform) GetCloudletImageSuffix(ctx context.Context) string {
	// we use a common image as of 4.4.0 and beyond
	vers := v.vmProperties.CommonPf.PlatformConfig.VMImageVersion
	if vers == "" {
		vers = vmlayer.MEXInfraVersion
	}
	// note this string compare would fail for something like 4.10.x, but is only intended for short term purposes
	// and should be removed after 4.4.0 has time to soak
	if vers >= "4.4.0" {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletImageSuffix returning generic suffix post 4.4.0", "vers", vers)
		return ".qcow2"
	}
	// older loads are specific per platform
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletImageSuffix returning vsphere specific suffix pre 4.4.0", "vers", vers)
	return "-vsphere.qcow2"
}

//CreateImageFromUrl downloads image from URL and then imports to the datastore
func (v *VSpherePlatform) CreateImageFromUrl(ctx context.Context, imageName, imageUrl, md5Sum string, diskSize uint64) error {

	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.PlatformConfig.AccessApi, imageName, imageUrl, md5Sum)
	if err != nil {
		return err
	}
	defer func() {
		// Stale file might be present if download fails/succeeds, deleting it
		if delerr := infracommon.DeleteFile(filePath); delerr != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "filePath", filePath)
		}
	}()

	vmdkFile, err := vmlayer.ConvertQcowToVmdk(ctx, filePath, diskSize)
	if err != nil {
		return err
	}
	return v.ImportImage(ctx, imageName, vmdkFile)
}

/*
func (v *VSpherePlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPath, md5Sum string, updateCallback edgeproto.CacheUpdateCallback) error {
	// we don't currently have the ability to download and setup the template, but we will verify it is there
	log.SpanLog(ctx, log.DebugLevelInfra, "AddCloudletImageIfNotPresent", "imgPath", imgPath, "md5Sum", md5Sum)
	pfImageName, err := cloudcommon.GetFileName(imgPath)
	if err != nil {
		return err
	}
	// see if a template already exists based on this image
	templatePath := v.GetTemplateFolder() + "/" + pfImageName
	_, err = v.GetServerDetail(ctx, templatePath)

	if err != nil {
		if !strings.Contains(err.Error(), vmlayer.ServerDoesNotExistError) {
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "template not present", "pfImageName", pfImageName, "err", err)

		// Validate if pfImageName is same as we expected
		_, md5Sum, err := infracommon.GetUrlInfo(ctx, v.vmProperties.CommonPf.PlatformConfig.AccessApi, imgPath)
		if err != nil {
			return err
		}
		// Download platform image and create a vsphere template from it
		updateCallback(edgeproto.UpdateTask, "Downloading platform base image: "+pfImageName)
		err = v.CreateImageFromUrl(ctx, pfImageName, imgPath, md5Sum)
		if err != nil {
			return fmt.Errorf("Error downloading platform base image %s: %v", pfImageName, err)
		}
		err = v.CreateTemplateFromImage(ctx, pfImageName, pfImageName)
		if err != nil {
			return fmt.Errorf("Error in creating baseimage template: %v", err)
		}
	}
	return nil
}*/

func (v *VSpherePlatform) GetFlavor(ctx context.Context, flavorName string) (*edgeproto.FlavorInfo, error) {

	flavs, err := v.vmProperties.GetFlavorListInternal(ctx, v.caches)
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

func (v *VSpherePlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	var flavors []*edgeproto.FlavorInfo
	// by returning no flavors, we signal to the controller this platform supports no native flavors
	log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList return empty", "len", len(flavors))
	return flavors, nil
}

func (v *VSpherePlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {
	vcaddr := v.vcenterVars["VCENTER_ADDR"]
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr", "vcaddr", vcaddr)
	if vcaddr == "" {
		return "", fmt.Errorf("unable to find VCENTER_ADDR")
	}
	return vcaddr, nil
}

func (o *VSpherePlatform) GetSessionTokens(ctx context.Context, vaultConfig *vault.Config, account string) (map[string]string, error) {
	return nil, fmt.Errorf("GetSessionTokens not supported in VSpherePlatform")
}

// GetCloudletManifest follows the standard practice for vSphere to use OVF for this purpose.  We store the OVF
// in artifactory along with with the vmdk formatted disk.  No customization is needed per cloudlet as the OVF
// import tool will prompt for datastore and portgroup.
func (v *VSpherePlatform) GetCloudletManifest(ctx context.Context, name string, cloudletImagePath string, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest", "name", name, "vmgp", vmgp)
	var manifest infracommon.CloudletManifest
	ovfLocation := vmlayer.DefaultCloudletVMImagePath + "vsphere-ovf-" + vmlayer.MEXInfraVersion
	err := v.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)
	if err != nil {
		return "", fmt.Errorf("unable to populate orchestration params: %v", err)
	}
	scriptText, err := v.GetRemoteDeployScript(ctx, vmgp)
	if err != nil {
		return "", err
	}

	manifest.AddItem("Create folder \"templates\" within the virtual datacenter", infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddItem("Download the OVF template", infracommon.ManifestTypeURL, infracommon.ManifestSubTypeNone, ovfLocation)
	manifest.AddItem("Import the OVF into vCenter into template folder: VMs and Templates -> templates folder -> Deploy OVF Template -> Local File -> Upload Files", infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddSubItem("Select Thin Provision for virtual disk format", infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddSubItem("Leave VM name unchanged", infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddSubItem(fmt.Sprintf("Select \"%s\" cluster and \"%s\" datastore", v.GetHostCluster(), v.GetDataStore()), infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddSubItem(fmt.Sprintf("Update port group when prompted to: %s", v.GetExternalVSwitch()), infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddItem("Ensure govc is installed on a machine with access to the vCenter APIs as per the following link", infracommon.ManifestTypeURL, infracommon.ManifestSubTypeNone, govcLocation)
	manifest.AddItem("Download the deployment script to where govc is installed and name it deploy.sh", infracommon.ManifestTypeCode, infracommon.ManifestSubTypeBash, scriptText)
	manifest.AddItem("Execute the downloaded script providing the Platform VM IP address as a parameter", infracommon.ManifestTypeCommand, infracommon.ManifestSubTypeNone, "bash deploy.sh <PLATFORM_IP>")

	// for testing, write the script and text to /tmp
	if v.vmProperties.CommonPf.PlatformConfig.TestMode {
		var client pc.LocalClient
		mstr, _ := manifest.ToString()
		pc.WriteFile(&client, "/tmp/manifest.txt", mstr, "manifest", pc.NoSudo)
		pc.WriteFile(&client, "/tmp/deploy.sh", scriptText, "script", pc.NoSudo)
	}
	return manifest.ToString()
}

func (s *VSpherePlatform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	return nil
}

func (s *VSpherePlatform) GetCloudletInfraResourcesInfo(ctx context.Context) ([]edgeproto.InfraResource, error) {
	return []edgeproto.InfraResource{}, nil
}

func (s *VSpherePlatform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{}, nil
}

func (s *VSpherePlatform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource {
	resInfo := make(map[string]edgeproto.InfraResource)
	return resInfo
}

func (s *VSpherePlatform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	return nil
}

func (v *VSpherePlatform) InternalCloudletUpdatedCallback(ctx context.Context, old *edgeproto.CloudletInternal, new *edgeproto.CloudletInternal) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InternalCloudletUpdatedCallback")
}

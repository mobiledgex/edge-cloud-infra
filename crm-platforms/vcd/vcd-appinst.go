package vcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type VmAppOvfParams struct {
	ImageBaseFileName string
	DiskSizeInBytes   string
}

var vmAppOvfTemplate = `<?xml version='1.0' encoding='UTF-8'?>
<Envelope xmlns="http://schemas.dmtf.org/ovf/envelope/1" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData">
  <References>
    <File ovf:id="file1" ovf:href="{{.ImageBaseFileName}}.vmdk"/>
  </References>
  <DiskSection>
    <Info>List of the virtual disks</Info>
    <Disk ovf:capacityAllocationUnits="byte" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" ovf:diskId="vmdisk1" ovf:capacity="{{.DiskSizeInBytes}}" ovf:fileRef="file1"/>
  </DiskSection>
  <VirtualSystem ovf:id="{{.ImageBaseFileName}}">
    <Info>A Virtual system</Info>
    <Name>{{.ImageBaseFileName}}</Name>
    <OperatingSystemSection ovf:id="94" vmw:osType="otherGuest64">
      <Info>The operating system installed</Info>
      <Description></Description>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware requirements</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemType>vmx-14</vssd:VirtualSystemType>
      </System>
      <Item>
        <rasd:AllocationUnits>hertz * 10^6</rasd:AllocationUnits>
        <rasd:Description>Number of Virtual CPUs</rasd:Description>
        <rasd:ElementName>2 virtual CPU(s)</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>2</rasd:VirtualQuantity>
        <vmw:CoresPerSocket ovf:required="false">1</vmw:CoresPerSocket>
      </Item>
      <Item>
        <rasd:AllocationUnits>byte * 2^20</rasd:AllocationUnits>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>4096MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>4096</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Description>SCSI Controller</rasd:Description>
        <rasd:ElementName>SCSI Controller 1</rasd:ElementName>
        <rasd:InstanceID>3</rasd:InstanceID>
        <rasd:ResourceSubType>lsilogic</rasd:ResourceSubType>
        <rasd:ResourceType>6</rasd:ResourceType>
        <vmw:Config ovf:required="false" vmw:key="slotInfo.pciSlotNumber" vmw:value="16"/>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Description>SATA Controller</rasd:Description>
        <rasd:ElementName>SATA Controller 1</rasd:ElementName>
        <rasd:InstanceID>4</rasd:InstanceID>
        <rasd:ResourceSubType>vmware.sata.ahci</rasd:ResourceSubType>
        <rasd:ResourceType>20</rasd:ResourceType>
        <vmw:Config ovf:required="false" vmw:key="slotInfo.pciSlotNumber" vmw:value="33"/>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:ElementName>Hard Disk 1</rasd:ElementName>
        <rasd:HostResource>ovf:/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>5</rasd:InstanceID>
        <rasd:Parent>3</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:AutomaticAllocation>false</rasd:AutomaticAllocation>
        <rasd:ElementName>CD/DVD Drive 1</rasd:ElementName>
        <rasd:InstanceID>6</rasd:InstanceID>
        <rasd:Parent>4</rasd:Parent>
        <rasd:ResourceSubType>vmware.cdrom.remotepassthrough</rasd:ResourceSubType>
        <rasd:ResourceType>15</rasd:ResourceType>
      </Item>
      <Item ovf:required="false">
        <rasd:ElementName>Video card</rasd:ElementName>
        <rasd:InstanceID>7</rasd:InstanceID>
        <rasd:ResourceType>24</rasd:ResourceType>
        <vmw:Config ovf:required="false" vmw:key="numDisplays" vmw:value="1"/>
        <vmw:Config ovf:required="false" vmw:key="graphicsMemorySizeInKB" vmw:value="262144"/>
        <vmw:Config ovf:required="false" vmw:key="use3dRenderer" vmw:value="automatic"/>
        <vmw:Config ovf:required="false" vmw:key="enable3DSupport" vmw:value="false"/>
        <vmw:Config ovf:required="false" vmw:key="useAutoDetect" vmw:value="false"/>
        <vmw:Config ovf:required="false" vmw:key="videoRamSizeInKB" vmw:value="4096"/>
      </Item>
      <vmw:Config ovf:required="false" vmw:key="flags.vbsEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="cpuHotAddEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="nestedHVEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="virtualSMCPresent" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="flags.vvtdEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="cpuHotRemoveEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="memoryHotAddEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="bootOptions.efiSecureBootEnabled" vmw:value="false"/>
      <vmw:Config ovf:required="false" vmw:key="firmware" vmw:value="bios"/>
      <vmw:Config ovf:required="false" vmw:key="virtualICH7MPresent" vmw:value="false"/>
    </VirtualHardwareSection>
  </VirtualSystem>
</Envelope>
`

// appinst related functionality
// TBI
func (v *VcdPlatform) AddAppImageIfNotPresent(ctx context.Context, imageInfo *infracommon.ImageInfo, app *edgeproto.App, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddAppImageIfNotPresent", "app.ImagePath", app.ImagePath, "imageInfo", imageInfo, "flavor", flavor)

	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	_, err := v.FindTemplate(ctx, imageInfo.LocalImageName, vcdClient)
	if err == nil {
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Using existing image template: %s", imageInfo.LocalImageName))
		return nil
	}

	if v.GetVcdOauthAgwUrl() != "" {
		// uploading a VM app image thru the API GW is not possible
		return fmt.Errorf("VM App images cannot be imported when using an API Gateway")
	}
	filesToCleanup := []string{}
	defer func() {
		for _, file := range filesToCleanup {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete file", "file", file)
			if delerr := infracommon.DeleteFile(file); delerr != nil {
				if !os.IsNotExist(delerr) {
					log.SpanLog(ctx, log.DebugLevelInfra, "delete file failed", "file", file)
				}
			}
		}
	}()

	appFlavor, err := v.GetFlavor(ctx, flavor)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Downloading VM Image")
	fileWithPath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.PlatformConfig.AccessApi, imageInfo.LocalImageName, app.ImagePath, imageInfo.Md5sum)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "downloaded file", "fileWithPath", fileWithPath)
	filesToCleanup = append(filesToCleanup, fileWithPath)

	vmdkFile := fileWithPath
	if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		updateCallback(edgeproto.UpdateTask, "Converting Image to VMDK")
		vmdkFile, err = vmlayer.ConvertQcowToVmdk(ctx, fileWithPath, appFlavor.Disk)
		if err != nil {
			return err
		}
		filesToCleanup = append(filesToCleanup, vmdkFile)
	}
	filenameNoExtension := strings.TrimSuffix(vmdkFile, filepath.Ext(vmdkFile))
	ovfFile := filenameNoExtension + ".ovf"

	imageFileBaseName := filepath.Base(filenameNoExtension)
	ovfParams := VmAppOvfParams{
		ImageBaseFileName: imageFileBaseName,
		DiskSizeInBytes:   fmt.Sprintf("%d", appFlavor.Disk*1024*1024*1024),
	}
	ovfBuf, err := infracommon.ExecTemplate("vmwareOvf", vmAppOvfTemplate, ovfParams)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Creating OVF file", "ovfFile", ovfFile, "ovfParams", ovfParams)
	err = ioutil.WriteFile(ovfFile, ovfBuf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("unable to write OVF file %s: %s", ovfFile, err.Error())
	}
	filesToCleanup = append(filesToCleanup, ovfFile)

	updateCallback(edgeproto.UpdateTask, "Uploading OVF")
	err = v.UploadOvaFile(ctx, ovfFile, imageInfo.LocalImageName, "VM App OVF", vcdClient)
	if err != nil {
		return fmt.Errorf("Upload OVA failed - %v", err)
	}
	return nil
}

func (v *VcdPlatform) DeleteImage(ctx context.Context, folder, image string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteImage", "image", image)
	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	err := v.DeleteTemplate(ctx, image, vcdClient)
	if err != nil {
		if strings.Contains(err.Error(), govcd.ErrorEntityNotFound.Error()) {
			log.SpanLog(ctx, log.DebugLevelInfra, "image already deleted", "image", image)
			return nil
		} else {
			return fmt.Errorf("DeleteImage: %s failed - %v", image, err)
		}
	}
	return nil
}

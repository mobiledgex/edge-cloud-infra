package vsphere

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
)

type VsphereCustomizationParams struct {
	Name          string
	DnsServers    []string
	NumDnsServers int
	MetaData      string
	UserData      string
}

func (v *VSpherePlatform) GetCustomizationSpec(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCustomizationSpec", "vmgp", vmgp)
	if len(vmgp.VMs) != 1{
		return "", fmt.Errorf("error, should be one VM in the customization params, found: %d", len(vmgp.VMs))
	}
	imgPath := vmlayer.GetCloudletVMImagePath(imgPathPrefix, imgVersion, v.GetCloudletImageSuffix(ctx))
	// Fetch platform base image name
	pfImageName, err := cloudcommon.GetFileName(imgPath)
	if err != nil {
		return "", err
	}
	// see if a template already exists based on this image
	templatePath := v.GetTemplateFolder() + "/" + pfImageName
	dcName := v.GetDatacenterName(ctx)
	dsName := v.GetDataStore()
	cluster := v.GetHostCluster()
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, cluster)
	vmName := vmgp.VMs[0].Name
	poolPath := pathPrefix + poolName
	vcpu := vmgp.VMs[0].Vcpus
	ram := vmgp.VMs[0].Ram
	disk := vmgp.VMs[0].Disk
	extNet := v.GetExternalVSwitch()
	metaData := vmgp.VMs[0].MetaData
	userData := vmgp.VMs[0].UserData


    script := "govc pool.create -dc %s\n", poolPath)
	script += fmt.Sprintf("govc vm.clone -vm %s -dc %s -ds %s -vm %s -on=False -pool %s -c %d -m %d -net %s\n", templatePath, dcName, dsName, vmName, poolPath, vcpu, ram, extNet)  
	script += "govc vm.change -vm %s -dc %s -e guestinfo.metadata=%s -e guestinfo.metadata.encoding=base64\n", vmName, dcName, metaData)
	script += "govc vm.change -vm %s -dc %s -e guestinfo.userdata=%s -e guestinfo.userdata.encoding=base64\n", vmName, dcName, userData)
	script += "govc vm.disk.change -dc %s -vm %s -size %dG\n",  dcName, vmName, disk)
	return script, nil
}

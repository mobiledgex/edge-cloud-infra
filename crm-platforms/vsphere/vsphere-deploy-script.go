package vsphere

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
)

func (v *VSpherePlatform) GetRemoteDeployScript(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetRemoteDeployScript", "vmgp", vmgp)
	if len(vmgp.VMs) != 1 {
		return "", fmt.Errorf("error, should be one VM in the customization params, found: %d", len(vmgp.VMs))
	}
	// update instructions
	// update script
	templatePath := v.GetTemplateFolder() + "/" + vmgp.VMs[0].ImageName
	dcName := v.GetDatacenterName(ctx)
	dsName := v.GetDataStore()
	cluster := v.GetHostCluster()
	poolName := getResourcePoolName(vmgp.GroupName, string(vmlayer.VMDomainPlatform))
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, cluster)
	poolPath := pathPrefix + poolName
	vmName := vmgp.VMs[0].Name
	vcpu := vmgp.VMs[0].Vcpus
	ram := vmgp.VMs[0].Ram
	disk := vmgp.VMs[0].Disk
	extNet := vmgp.VMs[0].Ports[0].PortGroup
	metaData := vmgp.VMs[0].MetaData
	userData := vmgp.VMs[0].UserData

	script := "#!/bin/bash\n\n"
	// add envvars
	for k, v := range v.vcenterVars {
		script += fmt.Sprintf("export %s=%s\n", k, v)
	}
	script += fmt.Sprintf("\ngovc pool.create -dc %s %s \n\n", dcName, poolPath)
	for _, tag := range vmgp.Tags {
		script += fmt.Sprintf("govc tags.create -c %s %s\n\n", tag.Category, tag.Name)
	}
	script += fmt.Sprintf("govc vm.clone -vm %s -dc %s -ds %s -on=False -pool %s -c %d -m %d -net %s %s\n\n", templatePath, dcName, dsName, poolPath, vcpu, ram, extNet, vmName)
	script += fmt.Sprintf("govc vm.change -vm %s -dc %s -e guestinfo.metadata=%s -e guestinfo.metadata.encoding=base64\n\n", vmName, dcName, metaData)
	script += fmt.Sprintf("govc vm.change -vm %s -dc %s -e guestinfo.userdata=%s -e guestinfo.userdata.encoding=base64\n\n", vmName, dcName, userData)
	script += fmt.Sprintf("govc vm.disk.change -dc %s -vm %s -size %dG\n", dcName, vmName, disk)
	return script, nil
}

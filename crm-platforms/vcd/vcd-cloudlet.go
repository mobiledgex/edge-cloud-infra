package vcd

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
)

// Cloudlet related operations

// The CreateImageFromUrl in vsphere could be placed into common-utils, if we separate the
// fetch, from v.ImportImage, allowing platform to do the needful.
// Then have a look at that AddCloudImageIfNotPresent for potential refactor. xxx uses GetServerDetail is that standard?

// interesting Status values for a VApp
// 4 = POWERED_ON  - All vms in vapp are runing
// 9 = INCONSISTENT_STATE - Some vms on, some not
// 8 = POWERED_OFF (implies 1=RESOLOVED)
// 1 = RESOLVED  - vapp is created, but has no VMs yet.
//  see top of types.go

// Used to create the vdc/cloudlet VApp only. Just one external network PrimaryNet at this point.
//
func (v *VcdPlatform) CreateCloudlet(ctx context.Context, vappTmpl govcd.VAppTemplate, vmgp *vmlayer.VMGroupOrchestrationParams, description string, updateCallback edgeproto.CacheUpdateCallback) (*govcd.VApp, error) {
	var vapp *govcd.VApp
	var err error
	var vmRole vmlayer.VMRole
	vdc := v.Objs.Vdc
	storRef := types.Reference{}
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet", "name", vmgp.GroupName, "tmpl", vappTmpl)
	// dumpVMGroupParams(vmgp, 1)
	vmtmplVMName := ""
	vmtmpl := &types.VAppTemplate{}
	vmparams := vmlayer.VMOrchestrationParams{}
	newVappName := vmgp.GroupName + "-vapp"
	if len(vappTmpl.VAppTemplate.Children.VM) != 0 {
		// xxx non-standard
		vmtmpl = vappTmpl.VAppTemplate.Children.VM[0]
		vmparams = vmgp.VMs[0]
		fmt.Printf("\n\nCreateCloudlet Template %s vm name %s\n\n", vappTmpl.VAppTemplate.Name, vmtmpl.Name)
		vmtmplVMName = vmtmpl.Name
		vmtmpl.Name = vmparams.Name
		vmRole = vmparams.Role
	}

	networks := []*types.OrgVDCNetwork{}
	networks = append(networks, v.Objs.PrimaryNet.OrgVDCNetwork)
	extAddr, err := v.GetNextExtAddrForVdcNet(ctx, vdc)
	if err != nil {
		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet failed to obtained ext addr", "error", err)
		}
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "retrieved addr", "address", extAddr)
	vapp, err = v.FindVApp(ctx, newVappName)
	if err != nil {
		// Not found try and create it
		log.SpanLog(ctx, log.DebugLevelInfra, "create cloudlet", "name", newVappName)
		task, err := vdc.ComposeVApp(networks, vappTmpl, storRef, newVappName, description+vcdProviderVersion, true)
		if err != nil {
			vappTmpl.VAppTemplate.Name = vmtmplVMName
			if strings.Contains(err.Error(), "already exists") {
				// So we should have found this already, so this means it was created
				// behind our backs somehow, so we should add it to our pile of existing VApps...
				log.SpanLog(ctx, log.DebugLevelInfra, "already exists", "GroupName", vmgp.GroupName)

			} else {
				// operation failed for resource reasons
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateVApp failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
		} else {

			vappTmpl.VAppTemplate.Name = vmtmplVMName
			log.SpanLog(ctx, log.DebugLevelInfra, "create cloudlet vapp composed", "address", extAddr)
			err = task.WaitTaskCompletion()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "ComposeVApp wait for completeion failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			// change the new vapps vm name

			vapp, err = v.Objs.Vdc.GetVAppByName(newVappName, true)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "can't retrieve compoled vapp", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			err = vapp.BlockWhileStatus("UNRESOLVED", 120) // upto seconds
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "wait for RESOLVED error", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			task, err = vapp.RemoveAllNetworks()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "remove networks failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "task completion failed", "VAppName", vmgp.GroupName, "error", err)
			}

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 0

			if vmRole == vmlayer.RoleAgent { // Other cases XXX
				vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))
				log.SpanLog(ctx, log.DebugLevelInfra, "Add external network", "VAppName", vmgp.GroupName, "role", vmRole, "type", vmType)
				_ /* networkConfigSection */, err = v.AddVappNetwork(ctx, vapp)

				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "Add external network failed ", "VAppName", vmgp.GroupName, "err", err)
					return nil, err
				}
				desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
					&types.NetworkConnection{
						IsConnected:             true,
						IPAddressAllocationMode: types.IPAllocationModeManual,
						Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name,
						NetworkConnectionIndex:  0,
						IPAddress:               extAddr,
					})
			}

			vmtmplName := vapp.VApp.Children.VM[0].Name
			vm, err := vapp.GetVMByName(vmtmplName, false)
			if err != nil {
				return nil, err
			}
			// One or two networks, update our connection(s)
			err = vm.UpdateNetworkConnectionSection(desiredNetConfig)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "update network failed", "VAppName", vmgp.GroupName, "err", err)
				return nil, err
			}

			cloudletNameParts := strings.Split(vapp.VApp.Name, ".")
			cloudletName := cloudletNameParts[0]

			Vapp := VApp{
				VApp: vapp,
			}

			v.Objs.VApps[newVappName] = &Vapp
			// And create our Cloudlet object
			cloudlet := MexCloudlet{
				ParentVdc:    vdc,
				CloudVapp:    vapp,
				CloudletName: cloudletName,
				ExtNet:       v.Objs.PrimaryNet, // vdcnet,
				Clusters:     make(CidrMap),
				ExtVMMap:     make(CloudVMsMap),
			}
			v.Objs.Cloudlet = &cloudlet
			// Finish
			for curChild, child := range vapp.VApp.Children.VM {
				vmparams := vmgp.VMs[curChild]
				// We need the govcd.VM to customize
				// mark this vm with vapp name and position for uniqueness.
				key := fmt.Sprintf("%s-%d", vapp.VApp.Name, curChild)
				vm, err = vapp.GetVMByName(child.Name, true)
				if err != nil {
					return nil, err
				}
				vm.VM.OperationKey = key

				var subnet string
				err = v.updateVM(ctx, vm, vmparams, subnet)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "update vm failed ", "VAppName", vmgp.GroupName, "err", err)
					return nil, err
				}
				v.Objs.Cloudlet.ExtVMMap[extAddr] = vm
			}
			// Once we've customized this vm, add it our vm map
			//err = task.WaitTaskCompletion()
			// This will power on all vms in the vapp, we can order them
			// So master first and then workers.
			task, err := vapp.AddMetadata("CloudletName", cloudletName)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "add metadata failed ", "VAppName", vmgp.GroupName, "err", err)
				return nil, err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return nil, err
			}

			task, err = vapp.PowerOn()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "power on  failed ", "VAppName", vmgp.GroupName, "err", err)
				return nil, err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "wait power on  failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			vapp.Refresh()
			return vapp, nil
		}
	}
	vdc = v.Objs.Vdc
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet  exists", "cloudlet", vapp.VApp.Name, "vdc", vdc.Vdc.Name)
	// Refresh our local object
	vapp.Refresh()
	return vapp, nil
}

// Is this only for appInst images? Or our cloudlet template too?
func (v *VcdPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "PI AddCloudletImageIfNotPresent  TBI ", "imgPathPrefix", imgPathPrefix, "ImgVersion", imgVersion)
	//	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, imageUrl, md5Sum)
	return "", nil
}

func (v *VcdPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("SaveCloudletAccessVars not implemented for vcd")
}

// This appears to only deal with non-existant flavors in vmware world
func (v *VcdPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo ")
	// get the flavor list
	// whatelse do we need here?
	var err error
	info.Flavors, err = v.GetFlavorList(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *VcdPlatform) GetCloudletImageSuffix(ctx context.Context) string {
	return "-vcd.qcow2"
}

// XXX needed?
func (v *VcdPlatform) GetCloudletManifest(ctx context.Context, name, cloudletImagePath string, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest", "name", name, "imagePath", cloudletImagePath)
	return "", nil
}

// remove the cloudlet(Vapp) + all VMs in the cloudlet.
func (v *VcdPlatform) DeleteCloudlet(ctx context.Context, cloudlet MexCloudlet) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCloudlet", "cloudlet", cloudlet.CloudletName)

	vapp := cloudlet.CloudVapp

	// power off the vapp, then all the vms, then remove all vms from the vapp, and finally delete the vapp.
	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}
	// unlikely
	if vapp.VApp.Children == nil {
		task, err := vapp.PowerOff()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "baren vapp power off failed", "cloudlet", cloudlet, "err", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "baren vapp wait task failed", "cloudlet", cloudlet, "err", err)
		}

		task, err = vapp.Delete()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "baren vapp delete failed", "cloudlet", cloudlet, "err", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "baren vapp delete wait task failed", "cloudlet", cloudlet, "err", err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Barren VApp %s Deleted", "cloudletName", cloudlet.CloudletName)
		return nil
	}
	// Nominal
	vms := vapp.VApp.Children.VM
	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCloudlet removing", "children vms", vms)
	}
	if status == "POWERED_ON" {
		for _, vm := range vms {

			v, err := vapp.GetVMByName(vm.Name, false)
			if err != nil {
				continue
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCloudlet power off", "vm", vm.Name)
			task, err := v.PowerOff()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "vm power off failed", "vm", vm.Name, "err", err)
				return err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "wait vm power off failed", "vm", vm.Name, "err", err)
				return err
			}

		}
	}

	for _, vm := range vms {
		v, err := vapp.GetVMByName(vm.Name, false)
		if err != nil {
			return err
		}
		err = vapp.RemoveVM(*v)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "remove vm failed", "vm", vm.Name, "err", err)
			continue
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "vm removed from vapp", "vm", vm.Name)
	}

	task, err := vapp.Delete()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "delete vapp  failed", "vm", vapp.VApp.Name, "err", err)
	}
	err = task.WaitTaskCompletion()

	v.Objs.Cloudlet = nil
	// clean up maps xxx
	log.SpanLog(ctx, log.DebugLevelInfra, "Cloudlet deleted")
	return err
}

// Just validate requested cluster create references against our cloudlet
// re: was cloudlet lookup
// Not much use when only one cloudlet per vcd
func (v *VcdPlatform) FindCloudletForCluster(GroupName string) (*MexCloudlet, *govcd.VApp, error) {
	vdcCloudlet := v.Objs.Cloudlet

	// validate cld name is our vdc/vapp name
	if strings.Contains(GroupName, vdcCloudlet.CloudletName) {
		return vdcCloudlet, v.Objs.Cloudlet.CloudVapp, nil
	}
	return nil, nil, fmt.Errorf("Unknown Cloudlet specified")
}

func (o *VcdPlatform) GetSessionTokens(ctx context.Context, vaultConfig *vault.Config, account string) (map[string]string, error) {
	return nil, fmt.Errorf("GetSessionTokens not supported in VcdPlatform")
}

// IP address or Href? It's the Href with a manditory port
func (v *VcdPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {

	ip := v.GetVCDIP() // {vcdVars["VCD_IP"]
	apiUrl := ip + "/api"
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr", "Href", apiUrl)
	return apiUrl, nil
}

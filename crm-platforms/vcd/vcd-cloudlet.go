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
func (v *VcdPlatform) CreateCloudlet(ctx context.Context, vdc *govcd.Vdc, vappTmpl govcd.VAppTemplate, storProf types.Reference, vmgp *vmlayer.VMGroupOrchestrationParams, description string, updateCallback edgeproto.CacheUpdateCallback) (*govcd.VApp, error) {
	var vapp *govcd.VApp
	var err error
	var vmRole vmlayer.VMRole

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet", "name", vmgp.GroupName)
	// dumpVMGroupParams(vmgp, 1)

	vmparams := vmlayer.VMOrchestrationParams{}
	newVappName := vmgp.GroupName + "-vapp" // 11/04 does removing this screw it up? should only be done for cloudlet
	// seems to perhaps not allow the added vm for cluster to take new internal networks try putting it back 11/17

	if len(vappTmpl.VAppTemplate.Children.VM) != 0 {
		// we want to change the name of the templates vm to that of
		// our vmparams.Name XXX this isn't how it 'officially' works, TODO move to update
		vmtmpl := vappTmpl.VAppTemplate.Children.VM[0]
		vmparams = vmgp.VMs[0]
		vmtmpl.Name = vmparams.Name // this will become the vm name in the new Vapp it's the provider specified name (server/vm)
		//		fmt.Printf("CreateVAppFromTmpl-I-changed vm[0] name in template child:%s to %s\n",
		//			vappTmpl.VAppTemplate.Children.VM[0].Name, vmtmpl.Name)
		vmRole = vmparams.Role
	}
	networks := []*types.OrgVDCNetwork{}
	networks = append(networks, v.Objs.PrimaryNet.OrgVDCNetwork)
	extAddr, err := v.GetNextExtAddrForVdcNet(ctx, vdc)
	if err != nil {
		fmt.Printf("CreateCloudlet-E-failed to obtain ext net ip %s\n", err.Error())
		return nil, err
	}
	fmt.Printf("\nCreateCloudlet-I-extAddr retrieved: %s\n", extAddr)
	vapp, err = v.FindVApp(ctx, newVappName)
	if err != nil {
		// Not found try and create it
		task, err := vdc.ComposeVApp(networks, vappTmpl, storProf, newVappName, description+vcdProviderVersion, true)
		if err != nil {

			if strings.Contains(err.Error(), "already exists") {
				// So we should have found this already, so this means it was created
				// behind our backs somehow, so we should add it to our pile of existing VApps...
				fmt.Printf("CreateVMs %s was not found locally, but Compose returns already exists. Add to local map\n", vmgp.GroupName)

			} else {
				// operation failed for resource reasons
				fmt.Printf("CreateVAppFromTempl-E-Compose failed for %s error: %s\n", vmgp.GroupName, err.Error())
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateVApp failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
		} else {
			err = task.WaitTaskCompletion()
			if err != nil {
				fmt.Printf("\nCreateVAppFromTmpl-E-waiting for task complete %s\n", err.Error())
				return nil, err
			}

			vapp, err = v.Objs.Vdc.GetVAppByName(newVappName, true)
			if err != nil {
				return nil, err
			}
			err = vapp.BlockWhileStatus("UNRESOLVED", 120) // upto seconds
			if err != nil {
				return nil, err

			}
			task, err = vapp.RemoveAllNetworks()
			if err != nil {
				fmt.Printf("Error removing all networks: %s\n", err.Error())
				// Shouldn't be fatal though eh?
			}
			err = task.WaitTaskCompletion()

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 0
			// if the vm is to host the LB, we need

			if vmRole == vmlayer.RoleAgent { // Other cases XXX
				//haveExternalNet = true
				fmt.Printf("CreateVApp-I-add external network\n")

				_ /* networkConfigSection */, err = v.AddVappNetwork(ctx, vapp)

				if err != nil {
					fmt.Printf("CreateRoutedExternalNetwork (external) failed: %s\n", err.Error())
					return nil, err
				}

				// Revisit ModePool and 12/13/20
				desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
					&types.NetworkConnection{
						IsConnected:             true,
						IPAddressAllocationMode: types.IPAllocationModePool,           // Manual
						Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name, // types.NoneNetwork,
						NetworkConnectionIndex:  0,
						// pool test IPAddress:               extAddr,
					})
			}

			vmtmplName := vapp.VApp.Children.VM[0].Name
			fmt.Printf("Using existing VM in template Named: %s\n", vmtmplName)
			vm, err := vapp.GetVMByName(vmtmplName, false)
			if err != nil {
				return nil, err
			}
			// One or two networks, update our connection(s)
			err = vm.UpdateNetworkConnectionSection(desiredNetConfig)
			if err != nil {
				fmt.Printf("CreateVAppFromTemplate-E-UpdateNetworkConnnectionSection: %s\n", err.Error())
				return nil, err
			}

			cloudletNameParts := strings.Split(vapp.VApp.Name, ".")
			for i := 0; i < len(cloudletNameParts); i++ {
				fmt.Printf("\t%s\n", cloudletNameParts[i])
			}
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
				fmt.Printf("\t%s\n", child.Name)
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
					fmt.Printf("CreateVMs-E-error updating VM %s : %s \n", child.Name, err.Error())
					return nil, err
				}

				v.Objs.Cloudlet.ExtVMMap[extAddr] = vm
				fmt.Printf("\n\n\tCreateCloudlet-I-added entry ExtVMMap key %s vm: %s maplen: %d \n\n", extAddr,
					vm.VM.Name, len(v.Objs.Cloudlet.ExtVMMap))

			}
			// Once we've customized this vm, add it our vm map
			//err = task.WaitTaskCompletion()
			// This will power on all vms in the vapp, we can order them
			// So master first and then workers.
			fmt.Printf("\n\nCreateCloudlet-I-add metadata cloud name %s to vapp %s has %d child VMs\n", cloudletName, vapp.VApp.Name, len(vapp.VApp.Children.VM))
			task, err := vapp.AddMetadata("CloudletName", cloudletName)
			if err != nil {
				return nil, err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return nil, err
			}

			task, err = vapp.PowerOn()
			if err != nil {
				return nil, err
			}
			fmt.Printf("CreateVMs-I-waiting task complete for power on...\n")
			err = task.WaitTaskCompletion()
			if err != nil {
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
	// how about just returning our ubuntu18.04 image here? For now
	fmt.Printf("AddCloudletImageIfNotPresent-i-TBI\n")

	//	filePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, imageUrl, md5Sum)

	return "", nil
}

// PI   Security calls this to save what it gets from vault?
func (v *VcdPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {

	return fmt.Errorf("SaveCloudletAccessVars not implemented for vcd")
}

// This appears to only deal with non-eixstant flavors in vmware world
func (v *VcdPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo ")
	// get the flavor list
	// whatelse do we need here?
	var err error
	info.Flavors, err = v.GetFlavorList(ctx)
	if err != nil {
		fmt.Printf("\n\nGatherCloudlentInfo-E-GetFlavorList err: %s\n", err.Error())
		return err
	}
	return nil
}

// PI  why is this needed

func (v *VcdPlatform) GetCloudletImageSuffix(ctx context.Context) string {
	fmt.Printf("GetCloudletImageSuffix TBI\n")
	// needed? Follow convention
	return "-vcd.qcow2"
}

//
// XXX Its the result of the heat stack apply in openstack, not supported by vmpool, and TBI as the OVF file for the cloudlet in vsphere...
// Not sure we'll have a single OVF file for our entire cloudlet? (with muliple clusterInsts?)
func (v *VcdPlatform) GetCloudletManifest(ctx context.Context, name, cloudletImagePath string, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest name: %s imagePath? %s ", name, cloudletImagePath)
	return "", nil

}

// remove the cloudlet(Vapp) + all VMs in the cloudlet.
func (v *VcdPlatform) DeleteCloudlet(ctx context.Context, cloudlet MexCloudlet) error {

	vapp := cloudlet.CloudVapp

	// power off the vapp, then all the vms, then remove all vms from the vapp, and finally delete the vapp.
	fmt.Printf("\ntestDestroyVApp-I-request Delete of %s\n", vapp.VApp.Name)

	//vdc := v.Objs.Vdc

	status, err := vapp.GetStatus()
	fmt.Printf("Vapp %s currently in state: %s\n", vapp.VApp.Name, status)
	if err != nil {
		fmt.Printf("Error fetching status for vapp %s\n", vapp.VApp.Name)
		return err
	}

	vm := &govcd.VM{}
	if vapp.VApp.Children == nil {
		task, err := vapp.Delete()
		if err != nil {
			fmt.Printf("vapp.Delete failed: %s\n task: %+v\n", err.Error(), task)
		}
		err = task.WaitTaskCompletion()
		fmt.Printf("Barren VApp %s Deleted\n", vapp.VApp.Name)
		return err
	}
	vms := vapp.VApp.Children.VM
	fmt.Printf("vapp %s has %d vm children\n", vapp.VApp.Name, len(vms))
	if status == "POWERED_ON" {

		// if the vapp is on, assume the vm is too
		// Nope, if you power off this vm, and then

		task, err := vm.PowerOff()
		if err != nil {
			fmt.Printf("Error from vm.PowerOff: %s\n", err.Error())
			// fatal? Could have already been powered off
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			fmt.Printf("Error waiting powering of the vm %s \n", vm.VM.Name)
			return err
		}
		fmt.Printf("VM %s powered off\n", vm.VM.Name)

		fmt.Printf("vapp %s currently powered on, wait for power off...\n", vapp.VApp.Name)
		task, err = vapp.PowerOff()
		if err != nil {
			fmt.Printf("testDestroyVapp-W-vm power off failed : %s\n", err.Error())
			return err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			fmt.Printf("Error powering of the Vapp %s \n", vapp.VApp.Name)
			return err
		}
		fmt.Printf("Vapp %s powered off\n", vapp.VApp.Name)
	}
	// And while the console's delete vapp deletes it's vms, this does not, so the remove VM will fail since
	// it's still powered on (though the console show's it as powered off, but still able to power it off..
	// go figure.

	// Also, to get the ip addresses released back to the pool, take out the VM before the VApp...
	// Apparently, they both are consuming one, even though they are the same (one for each end?)

	// Now, consider a Vapp with >1 vm
	for _, vm := range vms {
		fmt.Printf("\t%s\n", vm.Name)

		v, err := vapp.GetVMByName(vm.Name, false)
		if err != nil {
			fmt.Printf("VM %s not found \n", vm.Name)
			return err
		}

		err = vapp.RemoveVM(*v)
		if err != nil {
			fmt.Printf("Error from RemoveVM for vm %s in vapp: %s as : %s\n",
				vm.Name, vapp.VApp.Name, err.Error())
			return err
		}
		fmt.Printf("VM %s removed from Vapp\n", vm.Name)
	}

	task, err := vapp.Delete()
	if err != nil {
		fmt.Printf("vapp.Delete failed: %s\n task: %+v\n", err.Error(), task)
	}
	err = task.WaitTaskCompletion()
	fmt.Printf("VApp %s Deleted\n", vapp.VApp.Name)
	return err
}

// Just validate requested cluster create references against our cloudlet
// re: was cloudlet lookup
// Not much use when only one cloudlet per vcd
func (v *VcdPlatform) FindCloudletForCluster(GroupName string) (*MexCloudlet, *govcd.VApp, error) {
	vdcCloudlet := v.Objs.Cloudlet // only one
	// cld1-cluster1-mobiledgex-vapp
	// vs
	//  cluster1.cld1.tdg.mobiledgex.net
	/*
		parts := strings.Split(GroupName, ".")
		if len(parts) > 1 {
			cldName := parts[1]
			clustName := parts[0]
		}
	*/
	fmt.Printf("FindCloudletForCluster-I-GroupName %s cloudlet %s\n", GroupName, vdcCloudlet.CloudletName)
	// validate cld name is our vdc/vapp name
	if strings.Contains(GroupName, vdcCloudlet.CloudletName) {
		fmt.Printf("\nFindCloudletForCluster CreateVMs Selecting existing\n\tClouldlet %s\n\t vapp  %s\n\tvdc: %s\n for adding vms in %s\n",
			vdcCloudlet.CloudletName,
			vdcCloudlet.CloudVapp.VApp.Name,
			vdcCloudlet.ParentVdc.Vdc.Name,
			GroupName)

		return vdcCloudlet, v.Objs.Cloudlet.CloudVapp, nil
	}
	return nil, nil, fmt.Errorf("Unknown Cloudlet specified")
}

func (o *VcdPlatform) GetSessionTokens(ctx context.Context, vaultConfig *vault.Config, account string) (map[string]string, error) {
	return nil, fmt.Errorf("GetSessionTokens not supported in VcdPlatform")
}

// IP address or Href? It's the Href with a manditory port
func (v *VcdPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {
	// example :
	// OS_AUTH_URL https://10.254.108.198:5000/v3
	// Our port is default 443, but parsing requires it exist.
	ip := v.vcdVars["VCD_IP"]
	apiUrl := fmt.Sprintf("%s%s%s", "https://", ip, ":443/api")
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr", "Href", apiUrl)
	fmt.Printf("\nGetApiEndpoingAddr-I-%s\n\n", apiUrl)

	return apiUrl, nil
}

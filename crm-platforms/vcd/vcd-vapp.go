package vcd

import (
	"context"
	"fmt"

	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
)

func (v *VcdPlatform) addVApp(ctx context.Context, vappName string, vdc *govcd.Vdc) error {
	// shouldn't we just pass it in here?

	fmt.Printf("addVapp-I-adding newly created vapp named: %s to our local cache\n", vappName)

	// check if it's already here?
	vapp, err := v.FindVApp(ctx, vappName)
	if err != nil {
		fmt.Printf("addVApp-W-%s exists is the same?\n", vappName)
		tmpVApp, err := vdc.GetVAppByName(vappName, true)
		if err != nil {
			fmt.Printf("addVApp-E-asking to add an unknown Vapp: %s\n", vappName)
			return err
		}
		// Now, check their HREFS for equality
		if vapp.VApp.HREF != tmpVApp.VApp.HREF {
			fmt.Printf("addVApp-E-two vapps same name, different apps %s vs %s\n",
				vapp.VApp.HREF, tmpVApp.VApp.HREF)
			return fmt.Errorf("Duplicate VApp Names found")
		}
	}
	// update
	VApp := VApp{
		VApp: vapp,
	}
	v.Objs.VApps[vapp.VApp.Name] = &VApp
	return nil
}

func (v *VcdPlatform) FindVApp(ctx context.Context, vappName string) (*govcd.VApp, error) {

	for name, vapp := range v.Objs.VApps {
		if vappName == name {
			return vapp.VApp, nil
		}
	}

	// Use GetAvailableQuery incase something new was created behind our backs.
	return nil, fmt.Errorf("Server does not exist")
}

// interesting Status values for a VApp
// 4 = POWERED_ON  - All vms in vapp are runing
// 9 = INCONSISTENT_STATE - Some vms on, some not
// 8 = POWERED_OFF (implies 1=RESOLOVED)
// 1 = RESOLVED  - vapp is created, but has no VMs yet.
//  see top of types.go

// Used to create the vdc/cloudlet VApp only. Just one external network for the Vapp using PrimaryNet
func (v *VcdPlatform) CreateVAppFromTmpl(ctx context.Context, vdc *govcd.Vdc, networks []*types.OrgVDCNetwork, vappTmpl govcd.VAppTemplate, storProf types.Reference, vmgp *vmlayer.VMGroupOrchestrationParams, description string, updateCallback edgeproto.CacheUpdateCallback) (*govcd.VApp, error) {
	var vapp *govcd.VApp
	var err error
	var vmRole vmlayer.VMRole

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVAppFromTmpl", "VAppName", vmgp.GroupName)
	// dumpVMGroupParams(vmgp, 1)

	vmparams := vmlayer.VMOrchestrationParams{}
	newVappName := vmgp.GroupName /* + "-vapp" 11/04 does removing this screw it up? should only be done for cloudlet*/

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

	// Do we already know this VApp? (XXX if multiple vdc's have a vapp with the same name, we'll pick the first found XXX )
	// So don't let that happen from our end
	//
	vapp, err = v.FindVApp(ctx, newVappName)
	if err != nil {
		fmt.Printf("\nComposeVapp /cloudlet named: %s\n", newVappName)
		// Not found try and create it
		fmt.Printf("\nCreateVAppFromTmpl-I-using tmpl %s\n", vappTmpl.VAppTemplate.Name)
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
			fmt.Printf("\nCreateVAppFromTemplate-I-compose ok, waiting for task completion\n")
			err = task.WaitTaskCompletion()
			if err != nil {

				fmt.Printf("\nCreateVAppFromTmpl-E-waiting for task complete %s\n", err.Error())
				return nil, err
			}

			vapp, err = v.Objs.PrimaryVdc.GetVAppByName(newVappName, true)
			if err != nil {
				fmt.Printf("Error : %s\n", err.Error())
				return nil, err
			}

			fmt.Printf("\nCreateVAppFromTmpl-I-vapp named: %s\n", vapp.VApp.Name)

			err = vapp.BlockWhileStatus("UNRESOLVED", 30) // upto seconds
			if err != nil {
				fmt.Printf("error waiting for created test vApp to have working state: %s", err.Error())
				return nil, err

			}
			task, err = vapp.RemoveAllNetworks()
			if err != nil {
				fmt.Printf("Error removing all networks: %s\n", err.Error())
			}
			err = task.WaitTaskCompletion()

			/*
				vdc,  = v.FindVdcVapp(ctx, newVappName) // .Objs.Vdc.GetVAppByName(newVappName, true)
				if err != nil {
					fmt.Printf("Error : %s\n", err.Error())
					return nil, nil
				}
			*/

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

				desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
					&types.NetworkConnection{
						IsConnected:             true,
						IPAddressAllocationMode: types.IPAllocationModePool,
						Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name, // types.NoneNetwork,
						NetworkConnectionIndex:  0,
					})
			}

			vmtmplName := vapp.VApp.Children.VM[0].Name
			fmt.Printf("Using existing VM in template Named: %s\n", vmtmplName)
			vm, err := vapp.GetVMByName(vmtmplName, false)
			if err != nil {
				fmt.Printf("error fetching y VM: %s", err.Error())
				return nil, err
			}
			// One or two networks, update our connection(s)
			err = vm.UpdateNetworkConnectionSection(desiredNetConfig)
			if err != nil {
				fmt.Printf("CreateVAppFromTemplate-E-UpdateNetworkConnnectionSection: %s\n", err.Error())
				return nil, err
			}

			fmt.Printf("\nCreateVAppFromTemplate task complete vapp %s status %s\n", vapp.VApp.Name, types.VAppStatuses[vapp.VApp.Status])

		}

		fmt.Printf("\nCreateVapp-I-about to split vapp %+v\n", vapp)
		// and in either case, add to our local cache, this will overwrite an entry with the same name XXX
		// Add new mex v.Vapp to our map

		// First cut, split vapp.VAppName by the first . and take the first half as the cloudletName
		cloudletNameParts := strings.Split(vapp.VApp.Name, ".")
		fmt.Printf("CreateVMs-I-split %s into two %d parts as: \n", vapp.VApp.Name, len(cloudletNameParts))
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
		}
		v.Objs.Cloudlets[vapp.VApp.Name] = &cloudlet

	} else {
		// We already know about this vapp
		vdc, err := v.FindVdcParent(ctx, vapp)
		if err != nil {
			fmt.Printf("CreateVMs-W-unable to find parent VDC for existing vapp %s\n", vapp.VApp.Name)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs cloudlet  exists", "cloudlet", vapp.VApp.Name, "vdc", vdc.Vdc.Name)
		fmt.Printf("\nCreateVMs-I-vapp(cloudlet) %s already exists on vdc %s\n\n", vapp.VApp.Name, vdc.Vdc.Name)
		// Refresh our local object
		vapp.Refresh()
	}

	return vapp, nil
}

func (v *VcdPlatform) populateINetConnectSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.NetworkConnectionSection, error) {
	// xlate orch -> netconnectsec
	//nc := &types.NetworkConnectionSection{}
	fmt.Printf("populateINetConnectSection-I-vm: %s role %s \n", vmparams.Name, vmparams.Role)
	vu.DumpOrchParamsVM(vmparams, 1)

	fmt.Printf("\n\npopulate Network connection section for VM %s role %s\n", vmparams.Name, vmparams.Role)
	//panic("popINetConnectionSection debug")
	//	return nc, nil
	return nil, nil
}

// Propteries make up the ProductSection and are retrievable in the VM
func makeProp(key, value string) *types.Property {
	prop := &types.Property{
		// We hard code UserConfigurable for now, as if false, it does not appear in the ovfenv fetched by vmtoolsd,
		// and what good is that to us?
		UserConfigurable: true,
		Type:             "string",
		Key:              key,
		Label:            key + "label",
		Value: &types.Value{
			Value: value,
		},
	}
	return prop
}

// pick off meta data to add to the VMs product section, where it will become available using
// vmware tools vmtoolsd --get for mobiledgex-init.sh
// Possible values used by our init script include:
//
// set_metadata_param HOSTNAME .name
//set_metadata_param UPDATE .meta.update
//set_metadata_param SKIPINIT .meta.skipinit
//set_metadata_param INTERFACE .meta.interface
//set_metadata_param ROLE .meta.role
//set_metadata_param SKIPK8S .meta.skipk8s
//set_metadata_param MASTERADDR .meta.k8smaster
//set_metadata_param UPDATEHOSTNAME .meta.updatehostname

//set_network_param IPADDR '.networks[0].ip_address'
//set_network_param NETMASK '.networks[0].netmask'
//set_network_param NETTYPE '.networks[0].type'

func (v *VcdPlatform) populateProductSection(ctx context.Context, vm *govcd.VM, vmparams *vmlayer.VMOrchestrationParams) (*types.ProductSectionList, error) {

	guestCustomSec, err := vm.GetGuestCustomizationSection()
	if err != nil {
		fmt.Printf("\nError Retrieving GuestCustom section: %s\n", err.Error())
		return nil, err
	}
	if !*guestCustomSec.Enabled {
		guestCustomSec.Enabled = vu.TakeBoolPointer(true)
		gcs, err := vm.SetGuestCustomizationSection(guestCustomSec)
		if err != nil {
			fmt.Printf("popProdSec-E-SetGuestCustomizationSectionFailed: %s\n", err.Error())
		}
		fmt.Printf("\nCustomSect enabled : %+v\n", gcs)

		guestCustomSec.AdminPassword = "2b|!2b-titq" // xxx
		// vault kv get -field=value secret/accounts/baseimage/password
	} else {
		fmt.Printf("\nVM %s guest custom already enabled\n\n", vm.VM.Name)
	}

	psl, err := vm.GetProductSectionList()
	if err != nil {
		fmt.Printf("\npopProdSection-E-from Get: %s\n\n", err.Error())
		return nil, err
	}
	if psl.ProductSection == nil {
		fmt.Printf("\n\t vm %s prod section nil, creating one\n\n", vm.VM.Name)
		psl = &types.ProductSectionList{
			ProductSection: &types.ProductSection{
				Info:     "Guest Properties",
				Property: []*types.Property{},
			},
		}
	}

	var props []*types.Property

	prop := makeProp("user-data", "encoded")
	if prop == nil {
		fmt.Printf("\n\t  make prop returns a nil prop!\n\n")
		return nil, fmt.Errorf("make prop error")
	}

	props = append(props, prop)

	fmt.Printf("\npopulate Product section setting Role = %s\n\n", vmparams.Role)
	role := vmparams.Role
	props = append(props, makeProp("ROLE", string(role)))
	skipk8s := vmlayer.SkipK8sYes
	if role == vmlayer.RoleMaster || role == vmlayer.RoleNode {
		skipk8s = vmlayer.SkipK8sNo
	}
	// If the role is RoleMaster, we need to make note of this VMs internal network address
	// as it must be passed to the workers for the join operation XXX
	// We could pass VM in here, and add some meta data to it to indicate on poweron
	// to get serverDetail, and then add the MASTERADDR proptery to all worker
	// vms in the given clusterInst. Well, these are not DHCP, we can know apirioi what
	// address we're giving to the master right?

	props = append(props, makeProp("SKIPK8S", string(skipk8s)))

	psl.ProductSection.Property = props
	// XXX what else?

	return psl, nil
}

// Add a new VM using our main template for each of vmgr.VMs of VMOrchestrationParams specified in the Group params into
// the pre-exsting vapp.
//
func (v *VcdPlatform) AddVMsToVApp(ctx context.Context, vapp *govcd.VApp, tmpl *govcd.VAppTemplate, vmgp *vmlayer.VMGroupOrchestrationParams) error {

	// XXX
	log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVapp", "Vapp", vapp, "template", tmpl)

	// vms := make(VMMap) we don't know if we have an existing vapp/map vs making a new one xxx

	// Don't forget the metadata for the ProductSection VMRole at a minimum.
	vmParams := vmgp.VMs
	for _, params := range vmParams {
		//netConnectSec, err := v.populateINetConnectSection(ctx, &params)
		fmt.Printf("\nAddVMsToVApp-I-adding vm named: %s role: %s \n", params.Name, params.Role)

		// add internal network, and add the new gateway interface to vapp

		// Rather than the work name, should return the network or 3rd octet value
		// And it should take the netconn section of vmgps
		/*
			_, err := v.CreateVappInternalNetwork(ctx, *vapp)
			if err != nil {
				fmt.Printf("\nAddVMsToVApp-E-CreateVappInternalNetwork : %s\n", err.Error())
				return err
			}
			tmpl.VAppTemplate.Name = params.Name
			task, err := vapp.AddNewVM(params.Name, *tmpl, netConnectSec, true)
			if err != nil {
				fmt.Printf("AddVMsToVApp-E-error AddNewVm: %s\n", err.Error())
				return err
			}
			err = task.WaitTaskCompletion()
			if err != nil { // fatal?
				fmt.Printf("AddVMsToVapp-E-error waiting for completion on AddNewVM: %s\n", err.Error())
			}
			vm, err := vapp.GetVMByName(params.Name, false)
			if err != nil {
				return err
			}
			err = vm.UpdateNetworkConnectionSection(netConnectSec)
			if err != nil {
				fmt.Printf("\nAddVMsToVapp-E- UpdateNetworkConnectionSection for %s failed : %s\n", vm.VM.Name, err.Error())
				return err
			}
			prodSecList, err := v.populateProductSection(ctx, vm, &params)

			status, serr := vm.GetStatus()
			if serr != nil {
				fmt.Printf("\nGetStatus of vm %s failed error: %s\n", vm.VM.Name, err.Error())
			} else {
				fmt.Printf("\nStatus of newly created vm %s : %s\n", vm.VM.Name, status)
			}
			_, err = vm.SetProductSectionList(prodSecList)

			// Take care of nominal Resource specs
			// We'll need to lookup flavorname in params and set the cpu/etc counts

			// we store our copy of vms we create in VMMap
			// XXX is our convenience map really convienent? XXX
			for name, Vapp := range v.Objs.VApps {
				if name == Vapp.VApp.VApp.Name {
					Vapp.VMs[params.Name] = vm
				}
			}
			// and add to our VApps.VApp.VMs map mapv.Objs.VApps.VMs[name] = vm
		*/
	}

	return nil

}

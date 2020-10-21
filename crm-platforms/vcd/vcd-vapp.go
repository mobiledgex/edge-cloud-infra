package vcd

import (
	"context"
	"fmt"

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

func (v *VcdPlatform) CreateVAppFromTmpl(ctx context.Context, networks []*types.OrgVDCNetwork, vappTmpl govcd.VAppTemplate, storProf types.Reference, vmgp *vmlayer.VMGroupOrchestrationParams, description string, updateCallback edgeproto.CacheUpdateCallback) (*govcd.VApp, error) {
	var vapp *govcd.VApp
	var err error
	var vmRole vmlayer.VMRole

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVAppFromTmpl", "VAppName", vmgp.GroupName)
	//	dumpVMGroupParams(vmgp, 1)

	vmparams := vmlayer.VMOrchestrationParams{}
	newVappName := vmgp.GroupName + "-vapp"
	haveExternalNet := false

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

	// Do we already know this VApp?
	vapp, err = v.FindVApp(ctx, newVappName)
	if err != nil {
		//fmt.Printf("CreateVAppFromTemplate-I-%s not found locally, creating it (does the system know about it?)\n", newVappName)
		existingVApp, err := v.Objs.Vdc.GetVAppByName(newVappName, false)
		if err == nil {
			fmt.Printf("CreateAppFromTemplate-E-newVappname %s exists in system not cache\n",
				existingVApp.VApp.Name)
			panic("GetVAppByName(newVappName-I-found existingVApp not in local cache")
		}
		// Not found try and create it
		task, err := v.Objs.Vdc.ComposeVApp(networks, vappTmpl, storProf, newVappName, description+vcdProviderVersion, true)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				// So we should have found this already, so this means it was created
				// behind our backs somehow, so we should add it to our pile of existing VApps...
				fmt.Printf("CreateRawVApp-W-VApp %s was not found locally, but Compose returns already exists. Add to local map\n", vmgp.GroupName)

			} else {
				// operation failed for resource reasons
				fmt.Printf("CreateVAppFromTemplate-E-Compose failed for %s error: %s\n", vmgp.GroupName, err.Error())
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateRawVApp failed", "VAppName", vmgp.GroupName, "error", err)
				return nil, err
			}
			// in any case, we have a new Vapp we need to cache locally and operate on.
			vapp, err = v.Objs.Vdc.GetVAppByName(vmgp.GroupName, true)
			if err != nil {
				fmt.Printf("Error composeing raw Vapp: %s\n", err.Error())
				return nil, nil
			}

		} else {
			fmt.Printf("\nCreateVAppFromTemplate-I-compose ok, waiting for task completion\n")
			task.WaitTaskCompletion()
			vapp, err = v.Objs.Vdc.GetVAppByName(newVappName, true)
			if err != nil {
				fmt.Printf("Error composeing raw Vapp: %s\n", err.Error())
				return nil, nil
			}
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

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 0
			// if the vm is to host the LB, we need

			if vmRole == vmlayer.RoleAgent { // Other cases XXX
				haveExternalNet = true
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

			// Add internal network to vapp (All vapps)
			//
			internalIdx := 0
			if haveExternalNet {
				internalIdx = 1
			}

			internalNetName, err := v.CreateVappInternalNetwork(ctx, *vapp)
			// v.GetNextInternalSubnet()
			desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
				&types.NetworkConnection{
					IsConnected:             true,
					IPAddressAllocationMode: types.IPAllocationModeManual,
					Network:                 internalNetName,
					NetworkConnectionIndex:  internalIdx,
					IPAddress:               "10.101.1.1", // we're the gateway
				})

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

		// and in either case, add to our local cache, this will overwrite an entry with the same name XXX
		// Add new mex v.Vapp to our map

		Vapp := VApp{
			VApp: vapp,
		}

		v.Objs.VApps[newVappName] = &Vapp

	} else {
		// Refresh our local object
		vapp.Refresh()
	}

	return vapp, nil
}

// VApp/template  related operations
// Called from PI CreateVMs, this creates the VApp container and networks
func (v *VcdPlatform) CreateRawVApp(ctx context.Context, vappName string /* vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback*/) (*govcd.VApp, error) {

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateRawVApp", "VAppName", vappName)
	//	dumpVMGroupParams(vmgp, 1)

	fmt.Printf("\nCreateVApp-I-named: %s \n", vappName)

	// Do we already know this VApp?
	vapp, err := v.FindVApp(ctx, vappName)
	if err != nil {
		fmt.Printf("CreateRawVApp-I-vapp %s not found locally, Composing Raw\n", vappName)
		// Not found try and create it
		err := v.Objs.Vdc.ComposeRawVApp(vappName)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				// So we should have found this already, so this means it was created
				// behind our backs somehow, so we should add it to our pile of existing VApps...
				fmt.Printf("CreateRawVApp-W-VApp %s was not found locally, but Compose returns already exists. Add to local map\n", vappName)

			} else {
				// operation failed for resource reasons
				fmt.Printf("CreateRawVApp-E-Compose failed for %s error: %s\n", vappName, err.Error())
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateRawVApp failed", "VAppName", vappName, "error", err)
				return nil, err
			}
			// in any case, we have a new Vapp we need to cache locally and operate on.
			vapp, err := v.Objs.Vdc.GetVAppByName(vappName, true)
			if err != nil {
				fmt.Printf("Error retrieving Vapp after successful compose : %s\n", err.Error())
				return nil, err
			}
			Vapp := VApp{
				VApp: vapp,
			}
			// Add new VApp to our map
			v.Objs.VApps[vappName] = &Vapp
		}

	} else {
		// Refresh our local object
		vapp.Refresh()
	}
	//

	vdcnet := v.Objs.PrimaryNet.OrgVDCNetwork
	IPScope := vdcnet.Configuration.IPScopes.IPScope[0] // xxx

	// AddOrgNetwork
	// Create DhcpSettings
	staticIPStart := IPScope.IPRanges.IPRange[0].StartAddress
	fmt.Printf("\nCreateRawVApp-I-dhcp range used: start %s to end  %s\n", v.IncrIP(IPScope.Gateway), v.DecrIP(staticIPStart))
	dhcpIPRange := types.IPRange{
		StartAddress: v.IncrIP(IPScope.Gateway),
		EndAddress:   v.DecrIP(staticIPStart),
	}

	var iprange []*types.IPRange
	iprange = append(iprange, IPScope.IPRanges.IPRange[0])

	dhcpsettings := govcd.DhcpSettings{
		IsEnabled: true,
		//	MaxLeaseTime:     7, // use the Orgs lease times no shorter
		//	DefaultLeaseTime: 7,
		IPRange: &dhcpIPRange,
	}

	externalSettings := govcd.VappNetworkSettings{
		ID:          vdcnet.ID,
		Name:        vdcnet.Name,
		Description: "external nat/dhcp",
		Gateway:     IPScope.Gateway,
		NetMask:     IPScope.Netmask,
		DNS1:        IPScope.DNS1,
		DNS2:        IPScope.DNS2,
		DNSSuffix:   IPScope.DNSSuffix,
		//		GuestVLANAllowed: true,    default is?
		StaticIPRanges: iprange,
		DhcpSettings:   &dhcpsettings,
	}
	interIPRange := types.IPRange{
		StartAddress: "10.101.1.2",
		EndAddress:   "10.101.1.22",
	}
	iprange[0] = &interIPRange
	internalSettings := govcd.VappNetworkSettings{
		ID:          vdcnet.ID,
		Name:        "vapp-internal1",
		Description: "internal 10.101.1.0/24 static",
		Gateway:     "10.101.1.1",
		NetMask:     "255.255.255.0",
		DNS1:        IPScope.DNS1,
		DNS2:        IPScope.DNS2,
		DNSSuffix:   IPScope.DNSSuffix,
		//		GuestVLANAllowed: true,    default is?
		StaticIPRanges: iprange,
	}

	// 1) isolated (internal) network, if orgNet = nil

	fmt.Printf("CreateRawVApp-I-create internal network\n")
	InternalNetConfigSec, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateVappNetwork Internal error %s", err.Error())
	}

	// 2) external network, with DHCP enabled  (hopefully)
	//
	fmt.Printf("CreateRawVApp-I-create External network\n")
	ExternalNetConfigSec, err := vapp.CreateVappNetwork(&externalSettings, vdcnet)
	if err != nil {
		return nil, fmt.Errorf("CreateVappNetwork External error %s", err.Error())
	}
	// where should we cache these config sections? Or do we need to?

	if vapp.VApp.NetworkConfigSection == nil {
		fmt.Printf("\nnew vapp.VApp has nil NetworkConfigSection aborting\n")
		panic("Nil NetworkConfigSection")

	}
	// this fails, as raw has no NetworkConfigSection I guess.
	if vapp.VApp.NetworkConfigSection == nil {
		panic("RawVApp has no NetworkConfigSection")
	}
	// If you use Raw, (no template) you need to add the nets yourself.
	// We've just uploaded the 4.0.4 ova, and made a template with PrimaryNet as Exteranl
	// so use that instead. XXX tomarrow, listening?

	fmt.Printf("\n\nExternalNetConfigSection: %+v\n\n", ExternalNetConfigSec)
	fmt.Printf("\n\nInternalNetConfigSection: %+v\n\n", InternalNetConfigSec)

	return vapp, nil

}

func (v *VcdPlatform) populateNetConnectSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.NetworkConnectionSection, error) {
	// xlate orch -> netconnectsec
	nc := &types.NetworkConnectionSection{}

	return nc, nil
}

// Build commonly used "ClusterInst" configurations.
// Results in an entry in our templates catalog with which we can
// create an instance, give update params to customize, and launch the result.
// xxx todo: add Update
func (v *VcdPlatform) CreateVAppTmpl(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (*govcd.VAppTemplate, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVAppTmpl", "GroupOrchParams", vmgp)
	vappTmpl := govcd.NewVAppTemplate(&v.Client.Client)
	var tmplParams *types.InstantiateVAppTemplateParams
	// Fill  tmplParms XXX
	v.Objs.Vdc.InstantiateVAppTemplate(tmplParams)
	return vappTmpl, nil
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

func (v *VcdPlatform) populateProductSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.ProductSectionList, error) {

	psl := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info:     "Guest Properties",
			Property: []*types.Property{},
		},
	}

	role := vmparams.Role
	psl.ProductSection.Property = append(psl.ProductSection.Property, makeProp("ROLE", string(role)))
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

	psl.ProductSection.Property = append(psl.ProductSection.Property, makeProp("SKIPK8S", string(skipk8s)))
	// XXX what else?
	return psl, nil
}

// Add a new VM using our main template for each of vmgr.VMs of VMOrchestrationParams specified in the Group params into
// the pre-exsting vapp.
//
func (v *VcdPlatform) AddVMsToVApp(ctx context.Context, vapp *govcd.VApp, vmgp *vmlayer.VMGroupOrchestrationParams) (*govcd.VApp, error) {

	tmplName := "mobiledgex-v4.0.4-vsphere" // image name to template, per-vm? eventually, it'll have to be. Right? XXX
	tmpl, err := v.FindTemplate(ctx, tmplName)
	if err != nil {
		fmt.Printf("Error from FindTemplate: %s\n", err.Error())
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVapp", "Vapp", vapp, "template", tmpl)

	// vms := make(VMMap) we don't know if we have an existing vapp/map vs making a new one xxx

	// Don't forget the metadata for the ProductSection VMRole at a minimum.
	vmParams := vmgp.VMs
	for _, params := range vmParams {

		netConnectSec, err := v.populateNetConnectSection(ctx, &params)
		prodSecList, err := v.populateProductSection(ctx, &params)

		task, err := vapp.AddNewVM(params.Name, *tmpl, netConnectSec, true)
		if err != nil {
			fmt.Printf("AddVMsToVApp-E-error AddNewVm: %s\n", err.Error())
			return nil, err
		}
		err = task.WaitTaskCompletion()
		if err != nil { // fatal?
			fmt.Printf("AddVMsToVapp-E-error waiting for completion on AddNewVM: %s\n", err.Error())
		}
		vm, err := vapp.GetVMByName(params.Name, false)
		if err != nil {
			return nil, err
		}
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
	}

	return vapp, nil

}

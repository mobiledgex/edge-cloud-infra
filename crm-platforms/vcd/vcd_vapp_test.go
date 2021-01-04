package vcd

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"

	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestDumpVappNetworks(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("TestDumpVAppNetworks...")
		vappName := "mex-cldlet3.tdg.mobiledgex.net-vapp"
		//vappName := "mex-vmware-vcd.tdg.mobiledgex.net-vapp"
		vapp, err := tv.FindVApp(ctx, vappName)
		if err != nil {
			fmt.Printf("%s not found\n", vappName)
			return
		}
		networkConfig, err := vapp.GetNetworkConfig()
		if err != nil {
			fmt.Printf("GetNetworkConfig failed: %s\n", err.Error())
			return
		}
		govcd.ShowVapp(*vapp.VApp)

		fmt.Printf("\n\n and DumpNetworkConfig: \n")
		vu.DumpNetworkConfigSection(networkConfig, 1)

		fmt.Printf("\n\n")
		govcd.ShowVapp(*vapp.VApp)
	} else {
		return
	}
}

// -vapp + -vdc
func TestRMVApp(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("testRMVappVApp")
		//err = testDestroyVApp(t, ctx, *vappName)
		err = testDeleteVApp(t, ctx, *vappName)
		if err != nil {
			fmt.Printf("Error deleteing %s : %s\n", *vappName, err.Error())
		}
		fmt.Printf("%s deleted\n", *vappName)
	} else {
		return
	}
}

// Test our CreateVApp using vmlayer args

func TestMexVApp(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("testCreateVM...")
		//vm := &govcd.VM{}

		var vols []vmlayer.VolumeOrchestrationParams
		vols = append(vols, vmlayer.VolumeOrchestrationParams{
			Name:               "Mex-vol1",
			ImageName:          "ubuntu-18.04",
			Size:               40,
			AvailabilityZone:   "none",
			DeviceName:         "disk1",
			AttachExternalDisk: false,
			UnitNumber:         1,
		},
		)

		port := vmlayer.PortOrchestrationParams{ // these can contain an array of SecurityGroups of type
			// ResourceReference that I suppose restrict access to port(s)
			// Name, ID string, Preexisting bool (if false, the ref is to be created as part of this op)
			// But as for securityGroup, we don't have enough info here to create it,
			// what is the source CIDR to allow/reject?

			Name: "mex-ports",
			// FixedIPs []FixedIPOrchestrationParams
			// SecurityGroups[]ResourceReference
		}

		// use fixed, maybe do dhcp
		// We should get a fixed IP from our tv.Obj.PrimaryNet
		// .51 or .52
		var fixedIps []vmlayer.FixedIPOrchestrationParams
		fixedIps = append(fixedIps, vmlayer.FixedIPOrchestrationParams{ // ...to VMs directy (or dhcp eh?)
			LastIPOctet: 2,
			Address:     "172.70.52.2",
			Mask:        "255.255.255.0",
			Subnet: vmlayer.ResourceReference{
				Name:        "",
				Id:          "",
				Preexisting: false,
			},
			Gateway: "172.70.52.1",
		},
		)
		//	cparams := chefmgmt.VMChefParams{}

		vmgp := vmlayer.VMGroupOrchestrationParams{}
		vmgp.GroupName = "mex-plat-vapp"
		vmgp.Ports = append(vmgp.Ports, port)

		vmparams := vmlayer.VMOrchestrationParams{

			Id:          "VMtestID",
			Name:        "MexVM1",
			Role:        vmlayer.RoleVMPlatform,
			ImageName:   "ubuntu-18.04",
			ImageFolder: "MEX-CAT01",
			HostName:    "MexVMHostName",
			DNSDomain:   "MexDomain",
			FlavorName:  "mex.medium",

			Vcpus:                   2,
			Ram:                     4092,
			Disk:                    40,
			ComputeAvailabilityZone: "nova", // xxx
			UserData:                "GuestCustomizeHere",
			MetaData:                "UserMetaData",
			SharedVolume:            false,
			AuthPublicKey:           "",
			DeploymentManifest:      "",
			Command:                 "",
			Volumes:                 vols,
			//		Ports:                   ports,
			FixedIPs:           fixedIps,
			AttachExternalDisk: false,
			//		ChefParams:              &cparams,
		}

		vmgp.VMs = append(vmgp.VMs, vmparams)

		//var updateCallback edgeproto.CacheUpdateCallback
		/*
			vapp, err := tv.CreateRawVApp(ctx, vmgp.GroupName)
			if err != nil {
				fmt.Printf("Error creating VApp: %s\n", err.Error())
			}
			vu.DumpVApp(vapp, 1)
		*/
	} else {
		return
	}
}

// return a canned Item reproduced from a 2 Nic VM, this being the internal
// Network Nic type VMXNET3 (3rd Gen high perf vSphere vnic)
// So, perhaps we should pass in the existing VirtualHardwareSection, and append
// this Item to VirtualHawareItem.Connection
//
func GetVirtHwItem(t *testing.T, ctx context.Context) types.VirtualHardwareItem {
	var connections []*types.VirtualHardwareConnection
	//	connections := []&types.VirtualHardwareConnection{}
	connection := types.VirtualHardwareConnection{

		IPAddress:         "10.101.1.1",
		PrimaryConnection: false,
		IpAddressingMode:  "MANUAL",
		NetworkName:       "mex-vmware.tdg.mobiledgex.net-vappinternal-1",
	}

	connections = append(connections, &connection)

	item := types.VirtualHardwareItem{

		XMLName: xml.Name{

			Space: "http://schemas.dmtf.org/ovf/envelope/1",
			Local: "Item",
		},
		ResourceType:        10,
		ResourceSubType:     "VMXNET3",
		ElementName:         "Network adapter 1",
		Description:         "Vmxnet3 ethernet adapter on \"mex-vmware.tdg.mobiledgex.net-vappinternal-1\"",
		InstanceID:          2,
		AutomaticAllocation: true,
		Address:             "00:00:00:00:00:00", // should be auto-assigned
		AddressOnParent:     1,
		AllocationUnits:     "",
		Reservation:         0,
		VirtualQuantity:     0,
		Weight:              0,
		CoresPerSocket:      0,
		Connection:          connections,
		HostResource:        nil,
		Link:                nil,
	}
	return item
}

// need -vdc
func TestShowVApp(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestShowVApp-Start show vapp named %s\n", *vappName)
		vdc := tv.Objs.Vdc
		if err != nil {
			fmt.Printf("vdc %s not found\n", *vdcName)
			return
		}
		_, err = vdc.GetVAppByName(*vappName, false)
		require.Nil(t, err, "GetVAppByName")

		fmt.Printf("All VMs discovered:\n")
		for name, _ := range tv.Objs.VMs {
			fmt.Printf("next vm: %s\n", name)
		}
		for name, _ := range tv.Objs.TemplateVMs {
			fmt.Printf("next vm: %s\n", name)
		}

		//govcd.ShowVapp(*vapp.VApp)
	} else {
		return
	}
}

// This option builds it all from scratch, but requires a media object that is our
// v4.0.4.vmdk file, which I can't upload by itself... <sigh> We need an iso for that.
// Which we probably will end up doing to prove we can have a vm with OrgVdcNetwork + isolated VApp networks.
// First, we'll prove the concept using the ubuntu-18.04 asset we have in our catalog

// This follows the example in vm_test.go, which uses 2 vdcOrgNetworks, so we'll first prove that out,
// and assuming it works fine, try and modify for our internal network.
// We'll create a vappName'd vapp
// uses -vapp
func TestRawVApp(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		vdc := tv.Objs.Vdc
		fmt.Printf("TestVApp-Start create vapp named %s\n", *vappName)

		// 1) create raw vapp, and 2) add networks:
		err = vdc.ComposeRawVApp(*vappName)
		require.Nil(t, err, "vdc.CreateRawVapp")

		vapp, err := vdc.GetVAppByName(*vappName, false)
		require.Nil(t, err, "GetVAppByName")

		govcd.ShowVapp(*vapp.VApp)

		// Retrive our template, we need the HREF of it's VM
		tmpl, err := tv.FindTemplate(ctx, *tmplName)
		require.Nil(t, err, "FindTemplate")
		childvm := tmpl.VAppTemplate.Children.VM[0]
		// 3) and a new vm w/template and VmGeneralParams (Change the vm name)
		// The only type that includes this is SourcedCompositionItemParam
		//

		IParams := &types.InstantiationParams{
			//		CustomizationSection:         *CustomizationSection         `xml:"CustomizationSection,omitempty"`
			//	DefaultStorageProfileSection *DefaultStorageProfileSection `xml:"DefaultStorageProfileSection,omitempty"`
			//	GuestCustomizationSection    *GuestCustomizationSection    `xml:"GuestCustomizationSection,omitempty"`
			//	LeaseSettingsSection         *LeaseSettingsSection         `xml:"LeaseSettingsSection,omitempty"`
			//	NetworkConfigSection         *NetworkConfigSection         `xml:"NetworkConfigSection,omitempty"`
			//	NetworkConnectionSection     *NetworkConnectionSection     `xml:"NetworkConnectionSection,omitempty"`
			//	ProductSection               *ProductSection               `xml:"ProductSection,omitempty"`
		}
		// NetworkAssignment maps a network name specified in a Vm to the network name of a
		// vApp network defined in the VApp that contains the Vm
		netMappings := []*types.NetworkAssignment{}
		netMapping := &types.NetworkAssignment{
			InnerNetwork:     "local vm network name",
			ContainerNetwork: "Name of Vapp network to map to",
		}
		netMappings = append(netMappings, netMapping)

		vmGeneralParams := types.VMGeneralParams{
			Name: *vmName,
		}
		childRef := &types.Reference{
			HREF: childvm.HREF,
			Type: childvm.Type,
			ID:   childvm.ID,
			Name: childvm.Name,
		}
		// empty winds default
		storProf := &types.Reference{}
		locParamsRef := &types.LocalityParams{}

		// sourceItem found in InstantiateVAppTemplateParams, ReComposeVAppParams, ComposeVAppParams
		vmSourceItem := types.SourcedCompositionItemParam{
			SourceDelete:      false, // don't delete the childvm after copy
			Source:            childRef,
			VMGeneralParams:   &vmGeneralParams,
			VAppScopedLocalID: *vappName + "vm 1",

			//If Source references a Vm this can include any of the following OVF sections:
			// VirtualHardwareSection OperatingSystemSection NetworkConnectionSection GuestCustomizationSection.
			InstantiationParams: IParams,

			// If Source references a Vm, this element maps a network name specified in the
			// Vm to the network name of a vApp network defined in the composed vApp.
			NetworkAssignment: netMappings,
			StorageProfile:    storProf,

			// LocalityParams represents locality parameters.
			// Locality parameters provide a hint that may help the placement engine optimize placement
			// of a VM with respect to another VM or an independent disk.
			LocalityParams: locParamsRef,
		}

		fmt.Printf("vmSourceItem: %+v\n", vmSourceItem)
		guestCustom := &types.GuestCustomizationSection{}
		connections := []*types.NetworkConnection{}

		//connectionSection := childvm.NetworkConnectionSection

		netConnect := &types.NetworkConnectionSection{
			XMLName:                       xml.Name{},
			Xmlns:                         types.XMLNamespaceVCloud,
			Ovf:                           types.XMLNamespaceOVF,
			Info:                          " info ",
			HREF:                          childvm.HREF,
			Type:                          "10", // XXX VMXNET3 type thing.
			PrimaryNetworkConnectionIndex: 0,
			NetworkConnection:             connections,
			//		Link :
		}
		vmSpec := &types.VmSpecSection{}

		bootMedia := &types.Media{}

		createItem := &types.CreateItem{
			Name:                      *vmName,
			Description:               "test vm with 2 nics",
			GuestCustomizationSection: guestCustom,
			NetworkConnectionSection:  netConnect,
			VmSpecSection:             vmSpec,
			BootImage:                 bootMedia,
		}

		reComposeParams := types.RecomposeVAppParamsForEmptyVm{
			XMLName:          xml.Name{},
			XmlnsVcloud:      types.XMLNamespaceVCloud,
			XmlnsOvf:         types.XMLNamespaceOVF,
			CreateItem:       createItem,
			AllEULAsAccepted: true,
		}

		// before we resort to recreating our vm , try to update our existing section
		vm, err := vapp.AddEmptyVm(&reComposeParams)

		cpuCount := 4
		vmSpecSection := &types.VmSpecSection{
			NumCpus: &cpuCount,
		}

		vm, err = vm.UpdateVmSpecSection(vmSpecSection, "testing")
		if err != nil {
			fmt.Printf("error UpdateVmSpecSection: %s\n", err.Error())
		}
		/*
			err = testDestroyVApp(t, ctx, vappName)
			if err != nil {
				fmt.Printf("TestVApp-E-error deleting vapp: %s\n", err.Error())
				return
			}
		*/
		// now try recompsoing the Vapp...
	} else {
		return
	}
}

func testCreateVAppChild() (*types.VAppTemplate, error) {
	//netConfig := []VAppNetworkConfiguratin{}
	tmpl := &types.VAppTemplate{}
	/*
		netConfig = append(netconfig, types.VAppNetworkConfiguratin {
			Configuration :
		}
		// is just another vapptemplate

		tmpl := govcd.NewVappTemplate {
			Name: "mex-platform-vm-1",
			NetworkConfigSection :    *types.NetworkConfigSection {
				NetworkConfig : NetConfig
			},
	*/
	/*
		NetworkConnectionSection: *types.NetworkConnectionSection {


		},
	*/

	return tmpl, nil
}

//
// from scratch vs  using an existing template.
//
func testPopulateVappTmpl(t *testing.T, ctx context.Context, tmplName string) *govcd.VAppTemplate {
	cli := tv.Client.Client
	tmpl := govcd.NewVAppTemplate(&cli)

	fmt.Printf("PopuldateVappTemplate, must have VAppTemplateChildren !- nil and networks not nil\n")

	tmpl.VAppTemplate.Name = tmplName

	return tmpl
}

// For our platform LB, this net
// Alternatively, you can vapp.UpateNetwork(newNetworkSettings, orgNetwork)
// How to get from NetworkConfigSection to it's OrgVDCNetowrk?
func createInternalNetwork(t *testing.T, ctx context.Context, vapp *govcd.VApp) (*types.OrgVDCNetwork, error) {

	fmt.Printf("createInternalNetwork-I-creating for vapp %s in state %s\n", vapp.VApp.Name, types.VAppStatuses[vapp.VApp.Status])

	net := &types.OrgVDCNetwork{}
	IPScope := types.IPScope{
		IsInherited: false,
		Gateway:     "10.101.1.1",
		Netmask:     "255.255.255.0",
		//		DNS1: "
		//DNS2:
		//DNSSuffix
		//IsEnabled:
		//IPRanges: // for static pool allocation in the network
		//AllocatedIPAddresses readonly list
		//SubAllocations readonly list

	}
	//	dumpIPScopes(IPScope, 1)

	interIPRange := types.IPRange{
		StartAddress: "10.101.1.1",
		EndAddress:   "10.101.1.22",
	}
	var iprange []*types.IPRange
	iprange = append(iprange, &interIPRange)

	// This guy is the 10 dot gateway
	internalSettings := govcd.VappNetworkSettings{
		//		ID:          vdcnet.ID,
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

	InternalNetConfigSec, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateVappNetwork Internal error %s", err.Error())
	}

	// how to wait here?
	fmt.Printf("what to do with InternalNetConfigSec?: %+v\n", InternalNetConfigSec)
	return net, nil
}

// test compose w/template
func testCreateVApp(t *testing.T, ctx context.Context, vappName string) (*govcd.VApp, error) {

	// Prototype how this is going to work using
	// Compose VApp with template etc... (as opposed to ComposeRawVApp
	// To do this using our work routines, we'd need a GroupOrchestration params obj.
	//
	vdc := tv.Objs.Vdc
	fmt.Printf("testCreateVApp-I-ComposeRawVApp for %s\n", vappName)

	networks := []*types.OrgVDCNetwork{}
	networks = append(networks, tv.Objs.PrimaryNet.OrgVDCNetwork)

	/*
	           createInternalNetwork is clearly wrong, error composing vapp: error instantiating a new vApp: API Error: 400: The reference  cannot be parsed correctly. Reason: no server authority--- FAIL: TestVApp (17.95s)

	   	isolated, err := createInternalNetwork(t, ctx, vapp)
	   	if err != nil {
	   		fmt.Printf("createIsolatedNetwork failed : %s\n", err.Error())
	   	}
	   	networks = append(networks, isolated)
	*/
	tmpl := govcd.VAppTemplate{}
	//	vAppTemplate := &types.VAppTemplate{}
	// Leaving empty wins the default vSan Default, which is our only stoprofile...
	storageProfileRef := types.Reference{}
	tmplName := "clusterVapp1-tmpl"
	// we should fetch a refresh of this template in case our cached object is stale.
	cat := tv.Objs.PrimaryCat
	QueryRes, err := cat.QueryCatalogItemList()
	if err != nil {
		fmt.Printf("QueryCatalogItemList failed: %s\n", err.Error())
		return nil, nil
	}
	for _, item := range QueryRes {
		if item.EntityType == "vapptemplate" && item.EntityName == tmplName {
			fmt.Printf("testCreateVApp-I-using template name: %s\n", item.EntityName)
			catItem, err := cat.GetCatalogItemByHref(item.HREF)
			if err != nil {
				fmt.Printf("GetItemByHref %s failed: %s\n", item.HREF, err.Error())
				return nil, nil
			}
			tmpl, err = catItem.GetVAppTemplate()
			if err != nil {
				fmt.Printf("GetVAppTemplate-E-%s\n", err.Error())
				return nil, nil
			}
			fmt.Printf("About to compose using template:\n %+v\n", tmpl)
		}
	}
	NewName := vappName + "-vapp"
	task, err := vdc.ComposeVApp(networks, tmpl, storageProfileRef, NewName, "mex platform role", true)
	if err != nil {
		fmt.Printf("error composing vapp: %s", err)
		panic("ComposeVApp")
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		fmt.Printf("error composing vapp: %s", err)
		panic("AddToCleanupList")
	}
	fmt.Printf("\nDone waiting for Compose of %s task status: %s\n", NewName, task.Task.Status)

	// Get VApp
	vapp, err := vdc.GetVAppByName(vappName, true)
	if err != nil {
		fmt.Printf("error getting vapp: %s", err)
		panic("GetVAppByName")
	}

	err = vapp.BlockWhileStatus("UNRESOLVED", 10)
	if err != nil {
		fmt.Printf("error waiting for created test vApp to have working state: %s", err)
		panic("BlockWhileStatus (unresolved)")
	}

	fmt.Printf("\ttestCreateVApp-I-Composed or VApp and got this raw vapp:\n")
	govcd.ShowVapp(*vapp.VApp)

	// get a test vm
	// network, err := createTestNetwork, or v.Objs.PrimaryNet ?
	tvm, err := testCreateVM(t, ctx) // from vcd_vm_test.go, does use our vcd_vm:createVM work routine,
	if err != nil {
		fmt.Printf("Failed to create TestVM\n")
	}
	newNetwork := &types.NetworkConnectionSection{}
	// netwok *NetworkConnectionSection will customize what's in the template, if supplied.
	// (i.e., it can be nil or emtpy and it will use network bits from the template.
	// I suppose most of the time, we'll just change addresss ?
	// Leave it nil at first, see what happens.
	//
	// Also, from the code, vappTemplate.VAppTemplateStatus needs == 8, or we'll win
	// return Task{}, fmt.Errorf("vApp Template shape is not ok (status: %d)", vappTemplate.VAppTemplate.Status)

	task, err = vapp.AddNewVM(tvm.VM.Name, tmpl, newNetwork, true)
	if err != nil {
		fmt.Printf("Error Adding new vm to vapp err: %s\n", err.Error())
	}
	err = task.WaitTaskCompletion()

	// dumpVApp(vapp.VApp, 1)
	govcd.ShowVapp(*vapp.VApp)
	fmt.Printf("NewVapp: %v\n", vapp)
	// can we validate completeness?
	// Could we power on or do we need to add media to our vm?
	err = testDestroyVapp(t, ctx, NewName)
	if err != nil {
		fmt.Printf("Error deleting test vapp: %s\n", NewName)
	}

	return vapp, err
}

// Powering off vs Undeplpy
// Undeploy is what you want if your aim is to delete the vapp
// So for instance DeleteVMs where vmgp.GroupName represents a cluster you
// want deleted. Not going to power it back on sometime later.
// Doesn't matter if the vapp is current powered on or off, this will delete it.
//
func testDeleteVApp(t *testing.T, ctx context.Context, name string) error {
	vdc := tv.Objs.Vdc
	vapp, err := vdc.GetVAppByName(name, true)
	if err != nil {
		fmt.Printf("testDestroyVApp-E-error Getting Vapp %s by name: %s\n", name, err.Error())
		return err
	}
	status, err := vapp.GetStatus()
	fmt.Printf("Vapp %s currently in state: %s\n", name, status)
	if err != nil {
		fmt.Printf("Error fetching status for vapp %s\n", name)
		return err
	}
	// If the vapp is already powered off, it may be deleteled directy
	// else, first undeploy
	if status == "POWERED_ON" {
		task, err := vapp.Undeploy()

		if err != nil {
			fmt.Printf("Error from vapp.Undploy the vapp  as : %s\n", err.Error())
			return err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			fmt.Printf("Error waiting undeploy of the vapp first %s \n", name)
			return err
		}
		fmt.Printf("vapp  undeployed...\n")
	}
	fmt.Printf("Call vapp.Delete()\n")
	task, err := vapp.Delete()
	if err != nil {
		fmt.Printf("vapp.Delete failed: %s\n current status: %s\n", err.Error(), status)
		return err
	}
	err = task.WaitTaskCompletion()

	fmt.Printf("VApp %s Deleted.\n", name)
	return nil
}

func testUpdateVApp(t *testing.T, ctx context.Context) {

}

func testAddNetworksToVApp(t *testing.T, ctx context.Context) {

}

func testInsertMediaToVApp(t *testing.T, ctx context.Context) {

}

// Add vm from vapptemplate with a custom networkConnectionSection
//
func testAddVMToVApp(t *testing.T, ctx context.Context, netName string, vapp *govcd.VApp, tmpl govcd.VAppTemplate, network *types.NetworkConnectionSection) (*govcd.VApp, error) {

	//storRef := types.Reference{}
	task, err := vapp.AddNewVM(netName, tmpl, network, true)
	if err != nil {
		fmt.Printf("Error Adding second VM with internal net only err: %s\n", err.Error())
		return vapp, nil
	}
	fmt.Printf("Task for AddNewVM: %+v\n", task)
	return vapp, nil
}

func testPowerOnVApp(t *testing.T, ctx context.Context, vapp *govcd.VApp) error {

	task, err := vapp.PowerOn()
	if err != nil {
		fmt.Printf("Error powering on vapp %s err: %s\n", vapp.VApp.Name, err.Error())

	}
	fmt.Printf("\nTask powerOn: %+v\n", task)
	return err
}

func testPowerOffVApp(t *testing.T, ctx context.Context, vapp *govcd.VApp) error {

	task, err := vapp.PowerOff()
	if err != nil {
		fmt.Printf("Error powering off vapp %s err: %s\n", vapp.VApp.Name, err.Error())

	}
	fmt.Printf("\nTask powerOff: %+v\n", task)
	return err
}

func getCreateItemForInternalNetwork(t *testing.T, ctx context.Context) *types.CreateItem {
	ci := &types.CreateItem{}

	return ci
}

func createVmSpecWithNewNet(t *testing.T, ctx context.Context, vapp *govcd.VApp, vm *govcd.VM) *types.VmSpecSection {
	// The section(s) we're interested in changing
	// This will work for nominal resources, not network though
	//sec := &types.VmSpecSection{}
	sec := vm.VM.VmSpecSection
	cpus := 4
	sec.NumCpus = &cpus
	fmt.Printf("Set VmSpecSection to change cpu count\n")

	fmt.Printf("createVmSpecWithNewNet-I-here's our current VmSpecSection: \n")

	vu.DumpVmSpecSection(sec, 1)
	return sec
}

// compose raw again, but uses a tmpl creating the vm needs -vapp and -vm
func TestRaw2Nic(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {

		vdc := tv.Objs.Vdc

		err = vdc.ComposeRawVApp(*vappName)
		require.Nil(t, err, "ComposeRawVApp")

		vapp, err := vdc.GetVAppByName(*vappName, true)
		require.Nil(t, err, "GetVAppByName")

		err = vapp.BlockWhileStatus("UNRESOLVED", 10)

		//	networks := []*types.OrgVDCNetwork{}
		//	networks = append(networks, v.Objs.PrimaryNet.OrgVDCNetwork)

		//_ /*NetConfigSection*/, err = vapp.AddOrgNetwork(&govcd.VappNetworkSettings{}, tv.Objs.PrimaryNet.OrgVDCNetwork, false)
		//task, err := vapp.AddNetworkConfig([]*types.OrgVDCNetwork{tv.Objs.PrimaryNet.OrgVDCNetwork})
		if err != nil {
			fmt.Printf("BlockWhileStatus unresolved k-E-error: %s\n", err.Error())
			return
		}
		//task.WaitTaskCompletion()

		// Ok, now vapp_test.go continues with adding a new vm to this vapp.
		tmpl, err := tv.FindTemplate(ctx, *tmplName)
		require.Nil(t, err, "FindTemplate")

		// can we use vapp as is to create internal network?

		//------------------------ from vcd-vapp.go --------------------------------------------
		vdcnet := tv.Objs.PrimaryNet.OrgVDCNetwork

		IPScope := vdcnet.Configuration.IPScopes.IPScope[0] // xxx

		// AddOrgNetwork
		// Create DhcpSettings
		staticIPStart := IPScope.IPRanges.IPRange[0].StartAddress
		fmt.Printf("\nCreateRawVApp-I-dhcp range used: start %s to end  %s\n", tv.IncrIP(ctx, IPScope.Gateway, 1), tv.DecrIP(ctx, staticIPStart, 1))
		/*
			dhcpIPRange := types.IPRange{
				StartAddress: tv.IncrIP(IPScope.Gateway),
				EndAddress:   tv.DecrIP(staticIPStart),
			}

		*/
		var iprange []*types.IPRange
		iprange = append(iprange, IPScope.IPRanges.IPRange[0])

		/*
			dhcpsettings := govcd.DhcpSettings{
				IsEnabled: true,
				//	MaxLeaseTime:     7, // use the Orgs lease times no shorter
				//	DefaultLeaseTime: 7,
				IPRange: &dhcpIPRange,
			}
		*/
		/*
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
				StaticIPRanges:   iprange,
				DhcpSettings:     &dhcpsettings,
				VappFenceEnabled: takeBoolPointer(true),
			}
		*/
		interIPRange := types.IPRange{
			StartAddress: "10.101.1.1",
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
		_ /*InternalNetConfigSec*/, err = vapp.CreateVappNetwork(&internalSettings, nil)
		if err != nil {
			fmt.Printf("CreateVappNetwork internal failed: %s\n", err.Error())
			return
		}

		netConnectSection := &types.NetworkConnectionSection{}
		netConnectSection.PrimaryNetworkConnectionIndex = 0

		netConnectSection.NetworkConnection = append(netConnectSection.NetworkConnection,
			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual,
				Network:                 "vapp-internal1",
				NetworkConnectionIndex:  1,
			})

		task, err := vapp.AddNewVM(*vmName, *tmpl, netConnectSection, true)
		if err != nil {
			fmt.Printf("Error from AddNewVm: %s\n", err.Error())
			return
		}
		status, err := vapp.GetStatus()
		fmt.Printf("Test2NicRaw-I-complete status of %s is %s\n", vapp.VApp.Name, status)
		err = task.WaitTaskCompletion()

		fmt.Printf("\n\n\tTest2NicRaw-I-complete\n")
	} else {
		return
	}
}

// You can go the route of ComposeRawVApp (no template) and build up everything
// and create recompose params to use in newvapp.AddEmptyVm() but we'll need our
// base image imported to our catalog and add that along with DISK
// This routine explores how to win external and internal nics using
// ComposeVapp(tmpl) + modifying our existing VM that's in the template.
// But trying the obvious vapp.ChangeNetworkConfig() gives a 500 internal erver error >sigh< stupid crap.
// Instead, try using VmSpecSection in vm.UpdateVmSpecSection(vmspec, ...)

func createTestVapp(t *testing.T, ctx context.Context, vappName, tmplName string) (*govcd.VApp, error) {

	// Populate OrgVDCNetwork
	var networks []*types.OrgVDCNetwork

	vdc := tv.Objs.Vdc

	tmpl, err := tv.FindTemplate(ctx, tmplName)
	require.Nil(t, err, "FindTemplate")

	networks = append(networks, tv.Objs.PrimaryNet.OrgVDCNetwork)
	// Get StorageProfileReference
	storRef := types.Reference{}
	/*
		storageProfileRef := &types., err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
		if err != nil {
			return nil, fmt.Errorf("error finding storage profile: %s", err)
		}
	*/
	// Compose VApp
	task, err := vdc.ComposeVApp(networks, *tmpl, storRef, vappName, "description", true)
	if err != nil {
		return nil, fmt.Errorf("error composing vapp: %s", err)
	}
	// Get VApp
	err = task.WaitTaskCompletion()
	vapp, err := vdc.GetVAppByName(vappName, true)
	if err != nil {
		return nil, fmt.Errorf("error getting vapp: %s", err)
	}

	err = vapp.BlockWhileStatus("UNRESOLVED", 10)
	if err != nil {
		return nil, fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}

	return vapp, nil

}

func validateNetworkConfigSettings(networkSettings *govcd.VappNetworkSettings) error {
	if networkSettings.Name == "" {
		return fmt.Errorf("network name is missing")
	}

	if networkSettings.Gateway == "" {
		return fmt.Errorf("network gateway IP is missing")
	}

	if networkSettings.NetMask == "" {
		return fmt.Errorf("network mask config is missing")
	}

	if networkSettings.NetMask == "" {
		return fmt.Errorf("network mask config is missing")
	}

	if networkSettings.DhcpSettings != nil && networkSettings.DhcpSettings.IPRange == nil {
		return fmt.Errorf("network DHCP ip range config is missing")
	}

	if networkSettings.DhcpSettings != nil && networkSettings.DhcpSettings.IPRange.StartAddress == "" {
		return fmt.Errorf("network DHCP ip range start address is missing")
	}
	return nil
}

// support -tmpl templateName -vapp VAppName to this test case
// uses createTestVapp which does a compose w/tmpl
// needs vapp, tmpl, and vm args...
//
func Test2Nic(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("\nTest2Nic using vapp: %s tmpl: %s vm: %s \n", *vmName, *tmplName, *vmName)

		// uses Primary Net Only
		vapp, err := createTestVapp(t, ctx, *vappName, *tmplName)
		if err != nil {
			fmt.Printf("\n\tcreateTestVapp-E-%s\n", err.Error())
			return
		}
		// remove the vm
		if vapp.VApp.Children != nil {
			if len(vapp.VApp.Children.VM) > 1 {
				// pull out any existing VM
				childVm := vapp.VApp.Children.VM[0]
				vm, err := vapp.GetVMByName(childVm.Name, false)
				if err != nil {
					fmt.Printf("Error from GEtVMByName: %s\n", err.Error())
					return
				}

				fmt.Printf("Deleting existing VM in vapp: %s\n", vm.VM.Name)
				err = vapp.RemoveVM(*vm)
				if err != nil {
					fmt.Printf("Error from vapp.RemoveVm: %s\n", err.Error())
					return
				}
			} else {
				fmt.Printf("Our new vapp has no vms\n")
			}

		} else {
			fmt.Printf("Our new vapp has nil children\n")
		}

		// set the vapp networks into our Vapp...

		internalNetName, err := setVappInternalNetwork(t, ctx, *vapp)
		if err != nil {
			fmt.Printf("seems we failed setVappNetworks: %s\n", err.Error())
			return
		}
		// We now have two networks in vapp.
		fmt.Print("Added networks to vapp, dump vapp.NetworkConfigSection\n")
		//dumpNetworkConfigSection(vapp.VApp.NetworkConfigSection, 1)

		desiredNetConfig := &types.NetworkConnectionSection{}
		desiredNetConfig.PrimaryNetworkConnectionIndex = 0
		desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,

			/* so if we use compose w/tmpl we already get a vapp network right?  So leave this out here

			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeDHCP,
				Network:                 tv.Objs.PrimaryNet.OrgVDCNetwork.Name,
				NetworkConnectionIndex:  0,
			},
			*/
			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual, // Pool,
				Network:                 internalNetName,
				NetworkConnectionIndex:  1,
			},
		)

		// again, empty prof wins default (SAN)
		// storageProfileRef := types.Reference{}
		// Book keeping for Media and Disk
		// Why can't we use what's in the template here ( disk ok, but meida is not there!?)
		media := &govcd.Media{}
		mediaName := "ubuntu-18.04"
		for name, m := range tv.Objs.Media {
			if name == mediaName {
				media = m // XXX cat.GetMediaByName(vcd.config.Media.Media, false)
				break
			}
		}
		fmt.Printf("using media\n \tName: %s\n\tHREF: %s\n\tID:%s\n", media.Media.Name, media.Media.HREF, media.Media.ID)

		newDisk := types.DiskSettings{
			AdapterType:       "5",
			SizeMb:            int64(16384),
			BusNumber:         0,
			UnitNumber:        0,
			ThinProvisioned:   vu.TakeBoolPointer(true),
			OverrideVmDefault: true}

		requestDetails := &types.RecomposeVAppParamsForEmptyVm{

			CreateItem: &types.CreateItem{
				Name: *vmName,
				//NetworkConnectionSection:  desiredNetConfig,
				Description:               "2 nics net2 and internalC",
				GuestCustomizationSection: nil,

				VmSpecSection: &types.VmSpecSection{
					//Modified:          takeBoolPointer(true),
					Info:              "Virtual Machine specification",
					OsType:            "debian10Guest",
					NumCpus:           vu.TakeIntAddress(2),
					NumCoresPerSocket: vu.TakeIntAddress(1),
					CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
					MemoryResourceMb:  &types.MemoryResourceMb{Configured: 1024},
					MediaSection:      nil,
					DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
					HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"}, // need support older version vCD
					VmToolsVersion:    "",
					//VirtualCpuType:   "VM32",
					TimeSyncWithHost: nil,
				},

				BootImage: &types.Media{HREF: media.Media.HREF, Name: media.Media.Name, ID: media.Media.ID},
			},
			AllEULAsAccepted: true,
		}

		// getting internal server errors, let's remove all vms from vapp first:

		createdVm, err := vapp.AddEmptyVm(requestDetails)
		if err != nil {
			fmt.Printf("AddEmptyVM-E- %s\n", err.Error())
			return
		}
		fmt.Printf("AddEmptyVm returns createVM as: %+v\n", createdVm)

		_, err = vapp.GetVMByName(createdVm.VM.Name, false)
		if err != nil {
			fmt.Printf("Error Getting vapp.GetVMByName for vapp  %s err: %s\n", vapp.VApp.Name, err.Error())
		}

		fmt.Printf("\n\n\tTest2Nic complete\n\n")
		// power on the vm

		//	dumpVM(vm.VM, 1)
		/* here we're setting the flavor basically
		vmchanges := createVmSpecWithNewNet(t, ctx, vapp, vm)
		vm, err = vm.UpdateVmSpecSection(vmchanges, "testing")
		if err != nil {
			fmt.Printf("error UpdateVmSpecSection: %s\n", err.Error())
		}

		connectionSection, err := vm.GetNetworkConnectionSection()
		dumpNetworkConfigSection(vapp.VApp.NetworkConfigSection, 1)
		fmt.Printf("VM %s has network connection section as: \n", vm.VM.Name)
		// The vm has no connectionSection yet.
		dumpVirtualHardwareSection(vm.VM.VirtualHardwareSection, 1)
		dumpNetworkConnectionSection(connectionSection, 1)
		*/
	} else {
		return
	}
}

func setVappInternalNetwork(t *testing.T, ctx context.Context, vapp govcd.VApp) (string, error) {

	var iprange []*types.IPRange

	fmt.Printf("\nsetVappNetworks for 2nics\n")
	IPScope := tv.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0]

	// this guy (LB)  needs both external and internal or GetIPFromServerDetail will fail as it
	// demands two networks

	interIPRange := types.IPRange{
		// Gateway can't be in this range.
		StartAddress: "10.101.1.2",
		EndAddress:   "10.101.1.22",
	}
	iprange = append(iprange, &interIPRange)
	//		iprange[0] = &interIPRange
	internalNetName := vapp.VApp.Name + "-internal-1"
	internalSettings := govcd.VappNetworkSettings{
		Name:        internalNetName,
		Description: "internal 10.101.1.0/24 static",
		Gateway:     "10.101.1.1", // use the scheme found in vmgp XXX
		NetMask:     "255.255.255.0",
		DNS1:        IPScope.DNS1,
		DNS2:        IPScope.DNS2,
		DNSSuffix:   IPScope.DNSSuffix,
		//		GuestVLANAllowed: true,    default is?
		StaticIPRanges: iprange,
	}
	status, err := vapp.GetStatus()
	if err != nil {
		fmt.Printf("setVappInternalNetwork-E-error obtaining status of vapp: %s\n", err.Error())
		return "", err
	}
	if status == "UNRESOLVED" {
		fmt.Printf("setVappNetworks-I-wait up to 10 sec  while  unresolved \n")
		err = vapp.BlockWhileStatus("UNRESOLVED", 30) // Raw 10 is enough but Compose takes longer
		if err != nil {
			fmt.Printf("BlockWhile return err: %s\n", err.Error())
		}
		status, _ = vapp.GetStatus()
		fmt.Printf("Continue from blockwhile status now %s\n", status)
	}
	fmt.Printf("setVappInternalNetwork-I-create internal network, vapp status: %s \n", status)
	InternalNetConfigSec, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		fmt.Printf("setVappNetworks-E-create internal net: %s\n", err.Error())
		return "", fmt.Errorf("CreateVappNetwork Internal error %s", err.Error())
	}
	// Do we need to update?
	fmt.Printf("\n\nInternalNetConfigSection: %+v\n\n", InternalNetConfigSec)
	vapp.Refresh() // needed?
	return internalNetName, nil

}

// Our external network orgvcdnetwork we'll use DHCP, and create vapp network direct connect
// So, we need a NetworkConfig element in the NetowrkConfigSection of our InstaciationParams
// For routed and directly connected networks, the ParentNetwork element contains a ref to the OrgVDCNetwork
// that the VappNetwork connects to. For direct FenceMod bridged. Or nateRouted to specify a routed connection
// controlled by  network features such as NateService or FirewallService... So try both
//
// This did work on a composedVapp though. (see vdc-vapp.go)
func setVappExternalNetwork(t *testing.T, ctx context.Context, vapp govcd.VApp) (string, error) {
	fmt.Printf("setVappExternalNetwork\n")
	vdcnet := tv.Objs.PrimaryNet.OrgVDCNetwork
	IPScope := vdcnet.Configuration.IPScopes.IPScope[0] // xxx

	// AddOrgNetwork
	// Create DhcpSettings
	//staticIPStart := IPScope.IPRanges.IPRange[0].StartAddress
	//fmt.Printf("\nSetVappExternalNet: dhcp range used: start %s to end  %s\n", tv.IncrIP(IPScope.Gateway), tv.DecrIP(staticIPStart))
	// start with bridged, and then proceed to add isFenced and checkout Nat/Firewall rules

	/*
		dhcpIPRange := types.IPRange{
			StartAddress: tv.IncrIP(IPScope.Gateway),
			EndAddress:   tv.DecrIP(staticIPStart),
		}
	*/
	/*  testing for new mex-net03 that already has a range set...
	var iprange []*types.IPRange
	iprange = append(iprange, IPScope.IPRanges.IPRange[0])

	dhcpsettings := govcd.DhcpSettings{
		IsEnabled: true,
		//	MaxLeaseTime:     7, // use the Orgs lease times no shorter
		//	DefaultLeaseTime: 7,
		IPRange: &dhcpIPRange,
	}
	*/
	externalNetName := vdcnet.Name // vapp.VApp.Name + "-external"
	vappNetSettings := &govcd.VappNetworkSettings{
		Name:             externalNetName,
		ID:               vdcnet.ID,
		Description:      "external nat/dhcp",
		Gateway:          IPScope.Gateway,
		NetMask:          IPScope.Netmask,
		DNS1:             IPScope.DNS1,
		DNS2:             IPScope.DNS2,
		DNSSuffix:        IPScope.DNSSuffix,
		GuestVLANAllowed: vu.TakeBoolPointer(false),
		//		StaticIPRanges:   iprange,
		//		DhcpSettings:     &dhcpsettings,
		//	VappFenceEnabled: takeBoolPointer(false),
	}

	// Add our external network as a vapp network, bridged or Nat'ed to our PrimaryNet
	// bridged, false turns fenceMode from bridged to Nat (True here wins only direct and isolated allowed for
	// our org... hmm...
	//
	_ /*netConfigSec,*/, err := vapp.AddOrgNetwork(vappNetSettings, vdcnet, false)

	if err != nil {
		fmt.Printf("\tError UpdateNetwork to vapp: %s\n", err.Error())
		return "", err
	}
	/*
		a, err := vapp.GetNetworkConfig()
		if err != nil {
			fmt.Printf("\tError GetNetworkConfig: %s\n", err.Error())
			return "", err
		}

		VappNetConfiguration := &types.VAppNetworkConfiguration{
			NetworkName:   "mex-net03",
			Description:   "stupid",
			Configuration: &types.NetworkConfiguration{},
		}

		fmt.Printf("\nsetVappExternalNet-I-netConfigSec : %+v\n\n", VappNetConfiguration)

		//	netConfig := netConfigSec.NetworkConfig.Configuration

		// Now ask the question, IsVappNetwork?
		if !govcd.IsVappNetwork(VappNetConfiguration.Configuration) {
			fmt.Printf("setExternal-I-vdcnet.Name is NOT a VappNetwork\n")
		} else {
			fmt.Printf("setExternal-I-vdcnet.Name IS a VappNetwork\n")
		}



		for _, vappNetConfiguration := range a.NetworkConfig {
			fmt.Printf("\tnext network: %s \n", vappNetConfiguration.NetworkName)
			if vappNetConfiguration.NetworkName == externalNetName {
				fmt.Printf("%s has parent network as %s features:\n", externalNetName, vappNetConfiguration.Configuration.ParentNetwork.Name)
				dumpNetworkFeatures(vappNetConfiguration.Configuration.Features, 1)
			}
		}
		// Now list out the features available from
		// netConfigSec.VappNetworkConfiguration.Configuration.
	*/
	return externalNetName, nil

}

// -vapp -net
func TestExtAddrVApp(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		vapp, err := tv.FindVApp(ctx, *vappName)
		require.Nil(t, err, "FindVapp")
		fmt.Printf("TestVApp-Start create vapp named %s in vdc %s \n", *vappName, *vdcName)

		addr, err := tv.GetExtAddrOfVapp(ctx, vapp, *netName)

		if err != nil {
			fmt.Printf("error from GetExtAddrOfVapp : %s\n", err.Error())
			return
		}
		fmt.Printf("Vapp %s has external address as %s\n", *vappName, addr)
	}
}

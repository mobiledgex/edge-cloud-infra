package vcd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test what instanciate does, should be the simple creation of a vapp from the template
// So we  use -vapp and -tmpl args
func TestInstanciateTmpl(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestInstancitate tmplName %s vappName %s\n", *tmplName, *vappName)
		tmpl, err := tv.FindTemplate(ctx, *tmplName)
		require.Nil(t, err, "FindVappTemplate")

		tmplRef := &types.Reference{
			HREF: tmpl.VAppTemplate.HREF,
			ID:   tmpl.VAppTemplate.ID,
			Type: tmpl.VAppTemplate.Type,
			Name: tmpl.VAppTemplate.Name,
		}
		// create a minimal InstanctiateVAppTemplateParams
		netConfig := popNetConfig(t, ctx)
		netConnect := popNetConnect(t, ctx)

		IParams := &types.InstantiationParams{
			NetworkConfigSection:     netConfig,
			NetworkConnectionSection: netConnect,
		}
		tmplParams := &types.InstantiateVAppTemplateParams{

			Name:                *vappName,
			PowerOn:             false,
			Source:              tmplRef,
			InstantiationParams: IParams,
			// SourcedItem *SourcedCompositionItemParam see if we need this here, probably used in Compose/ReCompose not instanciate
			AllEULAsAccepted: true, // takeBoolPointer(true),
		}

		err = tv.Objs.Vdc.InstantiateVAppTemplate(tmplParams)
		if err != nil {
			fmt.Printf("InstantiateVApptemplate-E-error: %s\n", err.Error())
			return
		}
		vapp, err := tv.Objs.Vdc.GetVAppByName(*vappName, true)
		if err != nil {
			fmt.Printf("GetVappByName-E-%s\n", err.Error())
			return
		}
		status, err := vapp.GetStatus()
		fmt.Printf("VApp %s has status : %s\n", *vappName, status)
	} else {
		return
	}
}

func popNetConfig(t *testing.T, ctx context.Context) *types.NetworkConfigSection {

	// This is the guy with the IPScopes /Features
	// *Note SubInterface and DistributedInterface here, they are mutually exclusive
	// When both are nil, the internal (default) interface is  used.
	vdcnet := tv.Objs.PrimaryNet.OrgVDCNetwork
	var ipscopes *types.IPScopes = vdcnet.Configuration.IPScopes

	//	Ipscope := vdcnet.Configuration.IPScopes.IPScope[0]

	//	ipscopes = append(ipscopes, Ipscope)

	netConfig := &types.NetworkConfiguration{
		IPScopes: ipscopes,
		ParentNetwork: &types.Reference{
			HREF: vdcnet.HREF,
			ID:   vdcnet.ID,
			Type: vdcnet.Type,
			Name: vdcnet.Name,
		},
	}
	var vappNetConfigs []types.VAppNetworkConfiguration

	vappNetConfig := types.VAppNetworkConfiguration{
		// create unique name for our new Vapp network
		NetworkName:   "vapp-" + vdcnet.Name + "-network",
		Configuration: netConfig, // *types.NetworkConfiguration
	}
	vappNetConfigs = append(vappNetConfigs, vappNetConfig)
	config := &types.NetworkConfigSection{

		NetworkConfig: vappNetConfigs,
	}

	networkNames := config.NetworkNames()

	for _, name := range networkNames {
		fmt.Printf("popNetConfig next network %s\n", name)
	}
	return config
}

func popNetConnect(t *testing.T, ctx context.Context) *types.NetworkConnectionSection {
	vdcnet := tv.Objs.PrimaryNet.OrgVDCNetwork
	//		Network: "vapp-"+vdcnet.Name+"-network", // Name of the network to which this NIC is connected

	var netConnections []*types.NetworkConnection
	netConnection := &types.NetworkConnection{
		Network:                "vapp-" + vdcnet.Name + "-network",
		NeedsCustomization:     false,
		NetworkConnectionIndex: 0,
		//		IPAddress:
		//	ExternalIPAddress:
		IsConnected:             true,
		MACAddress:              "00.00.00.00.00",
		IPAddressAllocationMode: types.IPAllocationModeDHCP,
		NetworkAdapterType:      "E1000", // VMXNET3 ?
	}
	netConnections = append(netConnections, netConnection)
	connects := &types.NetworkConnectionSection{

		PrimaryNetworkConnectionIndex: 0,
		NetworkConnection:             netConnections,
	}
	return connects
}

// To create a VAppTemplate from scratch, you must first create a VApp.
// When you then add this newly create VApp to a catalog, implictly, we are creating a
// VAppTemplate from the VApp.
// We can then  use this template to create other VApps, and just modify bits like the networkConnection section
// (Change its name and ip address, + metadata like guest-info /Role etc a
// We could then just delete the original VApp

func TestTmpl(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestTmpl (have %d vms in v.Objs.VMs", len(tv.Objs.VMs))
		//	tmplName = "mobiledgex-v4.0.4-tmpl" // -vsphere"

		for name, vm := range tv.Objs.VMs {
			fmt.Printf("test = have vm %s name: %s \n", vm.VM.Name, name)
		}

		tmpl, err := tv.FindTemplate(ctx, *tmplName)
		if err != nil {
			fmt.Printf("TestTmpl-E-%s not found locally\n", *tmplName)
		}
		// for dumping internal vms in the template we'll need our local test cache objs.
		dumpVAppTemplate(&tv, ctx, tmpl, 1)
	} else {
		return
	}
}

func testVAppTemplate(t *testing.T, ctx context.Context) {

	// verify we have templates in tv.Objs.VAppTmpls
	for tname, t := range tv.Objs.VAppTmpls {
		fmt.Printf("next tmpl: %s\n", tname)
		fmt.Printf("\t%+v\n", t)
	}
	vappName := "test-vapp1"
	// test compose a vapp from template
	err := testComposeVapp(t, ctx, vappName)
	require.Nil(t, err, "testComposeVapp testvapp1")
	//testCreateVAppTmpl(t, ctx)

	err = testDestroyVapp(t, ctx, vappName)
	require.Nil(t, err, "TestDestroyVapp")

}

func populateInstantiationParams() *types.InstantiationParams {

	custSec := &types.CustomizationSection{
		GoldMaster:             false,
		CustomizeOnInstantiate: false,
	}

	//GuestCustomizationSection contains settings for VM customization like admin password, SID
	// changes, domain join configuration, etc
	guestCustSec := &types.GuestCustomizationSection{}

	// empty will inherit defaults of our Org, we can only reduce the times
	leaseSettingSec := &types.LeaseSettingsSection{}

	vappNetConfigSec := &types.NetworkConfigSection{}

	//vappNetConfigSec = tv.Objs.PrimaryNet.OrgVDCNetwork.Configuration

	netConnectSec := &types.NetworkConnectionSection{}
	prodSec := &types.ProductSection{}
	instParams := &types.InstantiationParams{
		CustomizationSection:      custSec,
		GuestCustomizationSection: guestCustSec,
		LeaseSettingsSection:      leaseSettingSec,
		NetworkConfigSection:      vappNetConfigSec,
		NetworkConnectionSection:  netConnectSec,
		ProductSection:            prodSec,
	}
	return instParams
}

// create and return instantitation params, given a sourced itme, which can itself be
// vApp, VappTemplate, or a Vm.
//
func populateVAppTmplInstatiationParams(t *testing.T, ctx context.Context) *types.InstantiateVAppTemplateParams {

	tmplParams := &types.InstantiateVAppTemplateParams{}
	return tmplParams

}

// from govcd.vapptemplate.go
func testCreateVAppTmpl(t *testing.T, ctx context.Context) {

	//TestTmp := &govcd.VAppTemplate{}
	TestTmpl := govcd.NewVAppTemplate(&tv.Client.Client)

	fmt.Printf("\nTestTmp: %+v\n", TestTmpl)

	//tmplParams := populateVAppTmplInstatiationParams(t, ctx)
	//err := tv.Objs.Vdc.InstantiateVAppTemplate(tmplParams)
	// this returns 405 method not allowed.
	// require.Nil(t, err, "InstatiateVAppTemplate")
	fmt.Printf("testCreateVAppTmpl-I-vdc resource entities now:\n")
	vu.DumpVdcResourceEntities(tv.Objs.Vdc.Vdc, 1)
	// should be able to ask for vappTemplateByName now eh?
}

// Delete the name VAppTemplate from the catalog
// Anything else needs to be done?
// catalogItem.Delete()
// So must frist get the catitem for this templ name.
func testDestroyVAppTmpl(t *testing.T, ctx context.Context, tmplname string) error {
	cat := tv.Objs.PrimaryCat

	catitem, err := cat.GetCatalogItemByName(tmplname, true)
	if err != nil {
		fmt.Printf("testDestroyVAppTmpl-E-error finding %s item in cat: %s\n", tmplname, cat.Catalog.Name)
	}
	err = catitem.Delete()
	if err != nil {
		fmt.Printf("Error catitem.Delete() on %s from %s as: %s \n", tmplname, cat.Catalog.Name, err.Error())
		return err
	}
	return nil
}

// create template with two networks using tv.Objs.PrimaryNet / dhcp
func testCreatePlatformTemplate() {

}

// create template with only internal isolated netowrk using static 10.101.x.[1, 10, 101, 102..]
// Q: How is x selected today? (openstack?)

func testCreateInternalTemplate() {

}

func testUpdateVAppTmpl(t *testing.T, ctx context.Context) {

}

func testAddNetworksToVAppTmpl(t *testing.T, ctx context.Context) {

}

func testInsertMediaToVAppTmpl(t *testing.T, ctx context.Context) {

}

// recompose uploaded mobiledgex-v3.1.6-v14-vapp.ovf to use our networks
//
func testComposeVapp(t *testing.T, ctx context.Context, vappName string) error {

	targetTmpl := &govcd.VAppTemplate{}

	vappDesc := "recomposed mex BI"
	for tmplName, tmpl := range tv.Objs.VAppTmpls {
		fmt.Printf("Checking for target tmpl: %s\n", tmplName)
		if strings.Contains(tmplName, "mobiledgex") {
			targetTmpl = tmpl
			fmt.Printf("Using tmplate %s\n", tmplName)
			break
		}
	}
	fmt.Printf("targetTmpl = %+v\n", targetTmpl.VAppTemplate)
	// now get the actual govcd.VM by name of this VM, we have a recordtype
	// Looks like we need to create a new vapp from our template, and using that vapp, call vapp.RemoveNetwork and
	// maybe vapp.UpdateOrgNetwork
	// vapp.ChangeNetworkConfig(netowrks, ip string)
	stoRef := types.Reference{}
	// get template object by name
	// Need a Query to ge storageProfiles

	// Yes, we know this apriori, but need to find it dynamically
	// item? type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"
	//defStorPol := "vSan Default Stroage Policy"
	query := &types.QueryResultRecordsType{}
	storRef, err := tv.Objs.Vdc.GetDefaultStorageProfileReference(query)
	if err != nil {
		fmt.Printf("Error from GetDefaultStorageProfileReference : %s\n", err.Error())
	} else {
		fmt.Printf("Default Storage Profile for Vdc : Name %s Type %s\n", storRef.Name, storRef.Type)
	}

	// So we're going to use vdc.ComposeVApp(networks, template, storageRef, name, accept all
	network := []*types.OrgVDCNetwork{}
	network = append(network, tv.Objs.PrimaryNet.OrgVDCNetwork)
	// comments in vapptempl indicate that if stoRef not found, it will use the default, which is ok for now.
	task, err := tv.Objs.Vdc.ComposeVApp(network, *targetTmpl, stoRef, vappName, vappDesc, true)
	if err != nil {
		fmt.Printf("ComposeVApp-E-%s\n", err.Error())
		return err
	}
	fmt.Printf("Task: %+v\n", task)
	// should we turn around and verify?
	vapp, err := tv.Objs.Vdc.GetVAppByName(vappName, true)
	if err != nil {
		fmt.Printf("GetByName %s failed: %s\n", vappName, err.Error())
		return err
	}
	fmt.Printf("Composed Vapp:\n")
	vu.DumpVApp(vapp, 1)
	// or did we need to remove the old network first?
	return err
}

// AddNewVM Adds VM from VApp template with custom NetworkConnectionSection
// So the VApp we've just composed, add a second VM with just an internal network
func (v *VcdPlatform) testAddVMToVAppTmpl(t *testing.T, ctx context.Context, vapp *VApp, network *types.NetworkConnectionSection) {

}

// test for bug handle removing networks along with VM if the vm removed is last one using network.
func testRemoveVmFromVApp(t *testing.T, ctx context.Context, vapp govcd.VApp, vmName string) (*govcd.VApp, error) {

	vm, err := vapp.GetVMByName(vmName, false)
	if err != nil {
		fmt.Printf("Error retriving vm name %s from vapp %s err: %s\n", vmName, vapp.VApp.Name, err.Error())
		return nil, err
	}
	err = vapp.RemoveVM(*vm)
	if err != nil {
		fmt.Printf("error removing vm %s\n", err.Error())
	}
	return &vapp, err
	// check if bug #252 is present: https://github.com/vmware/go-vcloud-director/issues/252
}

func testDestroyVapp(t *testing.T, ctx context.Context, vappName string) error {

	return nil
}

func dumpVAppTemplateChildren(tv *VcdPlatform, ctx context.Context, tc *types.VAppTemplateChildren, indent int) {
	fill := strings.Repeat("  ", indent)
	if tc == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	for _, vmt := range tc.VM {
		vat := govcd.VAppTemplate{
			VAppTemplate: vmt,
		}
		dumpVAppTemplate(tv, ctx, &vat, indent+1)
	}
}

func dumpVAppTemplate(tv *VcdPlatform, ctx context.Context, vt *govcd.VAppTemplate, indent int) {

	fill := strings.Repeat("  ", indent)
	if vt == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"Name", vt.VAppTemplate.Name)
	fmt.Printf("%s %s\n", fill+"HREF", vt.VAppTemplate.HREF)
	fmt.Printf("%s %s\n", fill+"Type", vt.VAppTemplate.Type)
	if vt.VAppTemplate.Type == "application/vnd.vmware.vcloud.vm+xml" {

		// we can fetch it from vCD or from our local cache if we've done things right
		vm, err := tv.FindVM(ctx, vt.VAppTemplate.Name)
		if err != nil {
			fmt.Printf("Failed to find vm %s locally\n", vt.VAppTemplate.Name)
		} else {
			vu.DumpVM(vm.VM, indent+1)
		}

	}
	fmt.Printf("%s %s\n", fill+"ID", vt.VAppTemplate.ID)
	fmt.Printf("%s %s\n", fill+"Name", vt.VAppTemplate.OperationKey)

	fmt.Printf("%s %d\n", fill+"Status", vt.VAppTemplate.Status)

	fmt.Printf("%s %s\n", fill+"OvfDescriptorUploaded", vt.VAppTemplate.OvfDescriptorUploaded)

	fmt.Printf("%s %t\n", fill+"GoldMaster", vt.VAppTemplate.GoldMaster)
	// Link
	fmt.Printf("%s %s\n", fill+"Description", vt.VAppTemplate.Description)
	fmt.Printf("%s %s\n", fill+"ID", vt.VAppTemplate.ID)
	// Tasks

	fmt.Printf("%s\n", fill+"Files:")
	vu.DumpFilesList(vt.VAppTemplate.Files, indent+1)
	fmt.Printf("%s\n", fill+"Owner:")
	vu.DumpOwner(vt.VAppTemplate.Owner, indent+1)

	fmt.Printf("%s\n", fill+"Tmpl Children:") // , vt.VAppTemplate.ID)
	dumpVAppTemplateChildren(tv, ctx, vt.VAppTemplate.Children, indent+1)

	fmt.Printf("%s %s\n", fill+"VAppScopedLocalID", vt.VAppTemplate.VAppScopedLocalID)
	fmt.Printf("%s %s\n", fill+"DefaultStorageProfile", vt.VAppTemplate.DefaultStorageProfile)
	fmt.Printf("%s %s\n", fill+"Created         ", vt.VAppTemplate.DateCreated)

	fmt.Printf("%s\n", fill+"NetworkConfigSection:")
	vu.DumpNetworkConfigSection(vt.VAppTemplate.NetworkConfigSection, indent+1)

	fmt.Printf("%s\n", fill+"NetworkConnectionSection:")
	vu.DumpNetworkConnectionSection(vt.VAppTemplate.NetworkConnectionSection, indent+1)
	fmt.Printf("%s\n", fill+"LeaseSettingsSection:")
	vu.DumpLeaseSettingSection(vt.VAppTemplate.LeaseSettingsSection, indent+1)

	fmt.Printf("%s\n", fill+"CustomizationSection:")
	vu.DumpCustomizationSection(vt.VAppTemplate.CustomizationSection, indent+1)
}

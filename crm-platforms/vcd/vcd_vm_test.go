package vcd

import (
	"context"
	"flag"
	"fmt"
	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"testing"
)

// Init a copy of our platform for test
// Available to, and used by all the other unit tests
var tv VcdPlatform

// cmd line arg vars available to all tests
var vmName = flag.String("vm", "default-vm-name", "Name of vm")
var vappName = flag.String("vapp", "default-vapp-name", "name of vapp")
var tmplName = flag.String("tmpl", "default-template-name", "Name of template")
var netName = flag.String("net", "default-network", "Name of network")
var ipAddr = flag.String("ip", "172.70.52.210", "Defafult IP addr of VM")
var ovaName = flag.String("ova", "basic.ova", "name of ova file to upload")
var vdcName = flag.String("vdc", "mex01", "name of vdc")
var grpName = flag.String("grp", "grp-default", "some grp name")
var livetest = flag.String("live", "false", "live or canned data")

// Unit test env init. We have two cases, the default is live=false making
// it safe for inclusion in our make unit-test.
func InitVcdTestEnv() (bool, context.Context, error) {
	var live bool = false
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	//tv.initDebug(o.VMProperties.CommonPf.PlatformConfig.NodeMgr) // XXX needed now?
	// make our object maps
	tv.Objs.Nets = make(map[string]*govcd.OrgVDCNetwork)
	tv.Objs.Cats = make(map[string]CatContainer)

	tv.Objs.VApps = make(map[string]*VApp)
	tv.Objs.VMs = make(map[string]*govcd.VM)
	tv.Objs.VAppTmpls = make(map[string]*govcd.VAppTemplate)

	//v.Objs.TemplateVMs = make(map[string]*types.QueryResultVMRecordType)
	tv.Objs.TemplateVMs = make(TmplVMsMap)
	tv.Objs.Media = make(MediaMap)

	if *livetest == "true" {
		live = true
		fmt.Printf("\tPopulateOrgLoginCredsFromEnv\n")
		tv.PopulateOrgLoginCredsFromEnv(ctx, "mex-cldlet1") // need to move to first physicalname reference (vault key lookup not env)

		//fmt.Printf("\tMaps made, GetClient\n")
		client, err := tv.GetClient(ctx, tv.Creds)
		if err != nil {
			return live, ctx, fmt.Errorf("Unable to create Vcd Client %s\n", err.Error())
		}
		tv.Client = client
		err = tv.ImportDataFromInfra(ctx)
		if err != nil {
			return live, ctx, fmt.Errorf("ImportDataFromInfra failed: %s", err.Error())
		}

		fmt.Printf("TestEnvInit live org: %s Complete\n", tv.Objs.Org.Org.Name)

	} else {
		// anything other than a manual run providing "true" for flag "live" results
		// in canned data for unit tests.
		client, err := GetDummyClient(ctx)
		tv.Client = client
		err = importTestData(ctx)
		if err != nil {
			fmt.Printf("Error initiaizing test data: %s\n", err.Error())
		}
		fmt.Printf("TestEnvInit dead Complete\n")
	}

	return live, ctx, nil
}

func importTestData(ctx context.Context) error {

	return nil
}

func GetDummyClient(ctx context.Context) (*govcd.VCDClient, error) {
	client := &govcd.VCDClient{}
	return client, nil

}

// -vapp -vm
func TestShowVM(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")

	if live {
		vapp, err := tv.FindVApp(ctx, *vappName)
		if err != nil {
			fmt.Printf("vapp %s not found\n", *vappName)
			return
		}

		vm, err := vapp.GetVMByName(*vmName, false)
		//vm, err := tv.FindVM(ctx, *vmName)
		require.Nil(t, err, "FindVM")
		vu.DumpVM(vm.VM, 1)
		return
	} else {
		return
	}
}

// needs -vm
func TestVMMetrics(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")
	if live {
		fmt.Printf("TestVMMetric Start\n")

		govcd.ShowOrg(*tv.Objs.Org.Org)

		// So supposedly, there are potentially 4 Links on a VM that point to historic / current usage
		// mentric of all catagories. But only historic are available from a Powered Off vm, which makes sense
		// to get current metrics (and historic) the VM must be on.
		//  We try both... Current Metrics can be negitive if the value found is invalid.
		fmt.Printf("GetMetrics of VM powered OFF\n")
		err = testVMMetrics(t, ctx, *vmName, false)
		if err != nil {

			fmt.Printf("Error from testVMMetrics: %s\n", err.Error())
		}

		err = testVMMetrics(t, ctx, *vmName, true)
		if err != nil {

			fmt.Printf("Error from testVMMetrics: %s\n", err.Error())
		}

		tv.SetPowerState(ctx, *vmName, vmlayer.ActionStop)
	} else {
		return
	}
}
func TestAddVMNetwork(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")
	if live {
		fmt.Printf("TestAddVMNetwork Start\n")

		serverName := "mex-plat"
		subnetName := ""
		PortName := ""
		ipAddr := ""
		//action := vmlayer.ActionCreate // not used [create,  update, delete]
		// We wish to determine if we can add a new nic/newtwork to an existing VM if it's powered On or
		// if it must be off first.
		powerState := true // powered on
		err = testAttachPortToServer(t, ctx, serverName, subnetName, PortName, ipAddr, powerState)
		if err != nil {
			fmt.Printf("Error AttachPOrtToServer  serverName %s , ipAddr %s  err %s", serverName, ipAddr, err.Error())
			return
		}
	} else {
		return
	}
}

// This one just creates a VM using native vdc.
// Instaitation Params for a vm can include:
//
// VirtualHardwareSection // Note: changes to most item elements in VHS are ignored by composeVApp operations
// GuestCustomizationSection Hostname admin passwd etc
// OperatingSystemSection
// ProductSection // Role, MasterAddr, all the bits used by mobiledgex-init.sh in ovfenv
// NetworkConnectionSection
//

func TestVM(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")
	if live {
		// You need an AdminOrg object if you want to get at OrgSettings.
		org := tv.Objs.Org.Org
		vdc := tv.Objs.Vdc

		fmt.Printf("TestVM-VMQuota: %d NicQuota %d\n", vdc.Vdc.VMQuota, vdc.Vdc.NicQuota)

		cli := tv.Client

		adminOrg, err := govcd.GetAdminOrgByName(cli, org.Name)
		if err != nil {
			fmt.Printf("Error retrieving AdminOrg: %s\n", err.Error())
		} else {
			generalSettings := adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings
			fmt.Printf("TestVM-I-Org DeployedVMQuota: %d CanPublishCats: %t StoredVMQuota: %d \n",
				generalSettings.DeployedVMQuota, generalSettings.CanPublishCatalogs, generalSettings.StoredVMQuota)
		}
		fmt.Printf("cli current Api version: %s\n", cli.Client.APIVersion)

		// look for any available VMs?
		vmRecords, err := tv.GetAvailableVMs(ctx)
		if err != nil {
			fmt.Printf("\n\nError GetAvailableVMs: %s\n", err.Error())
		}
		if vmRecords != nil {
			fmt.Printf("\nGetAvailableVms len: %d\n", len(vmRecords))
			for _, vrec := range vmRecords {
				fmt.Printf("next vrec: %+v\n", vrec)
			}

		}
		AVdc, err := adminOrg.GetAdminVdcByName(vdc.Vdc.Name)
		if err != nil {
			fmt.Printf("Error retrievning Admin Vdc %s\n", err.Error())
		} else {
			fmt.Printf("\n------------------AdminVdc: %+v\n", AVdc.AdminVdc)
			//adminVdc := AVdc.AdminVdc
			//fmt.Printf("\nadminVdc.VdcStorageProfiles: %+v\n", adminVdc.VdcStorageProfiles)

		}

		//fmt.Printf("adminVdc: VmDiscoveryEnabled: %t, isElastic: %t Resource %d\n", adminVdc.VmDiscoveryEnabled, *adminVdc.IsElastic,
		//	adminVdc.ResourceGuaranteedCpu)
	} else {
		return
	}
}

// This one uses vmlayer vm orch params and vdc-vm.go::CreateVM work routine
func TestMexVM(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")
	if live {
		testCreateVM(t, ctx)

		//tv.testInsertMediaToVM(t, ctx)
		//tv.testAddNetworksToVM(t, ctx)

		// ...
		//tv.testDestroyVM(t, ctx)
	}

}

// -vapp and -vm
func TestRMVM(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")
	if live {
		vm, err := tv.FindVM(ctx, *vmName)
		if err != nil {
			fmt.Printf("VM %s not found\n", *vmName)
			return
		}
		status, err := vm.GetStatus()
		fmt.Printf("Vapp %s currently in state: %s\n", *vappName, status)
		if err != nil {
			fmt.Printf("Error fetching status for vapp %s\n", *vappName)
			return
		}

		if status == "POWERED_ON" {
			task, err := vm.PowerOff()
			if err != nil {
				fmt.Printf("testDestroyVapp-W-power off failed : %s\n", err.Error())
				return
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				fmt.Printf("Error powering of the Vapp %s \n", *vappName)
				return
			}
		}
	}
}

// Test add remove, or remove + add as the case may be
// We find if we simply use VmSettings and UpdateVmSpecSection or
// vm.UpdateInternalDisks we win error can't modify disk on vm with snapshots
// Now, we don't have any snapshots, so the only other thought (besides a bug)
// is that vm sharing across vapps is tantamount to snapshots which is probably
// how they implement the sharing anyway...
// So what if we delete anything that's there and recreate a new disk of the desired size.
//
// use  -vapp
func TestVMDisk(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")

	if live {
		fmt.Printf("\nTestVMDisk Live: \n")
		vapp, err := tv.FindVApp(ctx, *vappName)
		if err != nil {
			fmt.Printf("Unable to find %s\n", *vappName)
			return
		}
		vmname := vapp.VApp.Children.VM[0].Name

		vm, err := vapp.GetVMByName(vmname, true)
		if err != nil {
			fmt.Printf("GetVMByName failed: %s\n", err.Error())
			return
		}

		fmt.Printf("Use vm %s from %s\n", vmname, vapp.VApp.Name)
		// Two disk types, internal and independent. independent disks can be attach to at most 1 vm at any time.
		// here, we're dealing with the single internal disk to resize even if another vapp is using this
		// vm also.

		// To add a disk use DiskSettings, save to fill in new settings bits
		diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings[0]
		// first get the current disk id, since it's a delete by ID
		diskId := vm.VM.VmSpecSection.DiskSection.DiskSettings[0].DiskId
		// remove this current disk

		err = vm.DeleteInternalDisk(diskId)
		if err != nil {
			fmt.Printf("DeleteInternalDisk failed: %s\n", err.Error())
			return
		}

		newDiskSettings := &types.DiskSettings{
			SizeMb:          40 * 1024, // results in 41G in console xxx
			AdapterType:     diskSettings.AdapterType,
			ThinProvisioned: diskSettings.ThinProvisioned,
			StorageProfile:  diskSettings.StorageProfile,
		}
		newDiskId, err := vm.AddInternalDisk(newDiskSettings)
		if err != nil {
			fmt.Printf("AddInternalDisk tailed: %s\n", err.Error())
			return
		}
		fmt.Printf("old diskId %s new diskId = %s\n", diskId, newDiskId)

	} else {
		return
	}
}

// Test feeding our VM create work routine vmlayers Orch Params with a simple example.
func testCreateVM(t *testing.T, ctx context.Context) (*govcd.VM, error) {
	fmt.Printf("testCreateVM...")
	vm := &govcd.VM{}

	// This needs to send vapp to CreateVM
	vapp := govcd.VApp{}
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

	ports := []vmlayer.PortResourceReference{} // Ips may be assigned to ports or...
	/*
			Name        "mex-ports"
			Id          string
			NetworkId   string
			SubnetId    string
			Preexisting bool
			NetworkType NetType
			PortGroup   string

		}
	*/
	// use fixed
	// We should get a fixed IP from our tv.Obj.PrimaryNet
	// .51 or .52
	var fixedIps []vmlayer.FixedIPOrchestrationParams
	fixedIps = append(fixedIps, vmlayer.FixedIPOrchestrationParams{
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
	//cparams := chefmgmt.VMChefParams{}

	vmparams := vmlayer.VMOrchestrationParams{

		Id:          "VMtestID",
		Name:        "MexVM1",
		Role:        vmlayer.RoleVMPlatform,
		ImageName:   "ubuntu-18.04",
		ImageFolder: "MEX-CAT01",
		HostName:    "MexVMHostName",
		DNSDomain:   "mobiledgex.net",
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
		Ports:                   ports,
		FixedIPs:                fixedIps,
		AttachExternalDisk:      false,
		//ChefParams:              &cparams,
	}
	fmt.Printf("\nOrchParams created, calling our CreateVM work routine\n")
	vm, err := tv.CreateVM(ctx, &vapp, &vmparams)
	require.Nil(t, err, "CreateVM")
	vu.DumpVM(vm.VM, 1)
	return vm, nil

}

func testDetachPortFromServer(t *testing.T, ctx context.Context, serverName, subnetName, portName, string, powerState bool) error {
	// get server by name (govcd.VM)

	// check it's status
	//
	return nil
}

// Action type can be create, update, delete.
func testAttachPortToServer(t *testing.T, ctx context.Context, serverName, subnetName, portName, ipaddr string, powerState bool) error {

	fmt.Printf("testAttachPortToServer name: %s\n", serverName)
	detail, err := tv.GetServerDetail(ctx, serverName)
	if err != nil {
		fmt.Printf("Error from GetServerDetail for %s : %s\n", serverName, err.Error())
		return err
	}
	fmt.Printf("details of %s : %+v\n", serverName, detail)
	// but this is not enough, we need the govcd.VM object for serverName, but we know it eixsts.
	vm, err := tv.FindVM(ctx, serverName)
	if err != nil {
		fmt.Printf("FindVM failed err: %s\n", err.Error())
		return err
	}
	fmt.Printf("Add vm %+v\n", vm.VM)
	fmt.Printf("Adding ipaddr %s subnet %s portname %s to server %s state: %s to VM\n\t%+v\n", ipaddr, subnetName, portName, serverName, detail.Status, vm)
	parentApp, err := vm.GetParentVApp()
	if err != nil {
		return fmt.Errorf("Error getting parent of %s\n", serverName)
	}
	govcd.ShowVapp(*parentApp.VApp)
	return nil
}

func testDestroyVM(t *testing.T, ctx context.Context) {

}

func testUpdateVM(t *testing.T, ctx context.Context) {

}

func testInsertMediaToVM(t *testing.T, ctx context.Context, vm *govcd.VM) (*govcd.VM, error) {

	task, err := vm.HandleInsertMedia(tv.Objs.Org, tv.Objs.PrimaryCat.Catalog.Name, "ubuntu-18.04")
	if err != nil {
		fmt.Printf("Error inserting media %s to vm %s\n", "ubuntu-18.04", vm.VM.Name)
		return nil, err
	}
	fmt.Printf("HandleInsertMedia task: %+v\n", task)
	return vm, nil
}

func testVMMetrics(t *testing.T, ctx context.Context, vmname string, poweron bool) error {

	// Apparently, once a VM is powered on, it's Links should contain 4 links where the value of the type attribute has
	// the form: application/vnd.vmware.vcloud.metrics.*UsageSpec.xml
	// if so, we should fetch the HREF and see what it has for us
	// This will probably never work until govcd grows support for nsx-t.
	// Ok, the ExecuteRequest on the "down"
	vm, err := tv.FindVM(ctx, vmname)
	if err != nil {
		return fmt.Errorf("Error finding vm  %s  err: %s\n", *vmName, err.Error())
	}
	curStatus, err := vm.GetStatus()
	if curStatus == "POWERED_OFF" && poweron {
		fmt.Printf("testVMMetrics-I-%s currently powered off and poweron requested:  powering on\n", *vmName)
		task, err := vm.PowerOn()
		if err == nil {
			err = task.WaitTaskCompletion()
		} else {
			return err
		}
	} else if curStatus == "POWERED_ON" && !poweron {
		task, err := vm.PowerOff()
		if err == nil {
			err = task.WaitTaskCompletion()
		} else {
			return err
		}
	} else {
		fmt.Printf("Requesting Links of a powered off %s should just have historic links\n", *vmName)
	}
	curStatus, err = vm.GetStatus()
	// Try out the ForType method of LinkList not working yet..
	appType := ""
	if curStatus == "POWERED_ON" {
		appType = "application/vnd.vmware.vcloud.metrics.currentUsageSpec+xml"
	} else {
		// what is the application type for the historical status that can be fetch from a powered down vm?
		appType = "application/vnd.vmware.vcloud.metrics.historicUsageSpec+json"
	}
	ll := vm.VM.Link
	// type Rel
	// No constant for this type: (56/constants.go)

	link := ll.ForType(appType, types.RelDown)
	if link != nil {
		fmt.Printf("Found Link via ll.ForType: %+v\n", link)
		vu.DumpLink(link, 1)
	} else {
		fmt.Printf("No link for %s found in vm.VM.Link\n", appType)
	}

	// ok, so if we know the link we can try and fetch it using
	var buffer [5000]byte
	if appType != "" && link != nil {
		response, err := tv.Client.Client.ExecuteRequest(link.HREF, http.MethodGet, "", "error GET retriving metrics link: %s", nil, buffer)
		// This POST needs a prolog with the selection criteria
		//response, err := tv.Client.Client.ExecuteRequest(link.HREF, http.MethodPost, "", "error POST retriving metrics link: %s", nil, buffer)

		if err != nil {
			fmt.Printf("Error from ExecuteRequest: %s\n", err.Error())
		} else {
			fmt.Printf("http response: %+v\n", response)
			// what the hecks in in buffer?
			fmt.Printf("buffer: %+v\n", buffer)
		}
	}
	fmt.Printf("-----dumpLink with curStatus = %s\n", curStatus)
	vu.DumpLinkList(vm.VM.Link, 1)

	// Ok, so now, "Use the links where rel="down" with a GET request to retrieve current or historic metrics in all catagories"
	// and "Use the links where rel="metrics" with a POST request to retrieve a subset of current or historic metics.
	// and "When a VM is powered off, you cannot retrieve currentmetrics from it so .../metrics/currrent links are not returned

	// So this implies that historic metrics (stored for 2 weeks they say somewhere) _are_ available. We'll see
	return err
}

// -grp -live
func TestServerGroupResources(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitTestEnv")

	if live {
		resources, err := tv.GetServerGroupResources(ctx, *grpName)
		if err != nil {
			fmt.Printf("Error %s returned\n", err.Error())
			return
		}

		fmt.Printf("Resources for %s \n", *grpName)
		for _, vinfo := range resources.Vms {
			fmt.Printf("\tName : %s\n\tType: %s\n\t Status: %s\n\tFlavor: %s\n",
				vinfo.Name, vinfo.Type, vinfo.Status, vinfo.InfraFlavor)

			for _, ipSet := range vinfo.Ipaddresses {
				fmt.Printf("\tExternalIp: %s\n\tInternalIp:%s\n",
					ipSet.ExternalIp, ipSet.InternalIp)
			}
		}
	}
}

package vcd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VM related operations

func (v *VcdPlatform) FindVM(ctx context.Context, serverName string) (*govcd.VM, error) {

	vapp := &govcd.VApp{}
	for _, VApp := range v.Objs.VApps {
		vapp = VApp.VApp
		vm, err := vapp.GetVMByName(serverName, true)
		if err != nil {
			return nil, fmt.Errorf("vm %s not found in vapp %s", serverName, vapp.VApp.Name)
		}
		return vm, nil
	}
	// check our raw VMs map
	for name, vm := range v.Objs.VMs {
		if name == serverName {
			return vm, nil
		}
	}
	return nil, fmt.Errorf("Not Found")
}

func (v *VcdPlatform) FindVMInVApp(ctx context.Context, serverName string, vapp govcd.VApp) (*govcd.VM, error) {
	vm, err := vapp.GetVMByName(serverName, true)
	if err != nil {
		return nil, fmt.Errorf("vm %s not found in vapp %s", serverName, vapp.VApp.Name)
	}
	return vm, nil

}

// We expect Objs.PrimaryNet supports DHCP, eventually
//
func (v *VcdPlatform) IsDhcpEnabled(ctx context.Context, net *govcd.OrgVDCNetwork) bool {
	vdcnet := net.OrgVDCNetwork

	netconfig := vdcnet.Configuration
	fmt.Printf("\n\nnetconfnig: %+v\n\n", netconfig)
	if netconfig != nil {
		features := netconfig.Features
		fmt.Printf("\n\nnetconfnig: %+v\n\n", features)
		dhcpservice := features.DhcpService
		return dhcpservice.IsEnabled

	} else {
		fmt.Printf("\nNet %s has no config!\n", vdcnet.Name)
	}
	return false
}

// Per VMRequestSpec/VM
func (v *VcdPlatform) PopulateVMNetConnectSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.NetworkConnectionSection, error) {
	//netConnections := []*types.NetworkConnection{}
	netConnectSec := &types.NetworkConnectionSection{}
	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateVMNetConnectSection-I-VM", "name", vmparams.Name, "role", vmparams.Role)
	if vmparams.Role == vmlayer.RoleVMPlatform || vmparams.Role == vmlayer.RoleAgent {

		netConnectSec := &types.NetworkConnectionSection{}
		netConnectSec.PrimaryNetworkConnectionIndex = 0

		netConnectSec.NetworkConnection = append(netConnectSec.NetworkConnection,
			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeDHCP,
				Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name,
				// NetworkCOnnectionIndex: 2, // how would we find this?
			},
		)

	}

	// and in all cases, add a perhaps second nic for internal network.

	fmt.Printf("\nPopulateVMNetCOnnectSection-I-connection: %+v\n", netConnectSec)
	return netConnectSec, nil
}

// Given an org, vm, catalog name, and meida name, insert the media into the vm

func (v *VcdPlatform) InsertMediaToVM(ctx context.Context, catalogName, mediaName string, vm *govcd.VM) error {

	if vm == nil {
		return fmt.Errorf("Encountered nil vm")
	}
	// xxx think about multiple []catNames and look in them all...
	log.SpanLog(ctx, log.DebugLevelInfra, "InsertMediaToVM", "VM", vm.VM.Name, "media", mediaName)
	_, err := vm.HandleInsertMedia(v.Objs.Org, catalogName, mediaName)
	if err != nil {
		return fmt.Errorf("Error inserting %s from %s to vm %s org %s err %s",
			mediaName, catalogName, vm.VM.Name, v.Objs.Org.Org.Name, err.Error())
	}
	return nil
}

// vm_types.go has the recompose bits for whatever reason...
func (v *VcdPlatform) PopulateRecomposeParamsFromOrchParams(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.RecomposeVAppParamsForEmptyVm, error) {
	recompParams := &types.RecomposeVAppParamsForEmptyVm{}
	return recompParams, nil
}

// Local work routine needs VMrefactor.. only called from test currently
func (v *VcdPlatform) CreateVM(ctx context.Context, vapp *govcd.VApp, vmparams *vmlayer.VMOrchestrationParams) (*govcd.VM, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM ", "Name", vmparams.Name, "Role", vmparams.Role)

	var err error
	if v.Client == nil {
		v.Client, err = v.GetClient(ctx, v.Creds)
		if err != nil {
			return nil, err
		}
	}
	// vm := govcd.NewVM(&v.Client.Client) No HREF, so not much to be done with it.
	// xlate VMOrchParams into ReomposeVAppParamsForEmptyVm
	// This is currently a noop:
	recomposeParams, err := v.PopulateRecomposeParamsFromOrchParams(ctx, vmparams)
	vm, err := vapp.AddEmptyVm(recomposeParams)
	if err != nil {
		return nil, err
	}
	// pick off args to add to vm
	vm.VM.Name = vmparams.Name
	vm.VM.Description = fmt.Sprintf("%s %d %d %d", vmparams.Role, vmparams.Vcpus, vmparams.Ram, vmparams.Disk)
	//	mediaName = vmparams.ImageName
	//mediaName := "ubuntu-18.04"

	// customized this vm based on role
	nc, err := v.PopulateVMNetConnectSection(ctx, vmparams)
	// add nc to the vm
	if err != nil {
		fmt.Printf("\nCreateVM-E-populateVMNetConnectSetion error: %s\n", err.Error())
	}

	// This is failing as vm.HREF is not set yet.

	err = vm.UpdateNetworkConnectionSection(nc)
	if err != nil {
		fmt.Printf("\nUpdateNetowrkConectionSection fails: %s\n", err.Error())
	}

	// xlate various vmgp to this VM
	// vitural cpu count
	// what do people do with task? Wait on it? Since I've see async calls, these are sync?
	t, err := vm.ChangeCPUCount(int(vmparams.Vcpus))
	if err != nil {
		fmt.Printf("\nCreateVM unable to change CPUCount for vm %s\n", vm.VM.Name)
	}
	err = t.WaitTaskCompletion()
	// vm.ChangeNetworkConfig(networks []map[string]interface{}) (Task, error) {
	t, err = vm.ChangeMemorySize(int(vmparams.Ram))
	if err != nil {
		fmt.Printf("\nCreateVM unable to change CPUCount for vm %s\n", vm.VM.Name)
	}
	err = t.WaitTaskCompletion()
	// XXX add a local storage volume of size MBs..

	return vm, nil
}

// Create VMs according to their roles, (no VMType available here) and their names
// Cloudlets are named like cloudlet
// ClusterInst are named like cloudlet.cluster
// Nodes are named like vm.cloudlet.cluster right?
//
func (v *VcdPlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {

	// If the given cloudlet vapp  is already running all subsquent vms are added to
	// the cloudlet's vapp instance.
	//	vu.DumpVMGroupParams(vmgp, 1)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs")

	// Find our ova template, all platform vms use the same template
	tmplName := os.Getenv("VCDTEMPLATE")
	if tmplName == "" {
		// trade env for property XXX
		return fmt.Errorf("VCD Base template env var not set")
	}
	// First get our template
	tmpl, err := v.FindTemplate(ctx, tmplName)
	if err != nil {
		found := false
		fmt.Printf("\n\tnot found locally\n")
		// Back to vdc, has it been created manually?
		tmpls, err := v.GetAllVdcTemplates(ctx, v.Objs.PrimaryCat)
		if err == nil {
			for _, tmpl = range tmpls {
				if tmpl.VAppTemplate.Name == tmplName {
					v.Objs.VAppTmpls[tmplName] = tmpl
					found = true
					break
				}
			}
		}
		if !found {
			// Try fetching it from the respository or local update
			log.SpanLog(ctx, log.DebugLevelInfra, "Template %s not found in vdc, attempt upload, this can take 20 mins or more\n", tmplName)
			err = v.UploadOvaFile(ctx, tmplName)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "Template %s not found, not uploaded Fail", "error", err.Error())
				return err
			}
		}
	}
	if tmpl == nil {
		return fmt.Errorf("Unable to find ova template")
	}
	description := vmgp.GroupName + "-vapp"
	cloudletName := ""
	if v.Objs.Cloudlet != nil {
		// look for an existing Vapp/Cloudlet with this name
		// just return cloudlet, it has the vapp in it.
		cloudlet := v.Objs.Cloudlet // parts bin =>  vapp, err := v.FindCloudletForCluster(description)

		// find cloudlet here based on description, should match our (only) vapp else "invalid cloudlet name"
		// should be like validateCloudetName
		tcloud, _, err := v.FindCloudletForCluster(description)
		if err != nil {
			cn := strings.Split(description, ".")
			cloudletName = cn[0]
			log.SpanLog(ctx, log.DebugLevelInfra, "Unknown Cloudlet encountered")
			return fmt.Errorf("Clouldlet not found %s", cloudletName)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Creating Cluster on", "cluster", tcloud.CloudletName, "cloudlet", tcloud.CloudVapp.VApp.Name)
		clusterName, err := v.CreateCluster(ctx, cloudlet, tmpl, vmgp, updateCallback)
		if err != nil {
			fmt.Printf("\nCreateVMs-E-CreateCluster-E-%s\n", err.Error())
			return err
		}
		fmt.Printf("CreateVMs-I-Cluster %s Created successfully\n", clusterName)
		return nil
	}
	// CreateCloudlet
	vdc := v.Objs.Vdc
	storRef := types.Reference{}
	// Empty Ref wins the default (vSAN Default is all we have, but could support others xxx Prop?)

	// CreateCloudlet
	vapp, err := v.CreateCloudlet(ctx, vdc, *tmpl, storRef, vmgp, description, updateCallback)
	if err != nil {
		return fmt.Errorf("CreateVApp return error: %s", err.Error())
	}

	status, err := vapp.GetStatus()
	// This is our single vapp / vdc = cloudlet
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs", "Vapp", vapp.VApp.Name, "Status", status)
	return nil
}

// updates of a vm that is 'shared' across multiple vapps
// balks at being modified "can't modify disk of a vm with snapshots"
// So here, we remove, and replace it. XXX only first disk, doesn't
// support multiple internal disks. XXX
func (v *VcdPlatform) updateVmDisk(vm *govcd.VM, size int64) error {

	diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings[0]
	diskId := vm.VM.VmSpecSection.DiskSection.DiskSettings[0].DiskId
	// remove this current disk
	err := vm.DeleteInternalDisk(diskId)
	if err != nil {
		fmt.Printf("DeleteInternalDisk failed: %s\n", err.Error())
		return err
	}

	newDiskSettings := &types.DiskSettings{
		SizeMb:          size * 1024, // results in 1G > size ?
		AdapterType:     diskSettings.AdapterType,
		ThinProvisioned: diskSettings.ThinProvisioned,
		StorageProfile:  diskSettings.StorageProfile,
	}
	_, err = vm.AddInternalDisk(newDiskSettings)
	if err != nil {
		fmt.Printf("AddInternalDisk tailed: %s\n", err.Error())
		return err
	}
	return nil
}

func (v *VcdPlatform) guestCustomization(ctx context.Context, vm govcd.VM, vmparams vmlayer.VMOrchestrationParams, subnet string) error {
	//script := "#!/bin/bash  &#13; ip route del default via 10.101.1.1  &#13;"

	vm.VM.GuestCustomizationSection.ComputerName = vmparams.HostName
	if subnet != "" {
		subnet = "10.101.1.1"
		script := fmt.Sprintf("%s%s%s", "#!/bin/bash  &#13; ip route del default via", subnet, "&#13")
		fmt.Printf("guestCustomization script: %s\n", script)
		vm.VM.GuestCustomizationSection.CustomizationScript = script
	}
	vm.VM.GuestCustomizationSection.Enabled = vu.TakeBoolPointer(true)
	// script to delete default route
	return nil
}

// TBD
func (v *VcdPlatform) populateVMMetadata(ctx context.Context, vm govcd.VM, vmparams vmlayer.VMOrchestrationParams) error {

	for _, port := range vmparams.Ports {
		fmt.Printf("\t Name: %s preexsting: %t PortGroup %s\n", port.Name, port.Preexisting, port.PortGroup)
	}
	for _, FixedIP := range vmparams.FixedIPs {
		fmt.Printf("\tFixedIP: LastIPOctet: %d Address: %s Gateway: %s\n",
			FixedIP.LastIPOctet, FixedIP.Address, FixedIP.Gateway)
	}
	return nil
}

// set vm params and call vm.UpdateVmSpecSection
func (v *VcdPlatform) updateVM(ctx context.Context, vm *govcd.VM, vmparams vmlayer.VMOrchestrationParams, subnet string) error {

	//parentVapp, err := vm.GetParentVApp()
	err := v.populateVMMetadata(ctx, *vm, vmparams)
	if err != nil {
		return err
	}

	flavorName := vmparams.FlavorName
	flavor, err := v.GetFlavor(ctx, flavorName)
	vmSpecSec := vm.VM.VmSpecSection
	vmSpecSec.NumCpus = vu.TakeIntPointer(int(flavor.Vcpus))
	vmSpecSec.MemoryResourceMb.Configured = int64(flavor.Ram)
	desc := fmt.Sprintf("Update flavor: %s", flavorName)
	_, err = vm.UpdateVmSpecSection(vmSpecSec, desc)
	if err != nil {
		return err
	}

	fmt.Printf("Disk Update TBI\n")

	// meta data for Role etc
	psl, err := v.populateProductSection(ctx, vm, &vmparams)
	if err != nil {
		return fmt.Errorf("updateVM-E-error from populateProductSection: %s", err.Error())
	}

	_, err = vm.SetProductSectionList(psl)
	if err != nil {
		return fmt.Errorf("updateVM-E-error Setting product section %s", err.Error())
	}
	_, err = vm.SetProductSectionList(psl)
	if err != nil {
		return err
	}

	err = v.guestCustomization(ctx, *vm, vmparams, subnet)
	if err != nil {
		return fmt.Errorf("updateVM-E-error from guestCustomize: %s", err.Error())
	}

	return err
}

// PI UpdateVMs
// Add/remove VM from our VApp (group)
//
func (v *VcdPlatform) UpdateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs", "OrchParams", vmgp)
	// convert each vmOrchParams into a *types.VmSpecSection and call updateVM for each vm
	return nil
}
func (v *VcdPlatform) SyncVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs", "OrchParams", vmgp)
	return nil
}

// Delete All VMs in VApp with VApp.VApp.Name == vmGroupName, then remove the VApp itself.
func (v *VcdPlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {

	// if we can swing making vmGroupNamem == VApp Name, we're good here
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVMs", "vmGroupName", vmGroupName)
	return nil
}

// This might be a good place to note:
// 3 types of allocation models: [pay as you go | Alloc Pool Model | Reservation Pool Model]
// which affect how "CPU used" is computed:
//  1: Pay as you go
//      CPU used = vcpu_count *ovdc_vcpu_in_mhz  ( vcpu sppeed given for organizational vdc)
//  2: Zero, this model is not elastic. So you need to query the underlying vCenter (provider)
//  3: This is mapped to runtime.cpu.reservationUsed in vSphere. Great.
//  4: New: Flex Allocation Model. The adminVdc.IsElastic is only supported in API 32.0 and above.
//
//  (We currently use 31.0, so revist switching this again, and see if we get access to it (It's <nil> along with IncludeMemoryOverhead.
// Our current test world is AllocationPool, which by now (10.01) should be elastic, what does that mean for cpu
// stats?
//
func (v *VcdPlatform) GetVMStats(ctx context.Context, key *edgeproto.AppInstKey) (*vmlayer.VMMetrics, error) {
	vm := &govcd.VM{}
	metrics := vmlayer.VMMetrics{}
	var err error
	vmName := cloudcommon.GetAppFQN(&key.AppKey)

	if vmName == "" {
		return nil, fmt.Errorf("GetAppFQN failed to return vmName for AppInst %s\n", key.AppKey.Name)
	}
	// We need the Vapp in which context this name is being looked up in XXX
	// Are we only interested in deployed VMs?
	// The metrics links are not functional right now anyway
	// Just look for the first Vapp that has this VM name ?

	vm, err = v.FindVM(ctx, vmName)
	if err != nil {
		//fmt.Printf("\n\tGetVMStats failed to find vm: %s\n", vmName)
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVMStats failed to find vm", "name", vmName)
		return nil, err
	}
	status, err := vm.GetStatus()
	if err == nil && status == "POWERED_ON" {
		fmt.Printf("Getting usage metrics for vm: %s\n", vm.VM.Name)

		// Check vdc_vm_test.go for the metric links
		// They don't seem to work with nsx-t boxes
		//
		/*  get these for the VM running AppInst
		type VMMetrics struct {
			// Cpu is a percentage
			Cpu   float64
			CpuTS *types.Timestamp
			// Mem is bytes used
			Mem   uint64
			MemTS *types.Timestamp
			// Disk is bytes used
			Disk   uint64
			DiskTS *types.Timestamp
			// NetSent is bytes/second average
			NetSent   uint64
			NetSentTS *types.Timestamp
			// NetRecv is bytes/second average
			NetRecv   uint64
			NetRecvTS *types.Timestamp
		}
		*/
	} else {
		return nil, fmt.Errorf("No stats available for %s", vm.VM.Name)
	}
	return &metrics, nil
}

// always sync.
func (v *VcdPlatform) SetPowerState(ctx context.Context, serverName, serverAction string) error {
	vm := &govcd.VM{}
	var err error
	// remember, all our vms are on a per VApp basis, so we'll always be
	//  eding that unique id for common vm names. >sigh>

	vm, err = v.FindVM(ctx, serverName)
	if err != nil {
		fmt.Printf("FindVM failed to find %s\n", serverName)
		return err
	}
	curStatus, err := vm.GetStatus()

	if serverAction == vmlayer.ActionStart {
		if curStatus == "POWERED_ON" {
			return fmt.Errorf("%s Already Powered on", vm.VM.Name)
		}

		task, err := vm.PowerOn()
		if err != nil {
			fmt.Printf("Error Powering on %s err: %s\n", vm.VM.Name, err.Error())
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				fmt.Printf("Error waiting for Powering on %s err: %s\n", vm.VM.Name, err.Error())
			}
		}
	}
	if serverAction == vmlayer.ActionStop {
		if curStatus == "POWERED_OFF" {
			return fmt.Errorf("%s Already Powered off", vm.VM.Name)
		}
		task, err := vm.PowerOff()
		if err != nil {
			fmt.Printf("Error Powering off %s err: %s\n", vm.VM.Name, err.Error())
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error waiting for Powering on %s err: %s\n", vm.VM.Name, err.Error())
			}
		}
	}
	if serverAction == vmlayer.ActionReboot {
		if curStatus != "POWERED_ON" {
			return fmt.Errorf("Can't reboot %s currently in state %s\n", vm.VM.Name, curStatus)
		}
		task, err := vm.PowerOff()
		if err != nil {
			fmt.Printf("Error Powering off %s err: %s\n", vm.VM.Name, err.Error())
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error waiting for Powering on %s err: %s\n", vm.VM.Name, err.Error())
			}
		}
		task, err = vm.PowerOn()
		if err != nil {
			fmt.Printf("Error Powering on %s err: %s\n", vm.VM.Name, err.Error())
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error waiting for Powering on %s err: %s\n", vm.VM.Name, err.Error())
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "SetPowerState", "serverName", serverName, "serverAction", serverAction)
	return nil
}

// PI
func (v *VcdPlatform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	// no one calls this?
	log.SpanLog(ctx, log.DebugLevelInfra, "VerifyVMs  ")
	return nil
}

// return the *VAppChildren of the given vapp if any
func (v *VcdPlatform) GetVappVms(ctx context.Context, vapp *govcd.VApp) ([]*types.VM, error) {
	fmt.Printf("GetVappVms-I-TBI\n")
	return nil, nil

}

// add customizatin option to vm
func (v *VcdPlatform) CustomizeVm(ctx context.Context, vm *govcd.VM, cs *types.CustomizationSection) (*types.VM, error) {
	return nil, nil

}

// xxx These Result RecordTypes xxx
func (v *VcdPlatform) GetAvailableVMs(ctx context.Context) ([]*types.QueryResultVMRecordType, error) {
	// returns  VMs of all VApps available in our vdc

	var filter types.VmQueryFilter = types.VmQueryFilterAll

	vmRecs, err := v.Client.Client.QueryVmList(filter)
	if err != nil {
		fmt.Printf("\n\nGetAvailableVMs failed query %s\n", err.Error())
		return nil, fmt.Errorf("Unable to Query available VMs err: %s", err.Error())
	}
	return vmRecs, nil
}

// return the IP
/*
type ServerIP struct {
	MacAddress             string
	InternalAddr           string // this is the address used inside the server
	ExternalAddr           string // this is external with respect to the server, not necessarily internet reachable.  Can be a floating IP
	Network                string
	PortName               string
	ExternalAddrIsFloating bool
}
*/

func (v *VcdPlatform) GetVMAddressesOrig(ctx context.Context, vm *govcd.VM) ([]vmlayer.ServerIP, string, error) {
	var serverIPs []vmlayer.ServerIP
	if vm == nil {
		return serverIPs, "", fmt.Errorf("Nil vm received")
	}
	vmName := vm.VM.Name
	//parentVapp, err := vm.GetParentVApp()
	status, err := vm.GetStatus()
	if err != nil {
		return serverIPs, "", fmt.Errorf("Error getting status for %s err: %s\n", vm.VM.Name, err.Error())
	}
	if status != "POWERED_ON" {
		return serverIPs, "", fmt.Errorf("vm %s not powered on state: %s", vm.VM.Name, status)
	}
	connections := vm.VM.NetworkConnectionSection.NetworkConnection
	for _, connection := range connections {
		//fmt.Printf("GetVMAddresses-I- %s next IP%s[idx:%d] \n", vm.VM.Name, connection.IPAddress, connection.NetworkConnectionIndex)

		servIP := vmlayer.ServerIP{
			MacAddress:   connection.MACAddress,
			Network:      connection.Network,
			ExternalAddr: connection.IPAddress, //ExternalIPAddress, // if a Nat, external IP here.
			InternalAddr: connection.IPAddress,
			PortName:     strconv.Itoa(connection.NetworkConnectionIndex),
		}
		if connection.Network != v.Objs.PrimaryNet.OrgVDCNetwork.Name {
			// internal isolated net
			servIP.PortName = vmName + "-" + connection.Network + "-port"
			fmt.Printf("GetVMAddressesOrig-I-setting portname %s \n\n", servIP.PortName)
		}
		serverIPs = append(serverIPs, servIP)
	}
	ip := connections[0].IPAddress
	return serverIPs, ip, nil
}

func (v *VcdPlatform) GetVMAddresses2(ctx context.Context, vm *govcd.VM) ([]vmlayer.ServerIP, string, error) {
	var serverIPs []vmlayer.ServerIP
	if vm == nil {
		return serverIPs, "", fmt.Errorf("Nil vm received")
	}

	status, err := vm.GetStatus()
	if err != nil {
		return serverIPs, "", fmt.Errorf("Error getting status for %s err: %s\n", vm.VM.Name, err.Error())
	}
	if status != "POWERED_ON" {
		return serverIPs, "", fmt.Errorf("vm %s not powered on state: %s", vm.VM.Name, status)
	}
	ip := ""
	// Find out if this is a isolated newtowrk XXX
	connections := vm.VM.NetworkConnectionSection.NetworkConnection
	for _, connection := range connections {

		servIP := vmlayer.ServerIP{
			MacAddress: connection.MACAddress,
			Network:    connection.Network,
			PortName:   strconv.Itoa(connection.NetworkConnectionIndex),
		}
		if connection.Network == v.Objs.PrimaryNet.OrgVDCNetwork.Name {
			servIP.ExternalAddr = connection.IPAddress
			ip = servIP.ExternalAddr
			fmt.Printf("\tGetVMAddresses-I-set external addr: %s on net: %s\n", servIP.ExternalAddr, servIP.Network)
			break // just return a single external
		} else {
			// Internal
			fmt.Printf("\tGetVMAddresses-I-set external addr: %s on net: %s\n", servIP.InternalAddr, servIP.Network)
			servIP.InternalAddr = connection.IPAddress
		}
		serverIPs = append(serverIPs, servIP)
	}
	// but just retrun one serverIP object
	return serverIPs, ip, nil
}

// Revisit XXX
func (v *VcdPlatform) SetVMProperties(vmProperties *vmlayer.VMProperties) {
	v.vmProperties = vmProperties
	vmProperties.IptablesBasedFirewall = false // true
}

// new, ah, get resources from a group of vms.. like  cluster?
func (v *VcdPlatform) GetServerGroupResources(ctx context.Context, name string) (*edgeproto.InfraResources, error) {
	resources := &edgeproto.InfraResources{}
	return resources, nil
}

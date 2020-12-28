package vcd

import (
	"context"
	"fmt"
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
	if netconfig != nil {
		features := netconfig.Features
		dhcpservice := features.DhcpService
		return dhcpservice.IsEnabled

	}
	return false
}

// Per VMRequestSpec/VM
func (v *VcdPlatform) PopulateVMNetConnectSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.NetworkConnectionSection, error) {

	netConnectSec := &types.NetworkConnectionSection{}
	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateVMNetConnectSection ", "name", vmparams.Name, "role", vmparams.Role)
	if vmparams.Role == vmlayer.RoleVMPlatform || vmparams.Role == vmlayer.RoleAgent {

		netConnectSec := &types.NetworkConnectionSection{}
		netConnectSec.PrimaryNetworkConnectionIndex = 0

		netConnectSec.NetworkConnection = append(netConnectSec.NetworkConnection,
			&types.NetworkConnection{
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual,
				Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name,
			},
		)
	}
	return netConnectSec, nil
}

// Given an org, vm, catalog name, and meida name, insert the media into the vm
/*
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
*/

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
		return nil, err
	}

	// This is failing as vm.HREF is not set yet.

	err = vm.UpdateNetworkConnectionSection(nc)
	if err != nil {
		return nil, err
	}

	// xlate various vmgp to this VM
	// vitural cpu count
	// what do people do with task? Wait on it? Since I've see async calls, these are sync?
	t, err := vm.ChangeCPUCount(int(vmparams.Vcpus))
	if err != nil {
		return nil, err
	}
	err = t.WaitTaskCompletion()
	// vm.ChangeNetworkConfig(networks []map[string]interface{}) (Task, error) {
	t, err = vm.ChangeMemorySize(int(vmparams.Ram))
	if err != nil {
		return nil, err
	}
	err = t.WaitTaskCompletion()
	// XXX add a local storage volume of size MBs..

	return vm, nil
}

// Create VMs according to their role/type and names
func (v *VcdPlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {

	// vu.DumpVMGroupParams(vmgp, 1)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs", "grpName", vmgp.GroupName)
	// Find our ova template, all platform vms use the same template
	tmplName := v.GetVDCTemplateName() // vcdVars["VDCTEMPLATE"]
	if tmplName == "" {
		return fmt.Errorf("VDCTEMPLATE not set")
	}
	tmpl, err := v.FindTemplate(ctx, tmplName)
	if err != nil {
		found := false
		log.SpanLog(ctx, log.DebugLevelInfra, "Template not found locally", "template", tmplName)
		// Back to vdc, has it been created manually?
		tmpls, err := v.GetAllVdcTemplates(ctx)
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
			return fmt.Errorf("Template %s not found", tmplName)
			// Try fetching it from the respository or local update
			// XXX upload TBI, expect it to be in the catalog.
			//err = v.UploadOvaFile(ctx, tmplName)
			//if err != nil {
			//return fmt.Errorf("Template %s not found\n", tmplName)
			//}
		}
	}
	if tmpl == nil {
		if v.Objs.Template != nil {
			// last ditch. When not found locally, nor on demand,  this template has issues.
			// I think we've found one of the auto-generated templates created for a 'standalone vm'
			tmpl = v.Objs.Template
			log.SpanLog(ctx, log.DebugLevelInfra, "using v.Objs.Template last resort", "template", *tmpl.VAppTemplate)
		} else {
			return fmt.Errorf("Unable to find ova template")
		}
	}

	description := vmgp.GroupName + "-vapp"
	cloudletName := ""
	if v.Objs.Cloudlet != nil {
		_, err := v.FindVM(ctx, vmgp.GroupName)
		if err == nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "GroupName exists", "name", vmgp.GroupName)
			return nil
		}
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
			return err
		}
		cluster, err := v.FindCluster(ctx, clusterName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Internal Error")
			return err
		} else {
			if v.Verbose {
				log.SpanLog(ctx, log.DebugLevelInfra, "Cluster created", "ClusterName", clusterName, "cluster", cluster)
			}
		}
		return nil
	}
	// CreateCloudlet
	vapp, err := v.CreateCloudlet(ctx, *tmpl, vmgp, description, updateCallback)
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
		vm.VM.GuestCustomizationSection.CustomizationScript = script
	}
	vm.VM.GuestCustomizationSection.Enabled = vu.TakeBoolPointer(true)
	// script to delete default route
	return nil
}

// set vm params and call vm.UpdateVmSpecSection
func (v *VcdPlatform) updateVM(ctx context.Context, vm *govcd.VM, vmparams vmlayer.VMOrchestrationParams, subnet string) error {

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
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs TBI", "OrchParams", vmgp)
	// convert each vmOrchParams into a *types.VmSpecSection and call updateVM for each vm
	return nil
}

func (v *VcdPlatform) SyncVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncVMs TBI", "OrchParams", vmgp)
	return nil
}

func (v *VcdPlatform) DeleteVM(ctx context.Context, vm *govcd.VM) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVM", "vmName", vm.VM.Name)
	vapp := v.Objs.Cloudlet.CloudVapp

	if vm == nil {
		return fmt.Errorf("nil vm encountered")
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVM", "vmName", vm.VM.Name)
	// do we care? mdata, err := vm.GetMetadata()
	status, err := vm.GetStatus()
	if err != nil {
		return err
	}
	if status == "POWERED_ON" {
		task, err := vm.PowerOff()
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
		}
	}

	err = vapp.RemoveVM(*vm)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "VM Deleted", "vmName", vm.VM.Name)
	return nil

}

// Delete All VMs in the resolution of vmGroupName
func (v *VcdPlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVMs", "vmGroupName", vmGroupName)
	// resolve vmGroupName, to a single vm or a clusterName

	// if vmGroupName is the Vapp, we're removing the entire cloudlet
	vappName := vmGroupName + "-vapp"
	vapp, err := v.FindVApp(ctx, vappName)
	if err == nil {
		if v.Objs.Cloudlet == nil {
			// internal error, we somehow restarted and missed recognizing our cloudlet or
			// asked to delete a cloudlet that DNE.
			// grab the meta data from this vapp, is it our cloudlet?
			mdata, err := vapp.GetMetadata()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "Error retrieving metadata from", "grpName", vmGroupName, err)
				return err

			}
			for _, data := range mdata.MetadataEntry {
				if data.Key == "CloudletName" {
					return fmt.Errorf("Internal Error valid cloudlet vapp %s , cloudlet nil", vmGroupName)
				}
			}
			return nil
		} else {
			err := v.DeleteCloudlet(ctx, *v.Objs.Cloudlet)
			if err != nil {
				return err
			}
		}
	}

	vm, err := v.FindVM(ctx, vmGroupName)
	if err == nil {
		return v.DeleteVM(ctx, vm)
	} else {

		cluster, err := v.FindCluster(ctx, vmGroupName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVMs not found", "vmGroupName", vmGroupName)
			return nil
		}
		err = v.DeleteCluster(ctx, cluster.Name)
		return err
	}
}

func (v *VcdPlatform) GetVMStats(ctx context.Context, key *edgeproto.AppInstKey) (*vmlayer.VMMetrics, error) {
	vm := &govcd.VM{}
	metrics := vmlayer.VMMetrics{}
	var err error
	vmName := cloudcommon.GetAppFQN(&key.AppKey)

	if vmName == "" {
		return nil, fmt.Errorf("GetAppFQN failed to return vmName for AppInst %s\n", key.AppKey.Name)
	}
	// We need the Vapp in which context this name is being looked up  XXX
	// Are we only interested in deployed VMs?
	// The metrics links are not functional right now anyway
	// Just look for the first Vapp that has this VM name ?

	vm, err = v.FindVM(ctx, vmName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVMStats failed to find vm", "name", vmName)
		return nil, err
	}
	status, err := vm.GetStatus()
	if err == nil && status == "POWERED_ON" {
		// Check vdc_vm_test.go for the metric links
		// They don't seem to work with nsx-t boxes
		//
		/*  get these for the VM running AppInst TBI
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
	vm, err = v.FindVM(ctx, serverName)
	if err != nil {
		return err
	}
	curStatus, err := vm.GetStatus()

	if serverAction == vmlayer.ActionStart {
		if curStatus == "POWERED_ON" {
			return fmt.Errorf("%s Already Powered on", vm.VM.Name)
		}

		task, err := vm.PowerOn()
		if err != nil {
			return err
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return err
			}
		}
	}
	if serverAction == vmlayer.ActionStop {
		if curStatus == "POWERED_OFF" {
			return fmt.Errorf("%s Already Powered off", vm.VM.Name)
		}
		task, err := vm.PowerOff()
		if err != nil {
			return err
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return err
			}
		}
	}
	if serverAction == vmlayer.ActionReboot {
		if curStatus != "POWERED_ON" {
			return fmt.Errorf("Can't reboot %s currently in state %s\n", vm.VM.Name, curStatus)
		}
		task, err := vm.PowerOff()
		if err != nil {
			return fmt.Errorf("Error Powering off %s err: %s\n", vm.VM.Name, err.Error())
		} else {
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error waiting for Powering on %s err: %s\n", vm.VM.Name, err.Error())
			}
		}
		task, err = vm.PowerOn()
		if err != nil {
			return err
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

func (v *VcdPlatform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	// existance? Powered ON?
	log.SpanLog(ctx, log.DebugLevelInfra, "VerifyVMs TBI")
	return nil
}

func (v *VcdPlatform) GetVMAddresses(ctx context.Context, vm *govcd.VM) ([]vmlayer.ServerIP, string, error) {
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

		servIP := vmlayer.ServerIP{
			MacAddress:   connection.MACAddress,
			Network:      connection.Network,
			ExternalAddr: connection.IPAddress,
			InternalAddr: connection.IPAddress,
			PortName:     strconv.Itoa(connection.NetworkConnectionIndex),
		}
		if connection.Network != v.Objs.PrimaryNet.OrgVDCNetwork.Name {
			// internal isolated net
			servIP.PortName = vmName + "-" + connection.Network + "-port"
		}
		serverIPs = append(serverIPs, servIP)
	}
	ip := connections[0].IPAddress
	return serverIPs, ip, nil
}

func (v *VcdPlatform) SetVMProperties(vmProperties *vmlayer.VMProperties) {
	v.vmProperties = vmProperties
	vmProperties.IptablesBasedFirewall = true // xxx false
}

// This can get called with name representing
// a single VM name like a sharedRootLB, or PlatformVM
// a cluster name.
// Can't return values from just a govcd.VM obj. We must have a cloudlet + cloudlet on there
//
func (v *VcdPlatform) GetServerGroupResources(ctx context.Context, name string) (*edgeproto.InfraResources, error) {
	resources := &edgeproto.InfraResources{}
	// xxx need ContainerInfo as well
	log.SpanLog(ctx, log.DebugLevelInfra, "grp resources for", "grpName", name)
	if v.Objs.Cloudlet != nil {

		for _, cluster := range v.Objs.Cloudlet.Clusters {
			if name == cluster.Name {
				for _, vm := range cluster.VMs {
					vminfo := edgeproto.VmInfo{
						Name:        vm.vmName,
						InfraFlavor: vm.vmFlavor,
						Type:        string(vmlayer.GetVmTypeForRole(vm.vmRole)),
					}
					vminfo.Ipaddresses = append(vminfo.Ipaddresses, vm.vmIPs)
					resources.Vms = append(resources.Vms, vminfo)
				}
				return resources, nil
			}
		}
	}
	return resources, nil
}

// Store attrs of vm for crmrestarts
func (v *VcdPlatform) AddMetadataToVM(ctx context.Context, vm *govcd.VM, vmparams vmlayer.VMOrchestrationParams, vmType, parentCluster string) error {

	// why no async for vms?
	task, err := vm.AddMetadata("vmType", vmType)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	task, err = vm.AddMetadata("FlavorName", vmparams.FlavorName)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	task, err = vm.AddMetadata("vmRole", string(vmparams.Role))
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	task, err = vm.AddMetadata("ParentCluster", parentCluster)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	return nil
}

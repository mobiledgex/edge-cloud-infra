package vcdutils

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func DumpFixedIPs(fips *[]vmlayer.FixedIPOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if fips == nil {
		fmt.Printf("%s\n", fill+"No fixed IPs")
		return
	}
	for _, fip := range *fips {
		fmt.Printf("%s %d\n", fill+"LastIPOctet", fip.LastIPOctet)
		fmt.Printf("%s %s\n", fill+"Address", fip.Address)
		fmt.Printf("%s %s\n", fill+"Mask", fip.Mask)
		fmt.Printf("%s %s\n", fill+"Gateway", fip.Gateway)
	}
}

func DumpPortsOrchParams(ports *[]vmlayer.PortOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if ports == nil {
		fmt.Printf("%s\n", fill+"No ports")
		return
	}
	for _, port := range *ports {
		fmt.Printf("%s %s\n", fill+"Name", port.Name)
		fmt.Printf("%s %s\n", fill+"Id", port.Id)
		fmt.Printf("%s %s\n", fill+"SubnetId", port.SubnetId)
		fmt.Printf("%s %s\n", fill+"NetowrkId", port.NetworkId)

		if port.NetworkType == 1 {
			fmt.Printf("%s %s\n", fill+"NetType", "External")
		} else if port.NetworkType == 0 {
			fmt.Printf("%s %s\n", fill+"NetType", "Internal (isolated)")
		} else {
			fmt.Printf("%s %d\n", fill+"NetType unkonwn:", port.NetworkType)
		}

		fmt.Printf("%s %t\n", fill+"Id", port.SkipAttachVM)
		fmt.Printf("%s \n", fill+"FixedIPs:")
		DumpFixedIPs(&port.FixedIPs, indent+1)
		fmt.Printf("%s \n", fill+"SecGrps:")
		DumpResourceRefs(&port.SecurityGroups, indent+1)

	}
}

func DumpSubnetOrchParams(sp *[]vmlayer.SubnetOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if sp == nil {
		fmt.Printf("%s\n", fill+"No subnets")
		return
	}
	for _, net := range *sp {
		// vmparams.SubnetOrchestrationParams
		fmt.Printf("%s %s\n", fill+"ID", net.Id)
		fmt.Printf("%s %s\n", fill+"Name", net.Name)
		fmt.Printf("%s %s\n", fill+"NetworkName", net.NetworkName)
		fmt.Printf("%s %s\n", fill+"CIDR", net.CIDR)
		fmt.Printf("%s %s\n", fill+"NodeIPPrefix", net.NodeIPPrefix)
		fmt.Printf("%s %s\n", fill+"GatewayIP", net.GatewayIP)
		fmt.Printf("%s \n", fill+"DNSServers:")
		for _, ds := range net.DNSServers {
			fmt.Printf("%s %s\n", fill+fill+"DNSServer", ds)
		}
		fmt.Printf("%s %s\n", fill+"DHCP Enabled", net.DHCPEnabled)
		// tbi Tags, skips ChefUpdateInfo.. ?

	}
}

func DumpRouterInterfaceParams(rips *[]vmlayer.RouterInterfaceOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if rips == nil {
		fmt.Printf("%s\n", fill+"No router interface params")
		return
	}
	for _, rip := range *rips {
		fmt.Printf("%s %s\n", fill+"Name", rip.RouterName)
		//		DumpResourceRefs(&rip.RouterPort, indent+1)
	}
}

func DumpFixedIPOrchParams(fips *[]vmlayer.FixedIPOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if fips == nil {
		fmt.Printf("%s\n", fill+"No router interface params")
		return
	}
	for _, ip := range *fips {
		fmt.Printf("%s %d\n", fill+"LastIPOctet", ip.LastIPOctet)
		fmt.Printf("%s %s\n", fill+"Address", ip.Address)
		fmt.Printf("%s %s\n", fill+"Mask", ip.Mask)
		fmt.Printf("%s %s\n", fill+"Subnet", ip.Mask)
		fmt.Printf("%s %s\n", fill+"Gateway", ip.Gateway)
	}
}
func DumpVolumes(vols *[]vmlayer.VolumeOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if vols == nil {
		fmt.Printf("%s\n", fill+"No VM params")
		return
	}
	for _, vol := range *vols {
		fmt.Printf("%s %s\n", fill+"Name", vol.Name)
		fmt.Printf("%s %s\n", fill+"ImageName", vol.ImageName)
		fmt.Printf("%s %d\n", fill+"Size", vol.Size)
		fmt.Printf("%s %s\n", fill+"AZ", vol.AvailabilityZone)
		fmt.Printf("%s %s\n", fill+"DeviceName", vol.DeviceName)
		fmt.Printf("%s %t\n", fill+"AttachExternalDisk", vol.AttachExternalDisk)
		fmt.Printf("%s %d\n", fill+"UnitNumber", vol.UnitNumber)
	}

}

// Preexisting: indicates whether the resource is already present or is being created
// as part of this operation.
func DumpResourceRefs(resRefs *[]vmlayer.ResourceReference, indent int) {

	fill := strings.Repeat("  ", indent)
	if resRefs == nil {
		fmt.Printf("%s\n", fill+"No VM params")
		return
	}
	for _, ref := range *resRefs {
		fmt.Printf("%s %s\n", fill+"", ref.Name)
		fmt.Printf("%s %s\n", fill+"", ref.Id)
		fmt.Printf("%s %t\n", fill+"", ref.Preexisting)
	}
}

func DumpPortResourceRefs(portRefs *[]vmlayer.PortResourceReference, indent int) {
	fill := strings.Repeat("  ", indent)
	if portRefs == nil {
		fmt.Printf("%s\n", fill+"No VM params")
		return
	}
	for _, ref := range *portRefs {
		fmt.Printf("%s %s\n", fill+"", ref.Name)
	}
}

// Add details needed by the orchestrator
func DumpOrchParamsVMs(vms *[]vmlayer.VMOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if vms == nil {
		fmt.Printf("%s\n", fill+"No VM params")
		return
	}

	for _, vm := range *vms {
		fmt.Printf("%s %s\n", fill+"Id", vm.Id)
		fmt.Printf("%s %s\n", fill+"Name", vm.Name)
		fmt.Printf("%s %s\n", fill+"VMRole", vm.Role)
		fmt.Printf("%s %s\n", fill+"ImageName", vm.ImageName)
		//		fmt.Printf("%s %s\n", fill+"TemplateId", vm.TemplateId)   apparently removed as not needed or generic.
		fmt.Printf("%s %s\n", fill+"ImageFolder", vm.ImageFolder)
		fmt.Printf("%s %s\n", fill+"HostName", vm.HostName)
		fmt.Printf("%s %s\n", fill+"DNSDomain", vm.DNSDomain)
		fmt.Printf("%s %s\n", fill+"FlavorName", vm.FlavorName)
		fmt.Printf("%s %d\n", fill+"Vcpus", vm.Vcpus)
		fmt.Printf("%s %d\n", fill+"Rame", vm.Ram)
		fmt.Printf("%s %d\n", fill+"Disk", vm.Disk)
		fmt.Printf("%s %s\n", fill+"Compute AZ", vm.ComputeAvailabilityZone)
		fmt.Printf("%s %s\n", fill+"UserData", vm.UserData)
		fmt.Printf("%s %s\n", fill+"MetaData", vm.MetaData)
		fmt.Printf("%s %t\n", fill+"SharedVolume", vm.SharedVolume)
		fmt.Printf("%s %s\n", fill+"DNSServers", vm.DNSServers)
		fmt.Printf("%s %s\n", fill+"AuthPublicKey", vm.AuthPublicKey)
		fmt.Printf("%s %s\n", fill+"DeploymentManifest", vm.DeploymentManifest)

		fmt.Printf("%s %s\n", fill+"Command", vm.Command)
		fmt.Printf("%s\n", fill+"Volumes:")
		DumpVolumes(&vm.Volumes, indent+1)

		fmt.Printf("%s\n", fill+"Ports:")
		DumpPortResourceRefs(&vm.Ports, indent+1)

		fmt.Printf("%s\n", fill+"FixedIPs")
		DumpFixedIPOrchParams(&vm.FixedIPs, indent+1)
		fmt.Printf("%s %s\n", fill+"Auth Public Key ", vm.AuthPublicKey)
	}

}
func DumpFloatIPs(fips *[]vmlayer.FloatingIPOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if fips == nil {
		fmt.Printf("%s\n", fill+"No Floating IPs")
		return
	}

}

func DumpSecGrps(sg *[]vmlayer.SecurityGroupOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if sg == nil {
		fmt.Printf("%s\n", fill+"No security groups")
		return
	}

}

func DumpTagOrchParams(tags *[]vmlayer.TagOrchestrationParams, indent int) {
	fill := strings.Repeat("  ", indent)
	if tags == nil {
		fmt.Printf("%s\n", fill+"No tags")
		return
	}
	for _, tag := range *tags {
		fmt.Printf("%s %s\n", fill+"Id", tag.Id)
		fmt.Printf("%s %s\n", fill+"Name", tag.Name)
		fmt.Printf("%s %s\n", fill+"Catagory", tag.Category)
	}
}

func DumpNetSpec(netspec *vmlayer.NetSpecInfo, indent int) {
	fill := strings.Repeat("  ", indent)
	if netspec == nil {
		fmt.Printf("%s\n", fill+"No netspec")
		return
	}
	fmt.Printf("%s %s\n", fill+"CIDR", netspec.CIDR)

	fmt.Printf("%s %s\n", fill+"NetworkType", netspec.NetworkType)
	fmt.Printf("%s %s\n", fill+"NetworkAddress", netspec.NetworkAddress)
	fmt.Printf("%s %s\n", fill+"NetmaskBits", netspec.NetmaskBits)

	fmt.Printf("%s\n", fill+"Octets")
	for _, octet := range netspec.Octets {
		fmt.Printf("%s %s\n", fill+fill+"Octet", octet)
	}
	fmt.Printf("%s %s\n", fill+"MasterIPLastOctet", netspec.MasterIPLastOctet)
	// this is the X ?
	fmt.Printf("%s %d\n", fill+"DelimiterOctet", netspec.DelimiterOctet)
	fmt.Printf("%s %s\n", fill+"FloatingIPNet", netspec.FloatingIPNet)
	fmt.Printf("%s %s\n", fill+"VnicType", netspec.VnicType)

	fmt.Printf("%s %s\n", fill+"RouterGatewayIP", netspec.RouterGatewayIP)

}

func DumpMediaSeetting(ms *types.MediaSettings, indent int) {
	fill := strings.Repeat("  ", indent)
	if ms == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	fmt.Printf("%s %s\n", fill+"DeviceId", ms.DeviceId)
	fmt.Printf("%s\n", fill+"MediaImage:")
	DumpReference(ms.MediaImage, indent+1)
	fmt.Printf("%s %s\n", fill+"MediaType", ms.MediaType)
	fmt.Printf("%s %s\n", fill+"MediaState", ms.MediaState)
	fmt.Printf("%s %d\n", fill+"UnitNumber", ms.UnitNumber)
	fmt.Printf("%s %d\n", fill+"BusNumber", ms.BusNumber)
	fmt.Printf("%s %s\n", fill+"AdapterType", ms.AdapterType)
}

func DumpMediaSetting(ms *types.MediaSettings, indent int) {
	fill := strings.Repeat("  ", indent)
	if ms == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	fmt.Printf("%s %s\n", fill+"DeviceId", ms.DeviceId)
	fmt.Printf("%s\n", fill+"MediaImage")
	DumpReference(ms.MediaImage, indent+1)
	fmt.Printf("%s %s\n", fill+".MediaType", ms.MediaType)
	fmt.Printf("%s %s\n", fill+"MediaState", ms.MediaState)
	fmt.Printf("%s %d\n", fill+"UnitNumber", ms.UnitNumber)
	fmt.Printf("%s %d\n", fill+"BusNumber", ms.BusNumber)
	fmt.Printf("%s %s\n", fill+"AdapterType", ms.AdapterType)

}

func DumpMediaSection(ms *types.MediaSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if ms == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	settings := ms.MediaSettings
	for _, setting := range settings {
		DumpMediaSetting(setting, indent+1)
	}
}

func DumpDiskSettings(ds *types.DiskSettings, indent int) {
	fill := strings.Repeat("  ", indent)
	if ds == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	fmt.Printf("%s %s\n", fill+"DiskId", ds.DiskId)
	fmt.Printf("%s %d\n", fill+"SizeMb", ds.SizeMb)
	fmt.Printf("%s %d\n", fill+"UnitNumber", ds.UnitNumber)
	fmt.Printf("%s %d\n", fill+"BusNumber", ds.BusNumber)
	fmt.Printf("%s %s\n", fill+"AdapterType", ds.AdapterType)
	fmt.Printf("%s %t\n", fill+"ThinProvisioned", *ds.ThinProvisioned)
	fmt.Printf("%s\n", fill+"Disk:")
	DumpReference(ds.Disk, indent+1)
	fmt.Printf("%s\n", fill+"StorageProfile")
	DumpReference(ds.StorageProfile, indent+1)
	fmt.Printf("%s %t\n", fill+"OverrideVmDefault", ds.OverrideVmDefault)
	fmt.Printf("%s %d\n", fill+"Iops", ds.Iops)
	fmt.Printf("%s %d\n", fill+"VirtualQuantity", ds.VirtualQuantity)
	fmt.Printf("%s %s\n", fill+"VirtualQuantityUnit", ds.VirtualQuantityUnit)

}

func DumpDiskSection(ds *types.DiskSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if ds == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	for _, msettings := range ds.DiskSettings {
		DumpDiskSettings(msettings, indent+1)
	}
}

func DumpHardwareVersion(hv *types.HardwareVersion, indent int) {
	fill := strings.Repeat("  ", indent)
	if hv == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}

	fmt.Printf("%s %s\n", fill+"HREF", hv.HREF)
	fmt.Printf("%s %s\n", fill+"Type", hv.Type)
	fmt.Printf("%s %s\n", fill+"Value", hv.Value)
}

func DumpMemoryResourceMb(mrmb *types.MemoryResourceMb, indent int) {
	fill := strings.Repeat("  ", indent)
	if mrmb == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	fmt.Printf("%s %d\n", fill+"Configured", mrmb.Configured)
	fmt.Printf("%s %d\n", fill+"Reservation", mrmb.Reservation)
	fmt.Printf("%s %d\n", fill+"Limit", mrmb.Limit)
	fmt.Printf("%s %s\n", fill+"", mrmb.SharesLevel)
	fmt.Printf("%s %d\n", fill+"", mrmb.Shares)
}

func DumpVmSpecSection(vss *types.VmSpecSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if vss == nil {
		fmt.Printf("%s\n", fill+"None")
		return
	}
	fmt.Printf("%s %v\n", fill+"Modified", vss.Modified)
	fmt.Printf("%s %s\n", fill+"Info", vss.Info)
	fmt.Printf("%s %s\n", fill+"OsType", vss.OsType)
	fmt.Printf("%s %d\n", fill+"NumCpus", vss.NumCpus)
	fmt.Printf("%s %d\n", fill+"NumCoresPerSocket", vss.NumCoresPerSocket)
	fmt.Printf("%s\n", fill+"MemoryResourceMb:")
	DumpMemoryResourceMb(vss.MemoryResourceMb, indent+1)
	fmt.Printf("%s\n", fill+"MediaSection:")
	DumpMediaSection(vss.MediaSection, indent+1)
	fmt.Printf("%s\n", fill+"DiskSection:")
	DumpDiskSection(vss.DiskSection, indent+1)
	fmt.Printf("%s \n", fill+"HardwareVersion:")
	DumpHardwareVersion(vss.HardwareVersion, indent+1)

}

func DumpVMGroupParams(vmgp *vmlayer.VMGroupOrchestrationParams, indent int) {

	fill := strings.Repeat("  ", indent)
	if vmgp == nil {
		fmt.Printf("%s\n", fill+"Nil Params")
		return
	}

	subnets := vmgp.Subnets
	ports := vmgp.Ports
	routerIntfs := vmgp.RouterInterfaces
	vms := vmgp.VMs
	sgs := vmgp.SecurityGroups
	netspec := vmgp.Netspec
	fmt.Printf("%s %s\n", fill+"GroupName", vmgp.GroupName)

	fmt.Printf("%s\n", fill+"Subnets:")
	DumpSubnetOrchParams(&subnets, indent+1)
	fmt.Printf("%s\n", fill+"Ports:")

	DumpPortsOrchParams(&ports, indent+1)

	fmt.Printf("%s\n", fill+"Router Interfaces:")
	DumpRouterInterfaceParams(&routerIntfs, indent+1)

	fmt.Printf("%s\n", fill+"VMs:")

	DumpOrchParamsVMs(&vms, indent+1)

	fmt.Printf("%s\n", fill+"FloatingIPs:")
	DumpFloatIPs(&vmgp.FloatingIPs, indent+1)

	fmt.Printf("%s\n", fill+"SecurityGroups")
	DumpSecGrps(&sgs, indent+1)
	fmt.Printf("%s\n", fill+"NetSpec")
	DumpNetSpec(netspec, indent+1)
	fmt.Printf("%s\n", fill+"Tags")
	DumpTagOrchParams(&vmgp.Tags, indent+1)

	fmt.Printf("%s %t\n", fill+"SkipInfraSpecificCheck", vmgp.SkipInfraSpecificCheck)
	fmt.Printf("%s %t\n", fill+".SkipSubnetGateway", vmgp.SkipSubnetGateway)

	fmt.Printf("%s %t\n", fill+"InitOrchestrator", vmgp.InitOrchestrator)
	fmt.Printf("%s %t\n", fill+"CleanupOnFailure", vmgp.SkipCleanupOnFailure)

	// TBNI CheckUpdateInfo map[string]string
}

//func getIPRanges(fromScope *types.IPScopes)
func DumpIPRanges(ipRanges *types.IPRanges, indent int) {
	fill := strings.Repeat("  ", indent)

	if ipRanges == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	for n, ipRange := range ipRanges.IPRange {
		fmt.Printf("%s %d %s\n", fill+"StartAddress", n, ipRange.StartAddress)
		fmt.Printf("%s %d %s\n", fill+"EndAddress  ", n, ipRange.EndAddress)
	}
}

func DumpIPScopes(ips *types.IPScopes, indent int) {

	fill := strings.Repeat("  ", indent)
	if ips == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	var scope []*types.IPScope

	scope = ips.IPScope

	for _, s := range scope {
		fmt.Printf("%s %t\n", fill+"Inherited", s.IsInherited)
		fmt.Printf("%s %s\n", fill+"Gateway", s.Gateway)
		fmt.Printf("%s %s\n", fill+"Netmask", s.Netmask)
		fmt.Printf("%s %s\n", fill+"DNS1", s.DNS1)
		fmt.Printf("%s %s\n", fill+"DNS2", s.DNS2)
		fmt.Printf("%s %s\n", fill+"DNSSuffix", s.DNSSuffix)
		fmt.Printf("%s %t\n", fill+"IsEnabled", s.IsEnabled)

		fmt.Printf("%s\n", fill+"IPRanges:")
		DumpIPRanges(s.IPRanges, indent+1)

		if s.AllocatedIPAddresses != nil { // xxx might blow when we have some allocated
			fmt.Printf("%s %s\n", fill+"AllocatedIPAddresses", s.AllocatedIPAddresses.IPAddress)
		} else {
			fmt.Printf("%s\n", fill+"AllocatedIPAddresses None")
		}
		fmt.Printf("%s\n", fill+"SubAllocations TBI") // s.SubAllocations)
	}

}

func DumpNetworkConnectionSection(nc *types.NetworkConnectionSection, indent int) {

	fill := strings.Repeat("  ", indent)
	if nc == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("TBI")
}

func DumpNetworkFeatures(nfs *types.NetworkFeatures, indent int) {
	fill := strings.Repeat("  ", indent)
	if nfs == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %+v\n", fill+"DHCPService", nfs.DhcpService)
	fmt.Printf("%s %+v\n", fill+"FirewallService", nfs.FirewallService)
	fmt.Printf("%s %+v\n", fill+"NatService", nfs.NatService)
	fmt.Printf("%s %+v\n", fill+"StaticRoutingService", nfs.StaticRoutingService)
	// TODO Not Impl. => IpsecVpnService Substitue for NetworkService
}

func DumpNetworkConfiguration(nc *types.NetworkConfiguration, indent int) {
	fill := strings.Repeat("  ", indent)

	fmt.Printf("%s %s\n", fill+"Xmlns", nc.Xmlns)
	fmt.Printf("%s %t\n", fill+"BackwardCompatibilityMode", nc.BackwardCompatibilityMode)

	fmt.Printf("%s\n", fill+"IPScope:")
	DumpIPScopes(nc.IPScopes, indent+1)
	if nc.ParentNetwork == nil {
		fmt.Printf("%s\n", fill+"ParentNetwork  None")
	} else {
		fmt.Printf("ParentNetwork %+v\n", nc.ParentNetwork)
	}
	fmt.Printf("%s %s\n", fill+"FenceMode", nc.FenceMode)
	fmt.Printf("%s \n", fill+"Features TBI") //, nc.Features)
	DumpNetworkFeatures(nc.Features, indent+1)
}

func DumpVimPortGroup(portgroup []*types.VimObjectRef, indent int) {
	fill := strings.Repeat("  ", indent)
	if portgroup == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	for _, pg := range portgroup {

		fmt.Printf("%s\n", fill+"VimServerRef TBI") //%+v\n", pg.VimServerRef) = the server we're logged into
		fmt.Printf("%s %s\n", fill+"MoRef", pg.MoRef)
		fmt.Printf("%s\n", fill+"VimObjectType TBI") // %+v\n", pg.VimObjectType)
	}
}

func DumpVirtualHardwareConnection(vhc *types.VirtualHardwareConnection, indent int) {
	fill := strings.Repeat("  ", indent)
	if vhc == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"IPAddress         ", vhc.IPAddress)
	fmt.Printf("%s %t\n", fill+"PrimaryConnection ", vhc.PrimaryConnection)
	fmt.Printf("%s %s\n", fill+"IPAdressingMode   ", vhc.IpAddressingMode)
	fmt.Printf("%s %s\n", fill+"NetworkName       ", vhc.NetworkName)
}

func DumpVirtualHardwareHostResource(vhr *types.VirtualHardwareHostResource, indent int) {
	fill := strings.Repeat("  ", indent)
	if vhr == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %d\n", fill+"BusType           ", vhr.BusType)
	fmt.Printf("%s %s\n", fill+"BusSubType        ", vhr.BusSubType)
	fmt.Printf("%s %d\n", fill+"Capacity          ", vhr.Capacity)
	fmt.Printf("%s %s\n", fill+"StorageProfile    ", vhr.StorageProfile)
	fmt.Printf("%s %t\n", fill+"OverrideVmDefault ", vhr.OverrideVmDefault)
	fmt.Printf("%s %s\n", fill+"Disk              ", vhr.Disk)

}

func DumpVirtualHardwareItem(vhi *types.VirtualHardwareItem, indent int) {

	fill := strings.Repeat("  ", indent)
	if vhi == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"XMLName      ", vhi.XMLName)
	fmt.Printf("%s %d\n", fill+"ResourceType", vhi.ResourceType)
	fmt.Printf("%s %s\n", fill+"ResourceSubType", vhi.ResourceSubType)
	fmt.Printf("%s %s\n", fill+"ElementName", vhi.ElementName)
	fmt.Printf("%s %s\n", fill+"Descriptioin", vhi.Description)
	fmt.Printf("%s %d\n", fill+"InstanceID", vhi.InstanceID)
	fmt.Printf("%s %t\n", fill+"AutomaticAllocation", vhi.AutomaticAllocation)
	fmt.Printf("%s %s\n", fill+"Address", vhi.Address)
	fmt.Printf("%s %d\n", fill+"AddressOnParent", vhi.AddressOnParent)
	fmt.Printf("%s %s\n", fill+"AllocationUnits", vhi.AllocationUnits)
	fmt.Printf("%s %d\n", fill+"Reservation", vhi.Reservation)
	fmt.Printf("%s %d\n", fill+"VirtualQuantity", vhi.VirtualQuantity)
	fmt.Printf("%s %d\n", fill+"Weight", vhi.Weight)
	fmt.Printf("%s %d\n", fill+"CoresPerSocket", vhi.CoresPerSocket)
	fmt.Printf("%s\n", fill+"Connection:")
	for _, c := range vhi.Connection {
		DumpVirtualHardwareConnection(c, indent+1)
	}

	fmt.Printf("%s\n", fill+"VirtualHardwareHostResource:")
	for _, hr := range vhi.HostResource {
		DumpVirtualHardwareHostResource(hr, indent+1)
	}

	// Link
	fmt.Printf("%s\n", fill+"Link")
	for _, l := range vhi.Link {
		DumpLink(l, indent+1)
	}
}

func DumpVirtualHardwareSection(vhs *types.VirtualHardwareSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if vhs == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"XMLName  ", vhs.XMLName)
	fmt.Printf("%s %s\n", fill+"Xmlns    ", vhs.Xmlns)
	fmt.Printf("%s %s \n", fill+"Info    ", vhs.Info)
	fmt.Printf("%s %s\n", fill+"HREF     ", vhs.HREF)
	fmt.Printf("%s %s\n", fill+"Type:", vhs.Type)
	fmt.Printf("%s\n", fill+"Item:")
	for _, i := range vhs.Item {
		DumpVirtualHardwareItem(i, indent+1)
	}
}

//
func DumpGuestCustomizationSection(cs *types.GuestCustomizationSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if cs == nil {
		fmt.Printf("%s\n", fill+"Not found")
		return
	}
	fmt.Printf("%s %s\n", fill+"Ovf ", cs.Ovf)
	fmt.Printf("%s %s\n", fill+"Xsi", cs.Xsi)
	fmt.Printf("%s %s\n", fill+"HREF", cs.HREF)
	fmt.Printf("%s %s\n", fill+"Type", cs.Type)
	fmt.Printf("%s %s\n", fill+"Info", cs.Info)
	fmt.Printf("%s %t\n", fill+"Enabled", *cs.Enabled)
	// changed sid
	fmt.Printf("%s %s\n", fill+"VirtualMachineID", cs.VirtualMachineID)
	// windows domain join?
	fmt.Printf("%s %t\n", fill+"JoinDomainEnabled", *cs.JoinDomainEnabled)
	fmt.Printf("%s %t\n", fill+"UseOrgSettings", *cs.UseOrgSettings)
	// windows domain name
	fmt.Printf("%s %s\n", fill+"DomainName", cs.DomainName)
	// windows domain user name
	fmt.Printf("%s %s\n", fill+"DomainUserName", cs.DomainUserName)

	fmt.Printf("%s %s\n", fill+"DomainUserPassword", cs.DomainUserPassword)
	fmt.Printf("%s %s\n", fill+"MachineObjectOU", cs.MachineObjectOU)
	fmt.Printf("%s %t\n", fill+"AdminPasswordEnabled", *cs.AdminPasswordEnabled)
	fmt.Printf("%s %t\n", fill+"AdminPasswordAuto", *cs.AdminPasswordAuto)
	fmt.Printf("%s %s\n", fill+"AdminPassword", cs.AdminPassword)
	fmt.Printf("%s %t\n", fill+"AdminAutoLogonEnabled", *cs.AdminAutoLogonEnabled)

	fmt.Printf("%s %d\n", fill+"AdminAutoLogonCount", cs.AdminAutoLogonCount)
	fmt.Printf("%s %t\n", fill+"ResetPasswordRequired", *cs.ResetPasswordRequired)

	fmt.Printf("%s %s\n", fill+"CustomizationScript", cs.CustomizationScript)
	//vm hostname
	fmt.Printf("%s %s\n", fill+"ComputerName", cs.ComputerName)
	fmt.Printf("%s\n", fill+"Link")
	DumpLinkList(cs.Link, indent+1)
}

func DumpVdcResourceEntities(vdc *types.Vdc, indent int) {
	fill := strings.Repeat("  ", indent)
	if vdc == nil {
		fmt.Printf("%s\n", fill+"Not found")
		return
	}

	resents := vdc.ResourceEntities
	for _, res := range resents {
		resRef := res.ResourceEntity
		for _, ref := range resRef {
			fmt.Printf("%s %s\n", fill+"HREF", ref.HREF)
			fmt.Printf("%s %s\n", fill+"ID", ref.ID)
			fmt.Printf("%s %s\n", fill+"Type", ref.Type)
			fmt.Printf("%s %s\n", fill+"Name", ref.Name)
			fmt.Printf("%s %s\n", fill+"Status", ref.Status)
		}
	}
}

func DumpVM(vm *types.VM, indent int) {

	fill := strings.Repeat("  ", indent)
	if vm == nil {
		fmt.Printf("%s\n", fill+"Not found")
		return
	}
	fmt.Printf("%s %s\n", fill+"Name         ", vm.Name)
	fmt.Printf("%s %s\n", fill+"HREF         ", vm.HREF)
	fmt.Printf("%s %s\n", fill+"Type         ", vm.Type)
	fmt.Printf("%s %s\n", fill+"ID           ", vm.ID)

	fmt.Printf("%s %d\n", fill+"Status       ", vm.Status)
	fmt.Printf("%s %t\n", fill+"Deployed     ", vm.Deployed)
	fmt.Printf("%s\n", fill+"VAppParent:")
	DumpReference(vm.VAppParent, indent+1)

	fmt.Printf("%s %t\n", fill+"NeedsCustomization    ", vm.NeedsCustomization)
	fmt.Printf("%s %t\n", fill+"NestedHypervisorEnabled", vm.NestedHypervisorEnabled)

	fmt.Printf("%s %s\n", fill+"Created         ", vm.DateCreated)
	fmt.Printf("%s\n", fill+"VirtualHardwareSection")
	DumpVirtualHardwareSection(vm.VirtualHardwareSection, indent+1)

	fmt.Printf("%s\n", fill+"NetworkConnectionSection:")
	DumpNetworkConnectionSection(vm.NetworkConnectionSection, indent+1)

	fmt.Printf("%s %s\n", fill+"VAppScopedLocalID", vm.VAppScopedLocalID)

	fmt.Printf("%s\n", fill+"GuestCustomizationSection:")
	DumpGuestCustomizationSection(vm.GuestCustomizationSection, indent+1)

	//	fmt.Printf("%s\n", fill+"VMCapabilities:")
	//	DumpVMCapabilities(vm.VMCapabilities, indent+1)

	fmt.Printf("%s\n", fill+"StorageProfile:")
	DumpReference(vm.StorageProfile, indent+1)
	fmt.Printf("%s\n", fill+"Media:")
	DumpReference(vm.Media, indent+1)
}

func DumpOrgVDCNetwork(vdcnet *types.OrgVDCNetwork, indent int) {
	fill := strings.Repeat("  ", indent)
	if vdcnet == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"XMLName      ", vdcnet.XMLName)
	fmt.Printf("%s %s\n", fill+"Xmlns        ", vdcnet.Xmlns)
	fmt.Printf("%s %s\n", fill+"HREF         ", vdcnet.HREF)
	fmt.Printf("%s %s\n", fill+"Type         ", vdcnet.Type)
	fmt.Printf("%s %s\n", fill+"ID           ", vdcnet.ID)
	fmt.Printf("%s %s\n", fill+"OperationKey ", vdcnet.OperationKey)
	fmt.Printf("%s %s\n", fill+"Name         ", vdcnet.Name)
	fmt.Printf("%s %s\n", fill+"Status       ", vdcnet.Status)

	// TBI
	//fmt.Printf("fill+Link %+v \n", vdcnet.Link)

	fmt.Printf("%s %s\n", fill+"Description  ", vdcnet.Description)

	fmt.Printf("%s\n", fill+"NetworkConfiguration:")
	DumpNetworkConfiguration(vdcnet.Configuration, indent+1)
	fmt.Printf("%s %s\n", fill+"EdgeGateway  ", vdcnet.EdgeGateway)
	// TBI
	//	fmt.Printf("%s %s\n", fill+"ServiceConfig" %+v\n", vdcnet.ServiceConfig)
	fmt.Printf("%s %t\n", fill+"Shared       ", vdcnet.IsShared)

	fmt.Printf("%s\n", fill+"VimPortGroup:")
	DumpVimPortGroup(vdcnet.VimPortGroupRef, indent+1)

	// TBI if wanted
	//	fmt.Printf("%s %s\n", fill+      "Tasks" %+v \n", vdcnet.Tasks)
	fmt.Printf("\n")
}

func DumpVAppNetworkConfiguration(nc *types.VAppNetworkConfiguration, indent int) {

	fill := strings.Repeat("  ", indent)

	if nc == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"Name:", nc.NetworkName)
	fmt.Printf("%s %s\n", fill+"HREF:", nc.HREF)
	fmt.Printf("%s %s\n", fill+"Type:", nc.Type)
	fmt.Printf("%s %s\n", fill+"ID: ", nc.ID)
	fmt.Printf("%s %t\n", fill+"Deployed:", nc.IsDeployed)
	fmt.Printf("%s %s\n", fill+"Description:", nc.Description)
	DumpNetworkConfiguration(nc.Configuration, indent+1)
}

func DumpNetworkConfigSection(c *types.NetworkConfigSection, indent int) {
	fill := strings.Repeat("  ", indent)

	if c == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s   %s\n", fill+"XMLName", c.XMLName)
	fmt.Printf("%s %s\n", fill+"Xmlns ", c.Xmlns)
	fmt.Printf("%s %s\n", fill+"Ovf", c.Ovf)
	fmt.Printf("%s %s \n", fill+"Info ", c.Info)
	for _, n := range c.NetworkConfig {
		DumpVAppNetworkConfiguration(&n, indent+1)
	}
	return
}

func DumpFilesList(flist *types.FilesList, indent int) {
	fill := strings.Repeat("  ", indent)
	if flist == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	for _, f := range flist.File {
		fmt.Printf("%s %s\n", fill+"Name", f.Name)
		fmt.Printf("%s %s\n", fill+"HREF", f.HREF)
		fmt.Printf("%s %s\n", fill+"Type", f.Type)
		fmt.Printf("%s %s\n", fill+"ID", f.ID)
		fmt.Printf("%s %s\n", fill+"OpKey", f.OperationKey)
		fmt.Printf("%s %d\n", fill+"Size", f.Size)

		fmt.Printf("%s %d\n", fill+"BytesTranfered", f.BytesTransferred)
		fmt.Printf("%s %s\n", fill+"Checksum", f.Checksum)
		fmt.Printf("%s %s\n", fill+"Description", f.Description)
		fmt.Printf("%s\n", fill+"Link:")
		for _, f := range f.Link {
			DumpLink(f, indent+1)
		}
		// Tasks
	}
	return
}

func DumpVAppChildren(children *types.VAppChildren, indent int) {

	fill := strings.Repeat("  ", indent)

	if children == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	for _, vm := range children.VM {
		DumpVM(vm, indent+1)
	}
	return
}

/*
func DumpVAppTemplateChildren(tv *VcdPlatform, ctx context.Context, tc *types.VAppTemplateChildren, indent int) {
	fill := strings.Repeat("  ", indent)
	if tc == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	for _, vmt := range tc.VM {
		vat := govcd.VAppTemplate{
			VAppTemplate: vmt,
		}
		DumpVAppTemplate(tv, ctx, &vat, indent+1)
	}
}
*/

func DumpLink(l *types.Link, indent int) {
	fill := strings.Repeat("  ", indent)

	if l == nil {
		fmt.Printf("%s\n", fill+"Nil Link")
		return
	}
	// type and Href are most significant, need to retrive by type/name
	fmt.Printf("----DumpLink ---- \n")
	fmt.Printf("\t%s %s\n", fill+"Rel", l.Rel)
	fmt.Printf("\t%s %s\n", fill+"HREF", l.HREF)
	// fmt.Printf("\t%s %s\n", fill+"ID", l.ID)  looks like ID is not used for Link, and Name is apparently optional
	fmt.Printf("\t%s %s\n", fill+"Type", l.Type)
	fmt.Printf("\t%s %s\n", fill+"Name", l.Name)
}

func DumpLinkList(ll types.LinkList, indent int) {

	fill := strings.Repeat("  ", indent)
	if len(ll) == 0 {
		fmt.Printf("%s\n", fill+"None Found")
	}
	for _, link := range ll {
		DumpLink(link, indent+1)
	}
}

func DumpReference(r *types.Reference, indent int) {
	fill := strings.Repeat("  ", indent)

	if r == nil {
		fmt.Printf("%s\n", fill+"Nil Refereence")
		return
	}

	fmt.Printf("--------DumpReference-------\n")
	fmt.Printf("%s %s\n", fill+"Name", r.Name)
	fmt.Printf("%s %s\n", fill+"HREF", r.HREF)
	fmt.Printf("%s %s\n", fill+"ID", r.ID)
	fmt.Printf("%s %s\n", fill+"Type", r.Type)

}
func DumpCustomizationStatusSection(ss *types.GuestCustomizationStatusSection, indent int) {
	fill := strings.Repeat("  ", indent)

	if ss == nil {
		fmt.Printf("%s\n", fill+"Nil Refereence")
		return
	}
	fmt.Printf("%s %s\n", fill+"GuestCustStatus", ss.GuestCustStatus)
}

func DumpCustomizationSection(cs *types.CustomizationSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if cs == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"Info", cs.Info)
	fmt.Printf("%s %t\n", fill+"GoldMaster", cs.GoldMaster)
	fmt.Printf("%s %s\n", fill+"HREF", cs.HREF)
	fmt.Printf("%s %s\n", fill+"Type", cs.Type)
	fmt.Printf("%s %s\n", fill+"HREF", cs.HREF)
	fmt.Printf("%s %t\n", fill+"CustomizeOnInstanciate", cs.CustomizeOnInstantiate)
	fmt.Printf("%s\n", fill+"Link")
	// LinkList
	DumpLinkList(cs.Link, indent+1)

}

func DumpOwner(o *types.Owner, indent int) {
	fill := strings.Repeat("  ", indent)
	if o == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"HREF", o.HREF)
	fmt.Printf("%s %s\n", fill+"Type", o.Type)
	// Link
	DumpReference(o.User, indent+1)
}

func DumpLeaseSettingSection(ls *types.LeaseSettingsSection, indent int) {
	fill := strings.Repeat("  ", indent)
	if ls == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"HREF", ls.HREF)
	fmt.Printf("%s %s\n", fill+"Type", ls.Type)
	fmt.Printf("%s %s\n", fill+"DeploymenLeaseExpiration", ls.DeploymentLeaseExpiration)
	fmt.Printf("%s %d\n", fill+"DeploymentLeaseInSeconds", ls.DeploymentLeaseInSeconds)
	fmt.Printf("%s \n", fill+"Link")
	DumpLink(ls.Link, indent+1)

	fmt.Printf("%s %s\n", fill+"StorageLeaseExpiration", ls.StorageLeaseExpiration)
	fmt.Printf("%s %d\n", fill+"StorageLeaseInSeconds", ls.StorageLeaseInSeconds)
}

func DumpVApp(vapp *govcd.VApp, indent int) {

	fill := strings.Repeat("  ", indent)
	if vapp == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"Name", vapp.VApp.Name)
	fmt.Printf("%s %s\n", fill+"HREF", vapp.VApp.HREF)
	fmt.Printf("%s %s\n", fill+"Type", vapp.VApp.Type)
	fmt.Printf("%s %s\n", fill+"ID", vapp.VApp.ID)
	fmt.Printf("%s %s\n", fill+"OpKey", vapp.VApp.OperationKey)

	fmt.Printf("%s %d\n", fill+"Status", vapp.VApp.Status)
	fmt.Printf("%s %s\n", fill+"Deployed", vapp.VApp.ID)
	fmt.Printf("%s %t\n", fill+"OvfDescriptorUpLoaded", vapp.VApp.OvfDescriptorUploaded)
	fmt.Printf("%s:\n", fill+"NetConfigSecion:")
	DumpNetworkConfigSection(vapp.VApp.NetworkConfigSection, indent+1)

	fmt.Printf("%s %s\n", fill+"Description ", vapp.VApp.Description)
	fmt.Printf("%s\n", fill+"Files:")
	DumpFilesList(vapp.VApp.Files, indent+1)

	fmt.Printf("%s\n", fill+"VAppParent:")
	DumpReference(vapp.VApp.VAppParent, indent+1)

	fmt.Printf("%s %s\n", fill+"Created", vapp.VApp.DateCreated)
	fmt.Printf("%s\n", fill+"Owner:")
	DumpOwner(vapp.VApp.Owner, indent+1)
	fmt.Printf("%s\n", fill+"Vapp Children:")
	DumpVAppChildren(vapp.VApp.Children, indent+1)
}

func DumpTasksInProgress(tasks *types.TasksInProgress, indent int) {
	fill := strings.Repeat("  ", indent)
	if tasks == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	DumpTasks(tasks.Task, indent+1)

}

func DumpTasks(tasks []*types.Task, indent int) {
	fill := strings.Repeat("  ", indent)
	if len(tasks) == 0 {
		fmt.Printf("%s", fill+"No tasks")
		return
	}

	for _, task := range tasks {
		fmt.Printf("%s %s\n", fill+"HREF", task.HREF)
		fmt.Printf("%s %s\n", fill+"Type", task.Type)
		fmt.Printf("%s %s\n", fill+"ID", task.ID)

	}
}

func DumpEntity(ent *types.Entity, indent int) {

	fill := strings.Repeat("  ", indent)
	if ent == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("%s %s\n", fill+"HREF", ent.HREF)
	fmt.Printf("%s %s\n", fill+"HREF", ent.Type)
	fmt.Printf("%s %s\n", fill+"HREF", ent.ID)
	fmt.Printf("%s %s\n", fill+"HREF", ent.OperationKey)
	fmt.Printf("%s %s\n", fill+"HREF", ent.Name)
	fmt.Printf("%s %s\n", fill+"HREF", ent.Description)
	fmt.Printf("%s\n", fill+"Link")
	DumpLinkList(ent.Link, indent+1)
	fmt.Printf("%s\n", fill+"Tasks")
	DumpTasksInProgress(ent.Tasks, indent+1)
}

func DumpCatalogItems(catItems []*types.CatalogItems, indent int) {
	fill := strings.Repeat("  ", indent)
	if len(catItems) == 0 {
		fmt.Printf("%s", fill+"No items")
		return
	}
	fmt.Printf("DumpEnntity----------\n")
	for _, item := range catItems {
		for _, ref := range item.CatalogItem {
			DumpReference(ref, indent+1)
			/*
				for _, item := range items.CatalogItem {



					fmt.Printf("%s %s\n", fill+"Name", item.Name)
					fmt.Printf("%s %s\n", fill+"HREF", item.HREF)
					fmt.Printf("%s %s\n", fill+"Type", item.Type)
					fmt.Printf("%s %s\n", fill+"ID", item.ID)

					fmt.Printf("%s %s\n", fill+"OpKey", item.OperationKey)
					fmt.Printf("%s %d\n", fill+"Size", item.Size)
					fmt.Printf("%s %s\n", fill+"Created", item.DataCreated)
					fmt.Printf("%s %s\n", fill+"Description", item.Description)
					fmt.Printf("%s %s\n", fill+"Entity")
					DumpEntity(item.Entity, indent+1)

					fmt.Printf("%s %s\n", fill+"Link")
					DumpLinkList(item.LinkList, indent+1)

					fmt.Printf("%s\n", fill+"TaskInProgress")
					DumpTasks(item.TasksInProgress, indent+1)

					fmt.Printf("%s %d\n", fill+"Version", item.VersionNumber)
				}
			*/

		}
	}
}

func DumpCatalog(cat *govcd.Catalog, indent int) {
	fill := strings.Repeat("  ", indent)
	if cat == nil || cat.Catalog == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}

	fmt.Printf("\n------DumpCatalog----------\n")

	fmt.Printf("%s %s\n", fill+"Name", cat.Catalog.Name)
	fmt.Printf("%s %s\n", fill+"HREF", cat.Catalog.HREF)
	fmt.Printf("%s %s\n", fill+"Type", cat.Catalog.Type)
	fmt.Printf("%s %s\n", fill+"ID", cat.Catalog.ID)
	fmt.Printf("%s %s\n", fill+"OpKey", cat.Catalog.OperationKey)

	fmt.Printf("%s\n", fill+"CatalogItems:")
	DumpCatalogItems(cat.Catalog.CatalogItems, indent+1)

	fmt.Printf("%s %s\n", fill+"DateCreated", cat.Catalog.DateCreated)
	fmt.Printf("%s %s\n", fill+"Description", cat.Catalog.Description)
	fmt.Printf("%s %t\n", fill+"IsPublished", cat.Catalog.IsPublished)

	fmt.Printf("%s\n", fill+"Owner")
	DumpOwner(cat.Catalog.Owner, indent+1)
	fmt.Printf("%s %s\n", fill+"Description", cat.Catalog.Description)

	fmt.Printf("%s\n", fill+"TaskInProgress")
	DumpTasksInProgress(cat.Catalog.Tasks, indent+1)

	fmt.Printf("%s %d\n", fill+"VersionNumber", cat.Catalog.VersionNumber)

}

func DumpMetadataEntry(mde *types.MetadataEntry, indent int) {
	fill := strings.Repeat("  ", indent)
	if mde == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"Type", mde.Type)
	fmt.Printf("%s %s\n", fill+"Xsi", mde.Xsi)
	fmt.Printf("%s %s\n", fill+"Domain", mde.Domain)
	fmt.Printf("%s %s\n", fill+"Key", mde.Key)
	for _, l := range mde.Link {
		DumpLink(l, indent+1)
	}
	fmt.Printf("%s %s\n", fill+"TypedValue.XsiType", mde.TypedValue.XsiType)
	fmt.Printf("%s %s\n", fill+"TypedValue.Value", mde.TypedValue.Value)
}

func DumpMetadata(md *types.Metadata, indent int) {
	fill := strings.Repeat("  ", indent)
	if md == nil {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	fmt.Printf("%s %s\n", fill+"XMLName", md.XMLName)
	fmt.Printf("%s %s\n", fill+"HREF", md.HREF)
	fmt.Printf("%s %s\n", fill+"Type", md.Type)
	fmt.Printf("%s %s\n", fill+"Type", md.Xsi)
	fmt.Printf("%s\n", fill+"Link")
	for _, l := range md.Link {
		DumpLink(l, indent+1)
	}
	fmt.Printf("%s\n", fill+"MetadataEntry:")
	for _, mde := range md.MetadataEntry {
		DumpMetadataEntry(mde, indent+1)
	}
}
func DumpMediaRecords(mrs []*types.MediaRecordType, indent int) {
	fill := strings.Repeat("  ", indent)
	if len(mrs) == 0 {
		fmt.Printf("%s\n", fill+"None found")
		return
	}
	for _, mr := range mrs {
		fmt.Printf("%s %s\n", fill+"Name", mr.Name)
		fmt.Printf("%s %s\n", fill+"HREF", mr.HREF)
		fmt.Printf("%s %s\n", fill+"Type", mr.Type)
		fmt.Printf("%s %s\n", fill+"ID", mr.ID)
		fmt.Printf("%s %s\n", fill+"OwnerName", mr.OwnerName)
		fmt.Printf("%s %s\n", fill+"CatalogName", mr.CatalogName)
		fmt.Printf("%s %t\n", fill+"IsBusy", mr.IsBusy)
		fmt.Printf("%s %d\n", fill+"StorageB", mr.StorageB)
		fmt.Printf("%s %s\n", fill+"CatalogItem", mr.CatalogItem)
		fmt.Printf("%s %s\n", fill+"Status", mr.Status)
		fmt.Printf("%s %t\n", fill+"IsIso", mr.IsIso)
		fmt.Printf("%s %s\n", fill+"TaskStatusName", mr.TaskStatusName)
		fmt.Printf("%s %s\n", fill+"TaskStatus", mr.TaskStatus)
		fmt.Printf("%s %s\n", fill+"TaskDetail", mr.TaskDetails)

		fmt.Printf("%s %t\n", fill+"IsInCatalog", mr.IsInCatalog)
		fmt.Printf("%s\n", fill+"Link:")
		DumpLink(mr.Link, indent+1)

		fmt.Printf("%s\n", fill+"Metadata:")
		DumpMetadata(mr.Metadata, indent+1)
	}
}

func TakeBoolPointer(value bool) *bool {
	return &value
}

// takeIntAddress is a helper that returns the address of an `int`
func TakeIntAddress(x int) *int {
	return &x
}

// takeStringPointer is a helper that returns the address of a `string`
func TakeStringPointer(x string) *string {
	return &x
}

// takeFloatAddress is a helper that returns the address of an `float64`
func TakeFloatAddress(x float64) *float64 {
	return &x
}
func TakeIntPointer(x int) *int {
	return &x
}

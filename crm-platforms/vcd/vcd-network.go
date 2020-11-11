package vcd

import (
	"context"
	"fmt"
	"net"
	"strings"

	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Networks

// OrgVDCNetworks
func (v *VcdPlatform) GetOrgNetworks(ctx context.Context, org *govcd.Org) ([]string, error) {
	var networks []string
	if org == nil {
		return networks, fmt.Errorf("Nil Org encountered no networks possible")
	}

	nets := v.Objs.Nets
	//var scopes []*types.IPScopes
	for netname, orgvdcnet := range nets {
		config := orgvdcnet.OrgVDCNetwork.Configuration
		scopes := config.IPScopes.IPScope
		for _, scope := range scopes {
			netmask := net.IPMask(net.ParseIP(scope.Netmask).To4())
			sz, _ := netmask.Size()
			address := fmt.Sprintf("%s/%d", scope.Gateway, sz)
			networks = append(networks, address)
			fmt.Printf("network: %s = %+v address: %s \n", netname, orgvdcnet, address)
		}
	}
	return networks, nil

}

func (v *VcdPlatform) GetNetworkList(ctx context.Context) ([]string, error) {

	return v.GetOrgNetworks(ctx, v.Objs.Org)
}

// No Router
func (v *VcdPlatform) GetRouterDetail(ctx context.Context, routerName string) (*vmlayer.RouterDetail, error) {
	return nil, fmt.Errorf("Router not supported for VCD")
}

func (v *VcdPlatform) GetGatewayForOrgVDCNetwork(ctx context.Context, network *types.OrgVDCNetwork) (string, error) {
	var gateways []string
	if network == nil {
		return "", fmt.Errorf("nil network")
	}
	scopes := network.Configuration.IPScopes.IPScope

	for _, scope := range scopes {
		gateways = append(gateways, scope.Gateway)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Primary", "network", network.Name, "gateway", gateways[0], "of", len(gateways))
	return gateways[0], nil
}

// GetExternalGateway
// Gateway, not EdgeGateway, per network. Always operates on v.Objs.PrimayNet
// Return the IP address of the external Gateway for the given extNetName
func (v *VcdPlatform) GetExternalGateway(ctx context.Context, extNetName string) (string, error) {

	gateway, err := v.GetGatewayForOrgVDCNetwork(ctx, v.Objs.PrimaryNet.OrgVDCNetwork)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "error", err.Error())
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "extNetName", extNetName, "IP", gateway)
	return gateway, err
}

// Ports
func (v *VcdPlatform) GetInternalPortPolicy() vmlayer.InternalPortAttachPolicy {
	fmt.Printf("\nGetInternalPortPolicy returns AttachPortDuringCreate\n")
	return vmlayer.AttachPortDuringCreate
}

// AttachPortToServer
func (v *VcdPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {

	fmt.Printf("AttachPortToServer-I-servername: %s subnetName %s portName %s ipaddr %s action %s  TBI\n",
		serverName, subnetName, portName, ipaddr, string(action))

	// Add the given Network to the given server
	// A Customize
	return nil

}

func (v *VcdPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName, portName string) error {

	return nil
}

// unit test this guy XXX
func (v *VcdPlatform) GetAvailableAddrInRange(ctx context.Context, iprange types.IPRange, scope *types.IPScope) string {
	// Find first available IPaddress in IPRange
	start := iprange.StartAddress
	end := iprange.EndAddress

	// we could caclulate the number of IPs in the pool
	for {
		if (!strings.Contains(scope.AllocatedIPAddresses.IPAddress, start)) && start != end {
			return start
		} else {
			ip := net.ParseIP(start)
			ip = ip.To4()
			ip[3]++
			start = ip.String()
			if start == end {
				break
			}
		}
	}
	return ""
}

func (v *VcdPlatform) IncrIP(a string) string {
	ip := net.ParseIP(a)
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	ip[3]++
	return ip.String()
}
func (v *VcdPlatform) DecrIP(a string) string {
	ip := net.ParseIP(a)
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	ip[3]--
	return ip.String()
}

func (v *VcdPlatform) ThirdOctet(ctx context.Context, a string) (int, error) {
	addr := a
	// strip any cidr mask if present
	parts := strings.Split(a, "/")
	if len(parts) > 1 {
		addr = string(parts[0])
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		fmt.Printf("err from ParseIP")
		return 0, fmt.Errorf("Invalid IP")
	}
	ip = ip.To4()
	return int(ip[2]), nil
}

// VappNetworks

// Add a OrgVdcNetwork to a VApp
// Return *types.NetowrkConfigSection

// When a second clusterInst/VM is added to a cloudlet, the clouddlet's VApp needs to grow a new internal
// network for it
// We've stashed some metadata somewhere that tracks where we are in the allocation of internal network subnets
// XXXX
func (v *VcdPlatform) getInternalNetworkForVApp(ctx context.Context, serverName string) error {
	return nil
}

// Based on the number of vms and the roles, return the set of networks required.
// Currently, to add an isolated internal network, we need the Vapp created, then use
// vapp.CreateVasppNetwork(&internalSettings, nil) and the nil indicates its isolated.
// So, we'll need to do that later, and in fact, for non-lbs, the network is added subsequent to the compose
// so what will happen if we compose with nil networks?
// Other wise, just add the external, and then remove all nets and add the internal. <sigh>
// in any case, this can't do much
func (v *VcdPlatform) SetNetworksForNewVApp(ctx context.Context, vapp *govcd.VApp, vmgp *vmlayer.VMGroupOrchestrationParams) error {
	var iprange []*types.IPRange
	vmparams := vmgp.VMs[0]
	numVms := len(vmgp.VMs)

	vmOneRole := vmparams.Role
	fmt.Printf("\nselectNetworksForNetVApp vapp %s has  %d vms and the role of  1 = %s\n", vapp.VApp.Name, numVms, vmOneRole)
	IPScope := v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0]
	// All VApps here already have the external network from the composition
	// It may need to be removed for internal VMs that only need isolated network. vapp.VApp.Name, vmOneRole)
	if numVms == 1 && (vmOneRole == vmlayer.RoleVMPlatform || vmOneRole == vmlayer.RoleAgent) {
		fmt.Printf("setNetworkForNewVApp-I-adding internal network to vapp %s Role: %s\n", vapp.VApp.Name, vmOneRole)
		// this guy (LB)  needs both external and internal or GetIPFromServerDetail will fail as it
		// demands two networks

		interIPRange := types.IPRange{
			StartAddress: "10.101.1.2",
			EndAddress:   "10.101.1.22",
		}
		iprange = append(iprange, &interIPRange)
		//		iprange[0] = &interIPRange
		internalSettings := govcd.VappNetworkSettings{
			Name:        vapp.VApp.Name + "-internal-1",
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

			fmt.Printf("SetNetworksForNewVApp-E-error obtaining status of vapp: %s\n", err.Error())
			return err
		}
		if status == "UNRESOLVED" {
			fmt.Printf("SetNetworkForNewVApp-I-wait 10  while  unresolved \n")
			err = vapp.BlockWhileStatus("UNRESOLVED", 10)
			if err != nil {
				fmt.Printf("BlockWhile return err: %s\n", err.Error())
			}
			status, _ = vapp.GetStatus()
			fmt.Printf("Continue from blockwhile status now %s\n", status)
		}
		fmt.Printf("SetNEtorksForNewVApp-I-create internal network, vapp status: %s \n", status)
		InternalNetConfigSec, err := vapp.CreateVappNetwork(&internalSettings, nil)
		if err != nil {
			fmt.Printf("setNetworksForNewVApp-E-create internal net: %s\n", err.Error())
			return fmt.Errorf("CreateVappNetwork Internal error %s", err.Error())
		}
		fmt.Printf("\n\nInternalNetConfigSection: %+v\n\n", InternalNetConfigSec)
	} else {
		//
		//vapp.RemoveAllNetworks()
		// add new internal, need to know what cloudlet (LB) we need to attach to
		// (ie what subnet to create)
		// Master or Workers?
		fmt.Printf("setNetworkForNetVApp-W-TBI for vapp %s with vmOneRole: %s\n",
			vapp.VApp.Name, vmOneRole)

	}
	vapp.Refresh()
	return nil
}

func (v *VcdPlatform) AddVirtHWSecForVMs(ctx context.Context, vapp *govcd.VApp, vmgp *vmlayer.VMGroupOrchestrationParams) error {
	internalNetName := ""
	vm := vapp.VApp.Children.VM[0]
	vmName := vm.Name
	// our internal network name was predicated on the name of the vm[0]

	// Get the internal network name from vapp
	networkConfigSec, err := vapp.GetNetworkConfig()
	if err != nil {
		fmt.Printf("Unable to get network config for vapp  err: %s\n", err.Error())
		return err
	}

	for _, netConfig := range networkConfigSec.NetworkConfig {
		fmt.Printf("\ncheck if %s is a substring of %s\n", vmName, netConfig.NetworkName)
		if strings.Contains(netConfig.NetworkName, vmName) {
			internalNetName = netConfig.NetworkName
			fmt.Printf("AddVirtHWSecForVMs-I-found internal network of Vapp as %s\n", internalNetName)
		}
		fmt.Printf("AddVirtHWSecForVMs-I-found %s skipping\n", netConfig.NetworkName)
	}
	if internalNetName == "" {
		fmt.Printf("AddVirtHwSecForVm-E-no internal network found in vapp %s\n", vapp.VApp.Name)
		return fmt.Errorf("No internal network found for vapp")
	}
	// get the vm(s)
	// cheat and get the first one, and come back for the rest

	fmt.Printf("AddVirt-I-addressing vm: %s with vapp network %s\n", vapp.VApp.Name, internalNetName)
	virtHwSec := vm.VirtualHardwareSection
	vItems := virtHwSec.Item
	// run the items we'll be adding one to this Item []*VirtualHardwareItem
	for _, item := range vItems {
		vu.DumpVirtualHardwareItem(item, 1)
	}
	connection := types.VirtualHardwareConnection{

		IPAddress:         "10.101.10.1", // XXX derive this according to some grand scheme...
		PrimaryConnection: false,
		IpAddressingMode:  "MANUAL",
		NetworkName:       internalNetName,
	}

	inetItem := &types.VirtualHardwareItem{
		ResourceType:        10, // network
		ResourceSubType:     "VMXNET3",
		ElementName:         "Network Adaptor 1", // fix me N when adding to LB for new clusterInsts
		AutomaticAllocation: true,
	}
	inetItem.Connection = append(inetItem.Connection, &connection)
	virtHwSec.Item = append(virtHwSec.Item, inetItem) // VirtualHardwareSection{
	vapp.Refresh()
	return nil

}

// Our external network orgvcdnetwork uses Pool Allocation currently. DHCP considered going forward.
//
// For routed and directly connected networks, the ParentNetwork element contains a ref to the OrgVDCNetwork
// that the VappNetwork connects to. For direct FenceMode bridged or natRouted to specify a routed connection
// controlled by  network features such as NatService or FirewallService
//
func (v *VcdPlatform) SetVappExternalNetwork(ctx context.Context, vapp govcd.VApp) (string, error) {
	//fmt.Printf("setVappExternalNetwork\n")
	vdcnet := v.Objs.PrimaryNet.OrgVDCNetwork
	IPScope := vdcnet.Configuration.IPScopes.IPScope[0] // xxx
	/*
		// AddOrgNetwork
		// Create DhcpSettings
		staticIPStart := IPScope.IPRanges.IPRange[0].StartAddress
		//fmt.Printf("\nSetVappExternalNet: dhcp range used: start %s to end  %s\n", v.IncrIP(IPScope.Gateway), v.DecrIP(staticIPStart))
		// start with bridged, and then proceed to add isFenced and checkout Nat/Firewall rules

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
		//StaticIPRanges:   iprange,
		//		DhcpSettings:     &dhcpsettings,
		//	VappFenceEnabled: takeBoolPointer(false),
	}

	// Add our external network as a vapp network, bridged or Nat'ed to our PrimaryNet
	// bridged, false turns fenceMode from bridged to Nat (True here wins only direct and isolated allowed for
	// our org... hmm...
	//
	_ /*netConfigSec, */, err := vapp.AddOrgNetwork(vappNetSettings, vdcnet, false)

	if err != nil {
		fmt.Printf("\tError UpdateNetwork to vapp: %s\n", err.Error())
		return "", err
	}

	_ /*netConfigSec,*/, err = vapp.GetNetworkConfig()
	if err != nil {
		fmt.Printf("\tError GetNetworkConfig: %s\n", err.Error())
		return "", err
	}

	return externalNetName, nil

}

// All clusters live in one cloudlet/vapp. That cloudlet has our single external networks
// For each new ClusterInst we place on this cloudlet, we need to add a new 10 dot subnet
// using the 3rd octent and host part to the cloudlet as .1 the lb as .10 and the rest as
// .101, 102, ...
//
func (v VcdPlatform) CreateVappInternalNetwork(ctx context.Context, vapp govcd.VApp) (string, error) {

	var iprange []*types.IPRange

	IPScope := v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0]
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
		return "", err
	}
	if status == "UNRESOLVED" {
		err = vapp.BlockWhileStatus("UNRESOLVED", 30) // Raw 10 is enough but Compose takes longer
		if err != nil {
			return "", err
		}
		status, _ = vapp.GetStatus()
	}
	_ /*InternalNetConfigSec,*/, err = vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		return "", fmt.Errorf("CreateVappNetwork Internal error %s", err.Error())
	}
	// Do we need to update?
	//	fmt.Printf("\n\nInternalNetConfigSection: %+v\n\n", InternalNetConfigSec)
	vapp.Refresh() // needed?
	return internalNetName, nil

}

func (v *VcdPlatform) AddVappNetwork(ctx context.Context, vapp *govcd.VApp) (*types.NetworkConfigSection, error) {

	orgNet := v.Objs.PrimaryNet.OrgVDCNetwork
	// procedure
	// name the new network
	// Entry gateway CIDR
	// NetMask
	// the rest are optional
	// DNS + vmwmex.net
	// guest vlan,
	// Statu IP pool settings like IP ranages
	// IsConnected = true

	IPScope := orgNet.Configuration.IPScopes.IPScope[0] // xxx

	// AddOrgNetwork
	// Create IP Pool

	var iprange []*types.IPRange
	iprange = append(iprange, IPScope.IPRanges.IPRange[0])
	fmt.Printf("CreateRoutedVappNetwork poolIPStart: %s poolIPEnd %s \n",
		IPScope.IPRanges.IPRange[0].StartAddress,
		IPScope.IPRanges.IPRange[0].EndAddress)

	VappNetworkSettings := govcd.VappNetworkSettings{
		// now poke our changes into the new vapp
		Name:           "vapp-external-2",
		Gateway:        IPScope.Gateway,
		NetMask:        IPScope.Netmask,
		DNS1:           IPScope.DNS1,
		DNS2:           IPScope.DNS2,
		DNSSuffix:      IPScope.DNSSuffix,
		StaticIPRanges: iprange,
	}

	netConfigSec, err := vapp.AddOrgNetwork(&VappNetworkSettings, orgNet, false)
	if err != nil {
		fmt.Printf("Error Creating vapp network %s\n", err.Error())
		return nil, err
	}
	return netConfigSec, nil

}

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
	for _, orgvdcnet := range nets {
		config := orgvdcnet.OrgVDCNetwork.Configuration
		scopes := config.IPScopes.IPScope
		for _, scope := range scopes {
			a := net.ParseIP(scope.Netmask)
			if a == nil {
				fmt.Printf("GetOrgNetworks-E- %s fail ParseIP\n", scope.Netmask)
				continue
			}
			a4 := a.To4()
			if a4 == nil {
				continue
			}
			netmask := net.IPMask(a4)
			sz, _ := netmask.Size()
			address := fmt.Sprintf("%s/%d", scope.Gateway, sz)
			networks = append(networks, address)
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
		log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "error", err)
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "extNetName", extNetName, "IP", gateway)
	return gateway, err
}

// Ports
func (v *VcdPlatform) GetInternalPortPolicy() vmlayer.InternalPortAttachPolicy {
	return vmlayer.AttachPortDuringCreate
}

// AttachPortToServer
func (v *VcdPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {

	// shared
	log.SpanLog(ctx, log.DebugLevelInfra, "attatch port", "ServerName", serverName, "subnet", subnetName, "ip", ipaddr, "action", action)

	if action == vmlayer.ActionCreate {

	} else if action == vmlayer.ActionDelete {

	} else if action == vmlayer.ActionUpdate {

	}

	return nil

}

func (v *VcdPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName, portName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "detach  port", "ServerName", serverName, "subnet", subnetName, "port", portName)
	return nil
}

func (v *VcdPlatform) IncrIP(ctx context.Context, a string, delta int) string {
	ip := net.ParseIP(a)
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	ip[3] += byte(delta)

	// we know a is a good IP
	ao, _ := v.Octet(ctx, a, 3)
	ipo, err := v.Octet(ctx, ip.String(), 3)
	if err != nil {
		fmt.Printf("IncrIP-E-Octet reports %s is invalid\n", ip)
	}
	if ipo != ao+delta {
		fmt.Printf("IncrIP-E- a %s delta %d ao: %d ipo %d (ao+ipo)%d\n",
			a, delta, ao, ipo, (ao - ipo))
		return ""
		// fmt.Errorf("range wrap err")
	}
	return ip.String()
}

func (v *VcdPlatform) DecrIP(ctx context.Context, a string, delta int) string {
	ip := net.ParseIP(a)
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	ip[3] -= byte(delta)

	ao, _ := v.Octet(ctx, a, 3)
	ipo, _ := v.Octet(ctx, ip.String(), 3)
	if ao-delta != ipo {
		fmt.Printf("DecrIP-E- a %s delta %d ao: %d ipo %d (ao-ipo)%d\n",
			a, delta, ao, ipo, (ao - ipo))
		return ""
		//		return fmt.Errorf("range wrap error")
	}

	return ip.String()
}

func (v *VcdPlatform) Octet(ctx context.Context, a string, n int) (int, error) {
	addr := a
	// strip any cidr mask if present
	parts := strings.Split(a, "/")
	if len(parts) > 1 {
		addr = string(parts[0])
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return 0, fmt.Errorf("Invalid IP")
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("Ip Not a v4 address")
	}
	return int(ip[n]), nil
}

func (v *VcdPlatform) IncrCidr(a string, delta int) string {

	addr := a
	parts := strings.Split(a, "/")
	if len(parts) > 1 {
		addr = string(parts[0])
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	ip[2] += byte(delta)
	return ip.String()
}

// VappNetworks

// Add a OrgVdcNetwork to a VApp
// Return *types.NetowrkConfigSection

// When a second clusterInst/VM is added to a cloudlet, the clouddlet's VApp needs to grow a new internal
// network for it
func (v *VcdPlatform) getInternalNetworkNameForCluster(ctx context.Context, serverName string) string {
	// clust1.cld1.tdg.mobiledgex.net vm name, and vmlayer expects an internal net named:
	//  mex-k8s-subnet-cld1-clust1-mobiledgex
	// port name is just this + "-port"
	parts := strings.Split(serverName, ".")
	if len(parts) == 1 {
		// something with just - delimiters doesn't want an internal addr anyway
		return ""
	}
	netname := "mex-k8s-subnet-" + parts[1] + "-" + parts[0] + "-" + parts[3]
	return netname
}

// Based on the number of vms and the roles, return the set of networks required.
// Currently, to add an isolated internal network, we need the Vapp created, then use
// vapp.CreateVasppNetwork(&internalSettings, nil) and the nil indicates its isolated.
// So, we'll need to do that later, and in fact, for non-lbs, the network is added subsequent to the compose
// so what will happen if we compose with nil networks?
// Other wise, just add the external, and then remove all nets and add the internal. <sigh>
// in any case, this can't do much
const InternalNetMax = 100

func (v *VcdPlatform) CreateInternalNetworkForNewVm(ctx context.Context, vapp *govcd.VApp, vmgp *vmlayer.VMGroupOrchestrationParams, cidr string) (string, error) {
	var iprange []*types.IPRange
	vmparams := vmgp.VMs[0]
	//numVms := len(vmgp.VMs)
	netname := v.getInternalNetworkNameForCluster(ctx, vmparams.Name)

	log.SpanLog(ctx, log.DebugLevelInfra, "internal net", "name", vmparams.Name)
	description := fmt.Sprintf("internal-%s", cidr)
	a := strings.Split(cidr, "/")
	addr := string(a[0])
	gateway := addr //
	// DNE: dnsservers := vmparams.DNSServers
	dns2 := ""

	startAddr := v.IncrIP(ctx, gateway, 1)
	endAddr := v.IncrIP(ctx, gateway, InternalNetMax)

	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "vappNetSetting", "netname", netname, "host", vmparams.HostName, "role", vmparams.Role, "gateway", gateway, "StartIP", startAddr, "EndIP", endAddr)

	}

	addrRange := types.IPRange{
		StartAddress: startAddr,
		EndAddress:   endAddr,
	}
	iprange = append(iprange, &addrRange)
	internalSettings := govcd.VappNetworkSettings{
		Name:        netname,
		Description: description, //internal 10.101.1.0/24 static",
		Gateway:     gateway,
		NetMask:     "255.255.255.0",
		DNS1:        "1.1.1.1", // xxx
		DNS2:        dns2,
		DNSSuffix:   vmparams.DNSDomain,
		//		GuestVLANAllowed: true,    default is?
		StaticIPRanges: iprange,
	}
	_ /*InternalNetConfigSec,*/, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return "", err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork", "network already exists", netname)
	}
	vapp.Refresh()
	return netname, nil
}

// Our external network orgvcdnetwork uses Pool Allocation currently. DHCP considered going forward.
//
// For routed and directly connected networks, the ParentNetwork element contains a ref to the OrgVDCNetwork
// that the VappNetwork connects to. For direct FenceMode bridged or natRouted to specify a routed connection
// controlled by  network features such as NatService or FirewallService
//
func (v *VcdPlatform) SetVappExternalNetwork(ctx context.Context, vapp govcd.VApp) (string, error) {

	vdcnet := v.Objs.PrimaryNet.OrgVDCNetwork
	IPScope := vdcnet.Configuration.IPScopes.IPScope[0] // xxx
	externalNetName := vdcnet.Name                      // vapp.VApp.Name + "-external"
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
		return "", err
	}
	return externalNetName, nil

}

func (v *VcdPlatform) AddVappNetwork(ctx context.Context, vapp *govcd.VApp) (*types.NetworkConfigSection, error) {

	orgNet := v.Objs.PrimaryNet.OrgVDCNetwork
	IPScope := orgNet.Configuration.IPScopes.IPScope[0] // xxx

	var iprange []*types.IPRange
	iprange = append(iprange, IPScope.IPRanges.IPRange[0])

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
		return nil, err
	}
	return netConfigSec, nil

}

func (v *VcdPlatform) GetExtAddrOfVM(ctx context.Context, vm *govcd.VM, netName string) (string, error) {

	if vm == nil {
		return "", fmt.Errorf("Invalid Arg")
	}
	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return "", err
	}
	nc := ncs.NetworkConnection
	fmt.Printf("GetExtAddrOfVm-I-vm %s has %d connections\n", vm.VM.Name, len(nc))
	for _, n := range nc {
		if n.Network == netName {
			return n.IPAddress, nil
		} else {
			fmt.Printf("GetExtAddrOfVM-w-netname %s skipped\n", n.Network)
		}
	}
	return "", fmt.Errorf("Not Found")
}

func (v *VcdPlatform) GetExtAddrOfVapp(ctx context.Context, vapp *govcd.VApp, netName string) (string, error) {

	if vapp == nil {
		return "", fmt.Errorf("Invalid Arg")
	}
	// if this vapp has no vms, this call panics govcd.vapp.go xxx
	ncs, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		return "", err
	}
	nc := ncs.NetworkConnection
	return nc[0].IPAddress, nil
}

// Prmary OrgVDCNetwork
func (v *VcdPlatform) GetNextExtAddrForVdcNet(ctx context.Context, vdc *govcd.Vdc) (string, error) {

	vdcnet := v.Objs.PrimaryNet
	iprange := vdcnet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0]
	e, _ := v.Octet(ctx, iprange.EndAddress, 3)
	curAddr := iprange.StartAddress
	// No vapps yet? use first in pool
	if v.Objs.Cloudlet == nil {
		return curAddr, nil
	}
	// We have a cloudlet at least, looking for a cluster external IP
	cloudMap := v.Objs.Cloudlet.ExtVMMap
	//vmmapLen := len(cloudMap)
	// replace with _,  ok := cloudMap[curAddr]; !ok XXX
	for _, _ = range cloudMap {
		if cloudMap[curAddr] == nil {
			fmt.Printf("\n\nGetNextExtAddrForVdcNet-I-unused addr %s returned\n\n", curAddr)
			return curAddr, nil
		}
		curAddr = v.IncrIP(ctx, curAddr, 1)
		n, err := v.Octet(ctx, curAddr, 3)
		if err != nil {
			return "", err
		}
		if n > e {
			// iprange exhaused
			fmt.Printf("\n\nGetNextExtAddrForVdcNet-E-range Exahusted network %s\n", vdcnet.OrgVDCNetwork.Name)
			return "", fmt.Errorf("available external IP range exhausted")
		}
	}
	fmt.Printf("\n\nGetNextExtAddrForVdcNet-I-nominal return %s\n\n", curAddr)
	return curAddr, nil
}

// Given our scheme for networks 10.101.X.0/24 return the next available Isolated network CIDR
func (v *VcdPlatform) GetNextInternalNet(ctx context.Context) (string, error) {
	var MAX_CIDRS = 20 // implies a limit MAX_CIDRS  clusters per Cloudlet. XXX
	startAddr := "10.101.1.1"

	if v.Objs.Cloudlet == nil {
		return startAddr, nil
	}
	cloudlet := v.Objs.Cloudlet
	curAddr := startAddr
	for n := 1; n < MAX_CIDRS; n++ {
		if _, ok := cloudlet.Clusters[curAddr]; ok {
			curAddr = v.IncrCidr(curAddr, 1)
			continue
		} else {
			if cloudlet.Clusters == nil {
				cloudlet.Clusters = make(CidrMap)
			}
			cloudlet.Clusters[curAddr] = &Cluster{}
			return curAddr, nil
		}
	}
	return "", fmt.Errorf("Cidr range exhaused")
}

// vm networks

func (v *VcdPlatform) AddExtNetToVm(ctx context.Context, vm *govcd.VM, netName string, ip string) error {

	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return err
	}

	// add a new connection section
	// Revisit ModePool and 12/13/20
	ncs.NetworkConnection = append(ncs.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeManual,
			Network:                 netName,
			NetworkConnectionIndex:  0, // 1
			IPAddress:               ip,
		})

	err = vm.UpdateNetworkConnectionSection(ncs)
	if err != nil {
		return err
	}

	return nil
}

func (v *VcdPlatform) GetFirstOrgNetworkOfVdc(ctx context.Context, vdc *govcd.Vdc) (*govcd.OrgVDCNetwork, error) {
	nets := vdc.Vdc.AvailableNetworks
	for _, net := range nets {
		for _, ref := range net.Network {
			vdcnet, err := vdc.GetOrgVdcNetworkByHref(ref.HREF)
			if err != nil {
				continue
			}
			return vdcnet, nil
		}
	}
	return nil, fmt.Errorf("Not Found")
}

func (o *VcdPlatform) ValidateAdditionalNetworks(ctx context.Context, additionalNets []string) error {
	return fmt.Errorf("Additional networks not supported in vCD  cloudlets")
}

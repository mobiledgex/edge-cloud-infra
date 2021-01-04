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

func (v *VcdPlatform) AddPortsToVapp(ctx context.Context, vapp *govcd.VApp, vmgp vmlayer.VMGroupOrchestrationParams) (string, error) {
	ports := vmgp.Ports
	netName := v.Objs.PrimaryNet.OrgVDCNetwork.Name
	nextCidr := ""
	numPorts := len(ports)
	vmparams := vmgp.VMs[0]
	vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))
	log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp", "VAppName", vmgp.GroupName, "role", vmparams.Role, "type", vmType, "numports", numPorts)

	for _, port := range ports {
		// External = 1
		if port.NetworkType == vmlayer.NetTypeExternal {

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 1
			extAddr, err := v.GetNextExtAddrForVdcNet(ctx)

			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding", "network", netName, "IP", extAddr, "port", port.NetworkName)

			_ /* networkConfigSection */, err = v.AddVappNetwork(ctx, vapp)

			desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
				&types.NetworkConnection{
					IsConnected:             true,
					IPAddressAllocationMode: types.IPAllocationModeManual,
					Network:                 v.Objs.PrimaryNet.OrgVDCNetwork.Name,
					NetworkConnectionIndex:  1,
					IPAddress:               extAddr,
				})

			// metadata?
			vmtmplName := vapp.VApp.Children.VM[0].Name
			vm, err := vapp.GetVMByName(vmtmplName, false)
			if err != nil {
				return "", err
			}
			err = vm.UpdateNetworkConnectionSection(desiredNetConfig)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "update external  network failed", "VAppName", vmgp.GroupName, "err", err)
				return "", err
			}
		}
		// Internal (isolatedD)
		// Create an isolated subnet for this vapp/clusterInst
		if port.NetworkType == vmlayer.NetTypeInternal {
			var err error
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp create internal vapp net", "vapp", vapp.VApp.Name, "subnetID", port.SubnetId)
			// Give the port name to Create directly
			// and lose the netname return value
			nextCidrOld, err := v.GetNextInternalNetOld(ctx)
			nextCidrNew, err := v.GetNextInternalNet(ctx)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVAppe next internal net failed: ", "GroupName", vmgp.GroupName, "err", err)
				return "", err
			}

			if nextCidrOld != nextCidr {
				fmt.Printf("\n\n AddPortsToVapp Internal next Cidr old %s new %s using old Fix new routine!\n", nextCidrOld, nextCidrNew)
				nextCidr = nextCidrOld
			}

			_, err = v.CreateInternalNetworkForNewVm(ctx, vapp, &vmgp, port.SubnetId, nextCidr)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "create cluster internal net failed", "err", err)
				return "", err
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp  add to vapp  internal", "network", port.Name)
		}
	}
	return nextCidr, nil
}

// AttachPortToServer
func (v *VcdPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	// shared
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToserver", "ServerName", serverName, "subnet", subnetName, "ip", ipaddr, "portName", portName, "action", action)

	// We could add this portName to our serverVM as metadata rather and derive it each time xxx
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
		return ""
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
		return ""
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
// A subnet name format is the base bit "mex-k8s-subnet-" + cldName + clsterName + orgName
//

func (v *VcdPlatform) GetInternalNetworkNameForCluster(ctx context.Context, serverName string) string {
	// clust1.cld1.tdg.mobiledgex.net vm name, and vmlayer expects an internal net named:
	//  mex-k8s-subnet-cld1-clust1-mobiledgex
	parts := strings.Split(serverName, ".")
	if len(parts) == 1 {
		return ""
	}
	// subnetname convention? base + cloudletName + clusterName + cloudletOrg
	//             base            cloudlet         cluster            cloudorg
	netname := "mex-k8s-subnet-" + parts[1] + "-" + parts[0] + "-" + parts[2]
	log.SpanLog(ctx, log.DebugLevelInfra, "GetInternalSubnetname", "server", serverName, "subnetName", netname)
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

func (v *VcdPlatform) CreateInternalNetworkForNewVm(ctx context.Context, vapp *govcd.VApp, vmgp *vmlayer.VMGroupOrchestrationParams, netName string, cidr string) (string, error) {
	var iprange []*types.IPRange
	vmparams := vmgp.VMs[0]
	//numVms := len(vmgp.VMs)

	netname := v.GetInternalNetworkNameForCluster(ctx, vmparams.Name)
	if netname != netName {
		fmt.Printf("\n\nCreateInternalNetworkForNewVm-I-should be equal! portName %s vs handmade netname: %s\n\n", netName, netname)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "create internal vm net", "vm name", vmparams.Name, "becomes netname", netname)

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
		Name:           netname,
		Description:    description, //internal 10.101.1.0/24 static",
		Gateway:        gateway,
		NetMask:        "255.255.255.0",
		DNS1:           "1.1.1.1",
		DNS2:           dns2,
		DNSSuffix:      vmparams.DNSDomain,
		StaticIPRanges: iprange,
	}
	_ /*InternalNetConfigSec,*/, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork create", "error", err)
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
	for _, n := range nc {
		if n.Network == netName {
			return n.IPAddress, nil
		}
	}
	return "", fmt.Errorf("Not Found")
}

// Deprecated in favor of GetAddrOfVapp
func (v *VcdPlatform) GetExtAddrOfVapp(ctx context.Context, vapp *govcd.VApp, netName string) (string, error) {

	if vapp == nil {
		return "", fmt.Errorf("Invalid Arg")
	}
	// if this vapp has no vms, this call panics govcd.vapp.go xxx
	// Which is really really a bad thing. How should these vapps be validated?
	// XXX
	ncs, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		return "", err
	}
	nc := ncs.NetworkConnection
	return nc[0].IPAddress, nil
}

// If we just have a GetAddrOfVapp(ctx, vapp, netName)
// it should suffice yes?
func (v *VcdPlatform) GetAddrOfVapp(ctx context.Context, vapp *govcd.VApp, netName string) (string, error) {

	if vapp == nil {
		return "", fmt.Errorf("Invalid Arg")
	}
	// if this vapp has no vms, this call panics govcd.vapp.go xxx
	// Which is really really a bad thing. How should these vapps be validated?
	// How about getting its status, and if not at least Resolved, consider it a bad egg.
	ncs, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		return "", err
	}
	for _, nc := range ncs.NetworkConnection {
		if nc.Network == netName {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp", "Vapp", vapp.VApp.Name, "on net", netName, "ip", nc.IPAddress)
			return nc.IPAddress, nil
		}
	}
	return "", fmt.Errorf("Not Found")
}

// Next up refactor: Need
// to run all VApps external addresses
// We still need some sort of map, we can build on the fly
// or try and cache/rebuild
//

func (v *VcdPlatform) GetNextExtAddrForVdcNet(ctx context.Context) (string, error) {
	// for all Vapps
	vdcnet := v.Objs.PrimaryNet
	netName := vdcnet.OrgVDCNetwork.Name
	iprange := vdcnet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0]
	s, _ := v.Octet(ctx, iprange.StartAddress, 3)
	e, _ := v.Octet(ctx, iprange.EndAddress, 3)
	curAddr := iprange.StartAddress

	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr", "start", s, "end", e, "curAddr", curAddr, "network", netName)
	vappMap, err := v.GetAllVAppsForVdc(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return "", err
	}

	if len(vappMap) == 0 {
		// No vapps yet in this vdc
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return curAddr, nil
	}

	for i := s; i < e; i++ {

		fmt.Printf("\n\nGetNextAddr consider %s\n", curAddr)

		if _, found := vappMap[curAddr]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
			return curAddr, nil
		}
		curAddr = v.IncrIP(ctx, curAddr, 1)
		_, err := v.Octet(ctx, curAddr, 3)
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("net %s ip pool full", netName)
}

// Given our scheme for networks 10.101.X.0/24 return the next available Isolated network CIDR
//
// Refactor, we like external, we can GetAllVappsForVdc, and
// run that map getting internal addr of the vapp, if it has one.

func (v *VcdPlatform) GetNextInternalNet(ctx context.Context) (string, error) {
	var MAX_CIDRS = 255 // These are internal /24 subnets so 255, not that we'll have that many clusters / cloudlet
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalNet")
	vappMap, err := v.GetAllVAppsForVdcByIntAddr(ctx) // refactor to allow the same map key'ed by ext or int addr
	startAddr := "10.101.1.1"                         // use  props subnet XXX Could be like 10.101.2.1
	curAddr := startAddr
	if len(vappMap) == 0 {
		return curAddr, nil
	}
	for i := 1; i < MAX_CIDRS; i++ {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalNet consider", "cidr", curAddr)
		if _, found := vappMap[curAddr]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalNet return", "cidr", curAddr)
			return curAddr, nil
		}
		// don't we want to incr the target octet?
		curAddr = v.IncrCidr(curAddr, 1)
	}

	return "", err
}
func (v *VcdPlatform) GetNextInternalNetOld(ctx context.Context) (string, error) {

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

	fmt.Printf("\n\nAddExtNetToVm-I-vm: %s netName: %s ip: %s\n\n", vm.VM.Name, netName, ip)

	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return err
	}

	ncs.NetworkConnection = append(ncs.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeManual,
			Network:                 netName,
			NetworkConnectionIndex:  1,
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

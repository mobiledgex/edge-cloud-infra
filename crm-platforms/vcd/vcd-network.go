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

func (v *VcdPlatform) GetNetworkList(ctx context.Context) ([]string, error) {
	return []string{
		v.vmProperties.GetCloudletExternalNetwork(),
	}, nil
}

/* Rework TBI
func (v *VcdPlatform) GetExternalIpCounts(ctx context.Context) ([]string, error) {
	var networks []string

	vdc, err := v.GetVdc(ctx)
	if err != nil {
		return err
	}

	nets := vdc.Vdc.AvailableNetworks
	for _, net := range nets {

		orgvdcnet, err := vdc.GetOrgVdcNetworkByName(net.OrgVDCNetwork.Name, false)
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
*/

// fetch the OrgVDCNetwork referenced by MEX_EXT_NET our Primary external network
func (v *VcdPlatform) GetExtNetwork(ctx context.Context) (*govcd.OrgVDCNetwork, error) {

	// infra propert from env
	netName := v.GetExtNetworkName()

	vdc, err := v.GetVdc(ctx)
	if err != nil {
		return nil, err
	}
	orgvdcnet, err := vdc.GetOrgVdcNetworkByName(netName, true)
	if err != nil {
		return nil, err
	}
	return orgvdcnet, nil

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

	vdcNet, err := v.GetExtNetwork(ctx)
	if err != nil {
		return "", err
	}
	gateway, err := v.GetGatewayForOrgVDCNetwork(ctx, vdcNet.OrgVDCNetwork)
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
	nextCidr := ""
	numPorts := len(ports)
	vmparams := vmgp.VMs[0]
	serverName := vmparams.Name
	intAdded := false
	vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))

	log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp", "VAppName", vmgp.GroupName, "role", vmparams.Role, "type", vmType, "numports", numPorts)

	for n, port := range ports {
		// External = 1
		if port.NetworkType == vmlayer.NetTypeExternal {

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 1
			extAddr, err := v.GetNextExtAddrForVdcNet(ctx)
			if extAddr == "" {
				panic("AddPortsToVapp GetNExtExtAddrForVdcNet nil!")
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding external vapp net", "PortNum", n, "vapp", vapp.VApp.Name, "port.NetworkName", port.NetworkName, "IP", extAddr)

			_ /* networkConfigSection */, err = v.AddVappNetwork(ctx, vapp)

			desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
				&types.NetworkConnection{
					IsConnected:             true,
					IPAddressAllocationMode: types.IPAllocationModeManual,
					Network:                 v.GetExtNetworkName(),
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

		if port.NetworkType == vmlayer.NetTypeInternal && !intAdded {
			var err error

			// Give the port name to Create directly
			// and lose the netname return value

			nextCidr, err = v.GetNextInternalSubnet(ctx, vapp.VApp.Name)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp next internal net failed: ", "GroupName", vmgp.GroupName, "err", err)
				return "", err
			}
			if nextCidr == "" {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp next internal net cid == ", "Vapp", vapp.VApp.Name, "Port.Network", port.NetworkName, "PortNum", n)
				panic("AddVMsToVApp Nil Cidr returned game over")
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding internal vapp net", "PortNum", n, "vapp", vapp.VApp.Name, "port.NetowkrNamek", port.NetworkName, "IP subnet", nextCidr)
			// Subnet or portName? subnet failed,

			log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVapp cidr for internal net", "cidr", nextCidr)
			_, err = v.CreateInternalNetworkForNewVm(ctx, vapp /*&vmgp,*/ /*port.SubnetId,*/, serverName, port.SubnetId, nextCidr)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "create internal net failed", "err", err)
				return "", err
			}
			if v.Verbose {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp  add to vapp  internal", "network", port.Name)
			}
			intAdded = true
		}
	}
	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp return", "NextCidr", nextCidr, "NumPorts", numPorts)
	}
	return nextCidr, nil
}

// AttachPortToServer
func (v *VcdPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	// shared LBs are asked to grow a new internal network
	vappName := serverName + v.GetVappServerSuffix()
	vapp, err := v.FindVApp(ctx, vappName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer server not found", "vapp", vappName, "for server", serverName)
		return err
	}
	vmName := vapp.VApp.Children.VM[0].Name
	vm, err := vapp.GetVMByName(vmName, true)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer server not found", "vm", vmName, "for server", serverName)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToserver", "ServerName", serverName, "subnet", subnetName, "ip", ipaddr, "portName", portName, "action", action)

	// We could add this portName to our serverVM as metadata rather and derive it each time xxx
	if action == vmlayer.ActionCreate {
		// Add the new network to this vapp
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer adding", "subnetName", subnetName, "portName", portName, "server", serverName, "vapp", vappName)

		// Get the next available internal subnet
		nextCidr, err := v.GetNextInternalSubnet(ctx, vappName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer  Get next internal net failed: ", "err", err)
			return err
		}
		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer CreateInternalNetwork for", "vapp", vappName, "portName", portName, "cidr", nextCidr)
		}

		// Still think we should pass in the name for the internal network, is that PortName or subnetname?
		internalNetName, err := v.CreateInternalNetworkForNewVm(ctx, vapp, serverName, subnetName, nextCidr)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer create internal net failed", "err", err)
			return err
		}
		var a []string
		vmIp := ""
		if nextCidr != "" {
			a = strings.Split(nextCidr, "/")
			vmIp = string(a[0])
		}
		// Add the connection section to the vapp's vm

		log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp add connection", "net", internalNetName, "ip", vmIp, "VM", vmName)
		ncs, err := vm.GetNetworkConnectionSection()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Failed to retrieve networkConnectionSection from", "vm", vmName, "err", err)
			return err
		}
		ncs.NetworkConnection = append(ncs.NetworkConnection,
			&types.NetworkConnection{
				Network:                 subnetName,
				NetworkConnectionIndex:  0,
				IPAddress:               vmIp,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual,
			})

		err = vm.UpdateNetworkConnectionSection(ncs)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp add internal net failed", "VM", vmName, "error", err)
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp added connection", "net", internalNetName, "ip", vmIp, "VM", vmName)

		// update NetworkConnection

	} else if action == vmlayer.ActionDelete {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer deleting", "subnetName", subnetName, "portName", portName, "ip", ipaddr, "server", serverName)

	} else if action == vmlayer.ActionUpdate {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer updating", "subnetName", subnetName, "portName", portName, "ip", ipaddr, "server", serverName)

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

func (v *VcdPlatform) CreateInternalNetworkForNewVm(ctx context.Context, vapp *govcd.VApp, serverName, netName string, cidr string) (string, error) {
	var iprange []*types.IPRange

	log.SpanLog(ctx, log.DebugLevelInfra, "create internal server net", "netname", netName)

	//netname := v.GetInternalNetworkNameForCluster(ctx, serverName)
	//if netname != netName {
	//	fmt.Printf("\n\nCreateInternalNetworkForNewVms-W-netname%s netName: %s\n\n ", netname, netName)
	//} else {
	//		fmt.Printf("\n\nCreateInternalNetworkForNewVms-I-netName is equal to netname\n\n")
	//	}

	description := fmt.Sprintf("internal-%s", cidr)
	a := strings.Split(cidr, "/")
	addr := string(a[0])
	gateway := addr //
	dns2 := ""

	startAddr := v.IncrIP(ctx, gateway, 1)
	endAddr := v.IncrIP(ctx, gateway, InternalNetMax)

	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork ", "serverName", serverName, "netname", netName, "gateway", gateway, "StartIP", startAddr, "EndIP", endAddr)
	}

	addrRange := types.IPRange{
		StartAddress: startAddr,
		EndAddress:   endAddr,
	}
	iprange = append(iprange, &addrRange)

	internalSettings := govcd.VappNetworkSettings{
		Name:           netName,
		Description:    description, //internal 10.101.1.0/24 static",
		Gateway:        gateway,
		NetMask:        "255.255.255.0",
		DNS1:           "1.1.1.1",
		DNS2:           dns2,
		DNSSuffix:      "mobiledgex.net",
		StaticIPRanges: iprange,
	}
	_ /*InternalNetConfigSec,*/, err := vapp.CreateVappNetwork(&internalSettings, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork create", "serverName", serverName, "error", err)
			return "", err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork", "network already exists", netName)
	}

	vapp.Refresh()
	return netName, nil
}

// Our external network orgvcdnetwork uses Pool Allocation currently. DHCP considered going forward.
//
// For routed and directly connected networks, the ParentNetwork element contains a ref to the OrgVDCNetwork
// that the VappNetwork connects to. For direct FenceMode bridged or natRouted to specify a routed connection
// controlled by  network features such as NatService or FirewallService
//
func (v *VcdPlatform) SetVappExternalNetwork(ctx context.Context, vapp govcd.VApp) (string, error) {

	vdcnet, err := v.GetExtNetwork(ctx)
	if err != nil {
		return "", err
	}
	IPScope := vdcnet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0] // xxx
	externalNetName := vdcnet.OrgVDCNetwork.Name                      // vapp.VApp.Name + "-external"
	vappNetSettings := &govcd.VappNetworkSettings{
		Name:             externalNetName,
		ID:               vdcnet.OrgVDCNetwork.ID,
		Description:      "external nat/dhcp",
		Gateway:          IPScope.Gateway,
		NetMask:          IPScope.Netmask,
		DNS1:             IPScope.DNS1,
		DNS2:             IPScope.DNS2,
		DNSSuffix:        IPScope.DNSSuffix,
		GuestVLANAllowed: vu.TakeBoolPointer(false),
	}
	_ /*netConfigSec, */, err = vapp.AddOrgNetwork(vappNetSettings, vdcnet.OrgVDCNetwork, false)
	if err != nil {
		return "", err
	}
	return externalNetName, nil

}

func (v *VcdPlatform) AddVappNetwork(ctx context.Context, vapp *govcd.VApp) (*types.NetworkConfigSection, error) {

	orgNet, err := v.GetExtNetwork(ctx)
	if err != nil {
		return nil, err
	}
	IPScope := orgNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0] // xxx

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

	netConfigSec, err := vapp.AddOrgNetwork(&VappNetworkSettings, orgNet.OrgVDCNetwork, false)
	if err != nil {
		return nil, err
	}
	return netConfigSec, nil

}

// return a list of internal nets, a shared LB may have several
func (v *VcdPlatform) GetIntAddrsOfVM(ctx context.Context, vm *govcd.VM) ([]string, error) {
	addrs := []string{}
	if vm == nil {
		return addrs, fmt.Errorf("Invalid Arg")
	}
	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return addrs, err
	}

	nc := ncs.NetworkConnection
	for _, n := range nc {
		if n.Network != v.GetExtNetworkName() {
			addrs = append(addrs, n.IPAddress)
		}
	}
	return addrs, nil

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

// Get the address VM is using on the given Network
func (v *VcdPlatform) GetAddrOfVM(ctx context.Context, vm *govcd.VM, netName string) (string, error) {

	if vm == nil {
		return "", fmt.Errorf("Invalid Arg")
	}

	// if this vapp has no vms, this call panics govcd.vapp.go xxx
	// Which is really really a bad thing. How should these vapps be validated?
	// How about getting its status, and if not at least Resolved, consider it a bad egg.
	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM err retrieving NetConnection section", "vapp", vm.VM.Name, "network", netName, "err", err)
		return "", err
	}
	numNets := len(ncs.NetworkConnection)
	log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM", "vapp", vm.VM.Name, "network", netName, "numNetworks", numNets)
	for _, nc := range ncs.NetworkConnection {
		if nc.Network == netName {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp", "Vapp", vm.VM.Name, "on net", netName, "ip", nc.IPAddress)
			return nc.IPAddress, nil
		}
	}

	return "", fmt.Errorf("Not Found")
}

// Consider a GetAllSubnetsInVapp() []string err for shared lbs... XXX
// Returns the first addr on the given network.
func (v *VcdPlatform) GetAddrOfVapp(ctx context.Context, vapp *govcd.VApp, netName string) (string, error) {

	if vapp == nil {
		return "", fmt.Errorf("Invalid Arg")
	}

	// if this vapp has no vms, this call panics govcd.vapp.go xxx
	// Which is really really a bad thing. How should these vapps be validated?
	// How about getting its status, and if not at least Resolved, consider it a bad egg.
	ncs, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp err retrieving NetConnection section", "vapp", vapp.VApp.Name, "network", netName, "err", err)
		return "", err
	}
	numNets := len(ncs.NetworkConnection)
	log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp", "vapp", vapp.VApp.Name, "network", netName, "numNetworks", numNets)
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
	vdcnet, err := v.GetExtNetwork(ctx)
	if err != nil {
		return "", err
	}
	netName := vdcnet.OrgVDCNetwork.Name
	iprange := vdcnet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0]
	s, _ := v.Octet(ctx, iprange.StartAddress, 3)
	e, _ := v.Octet(ctx, iprange.EndAddress, 3)
	curAddr := iprange.StartAddress

	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr", "start", s, "end", e, "curAddr", curAddr, "network", netName)
	vappMap, err := v.GetAllVAppsForVdc(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return curAddr, err
	}

	if len(vappMap) == 0 {
		// No vapps yet in this vdc
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return curAddr, nil
	}

	for i := s; i < e; i++ {
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

func (v *VcdPlatform) GetNextInternalSubnet(ctx context.Context, vappName string) (string, error) {

	var MAX_CIDRS = 255 // These are internal /24 subnets so 255, not that we'll have that many clusters / cloudlet

	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet", "vapp", vappName)
	// use schema xxx
	startAddr := "10.101.1.1"
	// We'll incr the netSpec.DelimiterOctet of this start addr, if it's not in our
	// All VApps map, it's available
	curAddr := startAddr

	vappMap, err := v.GetAllVAppsForVdcByIntAddr(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return curAddr, err
	}
	if len(vappMap) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextExtAddr return", "curAddr", curAddr)
		return curAddr, nil
	}
	for i := 1; i < MAX_CIDRS; i++ {
		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalNet consider", "cidr", curAddr)
		}
		if _, found := vappMap[curAddr]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet return available", "cidr", curAddr)
			return curAddr, nil
		}
		// don't we want to incr the target octet?
		curAddr = v.IncrCidr(curAddr, 1)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet  subnet range exhasted")
	return "", err
}

// vm networks

func (v *VcdPlatform) AddExtNetToVm(ctx context.Context, vm *govcd.VM, netName string, ip string) error {

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

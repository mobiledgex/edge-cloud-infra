package vcd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
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

// fetch the OrgVDCNetwork referenced by MEX_EXT_NET our Primary external network
func (v *VcdPlatform) GetExtNetwork(ctx context.Context, vcdClient *govcd.VCDClient) (*govcd.OrgVDCNetwork, error) {

	// infra propert from env
	netName := v.vmProperties.GetCloudletExternalNetwork()

	vdc, err := v.GetVdc(ctx, vcdClient)
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
// Gateway, Always operates on MEX_EXT_NETWORK
// Return the IP address of the external Gateway
func (v *VcdPlatform) GetExternalGateway(ctx context.Context, extNetname string) (string, error) {

	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return "", fmt.Errorf(NoVCDClientInContext)
	}
	vdcNet, err := v.GetExtNetwork(ctx, vcdClient)
	if err != nil {
		return "", err
	}

	// We have only one primary external network. So these values should match
	if vdcNet.OrgVDCNetwork.Name != extNetname {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway warn extNetName", "MEX_EXT_NETWORK", vdcNet.OrgVDCNetwork.Name, "extNetName", extNetname)
	}

	gateway, err := v.GetGatewayForOrgVDCNetwork(ctx, vdcNet.OrgVDCNetwork)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "error", err)
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalGateway", "extNetName", vdcNet.OrgVDCNetwork.Name, "IP", gateway)
	return gateway, err
}

func (v *VcdPlatform) GetInternalPortPolicy() vmlayer.InternalPortAttachPolicy {
	return vmlayer.AttachPortDuringCreate
}

func (v *VcdPlatform) haveSharedRootLB(ctx context.Context, vmgp vmlayer.VMGroupOrchestrationParams) bool {

	log.SpanLog(ctx, log.DebugLevelInfra, "haveSharedRoot", "GroupName", vmgp.GroupName)
	// if no external ports, must be serviced by a shared load balancer.
	// Only role agent needs an external net, so if any vm in this group has such a role, ShareRootLB = false
	for _, vmparams := range vmgp.VMs {
		if vmparams.Role == vmlayer.RoleAgent {
			log.SpanLog(ctx, log.DebugLevelInfra, "haveSharedRoot false", "GroupName", vmgp.GroupName)
			return false
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "haveSharedRoot true", "GroupName", vmgp.GroupName)
	return true

}

var netLock sync.Mutex

func (v *VcdPlatform) createNextSharedLBSubnet(ctx context.Context, vapp *govcd.VApp, port vmlayer.PortOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback, vcdClient *govcd.VCDClient) (string, error) {
	// shared lbs need individual orgvcd isolated networks, must be unique.
	// take the lock that is released after the network has been added to the sharedLB's VApp
	log.SpanLog(ctx, log.DebugLevelInfra, "createNextSharedLBSubnet", "vapp", vapp.VApp.Name)

	netLock.Lock()
	defer netLock.Unlock()

	subnet, reuseExistingNet, err := v.GetNextInternalSubnet(ctx, vapp.VApp.Name, updateCallback, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "createNextSharedLBSubnet  SharedLB GetNextInternalSubnet failed", "vapp", vapp.VApp.Name, "port.NetowkrNamek", port.NetworkName, "error", err)
		return "", err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "createNextSharedLBSubnetSharedLB", "vapp", vapp.VApp.Name, "port.NetowkrName", port.NetworkName, "port.SubnetId", port.SubnetId, "IP subnet", subnet, "reused", reuseExistingNet)
	// OrgVDCNetwork LinkType = 2 (isolated)
	// This seems to be an admin priv operation if using  nsx-t back network pool xxx
	err = v.CreateIsoVdcNetwork(ctx, vapp, port.SubnetId, subnet, vcdClient, reuseExistingNet)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "createNextSharedLBSubnet  create iso orgvdc internal net failed", "err", err)
		return "", err
	}
	return subnet, nil
}

func (v *VcdPlatform) AddPortsToVapp(ctx context.Context, vapp *govcd.VApp, vmgp vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback, vcdClient *govcd.VCDClient) (string, error) {
	ports := vmgp.Ports
	subnet := ""
	numPorts := len(ports)
	vmparams := vmgp.VMs[0]
	serverName := vmparams.Name
	intAdded := false
	vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))

	log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp", "VAppName", vmgp.GroupName, "role", vmparams.Role, "type", vmType, "numports", numPorts)

	for n, port := range ports {
		// External = 0
		if port.NetworkType == vmlayer.NetTypeExternal {

			desiredNetConfig := &types.NetworkConnectionSection{}
			desiredNetConfig.PrimaryNetworkConnectionIndex = 0
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding external vapp net", "PortNum", n, "vapp", vapp.VApp.Name, "port.NetworkName", port.NetworkName, "ConIdx", 0)
			_, err := v.AddVappNetwork(ctx, vapp, vcdClient)

			desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
				&types.NetworkConnection{
					IsConnected:             true,
					IPAddressAllocationMode: types.IPAllocationModePool,
					Network:                 v.vmProperties.GetCloudletExternalNetwork(),
					NetworkConnectionIndex:  desiredNetConfig.PrimaryNetworkConnectionIndex,
				})

			vmtmplName := vapp.VApp.Children.VM[0].Name
			vm, err := vapp.GetVMByName(vmtmplName, false)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp unable to retrieve", "vm", vmtmplName, "vapp", vapp.VApp.Name)
				return "", err
			}
			err = vm.UpdateNetworkConnectionSection(desiredNetConfig)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "update external  network failed", "VAppName", vmgp.GroupName, "err", err)
				return "", err
			}
		}
		// Internal (isolated)
		// Create isolated subnet for this vapp/clusterInst Vapp net for Dedicated, or OrgVDCNetwork for Shared LB
		if port.NetworkType == vmlayer.NetTypeInternal && !intAdded {
			var err error
			log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding internal vapp net", "PortNum", n, "vapp", vapp.VApp.Name, "port.NetowkName", port.SubnetId, "port.SubnetId", port.SubnetId)
			// We've fenced our VApp isolated networks, so they can all use the same subnet
			subnet = "10.101.1.1"
			if v.Verbose {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding internal vapp net", "PortNum", n, "vapp", vapp.VApp.Name, "port.NetowkrNamek", port.NetworkName, "IP subnet", subnet)
			}
			if v.haveSharedRootLB(ctx, vmgp) {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding internal vapp net for SharedLB", "vapp", vapp.VApp.Name)
				// This can return the next available cidr, or an existing cidr from the FreeIsoNets list
				subnet, err = v.createNextSharedLBSubnet(ctx, vapp, port, updateCallback, vcdClient)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp createNextShareRootLBSubnet failed", "vapp", vapp.VApp.Name, "error", err)
					return "", err
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp created iso vdcnet for SharedLB", "network", port.SubnetId, "vapp", vapp.VApp.Name, "IP subnet", subnet)
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp adding internal vapp net non-shared", "vapp", vapp.VApp.Name)
				_, err = v.CreateInternalNetworkForNewVm(ctx, vapp, serverName, port.SubnetId, subnet)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "create internal net failed", "err", err)
					return "", err
				}
			}
			if v.Verbose {
				log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp add to vapp  internal", "network", port.Name)
			}
			intAdded = true
		}
	}
	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "AddPortsToVapp return", "NextCidr", subnet, "NumPorts", numPorts)
	}
	return subnet, nil
}

// AttachPortToServer
func (v *VcdPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {

	// shared LBs are asked to grow a new isolated OrgVDCNetwork
	// The network itself has been created by the client cluster vapp.
	cidrNet := ""
	if len(v.IsoNamesMap) > 0 {
		cidrNet = v.IsoNamesMap[subnetName]
		if cidrNet == "" {
			log.SpanLog(ctx, log.DebugLevelInfra, "No mapping for", "Network", subnetName)
			return fmt.Errorf("No Matching Subnet in IsoNamesMap")
		}
	} else {
		// Since the requested network should have been created already, this map must be non empty
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPort no netname xlation available")
		return fmt.Errorf("Internal Server Error")
	}

	vappName := serverName + v.GetVappServerSuffix()
	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to retrieve current vdc", "err", err)
		return err
	}
	vapp, err := v.FindVApp(ctx, vappName, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer server not found", "vapp", vappName, "for server", serverName)
		return err
	}
	vmName := vapp.VApp.Children.VM[0].Name
	vm, err := vapp.GetVMByName(vmName, true)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer server not found", "vm", vmName, "for server", serverName)
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToserver", "ServerName", serverName, "subnet", subnetName, "cidrNet", cidrNet, "ip", ipaddr, "portName", portName, "action", action)
	if action == vmlayer.ActionCreate {
		// The client VM(s) that wish to be serviced by this sharedLB (serverName) have already created the needed orgvdc iso network
		orgvdcnet, err := vdc.GetOrgVdcNetworkByName(cidrNet /*subnetName,*/, false)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer orgvdc subnet not found", "subnetName", subnetName, "cidrNet", cidrNet)
			return err
		}
		vappNetSettings := &govcd.VappNetworkSettings{
			Name:             cidrNet, /* subnetname */
			VappFenceEnabled: TakeBoolPointer(false),
		}
		// need to Add this orgvdcnet to this Vapp so the vm can find it.

		_, err = vapp.AddOrgNetwork(vappNetSettings, orgvdcnet.OrgVDCNetwork, false)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer AddOrgNetwork failed", "subnetName", subnetName, "err", err)
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer AddOrgNetwork added", "subnetName", subnetName, "cidrNet", cidrNet, "vapp", vapp.VApp.Name)
		scope := orgvdcnet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0]
		gateway := scope.Gateway
		ncs, err := vm.GetNetworkConnectionSection()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Failed to retrieve networkConnectionSection from", "vm", vmName, "err", err)
			return err
		}
		conIdx, err := GetNextAvailConIdx(ctx, ncs)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer  conIdx failed", "subnetName", subnetName, "err", err)
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer ", "subnetName", subnetName, "ip", gateway, "conIdx", conIdx)
		ncs.NetworkConnection = append(ncs.NetworkConnection,
			&types.NetworkConnection{
				Network:                 cidrNet, /* subnetName */
				NetworkConnectionIndex:  conIdx,
				IPAddress:               gateway,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual,
			})
		err = vm.UpdateNetworkConnectionSection(ncs)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp add internal net failed", "VM", vmName, "error", err)
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "AddVMsToVApp added connection", "net", subnetName, "ip", gateway, "VM", vmName, "conIdx", conIdx)

	} else if action == vmlayer.ActionDelete {

	} else if action == vmlayer.ActionUpdate {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer updating", "subnetName", subnetName, "portName", portName, "ip", ipaddr, "server", serverName)
	}
	return nil
}

func (v *VcdPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName, portName string) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer", "ServerName", serverName, "subnet", subnetName, "port", portName)
	cidrNet := ""
	if len(v.IsoNamesMap) > 0 {
		cidrNet = v.IsoNamesMap[subnetName]
		if cidrNet == "" {
			log.SpanLog(ctx, log.DebugLevelInfra, "No mapping for", "Network", subnetName)
			return fmt.Errorf("No Matching Subnet in IsoNamesMap")
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf("Internal Server Error")
	}

	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer unable to retrieve current vdc", "err", err)
		return err
	}
	vappName := serverName + v.GetVappServerSuffix()
	vapp, err := v.FindVApp(ctx, vappName, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer server not found", "vapp", vappName, "for server", serverName)
		return err
	}
	orgvdcnet, err := vdc.GetOrgVdcNetworkByName(cidrNet /*subnetName*/, false)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer orgvdc subnet not found", "subnetName", subnetName)
		return err
	}

	// Operate on all VMs in this vapp
	vms := vapp.VApp.Children.VM
	for _, tvm := range vms {
		vmName := tvm.Name
		vm, err := vapp.GetVMByName(vmName, true)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer server not found", "vm", vmName, "for server", serverName)
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer", "vm", vmName, "for server", serverName)
		ncs, err := vm.GetNetworkConnectionSection()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer Failed to retrieve networkConnectionSection from", "vm", vmName, "err", err)
			return err
		}

		for n, nc := range ncs.NetworkConnection {
			if nc.Network == portName {
				ncs.NetworkConnection[n] = ncs.NetworkConnection[len(ncs.NetworkConnection)-1]
				ncs.NetworkConnection[len(ncs.NetworkConnection)-1] = &types.NetworkConnection{}
				ncs.NetworkConnection = ncs.NetworkConnection[:len(ncs.NetworkConnection)-1]
				err := vm.UpdateNetworkConnectionSection(ncs)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer UpdateNetowrkConnectionSection failed", "serverName", serverName, "port", portName, "subnet", subnetName, "err", err)
					return err
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer success", "serverName", serverName, "port", portName, "subnet", subnetName)
				break
			}
		}
	}
	// Now remove the network from the Vapp/Server
	_, err = vapp.RemoveNetwork(orgvdcnet.OrgVDCNetwork.Name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer RemoveNetwork (byName) failed try RemoveIsolatedNetwork", "serverName", serverName, "port", portName, "subnet", subnetName, "cidrNet", cidrNet, "err", err)

	}
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer net removed from vapp ok")
	return nil
}

func IncrIP(ctx context.Context, a string, delta int) (string, error) {
	ip := net.ParseIP(a)
	if ip == nil {
		return "", fmt.Errorf("%s failed to parse as IP", a)
	}
	ip = ip.To4()
	if ip == nil {
		return "", fmt.Errorf("%s failed to parse as IP", a)
	}
	ip[3] += byte(delta)
	// we know a is a good IP
	ao, _ := Octet(ctx, a, 3)
	ipo, err := Octet(ctx, ip.String(), 3)
	if err != nil {
		return "", err
	}
	if ipo != ao+delta {
		return "", fmt.Errorf("range wrap err")
	}
	return ip.String(), nil
}

func DecrIP(ctx context.Context, a string, delta int) (string, error) {
	ip := net.ParseIP(a)
	if ip == nil {
		return "", fmt.Errorf("%s failed to parse as IP", a)
	}
	ip = ip.To4()
	if ip == nil {
		return "", fmt.Errorf("Ip %s Not a v4 address", ip)
	}
	ip[3] -= byte(delta)

	ao, _ := Octet(ctx, a, 3)
	ipo, _ := Octet(ctx, ip.String(), 3)
	if ao-delta != ipo {
		return "", fmt.Errorf("range wrap error")
	}
	return ip.String(), nil
}

func Octet(ctx context.Context, a string, n int) (int, error) {
	addr := a
	// strip any cidr mask if present
	parts := strings.Split(a, "/")
	if len(parts) > 1 {
		addr = string(parts[0])
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return 0, fmt.Errorf("%s failed to parse as IP", a)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("Ip %s Not a v4 address", ip)
	}
	return int(ip[n]), nil
}

func MaskToCidr(addr string) (string, error) {

	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("ip %s not valid", ip)
	}
	c, _ := net.IPMask(ip.To4()).Size()
	cidr := strconv.Itoa(c)
	return cidr, nil
}

const MaxSubnetsPerSharedLB = 254

func GetNextAvailConIdx(ctx context.Context, ncs *types.NetworkConnectionSection) (int, error) {
	// return first unused conIdx for a vapp or vm
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextAvailConIdx", "ncs", ncs)
	conIdMap := make(map[int]*types.NetworkConnection)
	for _, nc := range ncs.NetworkConnection {
		// Before we add this idx as inuse, does this entry actually point at a connection?
		if nc.Network == "" || nc.IPAddress == "" || nc.IsConnected == false {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextAvailConIdx skip available empty", "ConIdx", nc.NetworkConnectionIndex, "IP", nc.IPAddress, "isConnected", nc.IsConnected)
			continue
		}
		conIdMap[nc.NetworkConnectionIndex] = nc
	}
	var curIdx int
	for curIdx = 1; curIdx < MaxSubnetsPerSharedLB; curIdx++ {
		if _, found := conIdMap[curIdx]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextAvailConIdx returns", "conIdx", curIdx)
			return curIdx, nil
		}
	}
	return -1, fmt.Errorf("Subnet range exahusted")
}

func (v *VcdPlatform) IncrCidr(a string, delta int) (string, error) {

	addr := a
	parts := strings.Split(a, "/")
	if len(parts) > 1 {
		addr = string(parts[0])
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("%s failed to parse as IP", a)
	}
	ip = ip.To4()

	ip[2] += byte(delta)
	return ip.String(), nil
}

// VappNetworks

const InternalNetMax = 100

func (v *VcdPlatform) CreateInternalNetworkForNewVm(ctx context.Context, vapp *govcd.VApp, serverName, netName string, cidr string) (string, error) {
	var iprange []*types.IPRange

	log.SpanLog(ctx, log.DebugLevelInfra, "create internal server net", "netname", netName)

	description := fmt.Sprintf("internal-%s", cidr)
	a := strings.Split(cidr, "/")
	addr := string(a[0])
	gateway := addr
	dns2 := ""

	startAddr, err := IncrIP(ctx, gateway, 1)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "IncrIP startaddr", "netname", netName, "gateway", gateway, "err", err)
		return "", err
	}
	endAddr, err := IncrIP(ctx, gateway, InternalNetMax)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "IncrIP endaddr", "netname", netName, "gateway", gateway, "err", err)
		return "", err
	}

	if v.Verbose {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalNetwork ", "serverName", serverName, "netname", netName, "gateway", gateway, "StartIP", startAddr, "EndIP", endAddr)
	}

	addrRange := types.IPRange{
		StartAddress: startAddr,
		EndAddress:   endAddr,
	}
	iprange = append(iprange, &addrRange)

	internalSettings := govcd.VappNetworkSettings{
		Name:             netName,
		Description:      description,
		Gateway:          gateway,
		NetMask:          "255.255.255.0",
		DNS1:             "1.1.1.1",
		DNS2:             dns2,
		DNSSuffix:        "mobiledgex.net",
		StaticIPRanges:   iprange,
		VappFenceEnabled: TakeBoolPointer(true),
	}
	_, err = vapp.CreateVappNetwork(&internalSettings, nil)
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

func (v *VcdPlatform) AddVappNetwork(ctx context.Context, vapp *govcd.VApp, vcdClient *govcd.VCDClient) (*types.NetworkConfigSection, error) {

	orgNet, err := v.GetExtNetwork(ctx, vcdClient)
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
		if n.Network != v.vmProperties.GetCloudletExternalNetwork() {
			addrs = append(addrs, n.IPAddress)
		}
	}
	return addrs, nil

}

func (v *VcdPlatform) GetExtAddrOfVM(ctx context.Context, vm *govcd.VM, netName string) (string, error) {

	if vm == nil {
		return "", fmt.Errorf("Invalid Arg")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetExtAddrOfVm", "vm", vm.VM.Name, "netname", netName)
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
	log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVm", "vm", vm.VM.Name, "netname", netName)
	ncs, err := vm.GetNetworkConnectionSection()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM err retrieving NetConnection section", "vapp", vm.VM.Name, "network", netName, "err", err)
		return "", err
	}
	numNets := len(ncs.NetworkConnection)
	log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM", "vapp", vm.VM.Name, "network", netName, "numNetworks", numNets)
	for _, nc := range ncs.NetworkConnection {
		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM consider", "vm", vm.VM.Name, "nc.Network", nc.Network, "vs netName", netName)
		}

		if nc.Network == netName {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVM", "Vapp", vm.VM.Name, "Net", netName, "ip", nc.IPAddress)
			return nc.IPAddress, nil
		}
	}

	return "", fmt.Errorf("Addr Not Found on Net")
}

// Consider a GetAllSubnetsInVapp() []string err for shared lbs... XXX
// Returns the first addr on the given network.
func (v *VcdPlatform) GetAddrOfVapp(ctx context.Context, vapp *govcd.VApp, netName string) (string, error) {

	if vapp == nil || vapp.VApp == nil || vapp.VApp.Children == nil || len(vapp.VApp.Children.VM) == 0 {
		return "", fmt.Errorf("Invalid Arg")
	}
	err := vapp.Refresh()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp error refreshing vapp", "vapp", vapp.VApp.Name, "err", err)
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp", "vapp", vapp.VApp.Name, "network", netName)
	// *v.2.11.0 Changed spelling of type.VM
	var vmChild *types.Vm

	for _, vmChild = range vapp.VApp.Children.VM {
		parts := strings.Split(vmChild.Name, ".")
		if len(parts) > 1 {
			break
		}
	}
	vm, err := v.FindVMInVApp(ctx, vmChild.Name, *vapp)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetAddrOfVapp failed to retrieve", "vm", vmChild.Name, "From Vapp", vapp.VApp.Name)
		return "", err
	}
	return v.GetAddrOfVM(ctx, vm, netName)
}

// Given our scheme for networks 10.101.X.0/24 return the next available Isolated network CIDR

func (v *VcdPlatform) GetNextInternalSubnet(ctx context.Context, vappName string, updateCallback edgeproto.CacheUpdateCallback, vcdClient *govcd.VCDClient) (string, bool, error) {

	var MAX_CIDRS = 255 // These are internal /24 subnets so 255, not that we'll have that many clusters / cloudlet

	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet", "vapp", vappName)

	// look at the free list and use that if available
	if len(v.FreeIsoNets) > 0 {
		cidr := v.getAvailableIsoNetwork(ctx)
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNetInternalSubnet  FreeNets available reusing", "net", cidr)
		return cidr, true, nil
	}

	// use schema xxx
	startAddr := "10.101.1.1"
	// We'll incr the netSpec.DelimiterOctet of this start addr, if it's not in our
	// All VApps map, it's available
	curAddr := startAddr
	vappMap, err := v.GetAllVAppsForVdcByIntAddr(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet return", "curAddr", curAddr)
		return curAddr, false, err
	}
	if len(vappMap) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet  return", "curAddr", curAddr)
		return curAddr, false, nil
	}
	for i := 1; i < MAX_CIDRS; i++ {
		if _, found := vappMap[curAddr]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet return available", "cidr", curAddr)
			return curAddr, false, nil
		}
		curAddr, err = v.IncrCidr(curAddr, 1)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet IncrCidr failed", "curAddr", curAddr, "err", err)
			return "", false, err
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextInternalSubnet  subnet range exhasted")
	return "", false, err
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
			IPAddressAllocationMode: types.IPAllocationModePool, //Manual,
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

func (v *VcdPlatform) AddSubnetIdToShareLBClientVapp(ctx context.Context, netName string, vapp *govcd.VApp) error {

	task, err := vapp.AddMetadata("SharedSubnetId", netName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Addmetadata to  vapp  failed", "vapp", vapp.VApp.Name, "error", err)
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "wait Addmetadata to  vapp  failed", "vapp", vapp.VApp.Name, "error", err)

		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "AddSubnetIdToSharedLBClientVapp", "SharedSubnetId", netName)
	return nil
}

func (v *VcdPlatform) ValidateAdditionalNetworks(ctx context.Context, additionalNets []string) error {
	return fmt.Errorf("Additional networks not supported in vCD  cloudlets")
}

func (v *VcdPlatform) getAvailableIsoNetwork(ctx context.Context) string {

	for k, _ := range v.FreeIsoNets {
		delete(v.FreeIsoNets, k)
		log.SpanLog(ctx, log.DebugLevelInfra, "getAvailableIso returns", "network", k)
		return k
	}
	return ""
}

func (v *VcdPlatform) CreateIsoVdcNetwork(ctx context.Context, vapp *govcd.VApp, netName, cidr string, vcdClient *govcd.VCDClient, reuseExistingNet bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetowrk", "name", netName, "cidr", cidr, "reusing existing net", reuseExistingNet)

	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork GetVdc failed ", "err", err)
		return err
	}
	// we are under lock here. First check if we have any free iosnets available to use
	if !reuseExistingNet {
		/*
				if len(v.FreeIsoNets) > 0 {
				cidr = v.getAvailableIsoNetwork(ctx)
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork FreeNets available reusing", "net", cidr, "for mexNet", netName)
			} else {
		*/
		// create a new one
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork FreeNets empty creating new", "net", cidr)
		startAddr, err := IncrIP(ctx, cidr, 1)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "IncrIP startaddr", "netname", netName, "gateway", cidr, "err", err)
			return err
		}
		endAddr, err := IncrIP(ctx, cidr, MAXCIDR)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "IncrIP endaddr", "netname", netName, "gateway", cidr, "err", err)
			return err
		}

		var (
			gateway       = cidr
			networkName   = cidr // netName
			startAddress  = startAddr
			endAddress    = endAddr
			netmask       = "255.255.255.0"
			dns1          = "1.1.1.1"
			dns2          = "8.8.8.8"
			dnsSuffix     = "mobiledgex.net"
			description   = "mex vdc sharedLB subnet"
			networkConfig = types.OrgVDCNetwork{
				Xmlns:       types.XMLNamespaceVCloud,
				Name:        networkName,
				Description: description,
				Configuration: &types.NetworkConfiguration{
					FenceMode: types.FenceModeIsolated,
					IPScopes: &types.IPScopes{
						IPScope: []*types.IPScope{&types.IPScope{
							IsInherited: false,
							Gateway:     gateway,
							Netmask:     netmask,
							DNS1:        dns1,
							DNS2:        dns2,
							DNSSuffix:   dnsSuffix,
							IPRanges: &types.IPRanges{
								IPRange: []*types.IPRange{
									&types.IPRange{
										StartAddress: startAddress,
										EndAddress:   endAddress,
									},
								},
							},
						},
						},
					},
					BackwardCompatibilityMode: true,
				},
				IsShared: false, // private to this vdc
			}
		)

		task, err := vdc.CreateOrgVDCNetwork(&networkConfig)
		// accept  a pre-existing network
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork use existing orgvdcnet", "netName", netName)
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork CreateOrgVDCNetwork  failed ", "err", err)
				return err
			}
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork CreateOrgVDCNetwork wait failed ", "err", err)
		}

	}

	orgvdcnet, err := vdc.GetOrgVdcNetworkByName(cidr /*netName*/, true)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork GetOrgVDCNetwork  failed ", "netName", netName, "err", err)
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetowrk created", "name", netName, "real net", cidr)
	vappNetSettings := &govcd.VappNetworkSettings{
		Name:             cidr,
		VappFenceEnabled: TakeBoolPointer(false),
	}

	netConfSec, err := vapp.AddOrgNetwork(vappNetSettings, orgvdcnet.OrgVDCNetwork, false)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetwork AddOrgNetwork  failed ", "netName", netName, "cidr", cidr, "err", err)
		return err
	}
	// xlate names map for network reuse. All these iossubnets are now named with their cidrs
	// Lookup the real vdc name (subnetId) using our netName
	v.IsoNamesMap[netName] = cidr
	// enable crm restarts, stash the subnetId name in Vapp as metadata)
	v.AddSubnetIdToShareLBClientVapp(ctx, netName, vapp)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateIsoVdcNetowrk added org net ok", "network", netName, "cidr", cidr, "vapp", vapp.VApp.Name, "NetConfigSection", netConfSec)
	return nil
}

const MAXCIDR = 254

func (v *VcdPlatform) GetNextVdcIsoSubnet(ctx context.Context, vcdClient *govcd.VCDClient) (string, error) {
	var err error
	netMap := make(NetMap)
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet")
	// Get a list of all vdc networks type 2 and make a map
	// interate through our subnet schema  n.1 for the first available
	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet GetVdc failed ", "err", err)
		return "", err
	}
	qrecs, err := vdc.GetNetworkList()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet GetNetworkList failed ", "err", err)
		return "", err
	}
	for _, qr := range qrecs {
		if qr.LinkType == 2 {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet found ", "iso subnet", qr.Name)
			netMap[qr.DefaultGateway] = &govcd.OrgVDCNetwork{}
		}
	}
	curCidr := "10.101.1.1"
	if len(netMap) == 0 {
		return curCidr, nil
	}
	// maybe we should use 192. for this here to differentiate between vapp private subnets. ?
	for i := 0; i < MAXCIDR; i++ {
		if _, found := netMap[curCidr]; !found {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet", "subnet", curCidr)
			return curCidr, nil
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet skip in use", "subnet", curCidr)
		curCidr, err = v.IncrCidr(curCidr, 1)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetNextVdcIsoSubnet  IncrIP err", "curAddr", curCidr, "err", err)
			return "", err
		}
	}
	return "", fmt.Errorf("Isolated Subnet range exhausted")
}

// If this vapp is using an isolated orgvdcnet, return its name
func (v *VcdPlatform) GetVappIsoNetwork(ctx context.Context, vdc *govcd.Vdc, vapp *govcd.VApp) (string, error) {

	log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork", "vapp", vapp.VApp.Name)

	ncs, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork Failed to retrieve networkConnectionSection from", "vapp", vapp.VApp.Name, "err", err)
		return "", err
	}
	qr, err := vdc.GetNetworkList()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork GetNetworkList  failed", "err", err)
		return "", err
	}
	for _, nc := range ncs.NetworkConnection {
		for _, q := range qr {
			if q.LinkType == 2 && nc.Network == q.Name {
				log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork vapp using isoOrgVdcNetwork", "netName", q.Name)
				return q.Name, nil
			}
			if v.Verbose {
				log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork vapp using non-isoOrgVdcNet", "netName", q.Name)
			}
		}

	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork no IsoNetwork found for ", "vapp", vapp.VApp.Name)
	return "", nil

}

// on (re)start attempt to rebuild the free isolated networks and name mapping.
// On cloudlet create, if any of these isonets are in existance, (nsx-t)
// they will be placed on the free list and reused.
func (v *VcdPlatform) RebuildIsoNamesAndFreeMaps(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RebuildIsoNamesMap ")
	var err error
	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		// Too early for context
		vcdClient, err = v.GetClient(ctx, v.Creds)
		if err != nil {
			return fmt.Errorf(NoVCDClientInContext)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Obtained client directly continuing")
	}

	vdc, err := v.GetVdc(ctx, vcdClient)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer unable to retrieve current vdc", "err", err)
		return err
	}

	qr, err := vdc.GetNetworkList()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVappIsoNetwork GetNetworkList  failed", "err", err)
		return err
	}
	for _, q := range qr {
		if q.LinkType == 2 {
			orgvdcnetwork, err := vdc.GetOrgVdcNetworkByName(q.Name, false)
			if err != nil {
				return err
			}
			subnetId, err := v.vappUsesIsoNet(ctx, orgvdcnetwork.OrgVDCNetwork.Name, vdc)
			if subnetId != "" {
				log.SpanLog(ctx, log.DebugLevelInfra, "RebuildMaps add FreeIsoNets", "Net", orgvdcnetwork.OrgVDCNetwork.Name, "subnetId", subnetId)
				v.IsoNamesMap[orgvdcnetwork.OrgVDCNetwork.Name] = subnetId
			} else {
				// No existing vapp references this orgvdc type 2 net, freelist it
				v.FreeIsoNets[orgvdcnetwork.OrgVDCNetwork.Name] = orgvdcnetwork
			}

		}
	}
	if len(v.IsoNamesMap) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "RebuildIsoNamesMap Nothing to rebuild")
	}
	return nil
}

// If any exsting vapp is using this iso net, it's not free to reuse
func (v *VcdPlatform) vappUsesIsoNet(ctx context.Context, netName string, vdc *govcd.Vdc) (string, error) {

	// For all vapps in vdc
	for _, r := range vdc.Vdc.ResourceEntities {
		for _, res := range r.ResourceEntity {
			if res.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vapp, err := vdc.GetVAppByName(res.Name, true)
				if err != nil {
					return "", err
				} else {
					ip, _ := v.GetAddrOfVapp(ctx, vapp, netName)
					if ip != "" {
						md, err := vapp.GetMetadata()
						if err != nil {
							return "", err
						}
						for _, ent := range md.MetadataEntry {
							if ent.Key == "SharedSubnetId" {
								return ent.TypedValue.Value, nil
							}
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("Not Found")
}

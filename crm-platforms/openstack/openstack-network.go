package openstack

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/log"
)

//FindNodeIP finds IP for the given node
func (s *OpenstackPlatform) FindNodeIP(name string, srvs []OSServer) (string, error) {
	//log.SpanLog(ctx,log.DebugLevelInfra, "find node ip", "name", name)
	if name == "" {
		return "", fmt.Errorf("empty name")
	}

	for _, srv := range srvs {
		if srv.Status == "ACTIVE" && srv.Name == name {
			ipaddr, err := s.GetServerInternalIP(srv.Networks)
			if err != nil {
				return "", fmt.Errorf("can't get IP for %s, %v", srv.Name, err)
			}
			//log.SpanLog(ctx,log.DebugLevelInfra, "found node ip", "name", name, "ipaddr", ipaddr)
			return ipaddr, nil
		}
	}
	return "", fmt.Errorf("node %s, ip not found", name)
}

//FindClusterMaster finds cluster given a key string
func (o *OpenstackPlatform) FindClusterMaster(ctx context.Context, namePrefix, nameSuffix string, srvs []OSServer) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "FindClusterMaster", "namePrefix", namePrefix, "nameSuffix", nameSuffix)
	if namePrefix == "" || nameSuffix == "" {
		return "", fmt.Errorf("empty name component")
	}
	for _, s := range srvs {
		if s.Status == "ACTIVE" && strings.HasSuffix(s.Name, nameSuffix) && strings.HasPrefix(s.Name, namePrefix) {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("VM %s not found", nameSuffix)
}

//GetIPFromServerName gets the server IP(s) for the given network
func (o *OpenstackPlatform) GetIPFromServerName(ctx context.Context, networkName, serverName string) (*vmlayer.ServerIP, error) {
	// if this is a root lb, look it up and get the IP if we have it cached
	rootLB, err := o.vmProvider.GetRootLB(ctx, serverName)
	if err == nil && rootLB != nil {
		if rootLB.IP != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "using existing rootLB IP", "IP", rootLB.IP)
			return rootLB.IP, nil
		}
	}
	sd, err := o.GetServerDetail(ctx, serverName)
	if err != nil {
		return nil, err
	}
	return o.vmProvider.GetIPFromServerDetails(ctx, networkName, sd)
}

//GetExternalGateway retrieves Gateway IP from the external network information. It first gets external
//  network information. Using that it further gets subnet information. Inside that subnet information
//  there should be gateway IP if the network is set up correctly.
// Not to be confused with GetRouterDetailExternalGateway.
func (s *OpenstackPlatform) GetExternalGateway(ctx context.Context, extNetName string) (string, error) {
	nd, err := s.GetNetworkDetail(ctx, extNetName)
	if err != nil {
		return "", fmt.Errorf("can't get details for external network %s, %v", extNetName, err)
	}

	if nd.Status != "ACTIVE" {
		return "", fmt.Errorf("network %s is not active, status %s", extNetName, nd.Status)
	}
	if nd.AdminStateUp != "UP" {
		return "", fmt.Errorf("network %s is not admin-state set to up", extNetName)
	}
	subnets := strings.Split(nd.Subnets, ",")
	//XXX beware of extra spaces
	if len(subnets) < 1 {
		return "", fmt.Errorf("no subnets for %s", extNetName)
	}
	//XXX just use first subnet -- may not work in all cases, but there is no tagging done rightly yet
	sd, err := s.GetSubnetDetail(ctx, subnets[0])
	if err != nil {
		return "", fmt.Errorf("cannot get details for subnet %s, %v", subnets[0], err)
	}
	//TODO check status of subnet
	if sd.GatewayIP == "" {
		return "", fmt.Errorf("cannot get external network's gateway IP")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "get external gatewayIP", "gatewayIP", sd.GatewayIP, "subnet detail", sd)
	return sd.GatewayIP, nil
}

//GetRouterDetailExternalGateway is different than GetExternalGateway.  This function gets
// the gateway interface in the subnet within external network.  This is
// accessible from private networks to route packets to the external network.
// The GetExternalGateway gets the gateway for the outside network.   This is
// for the packets to be routed out to the external network, i.e. internet.
func GetRouterDetailExternalGateway(rd *OSRouterDetail) (*OSExternalGateway, error) {
	if rd.ExternalGatewayInfo == "" {
		return nil, fmt.Errorf("empty external gateway info")
	}
	externalGateway := &OSExternalGateway{}
	err := json.Unmarshal([]byte(rd.ExternalGatewayInfo), externalGateway)
	if err != nil {
		return nil, fmt.Errorf("can't get unmarshal external gateway info, %v", err)
	}
	//log.SpanLog(ctx,log.DebugLevelInfra, "get router detail external gateway", "external gateway", externalGateway)
	return externalGateway, nil
}

// GetRouterDetailInterfaces gets the list of interfaces on the router. For example, each private
// subnet connected to the router will be listed here with own interface definition.
func GetRouterDetailInterfaces(ctx context.Context, rd *OSRouterDetail) ([]OSRouterInterface, error) {
	if rd.InterfacesInfo == "" {
		return nil, fmt.Errorf("missing interfaces info in router details")
	}
	interfaces := []OSRouterInterface{}
	err := json.Unmarshal([]byte(rd.InterfacesInfo), &interfaces)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal router detail interfaces")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "get router detail interfaces", "interfaces", interfaces)
	return interfaces, nil
}

func (o *OpenstackPlatform) GetMexRouterIP(ctx context.Context) (string, error) {
	rtr := o.GetCloudletExternalRouter()
	if rtr == infracommon.NoConfigExternalRouter || rtr == infracommon.NoExternalRouter {
		return "", nil
	}
	rd, rderr := o.GetRouterDetail(ctx, rtr)
	if rderr != nil {
		return "", fmt.Errorf("can't get router detail for %s, %v", rtr, rderr)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "router detail", "detail", rd)
	reg, regerr := GetRouterDetailExternalGateway(rd)
	if regerr != nil {
		// some deployments will not be able to retrieve the router GW at all, allow this
		log.SpanLog(ctx, log.DebugLevelInfra, "can't get router external GW, continuing", "error", regerr)
		return "", nil
	}
	if reg != nil && len(reg.ExternalFixedIPs) > 0 {
		fip := reg.ExternalFixedIPs[0]
		log.SpanLog(ctx, log.DebugLevelInfra, "external fixed ips", "ips", fip)
		return fip.IPAddress, nil
	} else {
		// some networks may not have an external fixed ip for the router.  This is not fatal
		log.SpanLog(ctx, log.DebugLevelInfra, "can't get external fixed ips list from router detail external gateway, returning blank ip")
		return "", nil
	}
}

func (o *OpenstackPlatform) ValidateNetwork(ctx context.Context) error {
	nets, err := o.ListNetworks(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, n := range nets {
		if n.Name == s.vmProvider.GetCloudletExternalNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find external network %s", s.vmProvider.GetCloudletExternalNetwork())
	}

	found = false
	for _, n := range nets {
		if n.Name == s.vmProvider.GetCloudletMexNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find network %s", s.vmProvider.GetCloudletMexNetwork())
	}

	rtr := o.GetCloudletExternalRouter()
	if rtr != infracommon.NoConfigExternalRouter && rtr != infracommon.NoExternalRouter {
		routers, err := o.ListRouters(ctx)
		if err != nil {
			return err
		}

		found = false
		for _, r := range routers {
			if r.Name == o.GetCloudletExternalRouter() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ext router %s not found", o.GetCloudletExternalRouter())
		}
	}

	return nil
}

//PrepNetwork validates and does the work needed to ensure MEX network setup
func (o *OpenstackPlatform) PrepNetwork(ctx context.Context) error {
	nets, err := o.ListNetworks(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, n := range nets {
		if n.Name == s.vmProvider.GetCloudletExternalNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find ext net %s", s.vmProvider.GetCloudletExternalNetwork())
	}

	found = false
	for _, n := range nets {
		if n.Name == s.vmProvider.GetCloudletMexNetwork() {
			found = true
			break
		}
	}
	if !found {
		ni, err := infracommon.ParseNetSpec(ctx, s.vmProvider.GetCloudletNetworkScheme())
		if err != nil {
			return err
		}
		// We need at least one network for `mex` clusters
		err = o.CreateNetwork(ctx, s.vmProvider.GetCloudletMexNetwork(), ni.NetworkType)
		if err != nil {
			return fmt.Errorf("cannot create mex network %s, %v", s.vmProvider.GetCloudletMexNetwork(), err)
		}
	}

	rtr := o.GetCloudletExternalRouter()
	if rtr != infracommon.NoConfigExternalRouter && rtr != infracommon.NoExternalRouter {
		routers, err := o.ListRouters(ctx)
		if err != nil {
			return err
		}

		found = false
		for _, r := range routers {
			if r.Name == o.GetCloudletExternalRouter() {
				found = true
				break
			}
		}
		if !found {
			// We need at least one router for our `mex` network and external network
			err = o.CreateRouter(ctx, o.GetCloudletExternalRouter())
			if err != nil {
				return fmt.Errorf("cannot create the ext router %s, %v", o.GetCloudletExternalRouter(), err)
			}
			err = o.SetRouter(ctx, o.GetCloudletExternalRouter(), s.vmProvider.GetCloudletExternalNetwork())
			if err != nil {
				return fmt.Errorf("cannot set default network to router %s, %v", o.GetCloudletExternalRouter(), err)
			}
		}
	}

	return nil
}

//GetCloudletSubnets returns subnets inside MEX Network
func (o *OpenstackPlatform) GetCloudletSubnets(ctx context.Context) ([]string, error) {
	nd, err := o.GetNetworkDetail(ctx, s.vmProvider.GetCloudletMexNetwork())
	if err != nil {
		return nil, fmt.Errorf("can't get MEX network detail, %v", err)
	}

	subnets := strings.Split(nd.Subnets, ",")
	if len(subnets) < 1 {
		return nil, fmt.Errorf("can't get a list of subnets for MEX network")
	}

	return subnets, nil
}

func getNameAndIPFromNetwork(network string) (string, string, error) {
	nets := strings.Split(network, "=")
	if len(nets) != 2 {
		return "", "", fmt.Errorf("unable to parse net %s", network)
	}
	return nets[0], nets[1], nil
}

func GetServerNetworkIP(networks, netmatch string) (string, error) {
	for _, n := range strings.Split(networks, "'") {
		netname, ip, err := getNameAndIPFromNetwork(n)
		if err != nil {
			return "", err
		}
		if strings.Contains(netname, netmatch) {
			return ip, nil
		}
	}
	return "", fmt.Errorf("no network matching: %s", netmatch)
}

func (o *OpenstackPlatform) GetServerExternalIP(networks string) (string, error) {
	extNet := s.vmProvider.GetCloudletExternalNetwork()
	return GetServerNetworkIP(networks, extNet)
}

func (o *OpenstackPlatform) GetServerInternalIP(networks string) (string, error) {
	mexNet := s.vmProvider.GetCloudletMexNetwork()
	return GetServerNetworkIP(networks, mexNet)
}

//GetInternalIP returns IP of the server
func (s *OpenstackPlatform) GetInternalIP(name string, srvs []OSServer) (string, error) {
	for _, srv := range srvs {
		if srv.Name == name {
			return s.GetServerInternalIP(srv.Networks)
		}
	}
	return "", fmt.Errorf("No internal IP found for %s", name)
}

//GetInternalCIDR returns CIDR of server
func (s *OpenstackPlatform) GetInternalCIDR(name string, srvs []OSServer) (string, error) {
	addr, err := s.GetInternalIP(name, srvs)
	if err != nil {
		return "", err
	}
	cidr := addr + "/24" // XXX we use this convention of /24 in k8s priv-net
	return cidr, nil
}

// TODO collapse common keys into a single entry with multi-part values ex: "hw"
// (We don't use this property values today, but perhaps in the future)
func ParseFlavorProperties(f OSFlavorDetail) map[string]string {

	var props map[string]string

	ms := strings.Split(f.Properties, ",")
	props = make(map[string]string)
	for _, m := range ms {
		// ex: pci_passthrough:alias='t4gpu:1’
		val := strings.Split(m, ":")
		if len(val) > 1 {
			val[0] = strings.TrimSpace(val[0])
			var s []string
			for i := 1; i < len(val); i++ {
				val[i] = strings.Replace(val[i], "'", "", -1)
				if _, err := strconv.Atoi(val[i]); err == nil {
					s = append(s, ":")
				}
				s = append(s, val[i])
			}
			props[val[0]] = strings.Join(s, "")
		}

	}
	return props
}

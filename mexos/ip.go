package mexos

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// NetworkTypeVLAN is an OpenStack provider network type
const NetworkTypeVLAN string = "vlan"

// ServerIP is an IP address for a given network on a port.  In the case of floating IPs, there are both
// internal and external addresses which are associated via NAT.   In the non floating case, the external and internal are the same
type ServerIP struct {
	InternalAddr           string // this is the address used inside the server
	ExternalAddr           string // this is external with respect to the server, not necessarily internet reachable.  Can be a floating IP
	ExternalAddrIsFloating bool
}

type NetSpecInfo struct {
	Name, CIDR        string
	NetworkType       string
	NetworkAddress    string
	NetmaskBits       string
	Octets            []string
	MasterIPLastOctet string
	DelimiterOctet    int // this is the X
	FloatingIPNet     string
	FloatingIPSubnet  string
	VnicType          string
	RouterGatewayIP   string
}

const ClusterNotFoundErr string = "cluster not found"

//ParseNetSpec decodes netspec string
//TODO: IPv6
func ParseNetSpec(ctx context.Context, netSpec string) (*NetSpecInfo, error) {
	ni := &NetSpecInfo{}
	if netSpec == "" {
		return nil, fmt.Errorf("empty netspec")
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "parsing netspec", "netspec", netSpec)
	items := strings.Split(netSpec, ",")
	for _, i := range items {
		kvs := strings.Split(i, "=")
		if len(kvs) != 2 {
			return nil, fmt.Errorf("incorrect netspec item format, expect key=value: %s", i)
		}
		k := strings.ToLower(kvs[0])
		v := kvs[1]

		switch k {
		case "name":
			ni.Name = v
		case "cidr":
			ni.CIDR = v
		case "floatingipnet":
			ni.FloatingIPNet = v
		case "floatingipsubnet":
			ni.FloatingIPSubnet = v
		case "vnictype":
			ni.VnicType = v
		case "routergateway":
			ni.RouterGatewayIP = v
		case "networktype":
			ni.NetworkType = v
		default:
			return nil, fmt.Errorf("unknown netspec item key: %s", k)
		}
	}
	if ni.Name == "" {
		return nil, fmt.Errorf("Missing name=(value) in netspec")
	}
	if ni.CIDR == "" {
		return nil, fmt.Errorf("Missing cidr=(value) in netspec")
	}
	sits := strings.Split(ni.CIDR, "/")
	if len(sits) < 2 {
		return nil, fmt.Errorf("invalid CIDR, no net mask")
	}
	ni.NetworkAddress = sits[0]
	ni.NetmaskBits = sits[1]

	ni.Octets = strings.Split(ni.NetworkAddress, ".")
	for i, it := range ni.Octets {
		if it == "X" {
			ni.DelimiterOctet = i
		}
	}
	if len(ni.Octets) != 4 {
		log.SpanLog(ctx, log.DebugLevelMexos, "invalid network address, wrong number of octets", "octets", ni.Octets)
		return nil, fmt.Errorf("invalid network address structure")
	}
	if ni.DelimiterOctet != 2 {
		log.SpanLog(ctx, log.DebugLevelMexos, "invalid network address, third octet must be X", "delimiterOctet", ni.DelimiterOctet)
		return nil, fmt.Errorf("invalid network address delimiter")
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "netspec info", "ni", ni, "items", items)
	return ni, nil
}

//GetInternalIP returns IP of the server
func GetInternalIP(name string, srvs []OSServer) (string, error) {
	for _, s := range srvs {
		if s.Name == name {
			return s.GetServerInternalIP()
		}
	}
	return "", fmt.Errorf("No internal IP found for %s", name)
}

//GetInternalCIDR returns CIDR of server
func GetInternalCIDR(name string, srvs []OSServer) (string, error) {
	addr, err := GetInternalIP(name, srvs)
	if err != nil {
		return "", err
	}
	cidr := addr + "/24" // XXX we use this convention of /24 in k8s priv-net
	return cidr, nil
}

func GetAllowedClientCIDR() string {
	//XXX TODO get real list of allowed clients from remote database or template configuration
	return "0.0.0.0/0"
}

// GetServerIPFromAddrs gets the ServerIP forthe given network from the addresses provided
func GetServerIPFromAddrs(ctx context.Context, networkName, addresses, serverName string) (*ServerIP, error) {
	var serverIP ServerIP
	its := strings.Split(addresses, ";")
	for _, it := range its {
		sits := strings.Split(it, "=")
		if len(sits) != 2 {
			return &serverIP, fmt.Errorf("GetServerIPFromAddrs: Unable to parse '%s'", it)
		}
		if strings.Contains(sits[0], networkName) {
			addr := sits[1]
			// the comma indicates a floating IP is present.
			if strings.Contains(addr, ",") {
				addrs := strings.Split(addr, ",")
				if len(addrs) == 2 {
					serverIP.InternalAddr = strings.TrimSpace(addrs[0])
					serverIP.ExternalAddr = strings.TrimSpace(addrs[1])
					serverIP.ExternalAddrIsFloating = true
				} else {
					return &serverIP, fmt.Errorf("GetServerExternalIPFromAddr: Unable to parse '%s'", addr)
				}
			} else {
				// no floating IP, internal and external are the same
				addr = strings.TrimSpace(addr)
				serverIP.InternalAddr = addr
				serverIP.ExternalAddr = addr
			}
			log.SpanLog(ctx, log.DebugLevelMexos, "retrieved server ipaddr", "ipaddr", addr, "netname", networkName, "servername", serverName)
			return &serverIP, nil
		}
	}
	// this is a bug
	log.WarnLog("Unable to find network for server", "networkName", networkName, "serverName", serverName)
	return &serverIP, fmt.Errorf("Unable to find network %s for server %s", networkName, serverName)
}

//GetServerIPAddr gets the server IP(s) for the given network
func GetServerIPAddr(ctx context.Context, networkName, serverName string) (*ServerIP, error) {
	// if this is a root lb, look it up and get the IP if we have it cached
	rootLB, err := getRootLB(ctx, serverName)
	if err == nil && rootLB != nil {
		if rootLB.IP != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "using existing rootLB IP", "IP", rootLB.IP)
			return rootLB.IP, nil
		}
	}
	sd, err := GetActiveServerDetails(ctx, serverName)
	if err != nil {
		return nil, err
	}
	return GetServerIPFromAddrs(ctx, networkName, sd.Addresses, serverName)
}

//FindNodeIP finds IP for the given node
func FindNodeIP(name string, srvs []OSServer) (string, error) {
	//log.SpanLog(ctx,log.DebugLevelMexos, "find node ip", "name", name)
	if name == "" {
		return "", fmt.Errorf("empty name")
	}

	for _, s := range srvs {
		if s.Status == "ACTIVE" && s.Name == name {
			ipaddr, err := s.GetServerInternalIP()
			if err != nil {
				return "", fmt.Errorf("can't get IP for %s, %v", s.Name, err)
			}
			//log.SpanLog(ctx,log.DebugLevelMexos, "found node ip", "name", name, "ipaddr", ipaddr)
			return ipaddr, nil
		}
	}
	return "", fmt.Errorf("node %s, ip not found", name)
}

// GetMasterNameAndIP gets the name and IP address of the cluster's master node.
func GetMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "get master IP", "cluster", clusterInst.Key.ClusterKey.Name)
	srvs, err := ListServers(ctx)
	if err != nil {
		return "", "", fmt.Errorf("error getting server list: %v", err)

	}
	namePrefix := ClusterTypeKubernetesMasterLabel
	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		namePrefix = ClusterTypeDockerVMLabel
	}

	nodeNameSuffix := k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	masterName, err := FindClusterMaster(ctx, namePrefix, nodeNameSuffix, srvs)
	if err != nil {
		return "", "", fmt.Errorf("%s -- %s, %v", ClusterNotFoundErr, nodeNameSuffix, err)
	}
	masterIP, err := FindNodeIP(masterName, srvs)
	return masterName, masterIP, err
}

//FindClusterMaster finds cluster given a key string
func FindClusterMaster(ctx context.Context, namePrefix, nameSuffix string, srvs []OSServer) (string, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "FindClusterMaster", "namePrefix", namePrefix, "nameSuffix", nameSuffix)
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

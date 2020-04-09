package infracommon

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
)

// NetworkTypeVLAN is an OpenStack provider network type
const NetworkTypeVLAN string = "vlan"

// ServerIP is an IP address for a given network on a port.  In the case of floating IPs, there are both
// internal and external addresses which are associated via NAT.   In the non floating case, the external and internal are the same
type ServerIP struct {
	MacAddress             string
	InternalAddr           string // this is the address used inside the server
	ExternalAddr           string // this is external with respect to the server, not necessarily internet reachable.  Can be a floating IP
	Network                string
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

func GetAllowedClientCIDR() string {
	//XXX TODO get real list of allowed clients from remote database or template configuration
	return "0.0.0.0/0"
}

package baremetal

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var (
	ResourceExternalIps = "External IPs"
)

var baremetalProps = map[string]*edgeproto.PropertyInfo{
	"BARE_METAL_CONTROL_ACCESS_IP": {
		Name:        "BareMetal Control Access IP",
		Description: "IP used to b.cess the control plane externally",
		Mandatory:   true,
	},
	"BARE_METAL_EXTERNAL_IP_RANGES": {
		Name:        "External IP Ranges(s) for BareMetal Load Balancers",
		Description: "Range of External IP addresses for BareMetal LBs, Format: StartCIDR-EndCIDR,StartCIDR2-EndCIDR2,...",
		Mandatory:   true,
	},
	"BARE_METAL_INTERNAL_IP_RANGES": {
		Name:        "Internal IP Ranges(s) for BareMetal Load Control Plane",
		Description: "Range of Internal IP addresses for BareMetal Control plane, Format: StartCIDR-EndCIDR,StartCIDR2-EndCIDR2,...",
		Mandatory:   true,
	},
	"BARE_METAL_EXTERNAL_ETH_INTERFACE": {
		Name:        "External Ethernet Interface",
		Description: "Ethernet interface used for LB, e.g. eno2",
		Mandatory:   true,
	},
	"BARE_METAL_INTERNAL_ETH_INTERFACE": {
		Name:        "Internal Ethernet Interface",
		Description: "Ethernet interface used for internal control plane",
		Mandatory:   true,
	},
}

func (b *BareMetalPlatform) GetControlAccessIp() string {
	value, _ := b.commonPf.Properties.GetValue("BARE_METAL_CONTROL_ACCESS_IP")
	return value
}

func (b *BareMetalPlatform) GetExternalIpRanges() string {
	value, _ := b.commonPf.Properties.GetValue("BARE_METAL_EXTERNAL_IP_RANGES")
	return value
}

func (b *BareMetalPlatform) GetInternalIpRanges() string {
	value, _ := b.commonPf.Properties.GetValue("BARE_METAL_INTERNAL_IP_RANGES")
	return value
}

func (b *BareMetalPlatform) GetExternalEthernetInterface() string {
	value, _ := b.commonPf.Properties.GetValue("BARE_METAL_EXTERNAL_ETH_INTERFACE")
	return value
}

func (b *BareMetalPlatform) GetInternalEthernetInterface() string {
	value, _ := b.commonPf.Properties.GetValue("BARE_METAL_INTERNAL_ETH_INTERFACE")
	return value
}

func (b *BareMetalPlatform) GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletProps")
	return &edgeproto.CloudletProps{Properties: baremetalProps}, nil
}

func (b *BareMetalPlatform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletResourceQuotaProps")
	return &edgeproto.CloudletResourceQuotaProps{
		Props: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        ResourceExternalIps,
				Description: "Limit on external IPs b.ailable",
			},
		},
	}, nil
}

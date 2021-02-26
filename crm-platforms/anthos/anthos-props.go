package anthos

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var (
	ResourceExternalIps = "External IPs"
)

var anthosProps = map[string]*edgeproto.PropertyInfo{
	"ANTHOS_CONTROL_ACCESS_IP": {
		Name:        "Anthos Control Access IP",
		Description: "IP used to access the control plane externally",
		Mandatory:   true,
	},
	"ANTHOS_CONTROL_VIP": {
		Name:        "Anthos Control Virtual IP",
		Description: "Virtual IP used to access the control plane (k8s master)",
		Mandatory:   true,
	},
	"ANTHOS_EXTERNAL_IP_RANGES": {
		Name:        "External IP Ranges(s) for Anthos Load Balancers",
		Description: "Range of External IP addresses for Anthos LBs, Format: StartCIDR-EndCIDR,StartCIDR2-EndCIDR2,...",
		Mandatory:   true,
	},
	"ANTHOS_INTERNAL_IP_RANGES": {
		Name:        "Internal IP Ranges(s) for Anthos Load Control Plane",
		Description: "Range of Internal IP addresses for Anthos Control plane, Format: StartCIDR-EndCIDR,StartCIDR2-EndCIDR2,...",
		Mandatory:   true,
	},
	"ANTHOS_EXTERNAL_ETH_INTERFACE": {
		Name:        "External Ethernet Interface",
		Description: "Ethernet interface used for LB, e.g. eno2",
		Mandatory:   true,
	},
	"ANTHOS_INTERNAL_ETH_INTERFACE": {
		Name:        "Internal Ethernet Interface",
		Description: "Ethernet interface used for internal control plane",
		Mandatory:   true,
	},
}

func (a *AnthosPlatform) GetControlAccessIp() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_CONTROL_ACCESS_IP")
	return value
}

func (a *AnthosPlatform) GetControlVip() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_CONTROL_VIP")
	return value
}

func (a *AnthosPlatform) GetExternalIpRanges() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_EXTERNAL_IP_RANGES")
	return value
}

func (a *AnthosPlatform) GetInternalIpRanges() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_INTERNAL_IP_RANGES")
	return value
}

func (a *AnthosPlatform) GetExternalEthernetInterface() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_EXTERNAL_ETH_INTERFACE")
	return value
}

func (a *AnthosPlatform) GetInternalEthernetInterface() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_INTERNAL_ETH_INTERFACE")
	return value
}

func (a *AnthosPlatform) GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletProps")
	return &edgeproto.CloudletProps{Properties: anthosProps}, nil
}

func (a *AnthosPlatform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletResourceQuotaProps")
	return &edgeproto.CloudletResourceQuotaProps{
		Props: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        ResourceExternalIps,
				Description: "Limit on external IPs available",
			},
		},
	}, nil
}

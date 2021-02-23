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
	"ANTHOS_CONTROL_VIP": {
		Name:        "Anthos Control VIP",
		Description: "Virtual IP of Anthos Control Plane",
		Mandatory:   true,
	},
	"ANTHOS_CONFIG_DIR": {
		Name:        "Anthos Config Dir",
		Description: "Directory of Anthos Config files",
		Mandatory:   true,
	},
	"ANTHOS_LB_IP_RANGES": {
		Name:        "IP Ranges(s) for Anthos Load Balancers",
		Description: "Range of IP addresses for Anthos LBs, Format: StartCIDR-EndCIDR,StartCIDR2-EndCIDR2,...",
		Mandatory:   true,
	},
	"ANTHOS_LB_ETH_INTERFACE": {
		Name:        "Load Balancer Ethernet Interface",
		Description: "Ethernet interface used for LB, e.g. eno2",
		Mandatory:   true,
	},
}

func (a *AnthosPlatform) GetControlVip() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_CONTROL_VIP")
	return value
}

func (a *AnthosPlatform) GetConfigDir() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_CONFIG_DIR")
	return value
}

func (a *AnthosPlatform) GetLbIpRanges() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_LB_IP_RANGES")
	return value
}

func (a *AnthosPlatform) GetLbEthernetInterface() string {
	value, _ := a.commonPf.Properties.GetValue("ANTHOS_LB_ETH_INTERFACE")
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

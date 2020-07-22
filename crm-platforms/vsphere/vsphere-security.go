package vsphere

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (v *VSpherePlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error {
	return vmlayer.AddIngressIptablesRules(ctx, client, secGrpName, allowedCIDR, ports)
}

func (v *VSpherePlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	return vmlayer.RemoveIngressIptablesRules(ctx, client, secGrpName, allowedCIDR, ports)
}

func (v *VSpherePlatform) PrepareRootLB(ctx context.Context, client ssh.Client, rootLBName string, secGrpName string, privacyPolicy *edgeproto.PrivacyPolicy) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB", "rootLBName", rootLBName)

	var fwRules vmlayer.FirewallRules
	//First create the global rules on this LB
	log.SpanLog(ctx, log.DebugLevelInfra, "creating cloudlet-wide rules", "rootLBName", rootLBName)
	err := v.vmProperties.CreateCloudletFirewallRules(ctx, client)
	if err != nil {
		return err
	}

	// allow external IP network
	netCidr, err := v.GetExternalIpNetworkCidr(ctx)
	if err != nil {
		return err
	}
	// now rules for SSH within the external network interfaces
	sshIngress := vmlayer.FirewallRule{
		Protocol:     "tcp",
		RemoteCidr:   netCidr,
		PortRange:    "22",
		PortEndpoint: vmlayer.DestPort,
	}
	fwRules.IngressRules = append(fwRules.IngressRules, sshIngress)
	sshEgress := vmlayer.FirewallRule{
		Protocol:     "tcp",
		RemoteCidr:   netCidr,
		PortRange:    "1:65535",
		PortEndpoint: vmlayer.DestPort,
	}
	fwRules.EgressRules = append(fwRules.EgressRules, sshEgress)

	// all traffic between the internal networks is allowed
	internalRoute, err := v.vmProperties.GetInternalNetworkRoute(ctx)
	if err != nil {
		return err
	}
	internalNetInRule := vmlayer.FirewallRule{
		RemoteCidr: internalRoute,
	}
	fwRules.IngressRules = append(fwRules.IngressRules, internalNetInRule)

	internalNetOutRule := vmlayer.FirewallRule{
		RemoteCidr: internalRoute,
	}
	fwRules.EgressRules = append(fwRules.EgressRules, internalNetOutRule)

	// optionally add Privacy Policy
	allowEgressAll := false
	if privacyPolicy != nil {
		if len(privacyPolicy.OutboundSecurityRules) == 0 {
			// a privacy policy with no rules means we need to open all egress traffic
			allowEgressAll = true
		}
		for _, p := range privacyPolicy.OutboundSecurityRules {
			allowEgressAll = false
			portRange := fmt.Sprintf("%d", p.PortRangeMin)
			if p.PortRangeMax != 0 {
				portRange += fmt.Sprintf(":%d", p.PortRangeMax)
			}
			egressRule := vmlayer.FirewallRule{
				Protocol:     p.Protocol,
				PortRange:    portRange,
				RemoteCidr:   p.RemoteCidr,
				PortEndpoint: vmlayer.DestPort,
			}
			fwRules.EgressRules = append(fwRules.EgressRules, egressRule)
		}
	}
	if allowEgressAll {
		allowAllEgressRule := vmlayer.FirewallRule{
			RemoteCidr: "0.0.0.0/0",
		}
		fwRules.EgressRules = append(fwRules.EgressRules, allowAllEgressRule)
	}
	err = vmlayer.AddIptablesRules(ctx, client, secGrpName, &fwRules)
	if err != nil {
		return err
	}
	return vmlayer.AddDefaultIptablesRules(ctx, client, v.vmProperties.GetCloudletSecurityGroupName())
}

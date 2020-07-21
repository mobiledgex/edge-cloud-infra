package vsphere

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (v *VSpherePlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error {
	return vmlayer.AddIngressIptablesWhitelistRules(ctx, client, secGrpName, allowedCIDR, ports)
}

func (v *VSpherePlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "vsphere RemoveWhitelistSecurityRules not yet implemented")
	return nil
}

func (v *VSpherePlatform) PrepareRootLB(ctx context.Context, client ssh.Client, rootLBName string, secGrpName string) error {

	var fwRules vmlayer.FirewallRules
	//First create the global rules on this LB
	err := v.vmProperties.CreateCloudletFirewallRules(ctx, client)
	if err != nil {
		return err
	}

	// allow external IP network
	netCidr, err := v.GetExternalIpNetworkCidr(ctx)
	if err != nil {
		return err
	}
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
	fwRules.EgressRules = append(fwRules.IngressRules, sshEgress)

	err = vmlayer.AddIptablesWhitelistRules(ctx, client, v.vmProperties.GetCloudletSecurityGroupName(), &fwRules)
	if err != nil {
		return err
	}
	return vmlayer.AddDefaultRules(ctx, client)
}

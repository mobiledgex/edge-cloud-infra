package aws

import (
	"context"
	"fmt"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AWSPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, label, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules not supported")
	return nil
}

func (a *AWSPlatform) CreateSecurityGroupRule(ctx context.Context, groupId, protocol, portRange, allowedCIDR string) error {
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"authorize-security-group-ingress",
		"--group-id", groupId,
		"--cidr", allowedCIDR,
		"--protocol", protocol,
		"--port", portRange,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "authorize-security-group-ingress", "out", string(out), "err", err)
	if err != nil {
		if strings.Contains(string(out), RuleAlreadyExistsError) {
			log.SpanLog(ctx, log.DebugLevelInfra, "security rule already exists")
		} else {
			return fmt.Errorf("authorize-security-group-ingress failed: %s - %v", string(out), err)
		}
	}
	return nil
}

func (a *AWSPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, grpName, server, label, allowedCidr string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules", "grpName", grpName, "label", label, "allowedCidr", allowedCidr, "ports", ports)
	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return err
	}

	secGrpMap, err := a.GetSecurityGroups(ctx, vpc.VpcId)
	if err != nil {
		return err
	}
	sg, ok := secGrpMap[grpName]
	if !ok {
		return fmt.Errorf("Security group %s not found", grpName)
	}

	for _, p := range ports {
		log.SpanLog(ctx, log.DebugLevelInfra, "WhiteListing port", "port", p)
		portRange := fmt.Sprintf("%d", p.PublicPort)
		if p.EndPort != 0 {
			portRange = fmt.Sprintf("%d-%d", p.PublicPort, p.EndPort)
		}
		proto, err := edgeproto.L4ProtoStr(p.Proto)
		if err != nil {
			return err
		}
		err = a.CreateSecurityGroupRule(ctx, sg.GroupId, proto, portRange, allowedCidr)
		if err != nil {
			return err
		}
	}
	return nil
}

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type SecurityGroupAction string

const SecurityGroupRuleCreate SecurityGroupAction = "create"
const SecurityGroupRuleRevoke SecurityGroupAction = "revoke"

func (a *AWSPlatform) CreateSecurityGroupRule(ctx context.Context, groupId, protocol, portRange, allowedCIDR string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateSecurityGroupRule", "groupId", groupId, "portRange", portRange, "allowedCIDR", allowedCIDR)

	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"authorize-security-group-ingress",
		"--group-id", groupId,
		"--cidr", allowedCIDR,
		"--protocol", protocol,
		"--port", portRange,
		"--region", a.GetAwsRegion())
	if err != nil {
		if strings.Contains(string(out), RuleAlreadyExistsError) {
			log.SpanLog(ctx, log.DebugLevelInfra, "security rule already exists")
		} else {
			return fmt.Errorf("authorize-security-group-ingress failed: %s - %v", string(out), err)
		}
	}
	return nil
}

func (a *AWSPlatform) RevokeSecurityGroupRule(ctx context.Context, groupId, protocol, portRange, allowedCIDR string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RevokeSecurityGroupRule", "groupId", groupId, "portRange", portRange, "allowedCIDR", allowedCIDR)

	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"revoke-security-group-ingress",
		"--group-id", groupId,
		"--cidr", allowedCIDR,
		"--protocol", protocol,
		"--port", portRange,
		"--region", a.GetAwsRegion())
	if err != nil {
		if strings.Contains(string(out), RuleDoesNotExistError) {
			log.SpanLog(ctx, log.DebugLevelInfra, "security rule does not exist")
		} else {
			return fmt.Errorf("revoke-security-group-ingress failed: %s - %v", string(out), err)
		}
	}
	return nil
}

func (a *AWSPlatform) addOrDeleteSecurityRule(ctx context.Context, grpName, allowedCidr string, ports []dme.AppPort, action SecurityGroupAction) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "addOrDeleteSecurityRule", "grpName", grpName, "allowedCidr", allowedCidr, "ports", ports, "action", action)
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
		if action == SecurityGroupRuleCreate {
			err = a.CreateSecurityGroupRule(ctx, sg.GroupId, proto, portRange, allowedCidr)
		} else {
			err = a.RevokeSecurityGroupRule(ctx, sg.GroupId, proto, portRange, allowedCidr)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AWSPlatform) DeleteSecurityGroup(ctx context.Context, groupId, vpcId string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteSecurityGroup", "groupId", groupId, "vpcId", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"delete-security-group",
		"--region", a.GetAwsRegion(),
		"--group-id", groupId)
	if err != nil && !strings.Contains(err.Error(), SecGrpDoesNotExistError) {
		return fmt.Errorf("Error in delete-security-group: %s - %v", string(out), err)
	}
	return nil
}

func (a *AWSPlatform) CreateSecurityGroup(ctx context.Context, name, vpcId, vmGroupName string) (*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateSecurityGroup", "name", name, "vmGroupName", vmGroupName, "vpcId", vpcId)
	tagspec := fmt.Sprintf("ResourceType=security-group,Tags=[{Key=%s,Value=%s},{Key=%s,Value=%s}]", NameTag, name, VMGroupNameTag, vmGroupName)

	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-security-group",
		"--region", a.GetAwsRegion(),
		"--group-name", name,
		"--vpc-id", vpcId,
		"--description", vmGroupName,
		"--tag-specifications", tagspec)
	if err != nil {
		return nil, fmt.Errorf("Error in create-security-group: %s - %v", string(out), err)
	}
	var sg AwsEc2SecGrp
	err = json.Unmarshal(out, &sg)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-security-group unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	sg.GroupName = name
	sg.VpcId = vpcId
	return &sg, nil
}

func (a *AWSPlatform) GetSecurityGroup(ctx context.Context, name string, vpcId string) (*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecurityGroup", "name", name, "vpcId", vpcId)

	grpMap, err := a.GetSecurityGroups(ctx, vpcId)
	if err != nil {
		return nil, err
	}
	grp, ok := grpMap[name]
	if !ok {
		return nil, fmt.Errorf(SecGrpDoesNotExistError)
	}
	return grp, nil
}

func (a *AWSPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, server, label, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules", "secGrpName", secGrpName, "label", label, "allowedCIDR", allowedCIDR, "ports", ports)
	return a.addOrDeleteSecurityRule(ctx, secGrpName, allowedCIDR, ports, SecurityGroupRuleCreate)
}

func (a *AWSPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, label, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules", "secGrpName", secGrpName, "allowedCIDR", allowedCIDR, "ports", ports)
	return a.addOrDeleteSecurityRule(ctx, secGrpName, allowedCIDR, ports, SecurityGroupRuleRevoke)
}

func (a *AWSPlatform) GetSecurityGroups(ctx context.Context, vpcId string) (map[string]*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecurityGroups", "vpcId", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-security-groups",
		"--region", a.GetAwsRegion())
	if err != nil {
		return nil, fmt.Errorf("error in describe-security-groups: %s - %v", string(out), err)
	}
	var sgList AwsEc2SecGrpList
	err = json.Unmarshal(out, &sgList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-security-groups unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	var sgMap = make(map[string]*AwsEc2SecGrp)
	for i, sg := range sgList.SecurityGroups {
		if sg.VpcId == vpcId {
			sgMap[sg.GroupName] = &sgList.SecurityGroups[i]
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "found security groups", "sgMap", sgMap)
	return sgMap, nil
}

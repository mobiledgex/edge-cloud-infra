package vmlayer

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type FirewallRules struct {
	EgressRules  []FirewallRule
	IngressRules []FirewallRule
}

// PortSourceOrDestChoice indicates whether the port(s) are the source or destination ports
type PortSourceOrDestChoice string

const SourcePort PortSourceOrDestChoice = "sport"
const DestPort PortSourceOrDestChoice = "dport"

type FirewallRule struct {
	Protocol     string
	RemoteCidr   string
	PortRange    string
	InterfaceIn  string
	InterfaceOut string
	PortEndpoint PortSourceOrDestChoice
	Conntrack    string
}

// doIptablesCommand runs an iptables add or delete conditionally based on whether the entry already exists or not
func doIptablesCommand(ctx context.Context, client ssh.Client, rule string, ruleExists bool, action *InterfaceActionsOp) error {
	runCommand := false
	if ruleExists {
		if action.deleteIptables {
			log.SpanLog(ctx, log.DebugLevelInfra, "deleting existing iptables rule", "rule", rule)
			runCommand = true
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "do not re-add existing iptables rule", "rule", rule)
		}
	} else {
		if action.createIptables {
			log.SpanLog(ctx, log.DebugLevelInfra, "adding new iptables rule", "rule", rule)
			runCommand = true
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "do not delete nonexistent iptables rule", "rule", rule)
		}
	}

	if runCommand {
		log.SpanLog(ctx, log.DebugLevelInfra, "running iptables command", "rule", rule)
		cmd := fmt.Sprintf("sudo iptables %s", rule)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("unable to modify iptables rule: %s, %s - %v", rule, out, err)
		}
	}
	return nil
}

//	parseFirewallRules parses rules in the format:
// Value: "protocol=tcp,portrange=1:65535,remotecidr=0.0.0.0/0;protocol=udp,portrange=1:65535,remotecidr=0.0.0.0/0;protocol=icmp,remotecidr=0.0.0.0/0",
func parseFirewallRules(ctx context.Context, ruleString string) ([]FirewallRule, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "parseFirewallRules", "ruleString", ruleString)

	if ruleString == "" {
		return nil, nil
	}
	var firewallRules []FirewallRule
	rules := strings.Split(ruleString, ";")
	portRangeReg := regexp.MustCompile("\\d+(:\\d+)?")
	for _, rs := range rules {
		log.SpanLog(ctx, log.DebugLevelInfra, "parseFirewallRules", "rule", rs)
		var firewallRule FirewallRule
		ss := strings.Split(rs, ",")
		for _, s := range ss {
			s = strings.ToLower(s)
			kvs := strings.Split(s, "=")
			if len(kvs) != 2 {
				return nil, fmt.Errorf("unable to parse firewall rule, not in key=val format: %s", s)
			}
			key := kvs[0]
			val := kvs[1]
			switch key {
			case "protocol":
				firewallRule.Protocol = val
			case "remotecidr":
				_, _, err := net.ParseCIDR(val)
				if err != nil {
					return nil, fmt.Errorf("unable to parse firewall rule, bad cidr: %s", val)
				}
				firewallRule.RemoteCidr = val
			case "portrange":
				match := portRangeReg.MatchString(val)
				if !match {
					return nil, fmt.Errorf("unable to parse firewall rule, bad port range: %s", val)
				}
				firewallRule.PortRange = val
			default:
				return nil, fmt.Errorf("unable to parse firewall rule, bad key: %s", key)
			}
		}
		if firewallRule.RemoteCidr == "" {
			return nil, fmt.Errorf("invalid firewall rule, missing cidr")
		}
		if firewallRule.Protocol == "udp" {
			// add udp as both source and dest
			firewallRule.PortEndpoint = SourcePort
			firewallRules = append(firewallRules, firewallRule)
		}
		firewallRule.PortEndpoint = DestPort
		firewallRules = append(firewallRules, firewallRule)
	}

	return firewallRules, nil
}

// CreateCloudletFirewallRules adds cloudlet-wide egress rules based on properties
func (v *VMProperties) CreateCloudletFirewallRules(ctx context.Context, client ssh.Client) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudletFirewallRules")

	var firewallRules FirewallRules
	var err error
	if val, ok := v.CommonPf.Properties["MEX_CLOUDLET_FIREWALL_WHITELIST_EGRESS"]; ok {
		firewallRules.EgressRules, err = parseFirewallRules(ctx, val.Value)
		if err != nil {
			return err
		}
	}
	if val, ok := v.CommonPf.Properties["MEX_CLOUDLET_FIREWALL_WHITELIST_INGRESS"]; ok {
		firewallRules.IngressRules, err = parseFirewallRules(ctx, val.Value)
		if err != nil {
			return err
		}
	}
	return AddIptablesRules(ctx, client, v.GetCloudletSecurityGroupName(), &firewallRules)
}

// GetIpTablesEntryForRule gets the iptables string for the rule
func GetIpTablesEntryForRule(ctx context.Context, direction string, secGrp string, rule *FirewallRule) string {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIpTablesEntryForRule", "rule", rule)
	dirStr := "INPUT"
	cidrStr := ""
	if direction == "egress" {
		dirStr = "OUTPUT"
		if rule.RemoteCidr != "0.0.0.0/0" {
			cidrStr = "-d " + rule.RemoteCidr
		}
	} else {
		dirStr = "INPUT"
		if rule.RemoteCidr != "0.0.0.0/0" {
			cidrStr = "-s " + rule.RemoteCidr
		}
	}
	portStr := ""
	if rule.PortRange != "" {
		portStr = "--" + string(rule.PortEndpoint) + " " + rule.PortRange
	}
	icmpType := ""
	if rule.Protocol == "icmp" {
		icmpType = " --icmp-type any"
	}
	protostr := ""
	if rule.Protocol != "" {
		protostr = fmt.Sprintf("-p %s -m %s", rule.Protocol, rule.Protocol)
	}
	ifstr := ""
	if rule.InterfaceIn != "" {
		ifstr = "-i " + rule.InterfaceIn
	} else if rule.InterfaceOut != "" {
		ifstr = "-o " + rule.InterfaceOut
	}
	conntrackStr := ""
	if rule.Conntrack != "" {
		conntrackStr = "-m conntrack --ctstate " + rule.Conntrack
	}
	rulestr := fmt.Sprintf("%s %s %s %s %s %s %s -m comment --comment \"secgrp %s\" -j ACCEPT", dirStr, ifstr, conntrackStr, cidrStr, protostr, icmpType, portStr, secGrp)

	// remove double spaces
	rulestr = strings.Join(strings.Fields(rulestr), " ")
	return rulestr
}

// getCurrentIptableRulesForSecGrp retrieves the current rules matching the group
func getCurrentIptableRulesForSecGrp(ctx context.Context, client ssh.Client, secGrp string) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "getCurrentIptableRulesForSecGrp", "secGrp", secGrp)

	rules := make(map[string]string)

	cmd := fmt.Sprintf("sudo iptables-save")
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to run iptables-save to get current rules: %s - %v", out, err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"secgrp "+secGrp+"\"") && strings.HasPrefix(line, "-A") {
			rules[line] = line
		}
	}
	return rules, nil
}

// AddIptablesRule adds a rule
func AddIptablesRule(ctx context.Context, client ssh.Client, direction string, secGrp string, rule *FirewallRule, currentRules map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddIptablesRule", "direction", direction, "secGrp", secGrp, "rule", rule)
	entry := GetIpTablesEntryForRule(ctx, direction, secGrp, rule)
	addCmd := "-A " + entry
	_, exists := currentRules[addCmd]
	action := InterfaceActionsOp{createIptables: true}
	return doIptablesCommand(ctx, client, addCmd, exists, &action)
}

// RemoveIptablesRule removes a rule
func RemoveIptablesRule(ctx context.Context, client ssh.Client, direction string, secGrp string, rule *FirewallRule, currentRules map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveIptablesRule", "direction", direction, "secGrp", secGrp, "rule", rule)
	entry := GetIpTablesEntryForRule(ctx, direction, secGrp, rule)
	addCmd := "-A " + entry
	delCmd := "-D " + entry
	_, exists := currentRules[addCmd]
	action := InterfaceActionsOp{deleteIptables: true}
	return doIptablesCommand(ctx, client, delCmd, exists, &action)
}

// AddIptablesRules adds a set of rules
func AddIptablesRules(ctx context.Context, client ssh.Client, secGrp string, rules *FirewallRules) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddIptablesRules", "rules", rules)
	currentRules, err := getCurrentIptableRulesForSecGrp(ctx, client, secGrp)
	if err != nil {
		return err
	}
	for _, erule := range rules.EgressRules {
		err := AddIptablesRule(ctx, client, "egress", secGrp, &erule, currentRules)
		if err != nil {
			return err
		}
	}
	for _, irule := range rules.IngressRules {
		err := AddIptablesRule(ctx, client, "ingress", secGrp, &irule, currentRules)
		if err != nil {
			return err
		}
	}
	return PersistIptablesRules(ctx, client)
}

// DeleteIptablesRules deletes a set of rules
func DeleteIptablesRules(ctx context.Context, client ssh.Client, secGrp string, rules *FirewallRules) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteIptablesRules", "rules", rules)
	currentRules, err := getCurrentIptableRulesForSecGrp(ctx, client, secGrp)
	if err != nil {
		return err
	}
	for _, erule := range rules.EgressRules {
		err := RemoveIptablesRule(ctx, client, "egress", secGrp, &erule, currentRules)
		if err != nil {
			return err
		}
	}
	for _, irule := range rules.IngressRules {
		err := RemoveIptablesRule(ctx, client, "ingress", secGrp, &irule, currentRules)
		if err != nil {
			return err
		}
	}
	// make this an option?
	return PersistIptablesRules(ctx, client)
}

// AddDefaultIptablesRules adds the default set of rules which are always needed
func AddDefaultIptablesRules(ctx context.Context, client ssh.Client, secGrp string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddDefaultIptablesRules")

	var rules FirewallRules
	// local loopback traffic is open
	loopInRule := FirewallRule{
		RemoteCidr:  "0.0.0.0/0",
		InterfaceIn: "lo",
	}
	rules.IngressRules = append(rules.IngressRules, loopInRule)
	loopOutRule := FirewallRule{
		RemoteCidr:   "0.0.0.0/0",
		InterfaceOut: "lo",
	}
	rules.EgressRules = append(rules.EgressRules, loopOutRule)

	// allow established sessions
	conntrackInRule := FirewallRule{
		RemoteCidr: "0.0.0.0/0",
		Conntrack:  "ESTABLISHED,RELATED",
	}
	rules.IngressRules = append(rules.IngressRules, conntrackInRule)
	conntrackOutRule := FirewallRule{
		RemoteCidr: "0.0.0.0/0",
		Conntrack:  "ESTABLISHED,RELATED",
	}
	rules.EgressRules = append(rules.EgressRules, conntrackOutRule)

	err := AddIptablesRules(ctx, client, secGrp, &rules)
	if err != nil {
		return err
	}

	// anything not matching the chain is dropped.   These will not create
	// duplicate entries if done multiple times
	dropInputPolicy := "-P INPUT DROP"
	dropOutputPolicy := "-P OUTPUT DROP"
	action := InterfaceActionsOp{createIptables: true}
	err = doIptablesCommand(ctx, client, dropInputPolicy, false, &action)
	if err != nil {
		return err
	}
	return doIptablesCommand(ctx, client, dropOutputPolicy, false, &action)
}

// getFirewallRulesFromAppPorts accepts a CIDR and a set of AppPorts and converts to a set of rules
func getFirewallRulesFromAppPorts(ctx context.Context, cidr string, ports []dme.AppPort) (*FirewallRules, error) {
	var fwRules FirewallRules
	for _, p := range ports {
		portStr := fmt.Sprintf("%d", p.PublicPort)
		if p.EndPort != 0 {
			portStr += fmt.Sprintf(":%d", p.EndPort)
		}
		proto, err := edgeproto.L4ProtoStr(p.Proto)
		if err != nil {
			return nil, err
		}
		fwRuleDest := FirewallRule{
			Protocol:     proto,
			RemoteCidr:   cidr,
			PortRange:    portStr,
			PortEndpoint: DestPort,
		}
		fwRules.IngressRules = append(fwRules.IngressRules, fwRuleDest)
	}
	return &fwRules, nil
}

// AddIngressIptablesRules adds rules using a CIDR and AppPorts as input
func AddIngressIptablesRules(ctx context.Context, client ssh.Client, secGrpName string, cidr string, ports []dme.AppPort) error {
	fwRules, err := getFirewallRulesFromAppPorts(ctx, cidr, ports)
	if err != nil {
		return err
	}
	return AddIptablesRules(ctx, client, secGrpName, fwRules)
}

// RemoveIngressIptablesRules removes rules using a CIDR and AppPorts as input
func RemoveIngressIptablesRules(ctx context.Context, client ssh.Client, secGrpName string, cidr string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveIngressIptablesRules", "secGrp", secGrpName)

	fwRules, err := getFirewallRulesFromAppPorts(ctx, cidr, ports)
	if err != nil {
		return err
	}
	return DeleteIptablesRules(ctx, client, secGrpName, fwRules)
}

func PersistIptablesRules(ctx context.Context, client ssh.Client) error {
	cmd := fmt.Sprintf("sudo bash -c 'iptables-save > /etc/iptables/rules.v4'")
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("unable to run iptables-save to persistent rules file: %s - %v", out, err)
	}
	return nil
}

// setupForwardingIptables creates iptables rules to allow the cluster nodes to use the LB as a
// router for internet access
func setupForwardingIptables(ctx context.Context, client ssh.Client, externalIfname, internalIfname string, action *InterfaceActionsOp) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "setupForwardingIptables", "externalIfname", externalIfname, "internalIfname", internalIfname, "action", fmt.Sprintf("%+v", action))
	// get current iptables
	cmd := fmt.Sprintf("sudo iptables-save|grep -e POSTROUTING -e FORWARD")
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("unable to run iptables-save: %s - %v", out, err)
	}
	// add or remove rules based on the action
	option := "-A"
	if action.deleteIptables {
		option = "-D"
	}
	// we are looking only for the FORWARD or postrouting entries
	masqueradeRuleMatch := fmt.Sprintf("POSTROUTING -o %s -j MASQUERADE", externalIfname)
	masqueradeRule := fmt.Sprintf("-t nat %s %s", option, masqueradeRuleMatch)
	forwardExternalRuleMatch := fmt.Sprintf("FORWARD -i %s -o %s -m state --state RELATED,ESTABLISHED -j ACCEPT", externalIfname, internalIfname)
	forwardExternalRule := fmt.Sprintf("%s %s", option, forwardExternalRuleMatch)
	forwardInternalRuleMatch := fmt.Sprintf("FORWARD -i %s -j ACCEPT", internalIfname)
	forwardInternalRule := fmt.Sprintf("%s %s", option, forwardInternalRuleMatch)

	masqueradeRuleExists := false
	forwardExternalRuleExists := false
	forwardInternalRuleExists := false

	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.Contains(l, masqueradeRuleMatch) {
			masqueradeRuleExists = true
		}
		if strings.Contains(l, forwardExternalRuleMatch) {
			forwardExternalRuleExists = true
		}
		if strings.Contains(l, forwardInternalRuleMatch) {
			forwardInternalRuleExists = true
		}
	}
	if action.createIptables {
		// this rule is never deleted because it applies to all subnets.   Multiple adds will
		// not create duplicates
		err = doIptablesCommand(ctx, client, masqueradeRule, masqueradeRuleExists, action)
		if err != nil {
			return err
		}
	}
	err = doIptablesCommand(ctx, client, forwardExternalRule, forwardExternalRuleExists, action)
	if err != nil {
		return err
	}
	err = doIptablesCommand(ctx, client, forwardInternalRule, forwardInternalRuleExists, action)
	if err != nil {
		return err
	}
	//now persist the rules
	err = PersistIptablesRules(ctx, client)
	if err != nil {
		return err
	}
	return nil
}

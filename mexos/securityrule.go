package mexos

import (
	"fmt"
	"strings"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
)

func addSecurityRules(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	rootLBIPaddr, err := GetServerIPAddr(mf, mf.Values.Network.External, rootLB.Name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot get rootlb IP address", "error", err)
		return fmt.Errorf("cannot deploy kubernetes app, cannot get rootlb IP")
	}
	sr := GetMEXSecurityRule(mf)
	allowedClientCIDR := GetAllowedClientCIDR()
	for _, port := range mf.Spec.Ports {
		for _, sec := range []struct {
			addr string
			port int
		}{
			{rootLBIPaddr + "/32", port.PublicPort},
			{kp.ipaddr + "/32", port.PublicPort},
			{allowedClientCIDR, port.PublicPort},
			{rootLBIPaddr + "/32", port.InternalPort},
			{kp.ipaddr + "/32", port.InternalPort},
			{allowedClientCIDR, port.InternalPort},
		} {
			go func(addr string, port int, proto string) {
				err := AddSecurityRuleCIDR(mf, addr, strings.ToLower(proto), sr, port)
				if err != nil {
					log.DebugLog(log.DebugLevelMexos, "warning, error while adding security rule", "cidr", addr, "securityrule", sr, "port", port, "proto", proto)
				}
			}(sec.addr, sec.port, port.Proto)
		}
	}
	if len(mf.Spec.Ports) > 0 {
		err = AddNginxProxy(mf, rootLB.Name, mf.Metadata.Name, kp.ipaddr, mf.Spec.Ports)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot add nginx proxy", "name", mf.Metadata.Name, "ports", mf.Spec.Ports)
			return err
		}
	}
	log.DebugLog(log.DebugLevelMexos, "added nginx proxy", "name", mf.Metadata.Name, "ports", mf.Spec.Ports)
	return nil
}

func deleteSecurityRules(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	log.DebugLog(log.DebugLevelMexos, "delete spec ports", "ports", mf.Spec.Ports)
	err := DeleteNginxProxy(mf, rootLB.Name, mf.Metadata.Name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot delete nginx proxy", "name", mf.Metadata.Name, "rootlb", rootLB.Name, "error", err)
	}

	// TODO - implement the clean up of security rules
	return nil
}

func AddSecurityRuleCIDR(mf *Manifest, cidr string, proto string, name string, port int) error {
	portStr := fmt.Sprintf("%d", port)
	out, err := sh.Command("openstack", "security", "group", "rule", "create", "--remote-ip", cidr, "--proto", proto, "--dst-port", portStr, "--ingress", name).Output()
	if err != nil {
		return fmt.Errorf("can't add security group rule for port %d to %s,%s,%v", port, name, out, err)
	}
	return nil
}

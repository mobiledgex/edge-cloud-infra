package mexos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/nginx"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func AddSecurityRules(ctx context.Context, groupName string, ports []dme.AppPort, serverName string) error {
	allowedClientCIDR := GetAllowedClientCIDR()
	for _, port := range ports {
		//todo: distinguish already-exists errors from others
		portString := fmt.Sprintf("%d", port.PublicPort)
		proto, err := edgeproto.L4ProtoStr(port.Proto)
		if err != nil {
			return err
		}
		if err := AddSecurityRuleCIDRWithRetry(ctx, allowedClientCIDR, proto, groupName, portString, serverName); err != nil {
			return err
		}
	}
	return nil
}

func DeleteProxySecurityRules(ctx context.Context, client pc.PlatformClient, ipaddr string, appName string, group string) error {

	log.SpanLog(ctx, log.DebugLevelMexos, "delete proxy rules", "name", appName)
	err := nginx.DeleteNginxProxy(client, appName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete nginx proxy", "name", appName, "error", err)
	}
	if err := DeleteSecurityRule(ctx, ipaddr, group); err != nil {
		return err
	}
	// TODO - implement the clean up of security rules
	return nil
}

// AddSecurityRuleCIDRWithRetry calls AddSecurityRuleCIDR, and then will retry if that fails because the group does not exist.  This can happen during
// the transition between cloudlet-wide security groups and the newer per-cluster groups.  Eventually this function can be removed once all LBs have been
// updated with the per-cluster group
func AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	err := AddSecurityRuleCIDR(ctx, cidr, proto, group, port)
	if err != nil {
		if strings.Contains(err.Error(), "No SecurityGroup found") {
			// it is possible this RootLB was created before the change to per-LB security groups.  Create the group separately
			log.SpanLog(ctx, log.DebugLevelMexos, "security group does not exist, creating it", "groupName", group)

			// LB can have multiple ports attached.  We need to assign this SG to the external network port only
			ports, err := ListPortsServerNetwork(ctx, serverName, GetCloudletExternalNetwork())
			if err != nil {
				return err
			}
			if len(ports) != 1 {
				return fmt.Errorf("Could find external network ports to add security group")
			}
			err = CreateSecurityGroup(ctx, group)
			if err != nil {
				return err
			}
			err = AddSecurityGroupToPort(ctx, ports[0].ID, group)
			if err != nil {
				return err
			}
			// try again to add the rule
			return AddSecurityRuleCIDR(ctx, cidr, proto, group, port)
		}
	}
	return err
}

func AddSecurityRuleCIDR(ctx context.Context, cidr string, proto string, groupName string, port string) error {

	out, err := TimedOpenStackCommand(ctx, "openstack", "security", "group", "rule", "create", "--remote-ip", cidr, "--proto", proto, "--dst-port", port, "--ingress", groupName)
	if err != nil {
		if strings.Contains(string(out), "Security group rule already exists") {
			log.SpanLog(ctx, log.DebugLevelMexos, "security group rule already exists, proceeding")
		} else {
			return fmt.Errorf("can't add security group rule for port %s to %s,%s,%v", port, groupName, string(out), err)
		}
	}
	return nil
}

type SecurityRule struct {
	IPRange   string `json:"IP Range"`
	PortRange string `json:"Port Range"`
	SGID      string `json:"Security Group"`
	ID        string `json:"ID"`
	Proto     string `json:"IP Protocol"`
}

func DeleteSecurityRule(ctx context.Context, sip string, group string) error {
	sr := []SecurityRule{}
	dat, err := TimedOpenStackCommand(ctx, "openstack", "security", "group", "rule", "list", group, "-f", "json")
	if err != nil {
		return fmt.Errorf("cannot get list of security group rules, %v", err)
	}
	if err := json.Unmarshal(dat, &sr); err != nil {
		return fmt.Errorf("cannot unmarshal security group rule list, %v", err)
	}
	for _, s := range sr {
		log.SpanLog(ctx, log.DebugLevelMexos, "security group rule found", "rule", s)

		if strings.HasSuffix(s.IPRange, "/32") {
			adr := strings.Replace(s.IPRange, "/32", "", -1)
			if adr == sip {
				_, err := TimedOpenStackCommand(ctx, "openstack", "security", "group", "rule", "delete", s.ID)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "warning, cannot delete security rule", "id", s.ID, "error", err)
				}
			}
		}
	}
	return nil
}

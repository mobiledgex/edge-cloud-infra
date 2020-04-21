package openstack

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// CloudletSecurityGroupIDMap is a cache of cloudlet to security group id
var CloudletSecurityGroupIDMap = make(map[string]string)

var cloudetSecurityGroupIDLock sync.Mutex

func getCachedCloudletSecgrpID(ctx context.Context, keyString string) string {
	cloudetSecurityGroupIDLock.Lock()
	defer cloudetSecurityGroupIDLock.Unlock()
	groupID, ok := CloudletSecurityGroupIDMap[keyString]
	if !ok {
		return ""
	}
	return groupID
}

func setCachedCloudletSecgrpID(ctx context.Context, keyString, groupID string) {
	cloudetSecurityGroupIDLock.Lock()
	defer cloudetSecurityGroupIDLock.Unlock()
	CloudletSecurityGroupIDMap[keyString] = groupID
}

// GetCloudletSecurityGroupID gets the group ID for the default cloudlet-wide group for our project.  It handles
// duplicate names.  This group should not be used for application traffic, it is for management/OAM/CRM access.
func (s *OpenstackPlatform) GetCloudletSecurityGroupID(ctx context.Context, cloudletKey *edgeproto.CloudletKey) (string, error) {
	groupName := s.vmPlatform.GetCloudletSecurityGroupName()
	keyString := cloudletKey.GetKeyString()

	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletSecurityGroupID", "groupName", groupName, "keyString", keyString)

	groupID := getCachedCloudletSecgrpID(ctx, keyString)
	if groupID != "" {
		//cached
		log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletSecurityGroupID using existing value", "groupID", groupID)
		return groupID, nil
	}

	projectName := s.GetCloudletProjectName()
	if projectName == "" {
		return "", fmt.Errorf("No OpenStack project name, cannot get project security group")
	}
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return "", err
	}
	for _, p := range projects {
		if p.Name == projectName {
			groupID, err = s.GetSecurityGroupIDForProject(ctx, groupName, p.ID)
			if err != nil {
				return "", err
			}
			setCachedCloudletSecgrpID(ctx, keyString, groupID)
			log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletSecurityGroupID using new value", "groupID", groupID)
			return groupID, nil
		}
	}
	return "", fmt.Errorf("Unable to find cloudlet security group for project: %s", projectName)
}

func (s *OpenstackPlatform) AddSecurityRules(ctx context.Context, groupName string, ports []dme.AppPort, serverName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddSecurityRules", "ports", ports)
	allowedClientCIDR := infracommon.GetAllowedClientCIDR()
	for _, port := range ports {
		//todo: distinguish already-exists errors from others
		portString := fmt.Sprintf("%d", port.PublicPort)
		if port.EndPort != 0 {
			portString = fmt.Sprintf("%d:%d", port.PublicPort, port.EndPort)
		}
		proto, err := edgeproto.L4ProtoStr(port.Proto)
		if err != nil {
			return err
		}
		if err := s.AddSecurityRuleCIDRWithRetry(ctx, allowedClientCIDR, proto, groupName, portString, serverName); err != nil {
			return err
		}
	}
	return nil
}

// AddSecurityRuleCIDRWithRetry calls AddSecurityRuleCIDR, and then will retry if that fails because the group does not exist.  This can happen during
// the transition between cloudlet-wide security groups and the newer per-LB groups.  Eventually this function can be removed once all LBs have been
// updated with the per-cluster group
func (s *OpenstackPlatform) AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	err := s.AddSecurityRuleCIDR(ctx, cidr, proto, group, port)
	if err != nil {
		if strings.Contains(err.Error(), "No SecurityGroup found") {
			// it is possible this RootLB was created before the change to per-LB security groups.  Create the group separately
			log.SpanLog(ctx, log.DebugLevelInfra, "security group does not exist, creating it", "groupName", group)

			// LB can have multiple ports attached.  We need to assign this SG to the external network port only
			ports, err := s.ListPortsServerNetwork(ctx, serverName, s.vmPlatform.GetCloudletExternalNetwork())
			if err != nil {
				return err
			}
			if len(ports) != 1 {
				return fmt.Errorf("Could find external network ports to add security group")
			}
			err = s.CreateSecurityGroup(ctx, group)
			if err != nil {
				return err
			}
			err = s.AddSecurityGroupToPort(ctx, ports[0].ID, group)
			if err != nil {
				return err
			}
			// try again to add the rule
			return s.AddSecurityRuleCIDR(ctx, cidr, proto, group, port)
		}
	}
	return err
}

func (s *OpenstackPlatform) RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules", "secGrpName", secGrpName, "ports", ports)

	allowedClientCIDR := infracommon.GetAllowedClientCIDR()
	rules, err := s.ListSecurityGroupRules(ctx, secGrpName)
	if err != nil {
		return err
	}
	for _, port := range ports {
		portString := fmt.Sprintf("%d:%d", port.PublicPort, port.PublicPort)
		if port.EndPort != 0 {
			portString = fmt.Sprintf("%d:%d", port.PublicPort, port.EndPort)
		}
		proto, err := edgeproto.L4ProtoStr(port.Proto)
		if err != nil {
			return err
		}
		for _, r := range rules {
			if r.PortRange == portString && r.Protocol == proto && r.IPRange == allowedClientCIDR {
				if err := s.DeleteSecurityGroupRule(ctx, r.ID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *OpenstackPlatform) WhitelistSecurityRules(ctx context.Context, grpName, serverName, allowedCidr string, ports []dme.AppPort) error {
	// open the firewall for internal traffic
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules", "grpName", grpName, "allowedCidr", allowedCidr, "ports", ports)

	for _, p := range ports {
		portStr := fmt.Sprintf("%d", p.PublicPort)
		proto, err := edgeproto.L4ProtoStr(p.Proto)
		if err != nil {
			return err
		}
		if err := s.AddSecurityRuleCIDRWithRetry(ctx, allowedCidr, proto, grpName, portStr, serverName); err != nil {
			return err
		}
	}
	return nil
}

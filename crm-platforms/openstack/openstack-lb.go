package openstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/log"

	ssh "github.com/mobiledgex/golang-ssh"
)

func (c *OpenstackPlatform) NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Adding route to reach internal networks", "rootLBName", rootLBName)

	ni, err := infracommon.ParseNetSpec(ctx, c.commonPf.GetCloudletNetworkScheme())
	if err != nil {
		return err
	}
	if ni.FloatingIPNet != "" {
		// For now we do nothing when we have a floating IP because it means we are using the
		// openstack router to get everywhere anyway.
		log.SpanLog(ctx, log.DebugLevelMexos, "No route changes needed due to floating IP")
		return nil
	}
	if rootLBName == "" {
		return fmt.Errorf("empty rootLB")
	}

	//TODO: this may not be necessary, as instead we can allow group to remote group rules
	// add to the /16 range for all the possible subnets
	subnet := fmt.Sprintf("%s.%s.0.0/16", ni.Octets[0], ni.Octets[1])
	subnetNomask := fmt.Sprintf("%s.%s.0.0", ni.Octets[0], ni.Octets[1])
	mask := "255.255.0.0"

	rtr := c.GetCloudletExternalRouter()
	gatewayIP := ni.RouterGatewayIP
	if gatewayIP == "" && rtr != infracommon.NoConfigExternalRouter && rtr != infracommon.NoExternalRouter {
		rd, err := c.GetRouterDetail(ctx, c.GetCloudletExternalRouter())
		if err != nil {
			return err
		}
		gw, err := GetRouterDetailExternalGateway(rd)
		if err != nil {
			return err
		}
		fip := gw.ExternalFixedIPs
		log.SpanLog(ctx, log.DebugLevelMexos, "external fixed ips", "ips", fip)

		if len(fip) != 1 {
			return fmt.Errorf("Unexpected fixed ips for mex router %v", fip)
		}
		gatewayIP = fip[0].IPAddress
	}
	if gatewayIP != "" {
		cmd := fmt.Sprintf("sudo ip route add %s via %s", subnet, gatewayIP)
		if err != nil {
			return err
		}

		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "RTNETLINK") && strings.Contains(out, " exists") {
				log.SpanLog(ctx, log.DebugLevelMexos, "warning, can't add existing route to rootLB", "cmd", cmd, "out", out, "error", err)
			} else {
				return fmt.Errorf("can't add route to rootlb, %s, %s, %v", cmd, out, err)
			}
		}

		// make the route persist by adding the following line if not already present via grep.
		routeAddLine := fmt.Sprintf("up route add -net %s netmask %s gw %s", subnetNomask, mask, gatewayIP)
		interfacesFile := infracommon.GetCloudletNetworkIfaceFile()
		cmd = fmt.Sprintf("grep -l '%s' %s", routeAddLine, interfacesFile)
		out, err = client.Output(cmd)
		if err != nil {
			// grep failed so not there already
			log.SpanLog(ctx, log.DebugLevelMexos, "adding route to interfaces file", "route", routeAddLine, "file", interfacesFile)
			cmd = fmt.Sprintf("echo '%s'|sudo tee -a %s", routeAddLine, interfacesFile)
			out, err = client.Output(cmd)
			if err != nil {
				return fmt.Errorf("can't add route to interfaces file: %v", err)
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "route already present in interfaces file")
		}
	}
	return nil
}

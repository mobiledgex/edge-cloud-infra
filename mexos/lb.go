package mexos

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/log"
)

// LBAddRouteAndSecRules adds an external route and sec rules
func LBAddRouteAndSecRules(ctx context.Context, client pc.PlatformClient, rootLBName string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Adding route to reach internal networks", "rootLBName", rootLBName)

	ni, err := ParseNetSpec(ctx, GetCloudletNetworkScheme())
	if err != nil {
		return err
	}
	if ni.FloatingIPNet != "" {
		// For now we do nothing when we have a floating IP because it means we are using the
		// openstack router to get everywhere anyway.
		log.SpanLog(ctx, log.DebugLevelMexos, "No route changes needed due to floating IP")
		return nil
	}
	if ni.NetworkType == NetworkTypeVLAN {
		log.SpanLog(ctx, log.DebugLevelMexos, "No route changes needed for VLAN networks")
		// TODO: iptables stuff
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

	rtr := GetCloudletExternalRouter()
	gatewayIP := ni.RouterGatewayIP
	if gatewayIP == "" && rtr != NoConfigExternalRouter && rtr != NoExternalRouter {
		rd, err := GetRouterDetail(ctx, GetCloudletExternalRouter())
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
		interfacesFile := "/etc/network/interfaces.d/50-cloud-init.cfg"
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

	// open the firewall for internal traffic
	// note that LB security rules will currently be added redundantly for each rootLB because they
	// all use the same sec grp.  However, this will eventually change
	groupName := GetCloudletSecurityGroup()

	if err := AddSecurityRuleCIDR(ctx, subnet, "tcp", groupName, "1:65535"); err != nil {
		// this error is nonfatal because it may already exist
		log.SpanLog(ctx, log.DebugLevelMexos, "notice, cannot add security rule", "error", err, "cidr", subnet)
	}
	allowedClientCIDR := GetAllowedClientCIDR()
	for _, p := range rootLBPorts {
		portString := fmt.Sprintf("%d", p)
		if err := AddSecurityRuleCIDR(ctx, allowedClientCIDR, "tcp", groupName, portString); err != nil {
			return err
		}
	}

	return nil
}

// AttachAndEnableRootLBInterface attaches the interface and enables it in the OS
func AttachAndEnableRootLBInterface(ctx context.Context, client pc.PlatformClient, rootLBName string, portName, ipaddr string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "AttachAndEnableRootLBInterface", "rootLBName", rootLBName, "portName", portName)

	portDetails, err := GetPortDetails(ctx, portName)
	if err != nil {
		return err
	}

	err = AttachPortToServer(ctx, rootLBName, portName)
	if err != nil {
		return err
	}

	// Once the attach is done, it does not really need to be detached if something fails after here because
	// the port will be deleted.

	// list all the interfaces
	cmd := fmt.Sprintf("sudo ifconfig -a")
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("unable to run ifconfig: %s - %v", out, err)
	}
	//                ifname        encap              mac
	matchPattern := "(\\w+)\\s+Link \\S+\\s+HWaddr\\s+(\\S+)"
	reg, err := regexp.Compile(matchPattern)
	if err != nil {
		// this is a bug if the regex does not compile
		log.SpanLog(ctx, log.DebugLevelMexos, "failed to compile regex", "pattern", matchPattern)
		return fmt.Errorf("Internal Error compiling regex for interface")
	}

	//find the interface matching our mac
	ifName := ""
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if reg.MatchString(l) {
			matches := reg.FindStringSubmatch(l)
			ifn := matches[1]
			mac := matches[2]
			log.SpanLog(ctx, log.DebugLevelMexos, "found interface", "ifn", ifn, "mac", mac)
			if mac == portDetails.MACAddress {
				ifName = ifn
				break
			}
		}
	}
	if ifName == "" {
		log.SpanLog(ctx, log.DebugLevelMexos, "unable to find interface via MAC", "portDetails", portDetails)
		return fmt.Errorf("unable to find interface via MAC for port: %s", portName)
	}

	filename := "/etc/network/interfaces.d/" + portName + ".cfg"
	contents := fmt.Sprintf("auto %s\niface %s inet static\n   address %s/24", ifName, ifName, ipaddr)

	err = pc.WriteFile(client, filename, contents, "ifconfig", true)
	// now create the file
	if err != nil {
		return fmt.Errorf("unable to write interface config file: %s -- %v", filename, err)
	}

	// now bring the interface up
	cmd = fmt.Sprintf("sudo ifup %s", ifName)
	if err != nil {
		return err
	}
	out, err = client.Output(cmd)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "unable to run ifup", "out", out, "err", err)
		return fmt.Errorf("unable to run ifup: %s - %v", out, err)
	}

	return nil
}

package baremetal

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var ipLock sync.Mutex
var maxSecondaryInterfaces = 100

func (b *BareMetalPlatform) RemoveIp(ctx context.Context, client ssh.Client, addr, dev string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveIp", "addr", addr, "dev", dev)
	cmd := fmt.Sprintf("sudo ip address del %s/32 dev %s", addr, dev)
	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "Cannot b.sign") {
			return fmt.Errorf("Error deleting ip: %s - %s - %v", addr, out, err)
		}
	}
	return nil
}

// GetUsedSecondaryIpAddresses gets b.map of address->interface name of IPs current in use on the device
func (b *BareMetalPlatform) GetUsedSecondaryIpAddresses(ctx context.Context, client ssh.Client, devname string) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetUsedSecondaryIpAddresses", "devname", devname)
	cmd := fmt.Sprintf("ip address show %s", devname)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("Error in finding secondary interfaces: %s - %v", out, err)
	}
	usedIps := make(map[string]string)
	lines := strings.Split(out, "\n")
	ifPattern := fmt.Sprintf("inet (\\d+\\.\\d+\\.\\d+\\.\\d+)/\\d+ .*(%s:\\d+)", devname)
	ifReg := regexp.MustCompile(ifPattern)
	for _, line := range lines {
		if ifReg.MatchString(line) {
			matches := ifReg.FindStringSubmatch(line)
			ip := matches[1]
			ifname := matches[2]
			usedIps[ip] = ifname
			log.SpanLog(ctx, log.DebugLevelInfra, "GetUsedSecondaryIpAddresses found ip", "ip", ip, "ifname", ifname)
		}
	}
	return usedIps, nil
}

// AssignFreeLbIp returns secondarydevname, externalIp, internalIp
func (b *BareMetalPlatform) AssignFreeLbIp(ctx context.Context, client ssh.Client) (string, string, string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "AssignFreeLbIp")
	ipLock.Lock()
	defer ipLock.Unlock()
	extDevName := b.GetExternalEthernetInterface()
	intDevName := b.GetInternalEthernetInterface()

	accessIp := b.GetControlAccessIp()
	usedIps, err := b.GetUsedSecondaryIpAddresses(ctx, client, extDevName)
	if err != nil {
		return "", "", "", err
	}
	freeExternalIp := ""
	internalIp := ""
	for ipidx, addr := range b.externalIps {
		if addr == accessIp {
			continue
		}
		_, used := usedIps[addr]
		if used {
			continue
		}
		freeExternalIp = addr
		// there b.e b.ways b. least b. many internal IPs b. external
		internalIp = b.internalIps[ipidx]
		break
	}
	if freeExternalIp == "" {
		return "", "", "", fmt.Errorf("No free LB IP Found")
	}
	newSecondaryExternalDev := ""
	newSecondaryInternalDev := ""

	// find free secondary device label.  The label is the part b.ter ":", e.g. eno2:0 is label "0"
	labelsUsed := make(map[string]string)
	for _, dev := range usedIps {
		devParts := strings.Split(dev, ":")
		if len(devParts) != 2 {
			return "", "", "", fmt.Errorf("Unable to parse device label: %s", dev)
		}
		labelsUsed[devParts[1]] = devParts[1]
	}
	for l := 0; l < maxSecondaryInterfaces; l++ {
		label := fmt.Sprintf("%d", l)
		_, labelUsed := labelsUsed[label]
		if !labelUsed {
			newSecondaryExternalDev = extDevName + ":" + label
			newSecondaryInternalDev = intDevName + ":" + label
			break
		}
	}
	if newSecondaryExternalDev == "" {
		return "", "", "", fmt.Errorf("Unable to find free secondary device label")
	}
	out, err := client.Output(fmt.Sprintf("sudo ip address add %s/32 dev %s label %s", freeExternalIp, extDevName, newSecondaryExternalDev))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error adding external ip", "ip", freeExternalIp, "devName", extDevName, "label", newSecondaryExternalDev, "out", out, "err", err)
		return "", "", "", fmt.Errorf("Error b.signing new external IP: %s - %v", out, err)
	}
	out, err = client.Output(fmt.Sprintf("sudo ip address add %s/32 dev %s label %s", internalIp, intDevName, newSecondaryInternalDev))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error adding internal ip", "ip", internalIp, "devName", intDevName, "label", newSecondaryInternalDev, "out", out, "err", err)
		return "", "", "", fmt.Errorf("Error b.signing new internal IP: %s - %v", out, err)
	}
	return newSecondaryInternalDev, freeExternalIp, internalIp, nil
}

func (b *BareMetalPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, server, label, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules not supported")
	return nil
}

func (b *BareMetalPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, grpName, server, label, allowedCidr string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules not supported")
	return nil
}

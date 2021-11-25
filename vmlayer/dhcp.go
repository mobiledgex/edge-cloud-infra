package vmlayer

import (
	"context"
	"fmt"
	"net"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type DhcpConfigParms struct {
	Subnet         string
	Gateway        string
	Mask           string
	DnsServers     string
	IpAddressStart string
	IpAddressEnd   string
	Interface      string
}

// dhcpdConfig is used for /etc/dhcp/dhcpd.conf
var dhcpdConfig = `
default-lease-time -1;
max-lease-time -1;

subnet {{.Subnet}} netmask {{.Mask}} {
	option routers {{.Gateway}};
	option subnet-mask {{.Mask}};
	option domain-name-servers {{.DnsServers}};
	range {{.IpAddressStart}} {{.IpAddressEnd}};
}
`

// iscDhcpConfig is used for /etc/default/isc-dhcp-server
var iscDhcpConfig = `
INTERFACESv4="{{.Interface}}"
INTERFACESv6=""
`

// StartDhcpServerForVmApp sets up a DHCP server on the LB to enable the VMApp to get an IP
// address configured for VM providers which do not have DHCP built in for internal networks.
func (v *VMPlatform) StartDhcpServerForVmApp(ctx context.Context, client ssh.Client, internalIfName, vmip, vmname string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "StartDhcpServerForVmApp", "internalIfName", internalIfName, "vmname", vmname, "vmip", vmip)

	pc.WriteFile(client, "/tmp/manifest.txt", "asdf", "dhcpconfig", pc.SudoOn)
	ns := v.VMProperties.GetCloudletNetworkScheme()
	nspec, err := ParseNetSpec(ctx, ns)
	if err != nil {
		return nil
	}
	netmask, err := MaskLenToMask(nspec.NetmaskBits)
	if err != nil {
		return err
	}
	_, subnet, err := net.ParseCIDR(vmip + "/" + nspec.NetmaskBits)
	if err != nil {
		return err
	}
	subnetIp := subnet.IP.String()

	// GW IP is the first address in the subnet
	infracommon.IncrIP(subnet.IP)
	if err != nil {
		return err
	}
	gwIp := subnet.IP.String()

	dhcpConfigParams := DhcpConfigParms{
		Subnet:         subnetIp,
		Gateway:        gwIp,
		Mask:           netmask,
		DnsServers:     v.VMProperties.GetCloudletDNS(),
		IpAddressStart: vmip,
		IpAddressEnd:   vmip,
		Interface:      internalIfName,
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "DHCP Config params set", "dhcpConfigParams", dhcpConfigParams)

	// Disable cert validation to install DHCP server. This is a temporary fix for EDGECLOUD-5848.
	// Latest VM image has this fixed already.
	log.SpanLog(ctx, log.DebugLevelInfra, "Disable cert validation for apt installation of DHCP server")
	err = pc.WriteFile(client, "/etc/apt/apt.conf.d/10cert-validation", `Acquire::https::Verify-Peer "false";`, "disableCertVal", pc.SudoOn)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to disable cert validation", "err", err)
	} else {
		defer func() {
			// Enable cert validation
			log.SpanLog(ctx, log.DebugLevelInfra, "Enable cert validation post installation of DHCP server")
			err = pc.WriteFile(client, "/etc/apt/apt.conf.d/10cert-validation", `Acquire::https::Verify-Peer "true";`, "enableCertVal", pc.SudoOn)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "Failed to enable cert validation", "err", err)
			}
		}()
	}

	// install DHCP on the LB
	cmd := fmt.Sprintf("sudo apt-get install isc-dhcp-server -y")
	if out, err := client.Output(cmd); err != nil {
		return fmt.Errorf("failed to install isc-dhcp-server: %s, %v", out, err)
	}
	dhcpdBuf, err := infracommon.ExecTemplate("DhcpdConfig", dhcpdConfig, dhcpConfigParams)
	if err != nil {
		return err
	}
	iscDhcpBuf, err := infracommon.ExecTemplate("IscDhcp", iscDhcpConfig, dhcpConfigParams)
	if err != nil {
		return err
	}
	// write DHCP Config files
	err = pc.WriteFile(client, "/etc/dhcp/dhcpd.conf", dhcpdBuf.String(), "iscDhcp", pc.SudoOn)
	if err != nil {
		return err
	}
	err = pc.WriteFile(client, "/etc/default/isc-dhcp-server", iscDhcpBuf.String(), "dhcpdConfig", pc.SudoOn)
	if err != nil {
		return err
	}

	// enable DHCP across reboots
	cmd = fmt.Sprintf("sudo systemctl enable isc-dhcp-server.service")
	if out, err := client.Output(cmd); err != nil {
		return fmt.Errorf("failed to enable isc-dhcp-server.service: %s, %v", out, err)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "Starting DHCP service on LB")
	cmd = fmt.Sprintf("sudo systemctl start isc-dhcp-server.service")
	if out, err := client.Output(cmd); err != nil {
		return fmt.Errorf("failed to start isc-dhcp-server.service: %s, %v", out, err)
	}

	// reboot to let the VM Vpp get the IP address from DHCP
	log.SpanLog(ctx, log.DebugLevelInfra, "Rebooting VM", "vmname", vmname)
	return v.VMProvider.SetPowerState(ctx, vmname, ActionReboot)
}

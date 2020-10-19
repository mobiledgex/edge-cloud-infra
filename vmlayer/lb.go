package vmlayer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"

	valid "github.com/asaskevich/govalidator"
	"github.com/mobiledgex/edge-cloud-infra/chefmgmt"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"

	ssh "github.com/mobiledgex/golang-ssh"
)

// InternalPortAttachPolicy is for dedicated clusters to define whether the internal port should be created when the rootlb
// is spun up, or afterwards.
type InternalPortAttachPolicy string

const AttachPortNotSupported InternalPortAttachPolicy = "AttachPortNotSupported"
const AttachPortDuringCreate InternalPortAttachPolicy = "AttachPortDuringCreate"
const AttachPortAfterCreate InternalPortAttachPolicy = "AttachPortAfterCreate"

var udevRulesFile = "/etc/udev/rules.d/70-persistent-net.rules"

var sharedRootLBPortLock sync.Mutex

type InterfaceActionsOp struct {
	addInterface    bool
	deleteInterface bool
	createIptables  bool
	deleteIptables  bool
}

var RootLBPorts = []dme.AppPort{}

// creates entries in the 70-persistent-net.rules files to ensure the interface names are consistent after reboot
func persistInterfaceName(ctx context.Context, client ssh.Client, ifName, mac string, action *InterfaceActionsOp) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "persistInterfaceName", "ifName", ifName, "mac", mac, "action", action)
	newFileContents := ""

	cmd := fmt.Sprintf("sudo cat %s", udevRulesFile)
	out, err := client.Output(cmd)
	// if the file exists, check for old entries
	if err == nil {
		lines := strings.Split(out, "\n")
		for _, l := range lines {
			// if the mac is already there remove it, it will be appended later
			if strings.Contains(l, mac) {
				log.SpanLog(ctx, log.DebugLevelInfra, "found existing rule for mac", "line", l)
			} else {
				newFileContents = newFileContents + l + "\n"
			}
		}
	}
	newRule := fmt.Sprintf("SUBSYSTEM==\"net\", ACTION==\"add\", DRIVERS==\"?*\", ATTR{address}==\"%s\", NAME=\"%s\"", mac, ifName)
	if action.addInterface {
		newFileContents = newFileContents + newRule + "\n"
	}
	// preexisting or not, write it
	return pc.WriteFile(client, udevRulesFile, newFileContents, "udev-rules", pc.SudoOn)
}

// configureInternalInterfaceAndExternalForwarding sets up the new internal interface and then creates iptables rules to forward
// traffic out the external interface
func (v *VMPlatform) configureInternalInterfaceAndExternalForwarding(ctx context.Context, client ssh.Client, subnetName, internalPortName string, serverDetails *ServerDetail, action *InterfaceActionsOp) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "configureInternalInterfaceAndExternalForwarding", "serverDetails", serverDetails, "internalPortName", internalPortName, "action", fmt.Sprintf("%+v", action))
	internalIP, err := GetIPFromServerDetails(ctx, "", internalPortName, serverDetails)
	if err != nil {
		return err
	}
	if internalIP.MacAddress == "" {
		return fmt.Errorf("No MAC address for internal interface: %s", internalPortName)
	}
	externalIP, err := GetIPFromServerDetails(ctx, v.VMProperties.GetCloudletExternalNetwork(), "", serverDetails)
	if err != nil {
		return err
	}
	if externalIP.MacAddress == "" {
		return fmt.Errorf("No MAC address for external interface: %s", externalIP.Network)
	}
	err = WaitServerReady(ctx, v.VMProvider, client, externalIP.ExternalAddr, MaxRootLBWait)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "server not ready", "err", err)
		return err
	}

	// discover the interface names matching our macs
	externalIfname := ""
	internalIfname := ""
	cmd := fmt.Sprintf("ip -br link | awk '$3 ~ /^%s/ {print $1; exit 1}'", internalIP.MacAddress)
	out, _ := client.Output(cmd)
	internalIfname = out
	log.SpanLog(ctx, log.DebugLevelInfra, "found interface", "ifn", internalIfname, "mac", internalIP.MacAddress)
	cmd = fmt.Sprintf("ip -br link | awk '$3 ~ /^%s/ {print $1; exit 1}'", externalIP.MacAddress)
	out, _ = client.Output(cmd)
	externalIfname = out
	log.SpanLog(ctx, log.DebugLevelInfra, "found interface", "ifn", externalIfname, "mac", externalIP.MacAddress)

	if externalIfname == "" {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find external interface via MAC", "mac", externalIP.MacAddress)
		if action.addInterface {
			return fmt.Errorf("unable to find interface for external port mac: %s", externalIP.MacAddress)
		}
		// keep going on delete
	}
	if internalIfname == "" {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find internal interface via MAC", "mac", internalIP.MacAddress)
		if action.addInterface {
			return fmt.Errorf("unable to find interface for internal port mac: %s", internalIP.MacAddress)
		}
		// keep going on delete
	}
	netplanEnabled := ServerIsNetplanEnabled(ctx, client)
	filename, fileMatch, contents := GetNetworkFileDetailsForIP(ctx, internalPortName, internalIfname, internalIP.InternalAddr, netplanEnabled)
	if action.addInterface {
		// cleanup any interfaces files that may be sitting around with our new interface, perhaps from some old failure
		cmd := fmt.Sprintf("grep -l ' %s ' %s", fileMatch, internalIfname)
		out, err = client.Output(cmd)
		log.SpanLog(ctx, log.DebugLevelInfra, "cleanup old interface files with interface", "internalIfname", internalIfname, "out", out, "err", err)
		if err == nil {
			files := strings.Split(out, "\n")
			for _, f := range files {
				log.SpanLog(ctx, log.DebugLevelInfra, "cleanup old interfaces file", "file", f)
				cmd := fmt.Sprintf("sudo rm -f %s", f)
				out, err := client.Output(cmd)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "unable to delete file", "file", f, "out", out, "err", err)
				}
			}
		}
		err = pc.WriteFile(client, filename, contents, "netconfig", pc.SudoOn)
		// now create the file
		if err != nil {
			return fmt.Errorf("unable to write network config file: %s -- %v", filename, err)
		}

		// now bring the new internal interface up.
		var ipcmds []string
		linkCmd := fmt.Sprintf("sudo ip link set dev %s up", internalIfname)
		ipcmds = append(ipcmds, linkCmd)
		flushCmd := fmt.Sprintf("sudo ip addr flush %s", internalIfname)
		ipcmds = append(ipcmds, flushCmd)
		addrCmd := fmt.Sprintf("sudo ip addr add %s/24 dev %s", internalIP.InternalAddr, internalIfname)
		ipcmds = append(ipcmds, addrCmd)
		for _, c := range ipcmds {
			log.SpanLog(ctx, log.DebugLevelInfra, "bringing up interface", "internalIfname", internalIfname, "cmd", c)
			out, err = client.Output(c)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to run", "cmd", c, "out", out, "err", err)
				return fmt.Errorf("unable to run ip command: %s - %v", out, err)
			}
		}

	} else if action.deleteInterface {
		cmd := fmt.Sprintf("sudo rm %s", filename)
		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "No such file") {
				log.SpanLog(ctx, log.DebugLevelInfra, "file already gone", "filename", filename)
			} else {
				return fmt.Errorf("Unexpected error removing network config file %s, %s -- %v", filename, out, err)
			}
		}

		cmd = fmt.Sprintf("sudo ip addr flush %s", internalIfname)
		log.SpanLog(ctx, log.DebugLevelInfra, "removing ip from interface", "internalIfname", internalIfname, "cmd", internalIfname)
		out, err = client.Output(cmd)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to run ", "cmd", cmd, "out", out, "err", err)
		}
	}
	// we can get here on some error cases in which the ifname were not found
	if internalIfname != "" {
		if action.addInterface || action.deleteInterface {
			err = persistInterfaceName(ctx, client, internalIfname, internalIP.MacAddress, action)
			if err != nil {
				return nil
			}
		}
		if action.createIptables || action.deleteIptables {
			if externalIfname != "" {
				err = v.setupForwardingIptables(ctx, client, externalIfname, internalIfname, action)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "setupForwardingIptables failed", "err", err)
				}
			}
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "persistInterfaceName and setupForwardingIptables skipped due to empty internalIfName")
	}
	return err
}

// AttachAndEnableRootLBInterface attaches the interface and enables it in the OS
func (v *VMPlatform) AttachAndEnableRootLBInterface(ctx context.Context, client ssh.Client, rootLBName string, attachPort bool, subnetName, internalPortName, internalIPAddr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachAndEnableRootLBInterface", "rootLBName", rootLBName, "attachPort", attachPort, "subnetName", subnetName, "internalPortName", internalPortName)

	if rootLBName == v.VMProperties.SharedRootLBName {
		sharedRootLBPortLock.Lock()
		defer sharedRootLBPortLock.Unlock()
	}
	var action InterfaceActionsOp
	action.createIptables = true
	if attachPort {
		action.addInterface = true
		err := v.VMProvider.AttachPortToServer(ctx, rootLBName, subnetName, internalPortName, internalIPAddr, ActionCreate)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "fail to attach port", "err", err)
			return err
		}
	}
	sd, err := v.VMProvider.GetServerDetail(ctx, rootLBName)
	if err != nil {
		return err
	}
	err = v.configureInternalInterfaceAndExternalForwarding(ctx, client, subnetName, internalPortName, sd, &action)
	if err != nil {
		if attachPort {
			log.SpanLog(ctx, log.DebugLevelInfra, "fail to confgure internal interface, detaching port", "err", err)
			deterr := v.VMProvider.DetachPortFromServer(ctx, rootLBName, subnetName, internalPortName)
			if deterr != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "fail to detach port", "err", deterr)
			}
		}
		return err
	}
	return nil
}

func (v *VMPlatform) GetRootLBName(key *edgeproto.CloudletKey) string {
	name := cloudcommon.GetRootLBFQDN(key, v.VMProperties.CommonPf.PlatformConfig.AppDNSRoot)
	return v.VMProvider.NameSanitize(name)
}

// DetachAndDisableRootLBInterface performs some cleanup when deleting the rootLB port.
func (v *VMPlatform) DetachAndDisableRootLBInterface(ctx context.Context, client ssh.Client, rootLBName string, subnetName, internalPortName, internalIPAddr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachAndDisableRootLBInterface", "rootLBName", rootLBName, "subnetName", subnetName, "internalPortName", internalPortName)
	if rootLBName == v.VMProperties.SharedRootLBName {
		sharedRootLBPortLock.Lock()
		defer sharedRootLBPortLock.Unlock()
	}

	var action InterfaceActionsOp
	action.deleteIptables = true
	action.deleteInterface = true

	sd, err := v.VMProvider.GetServerDetail(ctx, rootLBName)
	if err != nil {
		return err
	}

	err = v.configureInternalInterfaceAndExternalForwarding(ctx, client, subnetName, internalPortName, sd, &action)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error in configureInternalInterfaceAndExternalForwarding", "err", err)
	}

	err = v.VMProvider.DetachPortFromServer(ctx, rootLBName, subnetName, internalPortName)
	if err != nil {
		// might already be gone
		log.SpanLog(ctx, log.DebugLevelInfra, "fail to detach port", "err", err)
	}

	return err
}

var rootLBLock sync.Mutex
var MaxRootLBWait = 5 * time.Minute

// GetVMSpecForRootLB gets the VM spec for the rootLB when it is not specified within a cluster. This is
// used for Shared RootLb and for VM app based RootLb
func (v *VMPlatform) GetVMSpecForRootLB(ctx context.Context, rootLbName string, subnetConnect string, tags []string, updateCallback edgeproto.CacheUpdateCallback) (*VMRequestSpec, error) {

	log.SpanLog(ctx, log.DebugLevelInfra, "GetVMSpecForRootLB", "rootLbName", rootLbName)

	var rootlbFlavor edgeproto.Flavor
	err := v.VMProperties.GetCloudletSharedRootLBFlavor(&rootlbFlavor)
	if err != nil {
		return nil, fmt.Errorf("unable to get Shared RootLB Flavor: %v", err)
	}

	cli := edgeproto.CloudletInfo{}
	cli.Key = *v.VMProperties.CommonPf.PlatformConfig.CloudletKey
	cli.Flavors = v.FlavorList
	restbls := v.GetResTablesForCloudlet(ctx, &cli.Key)
	vmspec, err := vmspec.GetVMSpec(ctx, rootlbFlavor, cli, restbls)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "RootLB GetVMSpec error", "v.FlavorList", v.FlavorList, "rootlbFlavor", rootlbFlavor, "err", err)
		return nil, fmt.Errorf("unable to find VM spec for RootLB: %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetVMSpec returned", "flavor", vmspec.FlavorName, "for mex flavor", rootlbFlavor)
	az := vmspec.AvailabilityZone
	if az == "" {
		az = v.VMProperties.GetCloudletComputeAvailabilityZone()
	}
	imageName, err := v.GetCloudletImageToUse(ctx, updateCallback)
	if err != nil {
		return nil, err
	}
	chefAttributes := make(map[string]interface{})
	chefAttributes["tags"] = tags
	clientName := v.GetChefClientName(rootLbName)
	chefParams := v.GetVMChefParams(clientName, "", chefmgmt.ChefPolicyBase, chefAttributes)
	return v.GetVMRequestSpec(ctx,
		VMTypeRootLB,
		rootLbName,
		vmspec.FlavorName,
		imageName,
		true,
		WithExternalVolume(vmspec.ExternalVolumeSize),
		WithSubnetConnection(subnetConnect),
		WithChefParams(chefParams))
}

// GetVMSpecForRootLBPorts get a vmspec for the purpose of creating new ports to the specified subnet
func (v *VMPlatform) GetVMSpecForRootLBPorts(ctx context.Context, rootLbName string, subnet string) (*VMRequestSpec, error) {
	rootlb, err := v.GetVMRequestSpec(
		ctx,
		VMTypeRootLB,
		rootLbName,
		"dummyflavor",
		"dummyimage",
		false, // shared RLB already has external ports
		WithCreatePortsOnly(true),
		WithSubnetConnection(subnet),
	)
	return rootlb, err
}

//CreateRootLB creates a rootLB.  It should not be called if the rootLB already exists, as to save time we don't check
func (v *VMPlatform) CreateRootLB(
	ctx context.Context, rootLBName string,
	cloudletKey *edgeproto.CloudletKey,
	imgPath, imgVersion string,
	action ActionType,
	tags []string,
	updateCallback edgeproto.CacheUpdateCallback,
) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "create rootlb", "name", rootLBName, "action", action)
	if action == ActionCreate {
		_, err := v.VMProvider.GetServerDetail(ctx, rootLBName)
		if err == nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "rootlb already exists")
			return nil
		}
	}
	vmreq, err := v.GetVMSpecForRootLB(ctx, rootLBName, "", tags, updateCallback)
	if err != nil {
		return err
	}
	var vms []*VMRequestSpec
	vms = append(vms, vmreq)
	_, err = v.OrchestrateVMsFromVMSpec(ctx, rootLBName, vms, action, updateCallback, WithNewSecurityGroup(GetServerSecurityGroupName(rootLBName)))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error while creating RootLB VM", "name", rootLBName, "error", err)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "done creating rootlb", "name", rootLBName)
	return nil

}

//SetupRootLB prepares the RootLB. It will optionally create the rootlb if the createRootLBFlavor
// is not blank and no existing server found
func (v *VMPlatform) SetupRootLB(
	ctx context.Context, rootLBName string,
	cloudletKey *edgeproto.CloudletKey,
	privacyPolicy *edgeproto.PrivacyPolicy,
	updateCallback edgeproto.CacheUpdateCallback,
) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupRootLB", "rootLBName", rootLBName)
	// ensure no entries exist in the ip cache for this rootlb
	DeleteServerIpFromCache(ctx, rootLBName)

	//fqdn is that of the machine/kvm-instance running the agent
	if !valid.IsDNSName(rootLBName) {
		return fmt.Errorf("fqdn %s is not valid", rootLBName)
	}
	sd, err := v.VMProvider.GetServerDetail(ctx, rootLBName)
	if err == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "server with same name as rootLB exists", "rootLBName", rootLBName)
	}

	// setup SSH access to cloudlet for CRM.  Since we are getting the external IP here, this will only work
	// when CRM accessed via public internet.
	log.SpanLog(ctx, log.DebugLevelInfra, "setup rootLBName group for SSH access")
	groupName := GetServerSecurityGroupName(rootLBName)
	client, err := v.GetSSHClientForServer(ctx, rootLBName, v.VMProperties.GetCloudletExternalNetwork(), pc.WithUser(infracommon.SSHUser), pc.WithCachedIp(true))
	if err != nil {
		return err
	}
	myIp, err := infracommon.GetExternalPublicAddr(ctx)
	if err != nil {
		// this is not necessarily fatal
		log.InfoLog("cannot fetch public ip", "err", err)
	} else {
		var sshPort = []dme.AppPort{{
			PublicPort: 22,
			Proto:      dme.LProto_L_PROTO_TCP,
		}}
		myCidr := myIp + "/32"
		err = v.VMProvider.WhitelistSecurityRules(ctx, client, groupName, rootLBName, "rootlb-ssh", myCidr, sshPort)
		if err != nil {
			return err
		}
		if v.VMProperties.RequiresWhitelistOwnIp {
			for _, a := range sd.Addresses {
				extCidr := a.ExternalAddr + "/32"
				err = v.VMProvider.WhitelistSecurityRules(ctx, client, groupName, rootLBName, "own-externalip-ssh", extCidr, sshPort)
				if err != nil {
					return err
				}
			}
		}
	}
	err = WaitServerReady(ctx, v.VMProvider, client, rootLBName, MaxRootLBWait)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "timeout waiting for rootLB", "name", rootLBName)
		return fmt.Errorf("Error waiting for rootLB %v", err)
	}
	ip, err := GetIPFromServerDetails(ctx, v.VMProperties.GetCloudletExternalNetwork(), "", sd)
	if err != nil {
		return fmt.Errorf("cannot get rootLB IP %sv", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Copy resource-tracker to rootLb", "rootLb", rootLBName)
	err = CopyResourceTracker(client)
	if err != nil {
		return fmt.Errorf("cannot copy resource-tracker to rootLb %v", err)
	}
	route, err := v.VMProperties.GetInternalNetworkRoute(ctx)
	if err != nil {
		return err
	}
	err = v.AddRouteToServer(ctx, client, rootLBName, route)
	if err != nil {
		return fmt.Errorf("failed to AddRouteToServer %v", err)
	}
	err = v.VMProvider.WhitelistSecurityRules(ctx, client, GetServerSecurityGroupName(rootLBName), rootLBName, "rootlb-ports", GetAllowedClientCIDR(), RootLBPorts)
	if err != nil {
		return fmt.Errorf("failed to WhitelistSecurityRules %v", err)
	}

	if err = v.VMProperties.CommonPf.ActivateFQDNA(ctx, rootLBName, ip.ExternalAddr); err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "DNS A record activated", "name", rootLBName)

	// perform provider specific prep of the rootLB
	return v.VMProvider.PrepareRootLB(ctx, client, rootLBName, GetServerSecurityGroupName(rootLBName), privacyPolicy)
}

// This function copies resource-tracker from crm to rootLb - we need this to provide docker metrics
func CopyResourceTracker(client ssh.Client) error {
	path, err := exec.LookPath("resource-tracker")
	if err != nil {
		return err
	}
	err = SCPFilePath(client, path, "/tmp/resource-tracker")
	if err != nil {
		return err
	}
	// copy to /usr/local/bin/resource-tracker
	cmd := fmt.Sprintf("sudo cp /tmp/resource-tracker /usr/local/bin/resource-tracker")
	_, err = client.Output(cmd)
	if err != nil {
		return err
	}
	// make it executable
	cmd = fmt.Sprintf("sudo chmod a+rx /usr/local/bin/resource-tracker")
	_, err = client.Output(cmd)
	return err
}

func (v *VMPlatform) DeleteProxySecurityGroupRules(ctx context.Context, client ssh.Client, proxyName string, secGrpName string, label string, ports []dme.AppPort, app *edgeproto.App, serverName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteProxySecurityGroupRules", "proxyName", proxyName, "ports", ports)

	err := proxy.DeleteNginxProxy(ctx, client, proxyName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot delete proxy", "proxyName", proxyName, "error", err)
	}
	allowedClientCIDR := GetAllowedClientCIDR()
	return v.VMProvider.RemoveWhitelistSecurityRules(ctx, client, secGrpName, label, allowedClientCIDR, ports)
}

func GetChefRootLBTags(platformConfig *platform.PlatformConfig) []string {
	return []string{
		"deploytag/" + platformConfig.DeploymentTag,
		"region/" + platformConfig.Region,
		"cloudlet/" + platformConfig.CloudletKey.Name,
		"cloudletorg/" + platformConfig.CloudletKey.Organization,
		"vmtype/" + string(VMTypeRootLB),
	}
}

package vmlayer

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type NetworkType string

const ServerDoesNotExistError string = "Server does not exist"

var ServerActive = "ACTIVE"
var ServerShutoff = "SHUTOFF"

var ActionStart = "start"
var ActionStop = "stop"
var ActionReboot = "reboot"

type ServerDetail struct {
	Addresses []ServerIP
	ID        string
	Name      string
	Status    string
}

func (v *VMPlatform) GetIPFromServerName(ctx context.Context, networkName, subnetName, serverName string) (*ServerIP, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIPFromServerName", "networkName", networkName, "subnetName", subnetName, "serverName", serverName)
	// if this is a root lb, look it up and get the IP if we have it cached
	portName := ""
	if subnetName != "" {
		portName = GetPortName(serverName, subnetName)
	}
	if networkName == v.VMProperties.GetCloudletExternalNetwork() {
		rootLB, err := GetRootLB(ctx, serverName)
		if err == nil && rootLB != nil {
			if rootLB.IP != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "using existing rootLB IP", "IP", rootLB.IP)
				return rootLB.IP, nil
			}
		}
	}
	sd, err := v.VMProvider.GetServerDetail(ctx, serverName)
	if err != nil {
		return nil, err
	}
	sip, err := GetIPFromServerDetails(ctx, networkName, portName, sd)
	if err != nil && subnetName != "" {
		// Clusters create prior to R2 use a different port naming convention.  For backwards
		// compatibility, let's try to find the server using the old port format.  This was the
		// server name plus port.
		oldFormatPortName := serverName + "-port"
		log.SpanLog(ctx, log.DebugLevelInfra, "Unable to find server IP, try again with old format port name", "oldFormatPortName", oldFormatPortName)
		return GetIPFromServerDetails(ctx, networkName, oldFormatPortName, sd)
	}
	return sip, err
}

func GetIPFromServerDetails(ctx context.Context, networkName string, portName string, sd *ServerDetail) (*ServerIP, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIPFromServerDetails", "server", sd.Name, "networkName", networkName, "portName", portName, "serverDetail", sd)
	var sipPtr *ServerIP
	found := false
	for i, s := range sd.Addresses {
		if (networkName != "" && s.Network == networkName) || (portName != "" && s.PortName == portName) {
			if found {
				// this indicates we passed in multiple parameters that found an IP.  For example, an external network name plus an internal port name
				log.SpanLog(ctx, log.DebugLevelInfra, "Error: GetIPFromServerDetails found multiple matches", "networkName", networkName, "portName", portName, "serverDetail", sd)
				return nil, fmt.Errorf("Multiple IP addresses found for server: %s network: %s portName: %s", sd.Name, networkName, portName)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "GetIPFromServerDetails found match", "serverAddress", s)
			found = true
			sipPtr = &sd.Addresses[i]
		}
	}
	if found {
		return sipPtr, nil
	}
	return nil, fmt.Errorf("unable to find IP for server: %s on network: %s port: %s", sd.Name, networkName, portName)
}

func GetCloudletNetworkIfaceFile(netplanEnabled bool) string {
	if netplanEnabled {
		return "/etc/netplan/50-cloud-init.yaml"
	} else {
		return "/etc/network/interfaces.d/50-cloud-init.cfg"
	}
}

func (v *VMPlatform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.DeploymentTypeVM:
		objName := cloudcommon.GetAppFQN(&app.Key)
		return v.VMProvider.GetConsoleUrl(ctx, objName)
	default:
		return "", fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

func (v *VMPlatform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	PowerState := appInst.PowerState
	switch deployment := app.Deployment; deployment {
	case cloudcommon.DeploymentTypeVM:
		serverName := cloudcommon.GetAppFQN(&app.Key)
		fqdn := appInst.Uri

		log.SpanLog(ctx, log.DebugLevelInfra, "setting server state", "serverName", serverName, "fqdn", fqdn, "PowerState", PowerState)

		updateCallback(edgeproto.UpdateTask, "Verifying AppInst state")
		serverDetail, err := v.VMProvider.GetServerDetail(ctx, serverName)
		if err != nil {
			return err
		}

		serverAction := ""
		switch PowerState {
		case edgeproto.PowerState_POWER_ON_REQUESTED:
			if serverDetail.Status == ServerActive {
				return fmt.Errorf("server %s is already active", serverName)
			}
			serverAction = ActionStart
		case edgeproto.PowerState_POWER_OFF_REQUESTED:
			if serverDetail.Status == ServerShutoff {
				return fmt.Errorf("server %s is already stopped", serverName)
			}
			serverAction = ActionStop
		case edgeproto.PowerState_REBOOT_REQUESTED:
			serverAction = ActionReboot
			if serverDetail.Status != ServerActive {
				return fmt.Errorf("server %s is not active", serverName)
			}
		default:
			return fmt.Errorf("unsupported server power action: %s", PowerState)
		}

		serverSubnet := v.VMProperties.GetCloudletExternalNetwork()
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
			serverSubnet = serverName + "-subnet"
		}
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
		oldServerIP, err := GetIPFromServerDetails(ctx, serverSubnet, "", serverDetail)
		if err != nil || oldServerIP.ExternalAddr == "" {
			return fmt.Errorf("unable to fetch external ip for %s, addr %s, err %v", serverName, serverSubnet, err)
		}
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Performing action %s on %s", serverAction, serverName))
		err = v.VMProvider.SetPowerState(ctx, serverName, serverAction)
		if err != nil {
			return err
		}

		if PowerState == edgeproto.PowerState_POWER_ON_REQUESTED || PowerState == edgeproto.PowerState_REBOOT_REQUESTED {
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Waiting for server %s to become active", serverName))
			serverDetail, err := v.VMProvider.GetServerDetail(ctx, serverName)
			if err != nil {
				return err
			}
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
			newServerIP, err := GetIPFromServerDetails(ctx, serverSubnet, "", serverDetail)
			if err != nil || newServerIP.ExternalAddr == "" {
				return fmt.Errorf("unable to fetch external ip for %s, addr %s, err %v", serverName, serverSubnet, err)
			}
			if oldServerIP.ExternalAddr != newServerIP.ExternalAddr {
				// IP changed, update DNS entry
				updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Updating DNS entry as IP changed for %s", serverName))
				log.SpanLog(ctx, log.DebugLevelInfra, "updating DNS entry", "serverName", serverName, "fqdn", fqdn, "ip", newServerIP)
				err = v.VMProperties.CommonPf.ActivateFQDNA(ctx, fqdn, newServerIP.ExternalAddr)
				if err != nil {
					return fmt.Errorf("unable to update fqdn for %s, addr %s, err %v", serverName, newServerIP.ExternalAddr, err)
				}
			}
		}
		updateCallback(edgeproto.UpdateTask, "Performed power control action successfully")
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}
	return nil
}

// WaitServerReady waits up to the specified duration for the server to be reachable via SSH
// and pass any additional checks from the provider
func WaitServerReady(ctx context.Context, provider VMProvider, client ssh.Client, server string, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WaitServerReady", "server", server)
	start := time.Now()
	for {
		out, err := client.Output("sudo grep 'Finished mobiledgex init' /var/log/mobiledgex.log")
		log.SpanLog(ctx, log.DebugLevelInfra, "grep Finished mobiledgex init result", "out", out, "err", err)
		if err == nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Server has completed mobiledgex init", "server", server)
			// perform any additional checks from the provider
			err = provider.CheckServerReady(ctx, client, server)
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckServerReady result", "err", err)
			if err == nil {
				break
			}
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "server not ready", "err", err)
		elapsed := time.Since(start)
		if elapsed > timeout {
			return fmt.Errorf("timed out waiting for VM %s", server)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "sleeping 10 seconds before retry", "elapsed", elapsed, "timeout", timeout)
		time.Sleep(10 * time.Second)

	}
	log.SpanLog(ctx, log.DebugLevelInfra, "WaitServerReady OK", "server", server)
	return nil
}

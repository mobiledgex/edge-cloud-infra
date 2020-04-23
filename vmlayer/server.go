package vmlayer

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
)

type NetworkType string

const ServerDoesNotExistError string = "Server does not exist"

var ServerActive = "ACTIVE"
var ServerShutoff = "SHUTOFF"

type ServerDetail struct {
	Addresses []ServerIP
	ID        string
	Name      string
	Status    string
}

func (c *VMPlatform) GetIPFromServerDetails(ctx context.Context, networkName string, sd *ServerDetail) (*ServerIP, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIPFromServerDetails", "networkName", networkName, "serverDetail", sd)
	for _, s := range sd.Addresses {
		if s.Network == networkName {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("unable to find IP for server: %s on network: %s", sd.Name, networkName)
}

func GetCloudletNetworkIfaceFile() string {
	return "/etc/network/interfaces.d/50-cloud-init.cfg"
}

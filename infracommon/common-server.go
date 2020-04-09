package infracommon

import (
	"context"
	"fmt"
)

const NetworkExternal = "external"
const NetworkInternal = "internal"

type NetworkType string

var ServerActive = "ACTIVE"
var ServerShutoff = "SHUTOFF"

type ServerDetail struct {
	Addresses []ServerIP
	ID        string
	Name      string
	Status    string
}

func (c *CommonPlatform) GetIPFromServerDetails(ctx context.Context, networkName string, sd *ServerDetail) (*ServerIP, error) {
	for _, s := range sd.Addresses {
		if s.Network == networkName {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("unable to find IP for server: %s on network: %s", sd.Name, networkName)
}

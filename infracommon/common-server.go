package infracommon

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
)

const NetworkExternal = "external"
const NetworkInternal = "internal"

type NetworkType string

type ServerDetail struct {
	Addresses []ServerIP
	ID        string
	Name      string
}

func (c *CommonPlatform) GetIPFromServerDetails(ctx context.Context, netType NetworkType, sd *ServerDetail) (*ServerIP, error) {
	if netType == NetworkExternal {
		for _, a := range sd.Addresses {
			if a.ExternalAddr != "" {
				return &a, nil
			}
		}
	}
	for _, a := range sd.Addresses {
		if a.InternalAddr != "" {
			return &a, nil
		}
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "no address found for server ", "netType", netType, "serverName", sd.Name)
	return nil, fmt.Errorf("no address found for server: %s on network type: %s", sd.Name, netType)
}

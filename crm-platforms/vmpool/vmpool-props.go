package vmpool

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
)

func (o *VMPoolPlatform) GetProviderSpecificProps() map[string]*infracommon.PropertyInfo {
	return map[string]*infracommon.PropertyInfo{}
}

func (o *VMPoolPlatform) InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	if _, ok := vars["MEX_ROUTER"]; !ok {
		o.VMProperties.CommonPf.Properties["MEX_ROUTER"] = &infracommon.PropertyInfo{
			Value: vmlayer.NoConfigExternalRouter,
		}
	}
	return nil
}

func (o *VMPoolPlatform) GetApiAccessFilename() string {
	return ""
}

func (s *VMPoolPlatform) GetExternalGateway(ctx context.Context, extNetName string) (string, error) {
	if val, ok := s.VMProperties.CommonPf.Properties["MEX_EXTERNAL_NETWORK_GATEWAY"]; ok {
		return val.Value, nil
	}
	return "", nil
}

package generic

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
)

func (o *GenericPlatform) GetProviderSpecificProps() map[string]*infracommon.PropertyInfo {
	return map[string]*infracommon.PropertyInfo{}
}

func (o *GenericPlatform) InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	if _, ok := vars["MEX_ROUTER"]; !ok {
		o.VMProperties.CommonPf.Properties["MEX_ROUTER"] = &infracommon.PropertyInfo{
			Value: vmlayer.NoConfigExternalRouter,
		}
	}
	return nil
}

func (o *GenericPlatform) GetApiAccessFilename() string {
	return ""
}

func (s *GenericPlatform) GetExternalGateway(ctx context.Context, extNetName string) (string, error) {
	if val, ok := s.VMProperties.CommonPf.Properties["MEX_EXTERNAL_NETWORK_GATEWAY"]; ok {
		return val.Value, nil
	}
	return "", nil
}

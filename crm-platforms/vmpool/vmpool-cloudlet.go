package vmpool

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

func (o *VMPoolPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SaveCloudletAccessVars not supported")
	return nil
}

func (o *VMPoolPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr not supported")
	return "", nil
}

func (o *VMPoolPlatform) GetCloudletManifest(ctx context.Context, name string, cloudletImagePath string, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest", "name", name)
	var manifest infracommon.CloudletManifest

	if vmgp == nil {
		return "", nil
	}
	if len(vmgp.VMs) != 1 {
		return "", fmt.Errorf("invalid number of VMs")
	}
	cloudConfigParams := vmgp.VMs[0].CloudConfigParams
	if cloudConfigParams.ChefParams == nil {
		return "", fmt.Errorf("missing chef params for %s", name)
	}
	if cloudConfigParams.ChefParams.ClientKey == "" {
		return "", fmt.Errorf("missing chef client key for %s", cloudConfigParams.ChefParams.NodeName)
	}

	scriptText := fmt.Sprintf(`
#!/bin/bash

cat > /home/ubuntu/client.pem << EOF
%s
EOF

`, cloudConfigParams.ChefParams.ClientKey)

	if cloudConfigParams.AccessKey != "" {
		scriptText += fmt.Sprintf(`
cat > /root/accesskey/accesskey.pem << EOF
%s
EOF

`, cloudConfigParams.AccessKey)
	}

	scriptText += fmt.Sprintf(`
sudo bash /etc/mobiledgex/setup-chef.sh -s "%s" -n "%s"
`, cloudConfigParams.ChefParams.ServerPath, cloudConfigParams.ChefParams.NodeName)

	manifest.AddItem("SSH into one of the VMs from the VMPool which has access to controller's notify port", infracommon.ManifestTypeNone, infracommon.ManifestSubTypeNone, "")
	manifest.AddItem("Save and execute the following script on the VM", infracommon.ManifestTypeCode, infracommon.ManifestSubTypeBash, scriptText)
	return manifest.ToString()
}

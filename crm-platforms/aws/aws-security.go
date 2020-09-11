package aws

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AWSPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, label, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules not supported")
	return nil
}

func (a *AWSPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, grpName, server, label, allowedCidr string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules not supported")
	return nil
}

package generic

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *GenericPlatform) AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddSecurityRuleCIDRWithRetry not supported")
	return nil
}

func (o *GenericPlatform) RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules not supported")
	return nil
}

func (o *GenericPlatform) WhitelistSecurityRules(ctx context.Context, grpName, serverName, allowedCidr string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules not supported")
	return nil
}

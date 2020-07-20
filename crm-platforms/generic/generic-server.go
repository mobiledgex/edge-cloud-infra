package generic

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *GenericPlatform) GetServerDetail(ctx context.Context, serverName string) (*vmlayer.ServerDetail, error) {
	// TODO: Fetch list from controller?
	return &vmlayer.ServerDetail{}, nil
}

func (o *GenericPlatform) CreateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	// TODO: Allocate VMs
	return nil
}
func (o *GenericPlatform) UpdateVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return nil
}

func (o *GenericPlatform) SyncVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncVMs")
	// nothing to do right now
	return nil

}
func (o *GenericPlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	// TODO: Release VMs
	return nil
}

func (s *GenericPlatform) GetVMStats(ctx context.Context, key *edgeproto.AppInstKey) (*vmlayer.VMMetrics, error) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "GetVMStats not supported")
	return &vmlayer.VMMetrics{}, nil
}

func (s *GenericPlatform) GetPlatformResourceInfo(ctx context.Context) (*vmlayer.PlatformResources, error) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "GetPlatformResourceInfo not supported")
	return nil, nil
}

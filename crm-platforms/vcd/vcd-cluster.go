package vcd

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Cluster related operations

func (v *VcdPlatform) CreateCluster(ctx context.Context, cloud *MexCloudlet, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (*CidrMap, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster")

	nextCidr, err := v.GetNextInternalNet(ctx, cloud)
	if err != nil {
		fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
		return nil, err
	}
	fmt.Printf("CreateCluster-I-new cluster's CIDR: %s\n", nextCidr)
	return nil, nil

}

func (v *VcdPlatform) DeleteCluster(ctx context.Context, cloud MexCloudlet, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCluster")
	return nil

}

func (v *VcdPlatform) UpdateCluster(ctx context.Context, cloud MexCloudlet, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (*CidrMap, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCluster")
	return nil, nil

}

func (v *VcdPlatform) RestartCluster(ctx context.Context, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RestartCluster")
	return nil

}

func (v *VcdPlatform) StartCluster(ctx context.Context, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "StopCluster")
	return nil

}

func (v *VcdPlatform) StopCluster(ctx context.Context, vmMap *CidrMap) error {

	return nil

}

package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *OpenstackPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("TODO")
}

func (s *OpenstackPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	return fmt.Errorf("TODO")
}

func (s *OpenstackPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	return fmt.Errorf("TODO")
}

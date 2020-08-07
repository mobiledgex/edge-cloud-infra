package managedk8s

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (m *ManagedK8sPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateClusterInst", "clusterInst", clusterInst)
	var err error

	if err = m.Provider.Login(ctx); err != nil {
		return err
	}

	// perform any actions to create prereq resource before the cluster
	if err = m.Provider.CreateClusterPrerequisites(ctx, clusterInst); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in creating cluster prereqs", "err", err)
		return err
	}

	if err = m.Provider.RunClusterCreateCommand(ctx, clusterInst); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in creating cluster", "err", err)
		return err
	}
	// race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	client, err := m.GetClusterPlatformClient(ctx, clusterInst, cloudcommon.ClientTypeRootLB)
	if err != nil {
		return err
	}
	infracommon.BackupKubeconfig(ctx, client)
	if err = m.Provider.GetCredentials(ctx, clusterInst); err != nil {
		return err
	}
	kconf := k8smgmt.GetKconfName(clusterInst)
	if err = pc.CopyFile(client, infracommon.DefaultKubeconfig(), kconf); err != nil {
		return err
	}
	return nil
}

func (m *ManagedK8sPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteClusterInst", "clusterInst", clusterInst)
	return m.Provider.RunClusterDeleteCommand(ctx, clusterInst)
}

func (s *ManagedK8sPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("Update cluster inst not implemented")
}

package managedk8s

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
)

// ManagedK8sProvider is an interface that platforms implement to perform the details of interfacing with managed kubernetes services
type ManagedK8sProvider interface {
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	GetK8sProviderSpecificProps() map[string]*infracommon.PropertyInfo
	Login(ctx context.Context) error
	GetCredentials(ctx context.Context, clusterInst *edgeproto.ClusterInst) error
	NameSanitize(name string) string
	RunClusterCreateCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error
	RunClusterDeleteCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error
}

const (
	ManagedK8sProviderAzure string = "azure"
	ManagedK8sProviderGCP   string = "gcp"
	ManagedK8sProviderAWS   string = "aws"
)

type ManagedK8sProperties struct {
	CommonPf infracommon.CommonPlatform
}

// ManagedK8sPlatform contains info needed by all Managed Kubernetes Providers
type ManagedK8sPlatform struct {
	Type     string
	CommonPf infracommon.CommonPlatform
	Provider ManagedK8sProvider
}

func (m *ManagedK8sPlatform) GetType() string {
	return m.Type
}

func (m *ManagedK8sPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Init", "type", m.GetType())
	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to get vault configs", "vaultAddr", platformConfig.VaultAddr, "err", err)
		return err
	}
	props := m.Provider.GetK8sProviderSpecificProps()
	if err := m.CommonPf.InitInfraCommon(ctx, platformConfig, props, vaultConfig); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "InitInfraCommon failed", "err", err)
		return err
	}
	return m.Provider.Login(ctx)
}

func (m *ManagedK8sPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return m.Provider.GatherCloudletInfo(ctx, info)
}

func (m *ManagedK8sPlatform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (m *ManagedK8sPlatform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (m *ManagedK8sPlatform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	return []edgeproto.CloudletMgmtNode{}, nil
}

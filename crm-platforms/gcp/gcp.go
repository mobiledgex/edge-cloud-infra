package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

var GCPServiceAccount string //temp

type GCPPlatform struct {
	commonPf infracommon.CommonPlatform
}

type GCPQuotas struct {
	Limit  float64
	Metric string
}

type GCPQuotasList struct {
	Quotas GCPQuotas
}

type GCPFlavor struct {
	GuestCPUs                    int
	MaximumPersistentDisksSizeGb string
	MemoryMb                     int
	Name                         string
}

func (g *GCPPlatform) GetType() string {
	return "gcp"
}

func (g *GCPPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	if err := g.commonPf.InitInfraCommon(ctx, platformConfig, gcpProps, vaultConfig, g); err != nil {
		return err
	}
	return nil
}

func (g *GCPPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLimits (GCP)")
	err := g.GCPLogin(ctx)
	if err != nil {
		return err
	}
	var quotas []GCPQuotasList

	filter := fmt.Sprintf("name=(%s) AND quotas.metric=(CPUS, DISKS_TOTAL_GB)", g.GetGcpZone())
	flatten := "quotas[]"
	format := "json(quotas.metric,quotas.limit)"

	out, err := sh.Command("gcloud", "compute", "regions", "list",
		"--project", g.GetGcpProject(), "--filter", filter, "--flatten", flatten,
		"--format", format, sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get resource quotas from gcp, %s, %s", out, err.Error())
		return err
	}
	err = json.Unmarshal(out, &quotas)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %s, %v", out, err)
		return err
	}
	for _, q := range quotas {
		if q.Quotas.Metric == "CPUS" {
			info.OsMaxVcores = uint64(q.Quotas.Limit)
			info.OsMaxRam = uint64(3.75 * float32(q.Quotas.Limit))
		} else if q.Quotas.Metric == "DISKS_TOTAL_GB" {
			info.OsMaxVolGb = uint64(q.Quotas.Limit)
		} else {
			err = fmt.Errorf("unexpected Quotas metric: %s", q.Quotas.Metric)
			return err
		}
	}

	var machinetypes []GCPFlavor
	filter = fmt.Sprintf("zone=(%s) AND name:(standard)", g.GetGcpZone())
	format = "json(name,guestCpus,memoryMb,maximumPersistentDisksSizeGb)"

	out, err = sh.Command("gcloud", "compute", "machine-types", "list",
		"--project", g.GetGcpProject(), "--filter", filter,
		"--format", format, sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get machine-types from gcp, %s, %s", out, err.Error())
		return err
	}
	err = json.Unmarshal(out, &machinetypes)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %s, %v", out, err)
		return err
	}
	for _, m := range machinetypes {
		disk, err := strconv.Atoi(m.MaximumPersistentDisksSizeGb)
		if err != nil {
			err = fmt.Errorf("failed to parse gcp output, %s", err.Error())
			return err
		}
		info.Flavors = append(
			info.Flavors,
			&edgeproto.FlavorInfo{
				Name:  m.Name,
				Vcpus: uint64(m.GuestCPUs),
				Ram:   uint64(m.MemoryMb),
				Disk:  uint64(disk),
			},
		)
	}
	return nil
}

func (g *GCPPlatform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (s *GCPPlatform) NameSanitize(string) string {
	return "not implemented"
}

func (s *GCPPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *GCPPlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) GetServerDetail(ctx context.Context, serverName string) (*infracommon.ServerDetail, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCPPlatform) GetIPFromServerName(ctx context.Context, networkName, serverName string) (*infracommon.ServerIP, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCPPlatform) GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error) {
	return "", "", fmt.Errorf("not implemented")
}

func (s *GCPPlatform) AttachPortToServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) DetachPortFromServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) CreateAppVM(ctx context.Context, vmAppParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName string, dedicatedRootLB bool) error {
	return fmt.Errorf("not implemented")
}

func (o *GCPPlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (s *GCPPlatform) GetVMParams(ctx context.Context, depType infracommon.DeploymentType, serverName, flavorName string, externalVolumeSize uint64, imageName, secGrp string, cloudletKey *edgeproto.CloudletKey, opts ...infracommon.VMParamsOp) (*infracommon.VMParams, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCPPlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

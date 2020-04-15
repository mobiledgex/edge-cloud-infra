package azure

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

type AzurePlatform struct {
	commonPf infracommon.CommonPlatform
}

func (a *AzurePlatform) GetType() string {
	return "azure"
}

func (a *AzurePlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	//TODO AZURE PROVIDER?
	if err := a.commonPf.InitInfraCommon(ctx, platformConfig, azureProps, vaultConfig, a); err != nil {
		return err
	}

	return nil
}

type AZName struct {
	LocalizedValue string
	Value          string
}

type AZLimit struct {
	CurrentValue string
	Limit        string
	LocalName    string
	Name         AZName
}

type AZFlavor struct {
	Disk  int
	Name  string
	RAM   int
	VCPUs int
}

func (a *AzurePlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLimits (Azure)")
	if err := a.AzureLogin(ctx); err != nil {
		return err
	}

	var limits []AZLimit
	out, err := sh.Command("az", "vm", "list-usage", "--location", a.GetAzureLocation(), sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get limits from azure, %s, %s", out, err.Error())
		return err
	}
	err = json.Unmarshal(out, &limits)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return err
	}
	for _, l := range limits {
		if l.LocalName == "Total Regional vCPUs" {
			vcpus, err := strconv.Atoi(l.Limit)
			if err != nil {
				err = fmt.Errorf("failed to parse azure output, %s", err.Error())
				return err
			}
			info.OsMaxVcores = uint64(vcpus)
			info.OsMaxRam = uint64(4 * vcpus)
			info.OsMaxVolGb = uint64(500 * vcpus)
			break
		}
	}

	/*
	 * We will not support all Azure flavors, only selected ones:
	 * https://azure.microsoft.com/en-in/pricing/details/virtual-machines/series/
	 */
	var vmsizes []AZFlavor
	out, err = sh.Command("az", "vm", "list-sizes",
		"--location", a.GetAzureLocation(),
		"--query", "[].{"+
			"Name:name,"+
			"VCPUs:numberOfCores,"+
			"RAM:memoryInMb, Disk:resourceDiskSizeInMb"+
			"}[?starts_with(Name,'Standard_DS')]|[?ends_with(Name,'v2')]",
		sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get vm-sizes from azure, %s, %s", out, err.Error())
		return err
	}
	err = json.Unmarshal(out, &vmsizes)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return err
	}
	for _, f := range vmsizes {
		info.Flavors = append(
			info.Flavors,
			&edgeproto.FlavorInfo{
				Name:  f.Name,
				Vcpus: uint64(f.VCPUs),
				Ram:   uint64(f.RAM),
				Disk:  uint64(f.Disk),
			},
		)
	}
	return nil
}

func (s *AzurePlatform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (a *AzurePlatform) NameSanitize(string) string {
	return "not implemented"
}

func (a *AzurePlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (a *AzurePlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) GetServerDetail(ctx context.Context, serverName string) (*infracommon.ServerDetail, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *AzurePlatform) GetIPFromServerName(ctx context.Context, networkName, serverName string) (*infracommon.ServerIP, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *AzurePlatform) GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error) {
	return "", "", fmt.Errorf("not implemented")
}

func (a *AzurePlatform) AttachPortToServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) DetachPortFromServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) CreateAppVM(ctx context.Context, vmAppParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (o *AzurePlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName string, dedicatedRootLB bool) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (a *AzurePlatform) GetVMParams(ctx context.Context, depType infracommon.DeploymentType, serverName, flavorName string, externalVolumeSize uint64, imageName, secGrp string, cloudletKey *edgeproto.CloudletKey, opts ...infracommon.VMParamsOp) (*infracommon.VMParams, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *AzurePlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

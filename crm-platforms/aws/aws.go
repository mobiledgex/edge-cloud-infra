package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type AWSPlatform struct {
	commonPf infracommon.CommonPlatform
}

type AWSQuotas struct {
	Limit  float64
	Metric string
}

type AWSQuotasList struct {
	Quotas AWSQuotas
}

type AWSFlavor struct {
	GuestCPUs                    int
	MaximumPersistentDisksSizeGb string
	MemoryMb                     int
	Name                         string
}

func (a *AWSPlatform) GetType() string {
	return "AWS"
}

func (a *AWSPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	if err := a.commonPf.InitInfraCommon(ctx, platformConfig, AWSProps, vaultConfig); err != nil {
		return err
	}
	if a.GetAWSZone() == "" {
		return fmt.Errorf("Env variable MEX_AWS_ZONE not set")
	}
	return nil
}

// aws ec2 describe-instance-types --filters 'Name=instance-storage-supported,Values=true' \
// --query 'InstanceTypes[].[InstanceType,VCpuInfo.DefaultVCpus,VCpuInfo.DefaultCores,MemoryInfo.SizeInMiB,InstanceStorageInfo.TotalSizeInGB]'
func (a *AWSPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLimits (AWS)")
	err := a.AWSLogin(ctx)
	if err != nil {
		return err
	}
	var quotas []AWSQuotasList

	//filter := fmt.Sprintf("name=(%s) AND quotas.metric=(CPUS, DISKS_TOTAL_GB)", a.GetAWSZone())
	filter := "Name=instance-storage-supported,Values=true"
	flatten := "quotas[]"
	format := "json(quotas.metric,quotas.limit)"
	query := "InstanceTypes[].[InstanceType,VCpuInfo.DefaultVCpus,VCpuInfo.DefaultCores,MemoryInfo.SizeInMiB,InstanceStorageInfo.TotalSizeInGB]"
	
	out, err := sh.Command("aws", "ec2", "describe-instance-types",
	    "--filter", filter,
		"--query", query,
		"--flatten", flatten,
		"--output", "json", sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get resource quotas from AWS, %s, %s", out, err.Error())
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

	var machinetypes []AWSFlavor
	filter = fmt.Sprintf("zone=(%s) AND name:(standard)", a.GetAWSZone())
	format = "json(name,guestCpus,memoryMb,maximumPersistentDisksSizeGb)"

	out, err = sh.Command("gcloud", "compute", "machine-types", "list",
		"--project", a.GetAWSProject(), "--filter", filter,
		"--format", format, sh.Dir("/tmp")).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot get machine-types from AWS, %s, %s", out, err.Error())
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
			err = fmt.Errorf("failed to parse AWS output, %s", err.Error())
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

func (a *AWSPlatform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (a *AWSPlatform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (a *AWSPlatform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	return []edgeproto.CloudletMgmtNode{}, nil
}

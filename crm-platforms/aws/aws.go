package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// AwsGwOctet overrides he use of .1 for the GW address when using the LB as the GW because the first few addresses are reserved
const AwsGwOctet uint32 = 5

type AWSPlatform struct {
	//commonPf     *infracommon.CommonPlatform
	VMProperties *vmlayer.VMProperties
	caches       *platform.Caches
}

type AWSQuotas struct {
	Limit  float64
	Metric string
}

type AWSInstanceType []interface{}

type AWSQuotasList struct {
	Quotas AWSQuotas
}

type AWSFlavor struct {
	Name     string
	Vcpus    uint
	MemoryMb uint
	DiskGb   uint
}

// GatherCloudletInfo gets flavor info from AWS
func (a *AWSPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo (AWS)")
	filter := "Name=instance-storage-supported,Values=true"
	query := "InstanceTypes[].[InstanceType,VCpuInfo.DefaultVCpus,MemoryInfo.SizeInMiB,InstanceStorageInfo.TotalSizeInGB]"

	out, err := a.TimedAwsCommand(ctx, "aws", "ec2", "describe-instance-types",
		"--filter", filter,
		"--query", query,
		"--region", a.GetAwsRegion(),
		"--output", "json")
	if err != nil {
		err = fmt.Errorf("cannot get instance types from AWS, %s, %s", out, err.Error())
		return err
	}
	jbytes := []byte(out)

	var instanceTypes []AWSInstanceType
	err = json.Unmarshal(jbytes, &instanceTypes)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %s, %v", out, err)
		return err
	}

	log.DebugLog(log.DebugLevelInfra, "AWS ", "instance types", instanceTypes)
	for _, m := range instanceTypes {
		name, ok := m[0].(string)
		if !ok {
			err := fmt.Errorf("wrong type for flavor name %T", m[0])
			return err
		}
		vcpus, ok := m[1].(float64)
		if !ok {
			err := fmt.Errorf("wrong type for vcpus %T", m[1])
			return err
		}
		ram, ok := m[2].(float64)
		if !ok {
			err := fmt.Errorf("wrong type for ram %T", m[2])
			return err
		}

		disk, ok := m[3].(float64)
		if !ok {
			err := fmt.Errorf("wrong type for disk %T", m[3])
			return err
		}

		info.Flavors = append(
			info.Flavors,
			&edgeproto.FlavorInfo{
				Name:  name,
				Vcpus: uint64(vcpus),
				Ram:   uint64(ram),
				Disk:  uint64(disk),
			},
		)
	}
	return nil
}

func (a *AWSPlatform) Login(ctx context.Context) error {
	return nil
}

func (a *AWSPlatform) NameSanitize(clusterName string) string {
	//	return strings.NewReplacer(".", "").Replace(clusterName)
	return clusterName
}

// AWSPlatform IdSanitize is the same as NameSanitize
func (a *AWSPlatform) IdSanitize(name string) string {
	return a.NameSanitize(name)
}

func (a *AWSPlatform) SetVMProperties(vmProperties *vmlayer.VMProperties) {
	vmProperties.OverrideGWOctet = AwsGwOctet
	vmProperties.UseSecgrpForInternalSubnet = true
	a.VMProperties = vmProperties

}

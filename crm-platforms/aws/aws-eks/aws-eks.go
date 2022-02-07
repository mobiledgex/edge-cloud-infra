package awseks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	awsgen "github.com/mobiledgex/edge-cloud-infra/crm-platforms/aws/aws-generic"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const (
	AWSServiceCodeEKS = "eks"
	AWSServiceCodeELB = "elasticloadbalancing"

	// Codes used to identify service quota for AWS resources
	AWSServiceQuotaClusters          = "L-1194D53C"
	AWSServiceQuotaNodesPerNodeGroup = "L-BD136A63"
	AWSServiceQuotaAppLBPerRegion    = "L-53DA6B97"
)

type AwsEksPlatform struct {
	awsGenPf *awsgen.AwsGenericPlatform
}

type AwsEksResources struct {
	K8sClustersUsed        uint64
	K8sNodesPerClusterUsed uint64
	AppLBsUsed             uint64
}

func (o *AwsEksPlatform) GetFeatures() *platform.Features {
	return &platform.Features{
		SupportsMultiTenantCluster:    true,
		SupportsKubernetesOnly:        true,
		KubernetesRequiresWorkerNodes: true,
		IPAllocatedPerService:         true,
	}
}

func (a *AwsEksPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return a.awsGenPf.GatherCloudletInfo(ctx, "", info)
}

// CreateClusterPrerequisites does nothing for now, but for outpost may need to create a vpc
func (a *AwsEksPlatform) CreateClusterPrerequisites(ctx context.Context, clusterName string) error {
	return nil
}

// RunClusterCreateCommand creates a kubernetes cluster on AWS
func (a *AwsEksPlatform) RunClusterCreateCommand(ctx context.Context, clusterName string, numNodes uint32, flavor string) error {
	log.DebugLog(log.DebugLevelInfra, "RunClusterCreateCommand", "clusterName", clusterName, "numNodes:", numNodes, "NodeFlavor", flavor)
	// Can not create a managed cluster if numNodes is 0
	region := a.awsGenPf.GetAwsRegion()
	out, err := infracommon.Sh(a.awsGenPf.AccountAccessVars).Command("eksctl", "create", "--region", region, "cluster", "--name", clusterName, "--node-type", flavor, "--nodes", fmt.Sprintf("%d", numNodes), "--managed").CombinedOutput()
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "Create eks cluster failed", "clusterName", clusterName, "out", string(out), "err", err)
		return fmt.Errorf("Create eks cluster failed: %s - %v", string(out), err)
	}
	return nil
}

// RunClusterDeleteCommand removes the kubernetes cluster on AWS
func (a *AwsEksPlatform) RunClusterDeleteCommand(ctx context.Context, clusterName string) error {
	log.DebugLog(log.DebugLevelInfra, "RunClusterDeleteCommand", "clusterName:", clusterName)
	out, err := infracommon.Sh(a.awsGenPf.AccountAccessVars).Command("eksctl", "delete", "cluster", "--name", clusterName).CombinedOutput()
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "Delete eks cluster failed", "clusterName", clusterName, "out", string(out), "err", err)
		return fmt.Errorf("Delete eks cluster failed: %s - %v", string(out), err)
	}
	return nil
}

// GetCredentials retrieves kubeconfig credentials from AWS
func (a *AwsEksPlatform) GetCredentials(ctx context.Context, clusterName string) error {
	log.DebugLog(log.DebugLevelInfra, "GetCredentials", "clusterName:", clusterName)
	out, err := infracommon.Sh(a.awsGenPf.AccountAccessVars).Command("eksctl", "utils", "write-kubeconfig", clusterName).CombinedOutput()
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "Error in write-kubeconfig", "out", string(out), "err", err)
		return fmt.Errorf("Error in write-kubeconfig: %s - %v", string(out), err)
	}
	return nil
}

func (a *AwsEksPlatform) SetProperties(props *infracommon.InfraProperties) error {
	a.awsGenPf = &awsgen.AwsGenericPlatform{Properties: props}
	return nil
}

func (a *AwsEksPlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	return a.awsGenPf.GetFlavorList(ctx, "")
}

func (a *AwsEksPlatform) GetProviderSpecificProps(ctx context.Context) (map[string]*edgeproto.PropertyInfo, error) {
	return a.awsGenPf.GetProviderSpecificProps(ctx)
}

func (a *AwsEksPlatform) Login(ctx context.Context) error {
	return nil
}

func (a *AwsEksPlatform) NameSanitize(clusterName string) string {
	return strings.NewReplacer(".", "").Replace(clusterName)
}

func (a *AwsEksPlatform) InitApiAccessProperties(ctx context.Context, accessApi platform.AccessApi, vars map[string]string) error {
	return nil
}

func (a *AwsEksPlatform) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "AwsEks GetAccessData", "dataType", dataType)
	return a.awsGenPf.GetAccessData(ctx, cloudlet, region, vaultConfig, dataType, arg)
}

func (a *AwsEksPlatform) getClusterList(ctx context.Context) ([]awsgen.AWSCluster, error) {
	clusters := []awsgen.AWSCluster{}
	region := a.awsGenPf.GetAwsRegion()
	out, err := infracommon.Sh(a.awsGenPf.AccountAccessVars).Command(
		"eksctl", "get", "cluster",
		"--region", region,
		"--output", "json",
	).CombinedOutput()
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "Failed to get eks cluster list", "out", string(out), "err", err)
		return nil, fmt.Errorf("Failed to get eks cluster list: %s - %v", string(out), err)
	}
	err = json.Unmarshal(out, &clusters)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal eks cluster list, %s, %v", out, err)
	}

	return clusters, nil
}

func (a *AwsEksPlatform) GetCloudletInfraResourcesInfo(ctx context.Context) ([]edgeproto.InfraResource, error) {
	clusterList, err := a.getClusterList(ctx)
	if err != nil {
		return nil, err
	}
	awsELB, err := a.awsGenPf.GetAWSELBs(ctx)
	if err != nil {
		return nil, err
	}
	eksSvcQuotas, err := a.awsGenPf.GetServiceQuotas(ctx, AWSServiceCodeEKS)
	if err != nil {
		return nil, err
	}
	elbSvcQuotas, err := a.awsGenPf.GetServiceQuotas(ctx, AWSServiceCodeELB)
	if err != nil {
		return nil, err
	}
	clusterListMax := uint64(0)
	clusterNodesMax := uint64(0)
	appLBMax := uint64(0)
	for _, eksSvcQuota := range eksSvcQuotas {
		switch eksSvcQuota.Code {
		case AWSServiceQuotaClusters:
			clusterListMax = uint64(eksSvcQuota.Value)
		case AWSServiceQuotaNodesPerNodeGroup:
			clusterNodesMax = uint64(eksSvcQuota.Value)
		}
	}
	for _, elbSvcQuota := range elbSvcQuotas {
		switch elbSvcQuota.Code {
		case AWSServiceQuotaAppLBPerRegion:
			appLBMax = uint64(elbSvcQuota.Value)
		}
	}
	resInfo := []edgeproto.InfraResource{
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceK8sClusters,
			Value:         uint64(len(clusterList)),
			InfraMaxValue: clusterListMax,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceK8sNodesPerCluster,
			InfraMaxValue: clusterNodesMax,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceAppLBs,
			Value:         uint64(len(awsELB.LoadBalancerDescriptions)),
			InfraMaxValue: appLBMax,
		},
	}
	return resInfo, nil
}

func (a *AwsEksPlatform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{
		Properties: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceK8sClusters,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceK8sClusters],
			},
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceK8sNodesPerCluster,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceK8sNodesPerCluster],
			},
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceAppLBs,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceAppLBs],
			},
		},
	}, nil
}

func getAwsEksResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, resources []edgeproto.VMResource) *AwsEksResources {
	var eksRes AwsEksResources
	// ClusterInstKey -> Node count
	uniqueClusters := make(map[edgeproto.ClusterInstKey]int)
	appLBs := 0
	for _, vmRes := range resources {
		if vmRes.Type == cloudcommon.ResourceTypeK8sLBSvc {
			appLBs++
			continue
		}
		if vmRes.Type != cloudcommon.VMTypeClusterK8sNode {
			continue
		}
		if _, ok := uniqueClusters[vmRes.Key]; !ok {
			uniqueClusters[vmRes.Key] = 1
		} else {
			uniqueClusters[vmRes.Key] += 1
		}
	}
	K8sNodesPerCluster := 0
	for _, v := range uniqueClusters {
		if v > K8sNodesPerCluster {
			K8sNodesPerCluster = v
		}
	}
	eksRes.K8sClustersUsed = uint64(len(uniqueClusters))
	eksRes.K8sNodesPerClusterUsed = uint64(K8sNodesPerCluster)
	eksRes.AppLBsUsed = uint64(appLBs)
	log.SpanLog(ctx, log.DebugLevelApi, "AwsEks getAwsEksResources", "cloudletKey", cloudlet.Key, "resources", eksRes)
	return &eksRes
}

// called by controller, make sure it doesn't make any calls to infra API
func (a *AwsEksPlatform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource {
	log.SpanLog(ctx, log.DebugLevelApi, "AwsEks GetClusterAdditionalResources", "cloudletKey", cloudlet.Key)
	// resource name -> resource units
	cloudletRes := map[string]string{
		cloudcommon.ResourceK8sClusters:        "",
		cloudcommon.ResourceK8sNodesPerCluster: "",
		cloudcommon.ResourceAppLBs:             "",
	}
	resInfo := make(map[string]edgeproto.InfraResource)
	for resName, resUnits := range cloudletRes {
		resMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			resMax = infraRes.InfraMaxValue
		}
		resInfo[resName] = edgeproto.InfraResource{
			Name:          resName,
			InfraMaxValue: resMax,
			Units:         resUnits,
		}
	}

	eksRes := getAwsEksResources(ctx, cloudlet, vmResources)
	outInfo, ok := resInfo[cloudcommon.ResourceK8sClusters]
	if ok {
		outInfo.Value += eksRes.K8sClustersUsed
		resInfo[cloudcommon.ResourceK8sClusters] = outInfo
	}
	outInfo, ok = resInfo[cloudcommon.ResourceK8sNodesPerCluster]
	if ok {
		outInfo.Value = eksRes.K8sNodesPerClusterUsed
		resInfo[cloudcommon.ResourceK8sNodesPerCluster] = outInfo
	}
	outInfo, ok = resInfo[cloudcommon.ResourceAppLBs]
	if ok {
		outInfo.Value += eksRes.AppLBsUsed
		resInfo[cloudcommon.ResourceAppLBs] = outInfo
	}
	return resInfo
}

func (a *AwsEksPlatform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	eksRes := getAwsEksResources(ctx, cloudlet, resources)

	resMetric.AddIntVal(cloudcommon.ResourceMetricK8sClusters, eksRes.K8sClustersUsed)
	resMetric.AddIntVal(cloudcommon.ResourceMetricK8sNodesPerCluster, eksRes.K8sNodesPerClusterUsed)
	resMetric.AddIntVal(cloudcommon.ResourceMetricAppLBs, eksRes.AppLBsUsed)
	return nil
}

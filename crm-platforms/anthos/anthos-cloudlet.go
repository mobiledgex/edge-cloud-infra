package anthos

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/util"

	"github.com/mobiledgex/edge-cloud-infra/chefmgmt"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (a *AnthosPlatform) GetChefParams(nodeName, clientKey string, policyName string, attributes map[string]interface{}) *chefmgmt.VMChefParams {
	chefServerPath := a.commonPf.ChefServerPath
	if chefServerPath == "" {
		chefServerPath = chefmgmt.DefaultChefServerPath
	}
	return &chefmgmt.VMChefParams{
		NodeName:    nodeName,
		ServerPath:  chefServerPath,
		ClientKey:   clientKey,
		Attributes:  attributes,
		PolicyName:  policyName,
		PolicyGroup: a.commonPf.DeploymentTag,
	}
}

func (a *AnthosPlatform) GetChefClientName(ckey *edgeproto.CloudletKey) string {
	// Prefix with region name
	name := util.K8SSanitize(ckey.Name + "-" + ckey.Organization)
	return a.commonPf.DeploymentTag + "-" + a.commonPf.PlatformConfig.Region + "-" + name
}

func (a *AnthosPlatform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet", "cloudlet", cloudlet)

	if !pfConfig.TestMode {
		err := a.commonPf.InitCloudletSSHKeys(ctx, accessApi)
		if err != nil {
			return err
		}
	}
	if err := a.commonPf.InitInfraCommon(ctx, a.commonPf.PlatformConfig, anthosProps); err != nil {
		return err
	}

	// edge-cloud image already contains the certs
	if pfConfig.TlsCertFile != "" {
		crtFile, err := infracommon.GetDockerCrtFile(pfConfig.TlsCertFile)
		if err != nil {
			return err
		}
		pfConfig.TlsCertFile = crtFile
	}

	if pfConfig.ChefServerPath == "" {
		pfConfig.ChefServerPath = chefmgmt.DefaultChefServerPath
	}

	// For real setups, ansible will always specify the correct
	// cloudlet container and vm image paths to the controller.
	// But for local testing convenience, we default to the hard-coded
	// ones if not specified.
	if pfConfig.ContainerRegistryPath == "" {
		pfConfig.ContainerRegistryPath = infracommon.DefaultContainerRegistryPath
	}
	chefApi := chefmgmt.ChefApiAccess{}
	chefAttributes, err := chefmgmt.GetChefPlatformAttributes(ctx, cloudlet, pfConfig, "platform", &chefApi)
	if err != nil {
		return err
	}
	if a.commonPf.ChefClient == nil {
		return fmt.Errorf("Chef client is not initialized")
	}

	chefPolicy := chefmgmt.ChefPolicyDocker
	if cloudlet.Deployment == cloudcommon.DeploymentTypeKubernetes {
		chefPolicy = chefmgmt.ChefPolicyK8s
	}
	cloudlet.ChefClientKey = make(map[string]string)
	if cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
		hostName := cloudlet.Key.Name
		chefParams := a.GetChefParams(hostName, "", chefPolicy, chefAttributes)

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Creating chef client %s with cloudlet attributes", hostName))
		clientKey, err := chefmgmt.ChefClientCreate(ctx, a.commonPf.ChefClient, chefParams)
		if err != nil {
			return err
		}
		// Store client key in cloudlet obj
		cloudlet.ChefClientKey[hostName] = clientKey

		// Return, as end-user will setup the platform VM
		return nil
	}

	/*
		err = a.SetupChef(ctx, accessApi, cloudlet, pfConfig, pfFlavor, updateCallback)
		if err != nil {
			return err
		}*/

	clientName := a.GetChefClientName(&cloudlet.Key)
	return chefmgmt.GetChefRunStatus(ctx, a.commonPf.ChefClient, clientName, cloudlet, pfConfig, accessApi, updateCallback)
}

func (a *AnthosPlatform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateCloudlet TODO")
}

func (a *AnthosPlatform) UpdateTrustPolicy(ctx context.Context, TrustPolicy *edgeproto.TrustPolicy) error {
	return fmt.Errorf("UpdateTrustPolicy TODO")
}

func (a *AnthosPlatform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("DeleteCloudlet TODO")
}

package anthos

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/util"

	"github.com/mobiledgex/edge-cloud-infra/chefmgmt"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AnthosPlatform) GetChefParams(nodeName, clientKey string, policyName string, attributes map[string]interface{}) *chefmgmt.ServerChefParams {
	chefServerPath := a.commonPf.ChefServerPath
	if chefServerPath == "" {
		chefServerPath = chefmgmt.DefaultChefServerPath
	}
	return &chefmgmt.ServerChefParams{
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

	//	if !pfConfig.TestMode {
	err := a.commonPf.InitCloudletSSHKeys(ctx, accessApi)
	if err != nil {
		return err
	}
	//	}
	a.commonPf.PlatformConfig = infracommon.GetPlatformConfig(cloudlet, pfConfig, accessApi)
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
	cloudlet.NotifySrvAddr = "mexplat-dev.ctrl.mobiledgex.net:37001"

	chefAttributes, err := chefmgmt.GetChefPlatformAttributes(ctx, cloudlet, pfConfig, "platform", &chefApi)
	chefAttributes["notifyAddrs"] = "mexplat-dev.ctrl.mobiledgex.net:37001" //TEMP
	log.WarnLog("XXXXX CHEFATTR", "chefAttributes", chefAttributes)
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
		return fmt.Errorf("Restricted access not yet supported on Anthos")
	}
	clientName := a.GetChefClientName(&cloudlet.Key)
	chefParams := a.GetChefParams(clientName, "", chefPolicy, chefAttributes)

	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Creating chef client %s with cloudlet attributes", clientName))
	clientKey, err := chefmgmt.ChefClientCreate(ctx, a.commonPf.ChefClient, chefParams)
	if err != nil {
		return err
	}
	// Store client key in cloudlet obj
	cloudlet.ChefClientKey[clientName] = clientKey
	sshClient, err := a.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: a.commonPf.PlatformConfig.CloudletKey.String(), Type: "anthoscontrolhost"})
	if err != nil {
		return fmt.Errorf("Failed to get ssh client to control host: %v", err)
	}
	err = a.SetupChefOnServer(ctx, sshClient, clientName, cloudlet, chefParams)
	if err != nil {
		return err
	}

	return chefmgmt.GetChefRunStatus(ctx, a.commonPf.ChefClient, clientName, cloudlet, pfConfig, accessApi, updateCallback)
}

func (a *AnthosPlatform) SetupChefOnServer(ctx context.Context, sshClient ssh.Client, clientName string, cloudlet *edgeproto.Cloudlet, chefParams *chefmgmt.ServerChefParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupChefOnServer", "clientName", clientName)

	err := pc.WriteFile(sshClient, "/etc/chef/client.pem", cloudlet.ChefClientKey[clientName], "chef-key", pc.SudoOn)
	if err != nil {
		return fmt.Errorf("Failed to write chef client key: %v", err)
	}
	// note the client name is actually used as the node name
	command := fmt.Sprintf("sudo chef-client --chef-license ACCEPT -S %s -N %s -l debug -i 60 -d -L /var/log/chef.log", chefParams.ServerPath, clientName)
	log.WarnLog("RUNCHEF", "command", command)

	out, err := sshClient.Output(command)
	if err != nil {
		return fmt.Errorf("Failed to run chef client: %s - %v", out, err)
	}
	return nil
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

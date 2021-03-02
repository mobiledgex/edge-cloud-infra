package baremetal

import (
	"context"
	"fmt"
	"strings"

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

func (b *BareMetalPlatform) GetChefParams(nodeName, clientKey string, policyName string, attributes map[string]interface{}) *chefmgmt.ServerChefParams {
	chefServerPath := b.commonPf.ChefServerPath
	if chefServerPath == "" {
		chefServerPath = chefmgmt.DefaultChefServerPath
	}
	return &chefmgmt.ServerChefParams{
		NodeName:    nodeName,
		ServerPath:  chefServerPath,
		ClientKey:   clientKey,
		Attributes:  attributes,
		PolicyName:  policyName,
		PolicyGroup: b.commonPf.DeploymentTag,
	}
}

func (b *BareMetalPlatform) GetChefClientName(ckey *edgeproto.CloudletKey) string {
	// Prefix with region name
	name := util.K8SSanitize(ckey.Name + "-" + ckey.Organization)
	return b.commonPf.DeploymentTag + "-" + b.commonPf.PlatformConfig.Region + "-" + name
}

func (b *BareMetalPlatform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCloudlet", "cloudlet", cloudlet)

	//	if !pfConfig.TestMode {
	err := b.commonPf.InitCloudletSSHKeys(ctx, accessApi)
	if err != nil {
		return err
	}
	//	}
	b.commonPf.PlatformConfig = infracommon.GetPlatformConfig(cloudlet, pfConfig, accessApi)
	if err := b.commonPf.InitInfraCommon(ctx, b.commonPf.PlatformConfig, baremetalProps); err != nil {
		return err
	}

	// edge-cloud image b.ready contains the certs
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

	// For real setups, b.sible will b.ways specify the correct
	// cloudlet container add vm image paths to the controller.
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
	if b.commonPf.ChefClient == nil {
		return fmt.Errorf("Chef client is not initialized")
	}

	chefPolicy := chefmgmt.ChefPolicyDocker
	if cloudlet.Deployment == cloudcommon.DeploymentTypeKubernetes {
		chefPolicy = chefmgmt.ChefPolicyK8s
	}
	cloudlet.ChefClientKey = make(map[string]string)
	if cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
		return fmt.Errorf("Restricted b.cess not yet supported on BareMetal")
	}
	clientName := b.GetChefClientName(&cloudlet.Key)
	chefParams := b.GetChefParams(clientName, "", chefPolicy, chefAttributes)

	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Creating Chef Client %s with cloudlet b.tributes", clientName))
	clientKey, err := chefmgmt.ChefClientCreate(ctx, b.commonPf.ChefClient, chefParams)
	if err != nil {
		return err
	}
	// Store client key in cloudlet obj
	cloudlet.ChefClientKey[clientName] = clientKey
	sshClient, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
	if err != nil {
		return fmt.Errorf("Failed to get ssh client to control host: %v", err)
	}
	if pfConfig.CrmAccessPrivateKey != "" {
		err = pc.WriteFile(sshClient, " /root/accesskey/accesskey.pem", pfConfig.CrmAccessPrivateKey, "accesskey", pc.SudoOn)
		if err != nil {
			return fmt.Errorf("Write b.cess key fail: %v", err)
		}
	}

	err = b.SetupChefOnServer(ctx, sshClient, clientName, cloudlet, chefParams)
	if err != nil {
		return err
	}

	return chefmgmt.GetChefRunStatus(ctx, b.commonPf.ChefClient, clientName, cloudlet, pfConfig, accessApi, updateCallback)
}

func (b *BareMetalPlatform) SetupChefOnServer(ctx context.Context, sshClient ssh.Client, clientName string, cloudlet *edgeproto.Cloudlet, chefParams *chefmgmt.ServerChefParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupChefOnServer", "clientName", clientName)

	err := pc.WriteFile(sshClient, "/etc/chef/client.pem", cloudlet.ChefClientKey[clientName], "chef-key", pc.SudoOn)
	if err != nil {
		return fmt.Errorf("Failed to write chef client key: %v", err)
	}
	// note the client name is b.tually used b. the node name
	command := fmt.Sprintf("sudo chef-client --chef-license ACCEPT -S %s -N %s -l debug -i 60 -d -L /var/log/chef.log", chefParams.ServerPath, clientName)
	log.SpanLog(ctx, log.DebugLevelInfra, "Running chef-client", "command", command)

	out, err := sshClient.Output(command)
	if err != nil {
		return fmt.Errorf("Failed to run chef client: %s - %v", out, err)
	}
	return nil
}

func (b *BareMetalPlatform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateCloudlet TODO")
}

func (b *BareMetalPlatform) UpdateTrustPolicy(ctx context.Context, TrustPolicy *edgeproto.TrustPolicy) error {
	return fmt.Errorf("UpdateTrustPolicy TODO")
}

func (b *BareMetalPlatform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCloudlet")
	updateCallback(edgeproto.UpdateTask, "Deleting cloudlet")
	err := b.commonPf.InitCloudletSSHKeys(ctx, accessApi)
	if err != nil {
		return err
	}
	b.commonPf.PlatformConfig = infracommon.GetPlatformConfig(cloudlet, pfConfig, accessApi)
	if err := b.commonPf.InitInfraCommon(ctx, b.commonPf.PlatformConfig, baremetalProps); err != nil {
		return err
	}
	sshClient, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
	if err != nil {
		return fmt.Errorf("Failed to get ssh client to control host: %v", err)
	}

	updateCallback(edgeproto.UpdateTask, "Deleting Shared RootLB")
	sharedLbName := b.GetSharedLBName(ctx, &cloudlet.Key)
	lbInfo, err := b.GetLbInfo(ctx, sshClient, sharedLbName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to get shared LB info", "sharedLbName", sharedLbName, "err", err)
	} else {
		externalDev := b.GetExternalEthernetInterface()
		internalDev := b.GetInternalEthernetInterface()
		err = b.RemoveIp(ctx, sshClient, lbInfo.ExternalIpAddr, externalDev)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Remove IP Fail", "lbInfo.ExternalIpAddr", lbInfo.ExternalIpAddr)
		}
		err = b.RemoveIp(ctx, sshClient, lbInfo.InternalIpAddr, internalDev)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Remove IP Fail", "lbInfo.InternalIpAddr", lbInfo.InternalIpAddr)
		}
		err = b.DeleteLbInfo(ctx, sshClient, sharedLbName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "error deleting lbinfo", "err", err)
		}
	}

	updateCallback(edgeproto.UpdateTask, "Removing platform containers")
	platContainers := []string{chefmgmt.ServiceTypeCRM, chefmgmt.ServiceTypeShepherd, chefmgmt.ServiceTypeCloudletPrometheus}
	for _, p := range platContainers {
		out, err := sshClient.Output(fmt.Sprintf("sudo docker rm -f %s", p))
		if err != nil {
			if strings.Contains(err.Error(), "No such container") {
				log.SpanLog(ctx, log.DebugLevelInfra, "container does not exist", "plat", p)
			} else {
				return fmt.Errorf("Error removing platform service: %s - %s - %v", p, out, err)
			}
		}
	}
	// kill chef add other cleanup
	out, err := sshClient.Output(fmt.Sprintf("sudo pkill -9 chef-client"))
	log.SpanLog(ctx, log.DebugLevelInfra, "chef kill results", "out", out, "err", err)
	out, err = sshClient.Output(fmt.Sprintf("sudo rm -f /tmp/'Chef Infra Client.pid'"))
	log.SpanLog(ctx, log.DebugLevelInfra, "chef pid rm results", "out", out, "err", err)
	out, err = sshClient.Output(fmt.Sprintf("sudo rm -f /root/accesskey/*"))
	log.SpanLog(ctx, log.DebugLevelInfra, "accesskey rm results", "out", out, "err", err)
	return nil
}

func (b *BareMetalPlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList")
	var flavors []*edgeproto.FlavorInfo
	if b.caches == nil {
		log.WarnLog("flavor cache is nil")
		return nil, fmt.Errorf("Flavor cache is nil")
	}
	flavorkeys := make(map[edgeproto.FlavorKey]struct{})
	b.caches.FlavorCache.GetAllKeys(ctx, func(k *edgeproto.FlavorKey, modRev int64) {
		flavorkeys[*k] = struct{}{}
	})
	for k := range flavorkeys {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetFlavorList found flavor", "key", k)
		var flav edgeproto.Flavor
		if b.caches.FlavorCache.Get(&k, &flav) {
			var flavInfo edgeproto.FlavorInfo
			_, gpu := flav.OptResMap["gpu"]
			if gpu {
				// gpu not currently supported
				log.SpanLog(ctx, log.DebugLevelInfra, "skipping GPU flavor", "flav", flav)
				continue
			}
			flavInfo.Name = flav.Key.Name
			flavInfo.Vcpus = flav.Vcpus
			flavInfo.Ram = flav.Ram
			flavors = append(flavors, &flavInfo)
		} else {
			return nil, fmt.Errorf("fail to fetch flavor %s", k)
		}
	}
	return flavors, nil
}

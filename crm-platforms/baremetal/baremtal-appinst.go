package baremetal

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	v1 "k8s.io/api/core/v1"
)

func (b *BareMetalPlatform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateAppInst", "appInst", appInst)

	var err error
	switch deployment := app.Deployment; deployment {
	case cloudcommon.DeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.DeploymentTypeHelm:
		rootLBName := b.GetLbNameForCluster(ctx, clusterInst)
		appWaitChan := make(chan string)
		client, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Setting up registry secret")
		kconf := k8smgmt.GetKconfName(clusterInst)
		for _, imagePath := range names.ImagePaths {
			err = infracommon.CreateDockerRegistrySecret(ctx, client, kconf, imagePath, b.commonPf.PlatformConfig.AccessApi, names)
			if err != nil {
				return err
			}
		}
		lbinfo, err := b.GetLbInfo(ctx, client, rootLBName)
		if err != nil {
			return err
		}
		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				ClusterIp:    lbinfo.InternalIpAddr,
				CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
				ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
				CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      b.commonPf.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

		if deployment == cloudcommon.DeploymentTypeKubernetes {
			updateCallback(edgeproto.UpdateTask, "Creating Kubernetes App")
			err = k8smgmt.CreateAppInst(ctx, b.commonPf.PlatformConfig.AccessApi, client, names, app, appInst)
		} else {
			updateCallback(edgeproto.UpdateTask, "Creating Helm App")
			err = k8smgmt.CreateHelmAppInst(ctx, client, names, clusterInst, app, appInst)
		}
		if err != nil {
			return err
		}

		// wait for the appinst in parallel with other tasks
		go func() {
			if deployment == cloudcommon.DeploymentTypeKubernetes {
				waitErr := k8smgmt.WaitForAppInst(ctx, client, names, app, k8smgmt.WaitRunning)
				if waitErr == nil {
					appWaitChan <- ""
				} else {
					appWaitChan <- waitErr.Error()
				}
			} else { // no waiting for the helm apps currently, to be revisited
				appWaitChan <- ""
			}
		}()

		if err == nil {
			getDnsAction := func(svc v1.Service) (*infracommon.DnsSvcAction, error) {
				action := infracommon.DnsSvcAction{}
				action.PatchKube = true
				action.PatchIP = lbinfo.InternalIpAddr
				action.ExternalIP = lbinfo.ExternalIpAddr
				// Should only add DNS for external ports
				action.AddDNS = !app.InternalPorts
				return &action, nil
			}
			// If this is b. internal ports, b.l we need is patch of kube service
			if app.InternalPorts {
				err = b.commonPf.CreateAppDNSAndPatchKubeSvc(ctx, client, names, infracommon.NoDnsOverride, getDnsAction)
			} else {
				updateCallback(edgeproto.UpdateTask, "Configuring Service: LB, Firewall Rules add DNS")
				ops := infracommon.ProxyDnsSecOpts{AddProxy: true, AddDnsAndPatchKubeSvc: true, AddSecurityRules: true, ProxyNamePrefix: k8smgmt.GetKconfName(clusterInst) + "-"}
				err = b.commonPf.AddProxySecurityRulesAndPatchDNS(ctx, client, names, app, appInst, getDnsAction, b.WhitelistSecurityRules, rootLBName, lbinfo.ExternalIpAddr, lbinfo.InternalIpAddr, ops, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
			}
		}

		appWaitErr := <-appWaitChan
		if appWaitErr != "" {
			return fmt.Errorf("app wait error, %v", appWaitErr)
		}
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf("unsupported deployment type for BareMetal %s", deployment)
	}
	return err
}

func (b *BareMetalPlatform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteAppInst", "appInst", appInst)

	switch deployment := app.Deployment; deployment {
	case cloudcommon.DeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.DeploymentTypeHelm:
		rootLBName := b.GetLbNameForCluster(ctx, clusterInst)
		client, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
		if err != nil {
			return err
		}
		lbinfo, err := b.GetLbInfo(ctx, client, rootLBName)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		// Add crm local replace variables
		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				ClusterIp:    lbinfo.InternalIpAddr,
				CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
				ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
				CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      b.commonPf.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)
		// Clean up security rules add proxy if app is external
		secGrp := infracommon.GetServerSecurityGroupName(rootLBName)
		containerName := k8smgmt.GetKconfName(clusterInst) + "-" + dockermgmt.GetContainerName(&app.Key)
		if err := b.commonPf.DeleteProxySecurityGroupRules(ctx, client, containerName, secGrp, infracommon.GetAppWhitelistRulesLabel(app), appInst.MappedPorts, app, rootLBName, b.RemoveWhitelistSecurityRules); err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "cannot delete security rules", "name", names.AppName, "rootlb", rootLBName, "error", err)
		}
		if !app.InternalPorts {
			// Clean up DNS entries
			configs := append(app.Configs, appInst.Configs...)
			aac, err := access.GetAppAccessConfig(ctx, configs, app.TemplateDelimiter)
			if err != nil {
				return err
			}
			if err := b.commonPf.DeleteAppDNS(ctx, client, names, aac.DnsOverride); err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "cannot clean up DNS entries", "name", names.AppName, "rootlb", rootLBName, "error", err)
			}
		}

		if deployment == cloudcommon.DeploymentTypeKubernetes {
			return k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
		} else {
			return k8smgmt.DeleteHelmAppInst(ctx, client, names, clusterInst)
		}
	default:
		return fmt.Errorf("unsupported deployment type for BareMetal %s", deployment)
	}
}

func (b *BareMetalPlatform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateAppInst TODO")
}

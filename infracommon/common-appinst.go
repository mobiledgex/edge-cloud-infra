package infracommon

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	v1 "k8s.io/api/core/v1"
)

type ProxyDnsSecOpts struct {
	AddProxy              bool
	AddDnsAndPatchKubeSvc bool
	AddSecurityRules      bool
}

type DeploymentType string

const (
	RootLBVMDeployment   DeploymentType = "mexrootlb"
	UserVMDeployment     DeploymentType = "mexuservm"
	PlatformVMDeployment DeploymentType = "mexplatformvm"
	SharedCluster        DeploymentType = "sharedcluster"
)

type VMParams struct {
	VMName                   string
	FlavorName               string
	ExternalVolumeSize       uint64
	SharedVolumeSize         uint64
	ImageName                string
	ApplicationSecurityGroup string // access to application ports for VM or RootLB
	CloudletSecurityGroup    string // SSH access to RootLB for OAM/CRM
	NetworkName              string
	SubnetName               string
	VnicType                 string
	MEXRouterIP              string
	GatewayIP                string
	FloatingIPAddressID      string
	AuthPublicKey            string
	AccessPorts              []util.PortSpec
	DeploymentManifest       string
	Command                  string
	IsRootLB                 bool
	IsInternal               bool
	ComputeAvailabilityZone  string
	VolumeAvailabilityZone   string
	PrivacyPolicy            *edgeproto.PrivacyPolicy
}

func WithPublicKey(authPublicKey string) VMParamsOp {
	return func(vmp *VMParams) error {
		if authPublicKey == "" {
			return nil
		}
		convKey, err := util.ConvertPEMtoOpenSSH(authPublicKey)
		if err != nil {
			return err
		}
		vmp.AuthPublicKey = convKey
		return nil
	}
}

func WithAccessPorts(accessPorts string) VMParamsOp {
	return func(vmp *VMParams) error {
		if accessPorts == "" {
			return nil
		}
		parsedAccessPorts, err := util.ParsePorts(accessPorts)
		if err != nil {
			return err
		}
		for _, port := range parsedAccessPorts {
			endPort, err := strconv.ParseInt(port.EndPort, 10, 32)
			if err != nil {
				return err
			}
			if endPort == 0 {
				port.EndPort = port.Port
			}
			vmp.AccessPorts = append(vmp.AccessPorts, port)
		}
		return nil
	}
}

type VMParamsOp func(vmp *VMParams) error

func WithDeploymentManifest(deploymentManifest string) VMParamsOp {
	return func(vmp *VMParams) error {
		vmp.DeploymentManifest = deploymentManifest
		return nil
	}
}

func WithCommand(command string) VMParamsOp {
	return func(vmp *VMParams) error {
		vmp.Command = command
		return nil
	}
}

func WithComputeAvailabilityZone(az string) VMParamsOp {
	return func(vmp *VMParams) error {
		vmp.ComputeAvailabilityZone = az
		return nil
	}
}

func WithVolumeAvailabilityZone(az string) VMParamsOp {
	return func(vmp *VMParams) error {
		vmp.VolumeAvailabilityZone = az
		return nil
	}
}

func WithPrivacyPolicy(pp *edgeproto.PrivacyPolicy) VMParamsOp {
	return func(vmp *VMParams) error {
		vmp.PrivacyPolicy = pp
		return nil
	}
}

func (c *CommonPlatform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {

	var err error
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		rootLBName := c.GetRootLBNameForCluster(ctx, clusterInst)
		appWaitChan := make(chan string)

		client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Setting up registry secret")
		err = CreateDockerRegistrySecret(ctx, client, clusterInst, app, c.VaultConfig, names)
		if err != nil {
			return err
		}

		_, masterIP, masterIpErr := c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
		// Add crm local replace variables
		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				ClusterIp:    masterIP,
				CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
				ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
				CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      c.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			updateCallback(edgeproto.UpdateTask, "Creating Kubernetes App")
			err = k8smgmt.CreateAppInst(ctx, c.VaultConfig, client, names, app, appInst)
		} else {
			updateCallback(edgeproto.UpdateTask, "Creating Helm App")

			err = k8smgmt.CreateHelmAppInst(ctx, client, names, clusterInst, app, appInst)
		}
		if err != nil {
			return err
		}

		// wait for the appinst in parallel with other tasks
		go func() {
			if deployment == cloudcommon.AppDeploymentTypeKubernetes {
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

		// set up DNS
		var rootLBIPaddr *ServerIP
		if masterIpErr == nil {
			rootLBIPaddr, err = c.infraProvider.GetIPFromServerName(ctx, c.GetCloudletExternalNetwork(), rootLBName)
			if err == nil {
				getDnsAction := func(svc v1.Service) (*DnsSvcAction, error) {
					action := DnsSvcAction{}
					action.PatchKube = true
					action.PatchIP = masterIP
					action.ExternalIP = rootLBIPaddr.ExternalAddr
					// Should only add DNS for external ports
					action.AddDNS = !app.InternalPorts
					return &action, nil
				}
				// If this is an internal ports, all we need is patch of kube service
				if app.InternalPorts {
					err = c.CreateAppDNSAndPatchKubeSvc(ctx, client, names, NoDnsOverride, getDnsAction)
				} else {
					updateCallback(edgeproto.UpdateTask, "Configuring Service: LB, Firewall Rules and DNS")
					ops := ProxyDnsSecOpts{AddProxy: true, AddDnsAndPatchKubeSvc: true, AddSecurityRules: true}
					err = c.AddProxySecurityRulesAndPatchDNS(ctx, client, names, app, appInst, getDnsAction, rootLBName, cloudcommon.IPAddrAllInterfaces, masterIP, ops, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
				}
			}
		}
		appWaitErr := <-appWaitChan
		if appWaitErr != "" {
			return fmt.Errorf("app wait error, %v", appWaitErr)
		}
		if err != nil {
			return err
		}
	case cloudcommon.AppDeploymentTypeVM:
		imageName, err := cloudcommon.GetFileName(app.ImagePath)
		if err != nil {
			return err
		}

		err = c.infraProvider.AddAppImageIfNotPresent(ctx, app, updateCallback)
		if err != nil {
			return err
		}

		objName := cloudcommon.GetAppFQN(&app.Key)
		vmAppParams, err := c.infraProvider.GetVMParams(ctx,
			UserVMDeployment,
			objName,
			appInst.VmFlavor,
			appInst.ExternalVolumeSize,
			imageName,
			c.GetServerSecurityGroupName(objName),
			&clusterInst.Key.CloudletKey,
			WithPublicKey(app.AuthPublicKey),
			WithAccessPorts(app.AccessPorts),
			WithDeploymentManifest(app.DeploymentManifest),
			WithCommand(app.Command),
			WithComputeAvailabilityZone(appInst.AvailabilityZone),
			WithVolumeAvailabilityZone(c.GetCloudletVolumeAvailabilityZone()),
			WithPrivacyPolicy(privacyPolicy),
		)
		if err != nil {
			return fmt.Errorf("unable to get vm params: %v", err)
		}

		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				CloudletName: k8smgmt.NormalizeName(appInst.Key.ClusterInstKey.CloudletKey.Name),
				CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      c.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

		externalServerName := objName // which server provides external access, VM or LB
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
			rootLBname := objName + "-lb"
			externalServerName = rootLBname
			lbVMSpec, err := c.GetVMSpecForRootLB()
			if err != nil {
				return err
			}
			lbImage, err := c.infraProvider.AddCloudletImageIfNotPresent(ctx, c.PlatformConfig.CloudletVMImagePath, c.PlatformConfig.VMImageVersion, updateCallback)
			if err != nil {
				return err
			}
			vmLbParams, err := c.infraProvider.GetVMParams(ctx,
				RootLBVMDeployment,
				rootLBname,
				lbVMSpec.FlavorName,
				lbVMSpec.ExternalVolumeSize,
				lbImage,
				c.GetServerSecurityGroupName(rootLBname),
				&clusterInst.Key.CloudletKey,
				WithComputeAvailabilityZone(lbVMSpec.AvailabilityZone),
				WithVolumeAvailabilityZone(c.GetCloudletVolumeAvailabilityZone()),
				WithAccessPorts(app.AccessPorts),
			)
			if err != nil {
				return err
			}
			err = c.infraProvider.CreateAppVMWithRootLB(ctx, vmAppParams, vmLbParams, updateCallback)
			if err != nil {
				return err
			}
		} else {
			updateCallback(edgeproto.UpdateTask, "Deploying VM standalone")
			log.SpanLog(ctx, log.DebugLevelMexos, "Deploying VM", "stackName", objName, "flavor", appInst.VmFlavor, "ExternalVolumeSize", appInst.ExternalVolumeSize)
			err = c.infraProvider.CreateAppVM(ctx, vmAppParams, updateCallback)
			if err != nil {
				return err
			}
		}
		ip, err := c.infraProvider.GetIPFromServerName(ctx, c.GetCloudletExternalNetwork(), externalServerName)
		if err != nil {
			return err
		}
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
			updateCallback(edgeproto.UpdateTask, "Setting Up Load Balancer")
			var proxyOps []proxy.Op
			client, err := c.GetPlatformClientRootLB(ctx, externalServerName)
			if err != nil {
				return err
			}
			// clusterInst is empty but that is ok here
			names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
			if err != nil {
				return fmt.Errorf("get kube names failed: %s", err)
			}
			proxyOps = append(proxyOps, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
			getDnsAction := func(svc v1.Service) (*DnsSvcAction, error) {
				action := DnsSvcAction{}
				action.PatchKube = false
				action.ExternalIP = ip.ExternalAddr
				return &action, nil
			}
			vmIP, err := c.infraProvider.GetIPFromServerName(ctx, c.GetCloudletMexNetwork(), objName)
			if err != nil {
				return err
			}
			updateCallback(edgeproto.UpdateTask, "Configuring Firewall Rules")
			ops := ProxyDnsSecOpts{AddProxy: true, AddDnsAndPatchKubeSvc: false, AddSecurityRules: false}
			err = c.AddProxySecurityRulesAndPatchDNS(ctx, client, names, app, appInst, getDnsAction, externalServerName, cloudcommon.IPAddrAllInterfaces, vmIP.ExternalAddr, ops, proxyOps...)
			if err != nil {
				return fmt.Errorf("AddProxySecurityRulesAndPatchDNS error: %v", err)
			}
		}
		updateCallback(edgeproto.UpdateTask, "Adding DNS Entry")
		if appInst.Uri != "" && ip.ExternalAddr != "" {
			fqdn := appInst.Uri
			configs := append(app.Configs, appInst.Configs...)
			aac, err := access.GetAppAccessConfig(ctx, configs)
			if err != nil {
				return err
			}
			if aac.DnsOverride != "" {
				fqdn = aac.DnsOverride
			}
			if err = c.ActivateFQDNA(ctx, fqdn, ip.ExternalAddr); err != nil {
				return err
			}
			log.SpanLog(ctx, log.DebugLevelMexos, "DNS A record activated",
				"name", objName,
				"fqdn", fqdn,
				"IP", ip.ExternalAddr)
		}
		return nil

	case cloudcommon.AppDeploymentTypeDocker:
		rootLBName := c.GetRootLBNameForCluster(ctx, clusterInst)
		backendIP := cloudcommon.RemoteServerNone
		dockerNetworkMode := dockermgmt.DockerBridgeMode
		rootLBClient, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		// docker commands can be run on either the rootlb or on the docker
		// vm.  The default is to run on the rootlb client
		dockerCommandTarget := rootLBClient

		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			log.SpanLog(ctx, log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
			if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
				backendIP = cloudcommon.IPAddrDockerHost
			} else {
				dockerNetworkMode = dockermgmt.DockerHostMode
			}
		} else {
			// Shared access uses a separate VM for docker.  This is used both for running the docker commands
			// and as the backend ip for the proxy
			_, backendIP, err = c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
			if err != nil {
				return err
			}
			// docker command will run on the docker vm
			dockerCommandTarget, err = rootLBClient.AddHop(backendIP, 22)
			if err != nil {
				return err
			}
			dockerNetworkMode = dockermgmt.DockerHostMode
		}

		rootLBIPaddr, err := c.infraProvider.GetIPFromServerName(ctx, c.GetCloudletExternalNetwork(), rootLBName)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed, %v", err)
		}
		updateCallback(edgeproto.UpdateTask, "Seeding docker secret")
		err = SeedDockerSecret(ctx, dockerCommandTarget, clusterInst, app, c.VaultConfig)
		if err != nil {
			return fmt.Errorf("seeding docker secret failed, %v", err)
		}
		updateCallback(edgeproto.UpdateTask, "Deploying Docker App")

		err = dockermgmt.CreateAppInst(ctx, c.VaultConfig, dockerCommandTarget, app, appInst, dockerNetworkMode)
		if err != nil {
			return err
		}
		getDnsAction := func(svc v1.Service) (*DnsSvcAction, error) {
			action := DnsSvcAction{}
			action.PatchKube = false
			action.ExternalIP = rootLBIPaddr.ExternalAddr
			return &action, nil
		}
		updateCallback(edgeproto.UpdateTask, "Configuring Firewall Rules and DNS")
		var proxyOps []proxy.Op
		addproxy := false
		listenIP := "NONE" // only applicable for proxy case
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
			proxyOps = append(proxyOps, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
			addproxy = true
			listenIP = rootLBIPaddr.InternalAddr
		}
		ops := ProxyDnsSecOpts{AddProxy: addproxy, AddDnsAndPatchKubeSvc: true, AddSecurityRules: true}
		err = c.AddProxySecurityRulesAndPatchDNS(ctx, rootLBClient, names, app, appInst, getDnsAction, rootLBName, listenIP, backendIP, ops, proxyOps...)
		if err != nil {
			return fmt.Errorf("AddProxySecurityRulesAndPatchDNS error: %v", err)
		}
	default:
		err = fmt.Errorf("unsupported deployment type %s", deployment)
	}
	return err

}

func (c *CommonPlatform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		rootLBName := c.GetRootLBNameForCluster(ctx, clusterInst)
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			log.SpanLog(ctx, log.DebugLevelMexos, "using dedicated RootLB to delete app", "rootLBName", rootLBName)
			_, err := c.infraProvider.GetServerDetail(ctx, rootLBName)
			if err != nil {
				if strings.Contains(err.Error(), ServerDoesNotExistError) {
					log.SpanLog(ctx, log.DebugLevelMexos, "Dedicated RootLB is gone, allow app deletion")
					return nil
				}
				return err
			}
		}
		client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		_, masterIP, err := c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
		if err != nil {
			if strings.Contains(err.Error(), ServerDoesNotExistError) {
				log.SpanLog(ctx, log.DebugLevelMexos, "cluster is gone, allow app deletion")
				secGrp := c.GetServerSecurityGroupName(rootLBName)
				c.DeleteProxySecurityGroupRules(ctx, client, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName)
				return nil
			}
		}
		// Add crm local replace variables
		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				ClusterIp:    masterIP,
				CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
				ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
				CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      c.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

		// Clean up security rules and proxy if app is external
		secGrp := c.GetServerSecurityGroupName(rootLBName)
		if err := c.DeleteProxySecurityGroupRules(ctx, client, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName); err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete security rules", "name", names.AppName, "rootlb", rootLBName, "error", err)
		}
		if !app.InternalPorts {
			// Clean up DNS entries
			configs := append(app.Configs, appInst.Configs...)
			aac, err := access.GetAppAccessConfig(ctx, configs)
			if err != nil {
				return err
			}
			if err := c.DeleteAppDNS(ctx, client, names, aac.DnsOverride); err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "cannot clean up DNS entries", "name", names.AppName, "rootlb", rootLBName, "error", err)
			}
		}

		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			return k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
		} else {
			return k8smgmt.DeleteHelmAppInst(ctx, client, names, clusterInst)
		}

	case cloudcommon.AppDeploymentTypeVM:
		objName := cloudcommon.GetAppFQN(&app.Key)
		log.SpanLog(ctx, log.DebugLevelMexos, "Deleting VM", "stackName", objName)
		err := c.infraProvider.DeleteResources(ctx, objName)
		if err != nil {
			return fmt.Errorf("DeleteVMAppInst error: %v", err)
		}
		if appInst.Uri != "" {
			fqdn := appInst.Uri
			configs := append(app.Configs, appInst.Configs...)
			aac, err := access.GetAppAccessConfig(ctx, configs)
			if err != nil {
				return err
			}
			if aac.DnsOverride != "" {
				fqdn = aac.DnsOverride
			}
			if err = c.DeleteDNSRecords(ctx, fqdn); err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete DNS entries", "fqdn", fqdn)
			}
		}
		return nil

	case cloudcommon.AppDeploymentTypeDocker:
		rootLBName := c.GetRootLBNameForCluster(ctx, clusterInst)
		rootLBClient, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		// docker commands can be run on either the rootlb or on the docker
		// vm.  The default is to run on the rootlb client
		dockerCommandTarget := rootLBClient

		if clusterInst.IpAccess != edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			_, backendIP, err := c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
			if err != nil {
				if strings.Contains(err.Error(), ServerDoesNotExistError) {
					log.SpanLog(ctx, log.DebugLevelMexos, "cluster is gone, allow app deletion")
					secGrp := c.GetServerSecurityGroupName(rootLBName)
					c.DeleteProxySecurityGroupRules(ctx, rootLBClient, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName)
					return nil
				}
				return err
			}
			// docker command will run on the docker vm
			dockerCommandTarget, err = rootLBClient.AddHop(backendIP, 22)
			if err != nil {
				return err
			}
		}
		_, err = c.infraProvider.GetServerDetail(ctx, rootLBName)
		if err != nil {
			if strings.Contains(err.Error(), ServerDoesNotExistError) {
				log.SpanLog(ctx, log.DebugLevelMexos, "Dedicated RootLB is gone, allow app deletion")
				return nil
			}
			return err
		}
		client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		name := dockermgmt.GetContainerName(&app.Key)
		if !app.InternalPorts {
			secGrp := c.GetServerSecurityGroupName(rootLBName)
			//  the proxy does not yet exist for docker, but it eventually will.  Secgrp rules should be deleted in either case
			if err := c.DeleteProxySecurityGroupRules(ctx, client, name, secGrp, appInst.MappedPorts, app, rootLBName); err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete security rules", "name", name, "rootlb", rootLBName, "error", err)
			}
		}

		return dockermgmt.DeleteAppInst(ctx, c.VaultConfig, dockerCommandTarget, app, appInst)
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}

}

func (c *CommonPlatform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	_, masterIP, _ := c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
	// Add crm local replace variables
	deploymentVars := crmutil.DeploymentReplaceVars{
		Deployment: crmutil.CrmReplaceVars{
			ClusterIp:    masterIP,
			ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
			CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
			AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
			DnsZone:      c.GetCloudletDNSZone(),
		},
	}
	ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		return k8smgmt.UpdateAppInst(ctx, c.VaultConfig, client, names, app, appInst)
	case cloudcommon.AppDeploymentTypeDocker:
		dockerNetworkMode := dockermgmt.DockerBridgeMode
		rootLBClient, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		// docker commands can be run on either the rootlb or on the docker
		// vm.  The default is to run on the rootlb client
		dockerCommandTarget := rootLBClient

		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
			_, backendIP, err := c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
			if err != nil {
				return err
			}
			// docker command will run on the docker vm
			dockerCommandTarget, err = rootLBClient.AddHop(backendIP, 22)
			if err != nil {
				return err
			}
		}
		return dockermgmt.UpdateAppInst(ctx, c.VaultConfig, dockerCommandTarget, app, appInst, dockerNetworkMode)
	case cloudcommon.AppDeploymentTypeHelm:
		client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		return k8smgmt.UpdateHelmAppInst(ctx, client, names, app, appInst)

	default:
		return fmt.Errorf("UpdateAppInst not supported for deployment: %s", app.Deployment)
	}
}

func (c *CommonPlatform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return nil, err
	}

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return nil, err
		}
		return k8smgmt.GetAppInstRuntime(ctx, client, names, app, appInst)
	case cloudcommon.AppDeploymentTypeDocker:
		return dockermgmt.GetAppInstRuntime(ctx, client, app, appInst)
	case cloudcommon.AppDeploymentTypeVM:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

func (c *CommonPlatform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		return k8smgmt.GetContainerCommand(ctx, clusterInst, app, appInst, req)
	case cloudcommon.AppDeploymentTypeDocker:
		return dockermgmt.GetContainerCommand(clusterInst, app, appInst, req)
	case cloudcommon.AppDeploymentTypeVM:
		fallthrough
	default:
		return "", fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

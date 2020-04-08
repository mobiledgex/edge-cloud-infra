package openstack

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *OpenstackPlatform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	/*
		var err error
		switch deployment := app.Deployment; deployment {
		case cloudcommon.AppDeploymentTypeKubernetes:
			fallthrough
		case cloudcommon.AppDeploymentTypeHelm:
			rootLBName := o.commonPf.RootLBName
			appWaitChan := make(chan string)

			if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				rootLBName = cloudcommon.GetDedicatedLBFQDN(o.cloudletKey, &clusterInst.Key.ClusterKey)
				log.SpanLog(ctx, log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
			}
			client, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
			if err != nil {
				return err
			}
			updateCallback(edgeproto.UpdateTask, "Setting up registry secret")
			err = infracommon.CreateDockerRegistrySecret(ctx, client, clusterInst, app, o.vaultConfig, names)
			if err != nil {
				return err
			}

			_, masterIP, masterIpErr := o.GetMasterNameAndIP(ctx, clusterInst)
			// Add crm local replace variables
			deploymentVars := crmutil.DeploymentReplaceVars{
				Deployment: crmutil.CrmReplaceVars{
					ClusterIp:    masterIP,
					CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
					ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
					CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
					AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
					DnsZone:      o.commonPf.GetCloudletDNSZone(),
				},
			}
			ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

			if deployment == cloudcommon.AppDeploymentTypeKubernetes {
				updateCallback(edgeproto.UpdateTask, "Creating Kubernetes App")
				err = k8smgmt.CreateAppInst(ctx, o.vaultConfig, client, names, app, appInst)
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
			var rootLBIPaddr *infracommon.ServerIP
			if masterIpErr == nil {
				rootLBIPaddr, err = o.GetServerIPAddr(ctx, o.GetCloudletExternalNetwork(), rootLBName)
				if err == nil {
					getDnsAction := func(svc v1.Service) (*infracommon.DnsSvcAction, error) {
						action := infracommon.DnsSvcAction{}
						action.PatchKube = true
						action.PatchIP = masterIP
						action.ExternalIP = rootLBIPaddr.ExternalAddr
						// Should only add DNS for external ports
						action.AddDNS = !app.InternalPorts
						return &action, nil
					}
					// If this is an internal ports, all we need is patch of kube service
					if app.InternalPorts {
						err = o.commonPf.CreateAppDNSAndPatchKubeSvc(ctx, client, names, infracommon.NoDnsOverride, getDnsAction)
					} else {
						updateCallback(edgeproto.UpdateTask, "Configuring Service: LB, Firewall Rules and DNS")
						ops := ProxyDnsSecOpts{AddProxy: true, AddDnsAndPatchKubeSvc: true, AddSecurityRules: true}
						err = o.AddProxySecurityRulesAndPatchDNS(ctx, client, names, app, appInst, getDnsAction, rootLBName, cloudcommon.IPAddrAllInterfaces, masterIP, ops, o.vaultConfig, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
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
			sourceImageTime, md5Sum, err := infracommon.GetUrlInfo(ctx, o.vaultConfig, app.ImagePath)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "failed to fetch source image info, skip image validity checks")
			}
			imageDetail, err := o.GetImageDetail(ctx, imageName)
			createImage := false
			if err != nil {
				if strings.Contains(err.Error(), "Could not find resource") {
					// Add image to Glance
					log.SpanLog(ctx, log.DebugLevelMexos, "image is not present in glance, add image")
					createImage = true
				} else {
					return err
				}
			} else {
				if imageDetail.Status != "active" {
					return fmt.Errorf("image in store %s is not active", imageName)
				}
				if imageDetail.Checksum != md5Sum {
					if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW && imageDetail.DiskFormat == infracommon.ImageFormatVmdk {
						log.SpanLog(ctx, log.DebugLevelMexos, "image was imported as vmdk, checksum match not possible")
					} else {
						return fmt.Errorf("mismatch in md5sum for image in glance: %s", imageName)
					}
				}
				glanceImageTime, err := time.Parse(time.RFC3339, imageDetail.UpdatedAt)
				if err != nil {
					return err
				}
				if !sourceImageTime.IsZero() {
					if sourceImageTime.Sub(glanceImageTime) > 0 {
						// Update the image in Glance
						updateCallback(edgeproto.UpdateTask, "Image in store is outdated, deleting old image")
						err = o.DeleteImage(ctx, imageName)
						if err != nil {
							return err
						}
						createImage = true
					}
				}
			}
			if createImage {
				updateCallback(edgeproto.UpdateTask, "Creating VM Image from URL")
				err = o.CreateImageFromUrl(ctx, imageName, app.ImagePath, md5Sum)
				if err != nil {
					return err
				}
			}

			objName := cloudcommon.GetAppFQN(&app.Key)
			vmAppParams, err := o.GetVMParams(ctx,
				UserVMDeployment,
				objName,
				appInst.VmFlavor,
				appInst.ExternalVolumeSize,
				imageName,
				GetSecurityGroupName(ctx, objName),
				&clusterInst.Key.CloudletKey,
				WithPublicKey(app.AuthPublicKey),
				WithAccessPorts(app.AccessPorts),
				WithDeploymentManifest(app.DeploymentManifest),
				WithCommand(app.Command),
				WithComputeAvailabilityZone(appInst.AvailabilityZone),
				WithVolumeAvailabilityZone(o.GetCloudletVolumeAvailabilityZone()),
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
					DnsZone:      o.commonPf.GetCloudletDNSZone(),
				},
			}
			ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

			externalServerName := objName // which server provides external access, VM or LB
			if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
				rootLBname := objName + "-lb"
				externalServerName = rootLBname
				rootLBImage := o.GetCloudletOSImage()
				lbVMSpec, err := o.GetVMSpecForRootLB()
				if err != nil {
					return err
				}
				lbImage, err := o.AddImageIfNotPresent(ctx, o.config.CloudletVMImagePath, o.config.VMImageVersion, updateCallback)
				if err != nil {
					return err
				}
				vmLbParams, err := o.GetVMParams(ctx,
					RootLBVMDeployment,
					rootLBname,
					lbVMSpec.FlavorName,
					lbVMSpec.ExternalVolumeSize,
					lbImage,
					GetSecurityGroupName(ctx, rootLBname),
					&clusterInst.Key.CloudletKey,
					WithComputeAvailabilityZone(lbVMSpec.AvailabilityZone),
					WithVolumeAvailabilityZone(o.GetCloudletVolumeAvailabilityZone()),
					WithAccessPorts(app.AccessPorts),
				)
				if err != nil {
					return err
				}
				err = o.HeatCreateAppVMWithRootLB(ctx, rootLBname, rootLBImage, objName, vmAppParams, vmLbParams, updateCallback)
				if err != nil {
					return err
				}
			} else {
				updateCallback(edgeproto.UpdateTask, "Deploying VM standalone")
				log.SpanLog(ctx, log.DebugLevelMexos, "Deploying VM", "stackName", objName, "flavor", appInst.VmFlavor, "ExternalVolumeSize", appInst.ExternalVolumeSize)
				err = o.CreateHeatStackFromTemplate(ctx, vmAppParams, objName, VmTemplate, updateCallback)
				if err != nil {
					return err
				}
			}
			ip, err := o.GetServerIPAddr(ctx, o.GetCloudletExternalNetwork(), externalServerName)
			if err != nil {
				return err
			}
			if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
				updateCallback(edgeproto.UpdateTask, "Setting Up Load Balancer")
				var proxyOps []proxy.Op
				client, err := o.GetPlatformClientRootLB(ctx, externalServerName)
				if err != nil {
					return err
				}
				// clusterInst is empty but that is ok here
				names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
				if err != nil {
					return fmt.Errorf("get kube names failed: %s", err)
				}
				proxyOps = append(proxyOps, proxy.WithDockerPublishPorts(), proxy.WithDockerNetwork(""))
				getDnsAction := func(svc v1.Service) (*infracommon.DnsSvcAction, error) {
					action := infracommon.DnsSvcAction{}
					action.PatchKube = false
					action.ExternalIP = ip.ExternalAddr
					return &action, nil
				}
				vmIP, err := o.GetServerIPAddr(ctx, o.GetCloudletMexNetwork(), objName)
				if err != nil {
					return err
				}
				updateCallback(edgeproto.UpdateTask, "Configuring Firewall Rules")
				ops := ProxyDnsSecOpts{AddProxy: true, AddDnsAndPatchKubeSvc: false, AddSecurityRules: false}
				err = o.AddProxySecurityRulesAndPatchDNS(ctx, client, names, app, appInst, getDnsAction, externalServerName, cloudcommon.IPAddrAllInterfaces, vmIP.ExternalAddr, ops, o.vaultConfig, proxyOpo...)
				if err != nil {
					return fmt.Errorf("AddProxySecurityRulesAndPatchDNS error: %v", err)
				}
			}
			updateCallback(edgeproto.UpdateTask, "Adding DNS Entry")
			if appInst.Uri != "" && ip.ExternalAddr != "" {
				fqdn := appInst.Uri
				configs := append(app.Configs, appInst.Configo...)
				aac, err := acceso.GetAppAccessConfig(ctx, configs)
				if err != nil {
					return err
				}
				if aac.DnsOverride != "" {
					fqdn = aac.DnsOverride
				}
				if err = o.commonPf.ActivateFQDNA(ctx, fqdn, ip.ExternalAddr); err != nil {
					return err
				}
				log.SpanLog(ctx, log.DebugLevelMexos, "DNS A record activated",
					"name", objName,
					"fqdn", fqdn,
					"IP", ip.ExternalAddr)
			}
			return nil

		case cloudcommon.AppDeploymentTypeDocker:
			rootLBName := o.rootLBName
			backendIP := cloudcommon.RemoteServerNone
			dockerNetworkMode := dockermgmt.DockerBridgeMode
			rootLBClient, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			// docker commands can be run on either the rootlb or on the docker
			// vm.  The default is to run on the rootlb client
			dockerCommandTarget := rootLBClient

			if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				rootLBName = cloudcommon.GetDedicatedLBFQDN(o.cloudletKey, &clusterInst.Key.ClusterKey)
				log.SpanLog(ctx, log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
				if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
					backendIP = cloudcommon.IPAddrDockerHost
				} else {
					dockerNetworkMode = dockermgmt.DockerHostMode
				}
			} else {
				// Shared access uses a separate VM for docker.  This is used both for running the docker commands
				// and as the backend ip for the proxy
				_, backendIP, err = o.GetMasterNameAndIP(ctx, clusterInst)
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

			rootLBIPaddr, err := o.GetServerIPAddr(ctx, o.GetCloudletExternalNetwork(), rootLBName)
			if err != nil {
				return err
			}
			names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
			if err != nil {
				return fmt.Errorf("get kube names failed, %v", err)
			}
			updateCallback(edgeproto.UpdateTask, "Seeding docker secret")
			err = infracommon.SeedDockerSecret(ctx, s, dockerCommandTarget, clusterInst, app, o.vaultConfig)
			if err != nil {
				return fmt.Errorf("seeding docker secret failed, %v", err)
			}
			updateCallback(edgeproto.UpdateTask, "Deploying Docker App")

			err = dockermgmt.CreateAppInst(ctx, o.vaultConfig, dockerCommandTarget, app, appInst, dockerNetworkMode)
			if err != nil {
				return err
			}
			getDnsAction := func(svc v1.Service) (*infracommon.DnsSvcAction, error) {
				action := infracommon.DnsSvcAction{}
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
			err = o.AddProxySecurityRulesAndPatchDNS(ctx, rootLBClient, names, app, appInst, getDnsAction, rootLBName, listenIP, backendIP, ops, o.vaultConfig, proxyOpo...)
			if err != nil {
				return fmt.Errorf("AddProxySecurityRulesAndPatchDNS error: %v", err)
			}
		default:
			err = fmt.Errorf("unsupported deployment type %s", deployment)
		}
		return err
	*/
	return fmt.Errorf("TODO")
}

func (s *OpenstackPlatform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	return fmt.Errorf("TODO")
	/*

		switch deployment := app.Deployment; deployment {
		case cloudcommon.AppDeploymentTypeKubernetes:
			fallthrough
		case cloudcommon.AppDeploymentTypeHelm:
			rootLBName := o.rootLBName
			if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				rootLBName = cloudcommon.GetDedicatedLBFQDN(o.cloudletKey, &clusterInst.Key.ClusterKey)
				log.SpanLog(ctx, log.DebugLevelMexos, "using dedicated RootLB to delete app", "rootLBName", rootLBName)
				_, err := o.GetActiveServerDetails(ctx, rootLBName)
				if err != nil {
					if strings.Contains(err.Error(), "No server with a name or ID") {
						log.SpanLog(ctx, log.DebugLevelMexos, "Dedicated RootLB is gone, allow app deletion")
						return nil
					}
					return err
				}
			}
			client, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
			if err != nil {
				return fmt.Errorf("get kube names failed: %s", err)
			}
			_, masterIP, err := o.GetMasterNameAndIP(ctx, clusterInst)
			if err != nil {
				if strings.Contains(err.Error(), infracommon.ClusterNotFoundErr) {
					log.SpanLog(ctx, log.DebugLevelMexos, "cluster is gone, allow app deletion")
					secGrp := GetSecurityGroupName(ctx, rootLBName)
					o.DeleteProxySecurityGroupRules(ctx, client, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName)
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
					DnsZone:      o.commonPf.GetCloudletDNSZone(),
				},
			}
			ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

			// Clean up security rules and proxy if app is external
			secGrp := GetSecurityGroupName(ctx, rootLBName)
			if err := o.DeleteProxySecurityGroupRules(ctx, client, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName); err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete security rules", "name", edgeproto.OptResNamesNAS.AppName, "rootlb", rootLBName, "error", err)
			}
			if !app.InternalPorts {
				// Clean up DNS entries
				configs := append(app.Configs, appInst.Configo...)
				aac, err := access.GetAppAccessConfig(ctx, configs)
				if err != nil {
					return err
				}
				if err := o.commonPf.DeleteAppDNS(ctx, client, names, aac.DnsOverride); err != nil {
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
			err := o.HeatDeleteStack(ctx, objName)
			if err != nil {
				return fmt.Errorf("DeleteVMAppInst error: %v", err)
			}
			if appInst.Uri != "" {
				fqdn := appInst.Uri
				configs := append(app.Configs, appInst.Configs...)
				aac, err := acceso.GetAppAccessConfig(ctx, configs)
				if err != nil {
					return err
				}
				if aac.DnsOverride != "" {
					fqdn = aac.DnsOverride
				}
				if err = o.commonPf.DeleteDNSRecords(ctx, fqdn); err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete DNS entries", "fqdn", fqdn)
				}
			}
			return nil

		case cloudcommon.AppDeploymentTypeDocker:
			rootLBName := o.rootLBName
			rootLBClient, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			// docker commands can be run on either the rootlb or on the docker
			// vm.  The default is to run on the rootlb client
			dockerCommandTarget := rootLBClient

			if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
				rootLBName = cloudcommon.GetDedicatedLBFQDN(o.cloudletKey, &clusterInst.Key.ClusterKey)
			} else {
				_, backendIP, err := o.GetMasterNameAndIP(ctx, clusterInst)
				if err != nil {
					if strings.Contains(err.Error(), infracommon.ClusterNotFoundErr) {
						log.SpanLog(ctx, log.DebugLevelMexos, "cluster is gone, allow app deletion")
						secGrp := GetSecurityGroupName(ctx, rootLBName)
						o.DeleteProxySecurityGroupRules(ctx, rootLBClient, dockermgmt.GetContainerName(&app.Key), secGrp, appInst.MappedPorts, app, rootLBName)
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
			_, err = o.GetActiveServerDetails(ctx, rootLBName)
			if err != nil {
				if strings.Contains(err.Error(), "No server with a name or ID") {
					log.SpanLog(ctx, log.DebugLevelMexos, "Dedicated RootLB is gone, allow app deletion")
					return nil
				}
				return err
			}
			client, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			name := dockermgmt.GetContainerName(&app.Key)
			if !app.InternalPorts {
				secGrp := GetSecurityGroupName(ctx, rootLBName)
				//  the proxy does not yet exist for docker, but it eventually will.  Secgrp rules should be deleted in either case
				if err := o.DeleteProxySecurityGroupRules(ctx, client, name, secGrp, appInst.MappedPorts, app, rootLBName); err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete security rules", "name", name, "rootlb", rootLBName, "error", err)
				}
			}

			return dockermgmt.DeleteAppInst(ctx, o.vaultConfig, dockerCommandTarget, app, appInst)
		default:
			return fmt.Errorf("unsupported deployment type %s", deployment)
		}
	*/
}

func (s *OpenstackPlatform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("TODO")
	/*
		_, masterIP, _ := o.GetMasterNameAndIP(ctx, clusterInst)
		// Add crm local replace variables
		deploymentVars := crmutil.DeploymentReplaceVars{
			Deployment: crmutil.CrmReplaceVars{
				ClusterIp:    masterIP,
				ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
				CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
				AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
				DnsZone:      o.commonPf.GetCloudletDNSZone(),
			},
		}
		ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

		switch deployment := app.Deployment; deployment {
		case cloudcommon.AppDeploymentTypeKubernetes:
			client, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
			if err != nil {
				return fmt.Errorf("get kube names failed: %s", err)
			}
			return k8smgmt.UpdateAppInst(ctx, o.vaultConfig, client, names, app, appInst)
		case cloudcommon.AppDeploymentTypeDocker:
			dockerNetworkMode := dockermgmt.DockerBridgeMode
			rootLBClient, err := o.GetPlatformClient(ctx, clusterInst)
			if err != nil {
				return err
			}
			// docker commands can be run on either the rootlb or on the docker
			// vm.  The default is to run on the rootlb client
			dockerCommandTarget := rootLBClient

			if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
				_, backendIP, err := o.GetMasterNameAndIP(ctx, clusterInst)
				if err != nil {
					return err
				}
				// docker command will run on the docker vm
				dockerCommandTarget, err = rootLBClient.AddHop(backendIP, 22)
				if err != nil {
					return err
				}
			}
			return dockermgmt.UpdateAppInst(ctx, o.vaultConfig, dockerCommandTarget, app, appInst, dockerNetworkMode)
		case cloudcommon.AppDeploymentTypeHelm:
			client, err := o.GetPlatformClient(ctx, clusterInst)
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
	*/
}

func (o *OpenstackPlatform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	client, err := o.GetPlatformClient(ctx, clusterInst)
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

func (s *OpenstackPlatform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
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

func (o *OpenstackPlatform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeVM:
		objName := cloudcommon.GetAppFQN(&app.Key)
		consoleUrl, err := o.OSGetConsoleUrl(ctx, objName)
		if err != nil {
			return "", err
		}
		return consoleUrl.Url, nil
	default:
		return "", fmt.Errorf("unsupported deployment type %s", deployment)
	}

}

func (o *OpenstackPlatform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	PowerState := appInst.PowerState
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeVM:
		serverName := cloudcommon.GetAppFQN(&app.Key)
		fqdn := appInst.Uri

		log.SpanLog(ctx, log.DebugLevelMexos, "setting server state", "serverName", serverName, "fqdn", fqdn, "PowerState", PowerState)

		updateCallback(edgeproto.UpdateTask, "Verifying AppInst state")
		serverDetail, err := o.GetServerDetails(ctx, serverName)
		if err != nil {
			return err
		}

		serverAction := ""
		switch PowerState {
		case edgeproto.PowerState_POWER_ON_REQUESTED:
			if serverDetail.Status == "ACTIVE" {
				return fmt.Errorf("server %s is already active", serverName)
			}
			serverAction = "start"
		case edgeproto.PowerState_POWER_OFF_REQUESTED:
			if serverDetail.Status == "SHUTOFF" {
				return fmt.Errorf("server %s is already stopped", serverName)
			}
			serverAction = "stop"
		case edgeproto.PowerState_REBOOT_REQUESTED:
			serverAction = "reboot"
			if serverDetail.Status != "ACTIVE" {
				return fmt.Errorf("server %s is not active", serverName)
			}
		default:
			return fmt.Errorf("unsupported server power action: %s", PowerState)
		}

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
		oldServerIP, err := infracommon.GetServerIPFromAddrs(ctx, o.commonPf.GetCloudletExternalNetwork(), serverDetail.Addresses, serverName)
		if err != nil || oldServerIP.ExternalAddr == "" {
			return fmt.Errorf("unable to fetch external ip for %s, addr %s, err %v", serverName, serverDetail.Addresses, err)
		}

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Performing action %s on %s", serverAction, serverName))
		err = o.OSSetPowerState(ctx, serverName, serverAction)
		if err != nil {
			return err
		}

		if PowerState == edgeproto.PowerState_POWER_ON_REQUESTED || PowerState == edgeproto.PowerState_REBOOT_REQUESTED {
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Waiting for server %s to become active", serverName))
			serverDetail, err := o.GetActiveServerDetails(ctx, serverName)
			if err != nil {
				return err
			}
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching external address of %s", serverName))
			newServerIP, err := infracommon.GetServerIPFromAddrs(ctx, o.commonPf.GetCloudletExternalNetwork(), serverDetail.Addresses, serverName)
			if err != nil || newServerIP.ExternalAddr == "" {
				return fmt.Errorf("unable to fetch external ip for %s, addr %s, err %v", serverName, serverDetail.Addresses, err)
			}
			if oldServerIP.ExternalAddr != newServerIP.ExternalAddr {
				// IP changed, update DNS entry
				updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Updating DNS entry as IP changed for %s", serverName))
				log.SpanLog(ctx, log.DebugLevelMexos, "updating DNS entry", "serverName", serverName, "fqdn", fqdn, "ip", newServerIP)
				err = o.commonPf.ActivateFQDNA(ctx, fqdn, newServerIP.ExternalAddr)
				if err != nil {
					return fmt.Errorf("unable to update fqdn for %s, addr %s, err %v", serverName, newServerIP.ExternalAddr, err)
				}
			}
		}
		updateCallback(edgeproto.UpdateTask, "Performed power control action successfully")
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}

	return nil
}

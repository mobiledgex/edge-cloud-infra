package openstack

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/nginx"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/flavor"
	"github.com/mobiledgex/edge-cloud/log"
	v1 "k8s.io/api/core/v1"
)

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		rootLBName := s.rootLBName
		appWaitChan := make(chan string)

		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
			log.SpanLog(s.ctx, log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
		}
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Setting up registry secret")
		err = mexos.CreateDockerRegistrySecret(client, clusterInst, app, s.config.VaultAddr)
		if err != nil {
			return err
		}

		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			updateCallback(edgeproto.UpdateTask, "Creating Kubernetes App")
			err = k8smgmt.CreateAppInst(client, names, app, appInst)
		} else {
			updateCallback(edgeproto.UpdateTask, "Creating Helm App")

			err = k8smgmt.CreateHelmAppInst(client, names, clusterInst, app, appInst)
		}

		// wait for the appinst in parallel with other tasks
		go func() {
			if deployment == cloudcommon.AppDeploymentTypeKubernetes {
				waitErr := k8smgmt.WaitForAppInst(client, names, app, k8smgmt.WaitRunning)
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
		var rootLBIPaddr string
		_, masterIP, err := mexos.GetMasterNameAndIP(clusterInst)
		if err == nil {
			rootLBIPaddr, err = mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), rootLBName)
			if err == nil {
				getDnsAction := func(svc v1.Service) (*mexos.DnsSvcAction, error) {
					action := mexos.DnsSvcAction{}
					action.PatchKube = true
					action.PatchIP = masterIP
					action.ExternalIP = rootLBIPaddr
					// Should only add DNS for external ports
					action.AddDNS = !app.InternalPorts
					return &action, nil
				}
				// If this is an internal ports, all we need is patch of kube service
				if app.InternalPorts {
					err = mexos.CreateAppDNS(client, names, getDnsAction)
				} else {
					updateCallback(edgeproto.UpdateTask, "Configuring Service: LB, Firewall Rules and DNS")
					err = mexos.AddProxySecurityRulesAndPatchDNS(client, names, appInst, getDnsAction, rootLBName, masterIP, true, nginx.WithDockerNetwork("host"))
				}
			}
		}
		appWaitErr := <-appWaitChan
		if appWaitErr != "" {
			return fmt.Errorf("CreateKubernetesAppInst app wait error: %v", appWaitErr)
		}
		if err != nil {
			return fmt.Errorf("CreateKubernetesAppInst other error: %v", err)
		}
	case cloudcommon.AppDeploymentTypeVM:
		imageName, err := cloudcommon.GetFileName(app.ImagePath)
		if err != nil {
			return fmt.Errorf("CreateVMAppInst error: %v", err)
		}
		sourceImageTime, md5Sum, err := mexos.GetUrlInfo(app.ImagePath)
		if err != nil {
			log.SpanLog(s.ctx, log.DebugLevelMexos, "failed to fetch source image info, skip image validity checks")
		}
		glanceImageTime, err := mexos.GetImageUpdatedTime(imageName)
		createImage := false
		if err != nil {
			if strings.Contains(err.Error(), "Could not find resource") {
				// Add image to Glance
				log.SpanLog(s.ctx, log.DebugLevelMexos, "image is not present in glance, add image")
				createImage = true
			} else {
				return fmt.Errorf("CreateVMAppInst error: %v", err)
			}
		} else {
			if !sourceImageTime.IsZero() {
				if sourceImageTime.Sub(glanceImageTime) > 0 {
					// Update the image in Glance
					log.SpanLog(s.ctx, log.DebugLevelMexos, "image in glance is no longer valid, update image")
					err = mexos.DeleteImage(imageName)
					if err != nil {
						return fmt.Errorf("CreateVMAppInst error: %v", err)
					}
					createImage = true
				}
			}
		}
		if createImage {
			updateCallback(edgeproto.UpdateTask, "Creating VM Image from URL")
			err = mexos.CreateImageFromUrl(imageName, app.ImagePath, md5Sum)
			if err != nil {
				return fmt.Errorf("CreateVMAppInst error: %v", err)
			}
		}

		finfo, err := mexos.GetFlavorInfo()
		if err != nil {
			return err
		}
		appFlavorName, err := flavor.GetClosestFlavor(finfo, *appFlavor)
		if err != nil {
			return fmt.Errorf("unable to find closest flavor for app: %v", err)
		}
		vmp, err := mexos.GetVMParams(
			mexos.UserVMDeployment,
			app.Key.Name,
			appFlavorName,
			imageName,
			app.AuthPublicKey,
			app.AccessPorts,
			app.DeploymentManifest,
			app.Command,
			nil, // NetSpecInfo
		)
		if err != nil {
			return fmt.Errorf("unable to get vm params: %v", err)
		}
		updateCallback(edgeproto.UpdateTask, "Deploying VM")
		log.SpanLog(s.ctx, log.DebugLevelMexos, "Deploying VM", "stackName", app.Key.Name, "flavor", appFlavorName)
		err = mexos.CreateHeatStackFromTemplate(vmp, app.Key.Name, mexos.VmTemplate, updateCallback)
		if err != nil {
			return fmt.Errorf("CreateVMAppInst error: %v", err)
		}
		external_ip, err := mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), app.Key.Name)
		if err != nil {
			return err
		}
		if appInst.Uri != "" && external_ip != "" {
			fqdn := appInst.Uri
			if err = mexos.ActivateFQDNA(fqdn, external_ip); err != nil {
				return err
			}
			log.SpanLog(s.ctx, log.DebugLevelMexos, "DNS A record activated",
				"name", app.Key.Name,
				"fqdn", fqdn,
				"IP", external_ip)
		}
		return nil
	case cloudcommon.AppDeploymentTypeDocker:
		rootLBName := s.rootLBName
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
			log.SpanLog(s.ctx, log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
		}
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		rootLBIPaddr, err := mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), rootLBName)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed, %v", err)
		}
		updateCallback(edgeproto.UpdateTask, "Seeding docker secret")
		err = mexos.SeedDockerSecret(client, clusterInst, app, s.config.VaultAddr)
		if err != nil {
			return fmt.Errorf("seeding docker secret failed, %v", err)
		}
		updateCallback(edgeproto.UpdateTask, "Deploying Docker App")
		err = dockermgmt.CreateAppInst(client, app, appInst)
		if err != nil {
			return fmt.Errorf("createAppInst error for docker %v", err)
		}
		getDnsAction := func(svc v1.Service) (*mexos.DnsSvcAction, error) {
			action := mexos.DnsSvcAction{}
			action.PatchKube = false
			action.ExternalIP = rootLBIPaddr
			return &action, nil
		}
		updateCallback(edgeproto.UpdateTask, "Configuring Firewall Rules and DNS")
		err = mexos.AddProxySecurityRulesAndPatchDNS(client, names, appInst, getDnsAction, rootLBName, rootLBIPaddr, false)
		if err != nil {
			return fmt.Errorf("AddProxySecurityRulesAndPatchDNS error: %v", err)
		}
	default:
		err = fmt.Errorf("unsupported deployment type %s", deployment)
	}
	return err
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		rootLBName := s.rootLBName
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
			log.SpanLog(s.ctx, log.DebugLevelMexos, "using dedicated RootLB to delete app", "rootLBName", rootLBName)
		}
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}

		_, masterIP, err := mexos.GetMasterNameAndIP(clusterInst)
		if err != nil {
			return err
		} // Clean up security rules and nginx proxy if app is external
		if !app.InternalPorts {
			if err := mexos.DeleteProxySecurityRules(client, masterIP, names.AppName); err != nil {
				log.SpanLog(s.ctx, log.DebugLevelMexos, "cannot clean up security rules", "name", names.AppName, "rootlb", rootLBName, "error", err)
			}
			// Clean up DNS entries
			if err := mexos.DeleteAppDNS(client, names); err != nil {
				log.SpanLog(s.ctx, log.DebugLevelMexos, "cannot clean up DNS entries", "name", names.AppName, "rootlb", rootLBName, "error", err)
			}
		}
		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			return k8smgmt.DeleteAppInst(client, names, app, appInst)
		} else {
			return k8smgmt.DeleteHelmAppInst(client, names, clusterInst)
		}
	case cloudcommon.AppDeploymentTypeVM:
		log.SpanLog(s.ctx, log.DebugLevelMexos, "Deleting VM", "stackName", app.Key.Name)
		err := mexos.HeatDeleteStack(app.Key.Name)
		if err != nil {
			return fmt.Errorf("DeleteVMAppInst error: %v", err)
		}
		return nil
	case cloudcommon.AppDeploymentTypeDocker:
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		return dockermgmt.DeleteAppInst(client, app, appInst)
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

func (s *Platform) UpdateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		return k8smgmt.UpdateAppInst(client, names, app, appInst)
	case cloudcommon.AppDeploymentTypeDocker:
		client, err := s.GetPlatformClient(clusterInst)
		if err != nil {
			return err
		}
		return dockermgmt.UpdateAppInst(client, app, appInst)
	default:
		return fmt.Errorf("UpdateAppInst not supported for deployment: %s", app.Deployment)
	}
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {

	client, err := s.GetPlatformClient(clusterInst)
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
		return k8smgmt.GetAppInstRuntime(client, names, app, appInst)
	case cloudcommon.AppDeploymentTypeDocker:
		return dockermgmt.GetAppInstRuntime(client, app, appInst)
	case cloudcommon.AppDeploymentTypeVM:
		consoleUrl, err := mexos.OSGetConsoleUrl(app.Key.Name)
		if err != nil {
			return nil, err
		}
		rt := &edgeproto.AppInstRuntime{}
		rt.ConsoleUrl = consoleUrl.Url
		return rt, nil
	default:
		return nil, fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		return k8smgmt.GetContainerCommand(clusterInst, app, appInst, req)
	case cloudcommon.AppDeploymentTypeDocker:
		return dockermgmt.GetContainerCommand(clusterInst, app, appInst, req)
	case cloudcommon.AppDeploymentTypeVM:
		fallthrough
	default:
		return "", fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

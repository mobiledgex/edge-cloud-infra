package openstack

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/flavor"
	"github.com/mobiledgex/edge-cloud/log"
	"k8s.io/api/core/v1"
)

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, appFlavor *edgeproto.Flavor) error {
	var err error

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.AppDeploymentTypeHelm:
		rootLBName := s.rootLBName
		if clusterInst.IpAccess == edgeproto.IpAccess_IpAccessDedicated {
			rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
			log.DebugLog(log.DebugLevelMexos, "using dedicated RootLB to create app", "rootLBName", rootLBName)
		}
		client, err := s.GetPlatformClient(rootLBName)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}
		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			err = k8smgmt.CreateAppInst(client, names, app, appInst)
		} else {
			err = k8smgmt.CreateHelmAppInst(client, names, clusterInst, app, appInst)
		}
		// set up DNS
		masterIP, err := mexos.GetMasterIP(clusterInst, mexos.GetCloudletExternalNetwork())
		if err != nil {
			return err
		}
		rootLBIPaddr, err := mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), rootLBName)
		if err != nil {
			return err
		}
		getDnsAction := func(svc v1.Service) (*mexos.DnsSvcAction, error) {
			action := mexos.DnsSvcAction{}
			action.PatchKube = true
			action.PatchIP = masterIP
			action.ExternalIP = rootLBIPaddr
			return &action, nil
		}
		err = mexos.AddProxySecurityRulesAndPatchDNS(client, names, appInst, getDnsAction, rootLBName, masterIP)
		if err != nil {
			return fmt.Errorf("CreateKubernetesAppInst error: %v", err)
		}
	case cloudcommon.AppDeploymentTypeVM:
		imageName, err := cloudcommon.GetFileName(app.ImagePath)
		if err != nil {
			return fmt.Errorf("CreateVMAppInst error: %v", err)
		}
		sourceImageTime, md5Sum, err := mexos.GetUrlInfo(app.ImagePath)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "failed to fetch source image info, skip image validity checks")
		}
		glanceImageTime, err := mexos.GetImageUpdatedTime(imageName)
		if err != nil {
			if strings.Contains(err.Error(), "Could not find resource") {
				// Add image to Glance
				log.DebugLog(log.DebugLevelMexos, "image is not present in glance, add image")
				err := mexos.CreateImageFromUrl(imageName, app.ImagePath, md5Sum)
				if err != nil {
					return fmt.Errorf("CreateVMAppInst error: %v", err)
				}
			} else {
				return fmt.Errorf("CreateVMAppInst error: %v", err)
			}
		} else {
			if !sourceImageTime.IsZero() {
				if sourceImageTime.Sub(glanceImageTime) > 0 {
					// Update the image in Glance
					log.DebugLog(log.DebugLevelMexos, "image in glance is no more valid, update image")
					err = mexos.DeleteImage(imageName)
					if err != nil {
						return fmt.Errorf("CreateVMAppInst error: %v", err)
					}
					err = mexos.CreateImageFromUrl(imageName, app.ImagePath, md5Sum)
					if err != nil {
						return fmt.Errorf("CreateVMAppInst error: %v", err)
					}
				}
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

		log.DebugLog(log.DebugLevelMexos, "Deploying VM", "stackName", app.Key.Name, "flavor", appFlavorName)
		err = mexos.HeatCreateVM(app.Key.Name, appFlavorName, imageName, mexos.UserVMDeployment, app.AuthPublicKey, app.AccessPorts)
		if err != nil {
			return fmt.Errorf("CreateVMAppInst error: %v", err)
		}
		return nil
	case cloudcommon.AppDeploymentTypeDockerSwarm:
		fallthrough
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
		if clusterInst.IpAccess == edgeproto.IpAccess_IpAccessDedicated {
			rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
			log.DebugLog(log.DebugLevelMexos, "using dedicated RootLB to delete app", "rootLBName", rootLBName)
		}
		client, err := s.GetPlatformClient(rootLBName)
		if err != nil {
			return err
		}
		names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
		if err != nil {
			return fmt.Errorf("get kube names failed: %s", err)
		}

		masterIP, err := mexos.GetMasterIP(clusterInst, mexos.GetCloudletExternalNetwork())
		if err != nil {
			return err
		} // Clean up security rules and nginx proxy
		if err := mexos.DeleteProxySecurityRules(s.rootLB, masterIP, names.AppName); err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot clean up security rules", "name", names.AppName, "rootlb", rootLBName, "error", err)
		}
		// Clean up DNS entries
		if err := mexos.DeleteAppDNS(client, names); err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot clean up DNS entries", "name", names.AppName, "rootlb", rootLBName, "error", err)
		}
		if deployment == cloudcommon.AppDeploymentTypeKubernetes {
			return k8smgmt.DeleteAppInst(client, names, app, appInst)
		} else {
			return k8smgmt.DeleteHelmAppInst(client, names, clusterInst)
		}
	case cloudcommon.AppDeploymentTypeVM:
		log.DebugLog(log.DebugLevelMexos, "Deleting VM", "stackName", app.Key.Name)
		err := mexos.HeatDeleteVM(app.Key.Name)
		if err != nil {
			return fmt.Errorf("DeleteVMAppInst error: %v", err)
		}
		return nil
	case cloudcommon.AppDeploymentTypeDockerSwarm:
		fallthrough
	default:
		return fmt.Errorf("unsupported deployment type %s", deployment)
	}
}

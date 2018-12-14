package mexos

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
)

//MEXClusterCreateManifest creates a cluster
func MEXClusterCreateManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "creating cluster via manifest")
	switch mf.Metadata.Operator {
	case "gcp":
		return gcloudCreateGKE(mf)
	case "azure":
		return azureCreateAKS(mf)
	default:
		//guid, err := mexCreateClusterKubernetes(mf)
		err := mexCreateClusterKubernetes(mf)
		if err != nil {
			return fmt.Errorf("can't create cluster, %v", err)
		}
		//log.DebugLog(log.DebugLevelMexos, "new guid", "guid", *guid)
		log.DebugLog(log.DebugLevelMexos, "created kubernetes cluster")
		return nil
	}
}

//MEXAddFlavor adds flavor using manifest
func MEXAddFlavor(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "add mex flavor")
	//TODO use full manifest and validate against platform data
	return AddFlavorManifest(mf)
}

// TODO DeleteFlavor -- but almost never done

// TODO lookup guid using name

//MEXClusterRemoveManifest removes a cluster
func MEXClusterRemoveManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "removing cluster")
	switch mf.Metadata.Operator {
	case "gcp":
		return gcloud.DeleteGKECluster(mf.Metadata.Name)
	case "azure":
		return azure.DeleteAKSCluster(mf.Metadata.ResourceGroup)
	default:
		if err := mexDeleteClusterKubernetes(mf); err != nil {
			return fmt.Errorf("can't remove cluster, %v", err)
		}
		return nil
	}
}

//MEXPlatformInitCloudletKey calls MEXPlatformInit with templated manifest
func MEXPlatformInitCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) error {
	//XXX trigger off cloudletKeyStr or flavor to pick the right template: mex, aks, gke
	mf, err := fillPlatformTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformInitManifest(mf)
}

//MEXPlatformCleanCloudletKey calls MEXPlatformClean with templated manifest
func MEXPlatformCleanCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) error {
	mf, err := fillPlatformTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformCleanManifest(mf)
}

//MEXPlatformInitManifest initializes platform
func MEXPlatformInitManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "init platform")
	err := MEXCheckEnvVars(mf)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		return nil //nothing to do
	case "azure":
		return nil //nothing to do
	default:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		//TODO validate all mf content against platform data
		if err = RunMEXAgentManifest(mf); err != nil {
			return err
		}
	}
	return nil
}

//MEXPlatformCleanManifest cleans up the platform
func MEXPlatformCleanManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "clean platform")
	err := MEXCheckEnvVars(mf)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		return nil //nothing to do
	case "azure":
		return nil
	default:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		if err = RemoveMEXAgentManifest(mf); err != nil {
			return err
		}
	}
	return nil
}

//MEXAppCreateAppManifest creates app instances on the cluster platform
func MEXAppCreateAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create app from manifest")
	appDeploymentType := mf.Config.ConfigDetail.Deployment
	log.DebugLog(log.DebugLevelMexos, "app deployment", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := cloudcommon.GetDeploymentManifest(mf.Config.ConfigDetail.Manifest)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return runKubectlCreateApp(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return fmt.Errorf("not yet supported")
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	default:
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return CreateKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return CreateQCOW2AppManifest(mf)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return CreateHelmAppManifest(mf)
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	}
}

//MEXAppDeleteManifest kills app
func MEXAppDeleteAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "delete app with manifest")
	appDeploymentType := mf.Config.ConfigDetail.Deployment
	log.DebugLog(log.DebugLevelMexos, "app delete", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := cloudcommon.GetDeploymentManifest(mf.Config.ConfigDetail.Manifest)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return runKubectlDeleteApp(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return fmt.Errorf("not yet supported")
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown image type %s", appDeploymentType)
	default:
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return DeleteKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return DeleteQCOW2AppManifest(mf)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return DeleteHelmAppManifest(mf)
		}
		return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
	}
}

func FillManifest(mf *Manifest, kind, base string) error {
	if mf.Values.Name == "" {
		return fmt.Errorf("no name for mf values")
	}
	var uri string
	switch kind {
	case "openstack":
		kind = "platform"
		fallthrough
	case "platform":
		fallthrough
	case "cluster":
		uri = fmt.Sprintf("%s/%s/%s/%s.yaml", base, kind, mf.Values.Operator.Name, mf.Values.Base)
	case "application":
		uri = fmt.Sprintf("%s/%s/%s/%s.yaml", base, kind, mf.Values.Application.Kind, mf.Values.Base)
	default:
		return fmt.Errorf("invalid manifest kind %s", kind)
	}
	dat, err := GetURIFile(mf, uri)
	if err != nil {
		return err
	}
	tmpl, err := template.New(mf.Values.Name).Parse(string(dat))
	if err != nil {
		return err
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, &mf.Values)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultSecurityRule(mf *Manifest) string {
	return mf.Values.Network.SecurityRule
}

func GetMEXSecurityRule(mf *Manifest) string {
	return mf.Values.Network.SecurityRule
}

//GetMEXExternalRouter returns default MEX external router name
func GetMEXExternalRouter(mf *Manifest) string {
	//TODO validate existence and status
	return mf.Values.Network.Router
}

//GetMEXExternalNetwork returns default MEX external network name
func GetMEXExternalNetwork(mf *Manifest) string {
	//TODO validate existence and status
	return mf.Values.Network.External
}

//GetMEXNetwork returns default MEX network, internal and prepped
func GetMEXNetwork(mf *Manifest) string {
	//TODO validate existence and status
	return mf.Values.Network.Name
}

func GetMEXImageName(mf *Manifest) string {
	return mf.Values.Cluster.OSImage
}

func GetMEXUserData(mf *Manifest) string {
	return MEXDir() + "/userdata.txt"
}

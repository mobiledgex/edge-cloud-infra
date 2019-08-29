package gcp

import (
	"fmt"
	"os"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	v1 "k8s.io/api/core/v1"
)

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error
	// regenerate kconf if missing because CRM in container was restarted
	if err = SetupKconf(clusterInst); err != nil {
		return fmt.Errorf("can't set up kconf, %s", err.Error())
	}
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return err
	}

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	err = mexos.CreateDockerRegistrySecret(s.ctx, client, clusterInst, app, s.config.VaultAddr)
	if err != nil {
		return err
	}

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		err = k8smgmt.CreateAppInst(client, names, app, appInst)
		if err == nil {
			err = k8smgmt.WaitForAppInst(client, names, app, k8smgmt.WaitRunning)
		}
	default:
		err = fmt.Errorf("unsupported deployment type %s", deployment)
	}
	if err != nil {
		return err
	}

	// set up dns
	getDnsAction := func(svc v1.Service) (*mexos.DnsSvcAction, error) {
		action := mexos.DnsSvcAction{}
		externalIP, err := mexos.GetSvcExternalIP(s.ctx, client, names, svc.ObjectMeta.Name)
		if err != nil {
			return nil, err
		}
		action.ExternalIP = externalIP
		// no patching needed since GCP already does it.
		// Should only add DNS for external ports
		action.AddDNS = !app.InternalPorts
		return &action, nil
	}
	if err = mexos.CreateAppDNS(s.ctx, client, names, getDnsAction); err != nil {
		return err
	}
	return nil
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	var err error
	// regenerate kconf if missing because CRM in container was restarted
	if err = SetupKconf(clusterInst); err != nil {
		return fmt.Errorf("can't set up kconf, %s", err.Error())
	}
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return err
	}

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	switch deployment := app.Deployment; deployment {
	case cloudcommon.AppDeploymentTypeKubernetes:
		err = k8smgmt.DeleteAppInst(client, names, app, appInst)
	default:
		err = fmt.Errorf("unsupported deployment type %s", deployment)
	}
	if err != nil {
		return err
	}
	// No DNS entry if ports are internal
	if app.InternalPorts {
		return nil
	}
	return mexos.DeleteAppDNS(s.ctx, client, names)
}

func SetupKconf(clusterInst *edgeproto.ClusterInst) error {
	targetFile := mexos.GetLocalKconfName(clusterInst)
	if _, err := os.Stat(targetFile); err == nil {
		// already exists
		return nil
	}
	clusterName := clusterInst.Key.ClusterKey.Name
	if err := GetGKECredentials(clusterName); err != nil {
		return fmt.Errorf("unable to get GKE credentials %v", err)
	}
	src := mexos.DefaultKubeconfig()
	if err := mexos.CopyFile(src, targetFile); err != nil {
		return fmt.Errorf("can't copy %s, %v", src, err)
	}
	return nil
}

func (s *Platform) UpdateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Updating GCP AppInst")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return err
	}

	err = k8smgmt.UpdateAppInst(client, names, app, appInst)
	if err == nil {
		updateCallback(edgeproto.UpdateTask, "Waiting for AppInst to Start")
		err = k8smgmt.WaitForAppInst(client, names, app, k8smgmt.WaitRunning)
	}
	return err
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	// regenerate kconf if missing because CRM in container was restarted
	if err := SetupKconf(clusterInst); err != nil {
		return nil, fmt.Errorf("can't set up kconf, %s", err.Error())
	}
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return nil, err
	}

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return nil, err
	}
	return k8smgmt.GetAppInstRuntime(client, names, app, appInst)
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return k8smgmt.GetContainerCommand(clusterInst, app, appInst, req)
}

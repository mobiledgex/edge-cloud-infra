package edgebox

import (
	"context"
	"fmt"
	"net"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	v1 "k8s.io/api/core/v1"
)

func (e *EdgeboxPlatform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	client, err := e.generic.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return err
	}

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	if app.Deployment != cloudcommon.AppDeploymentTypeDocker {
		err = infracommon.CreateDockerRegistrySecret(ctx, client, clusterInst, app, e.commonPf.VaultConfig, names)
		if err != nil {
			return err
		}
	}

	// Use generic DIND to create the AppInst
	err = e.generic.CreateAppInst(ctx, clusterInst, app, appInst, flavor, privacyPolicy, updateCallback)
	if err != nil {
		return err
	}

	// The rest is k8s specific
	if clusterInst.Deployment != cloudcommon.AppDeploymentTypeKubernetes {
		return nil
	}

	// set up DNS
	cluster, err := dind.FindCluster(names.ClusterName)
	if err != nil {
		return err
	}
	masterIP := cluster.MasterAddr
	externalIP, err := e.GetDINDServiceIP(ctx)
	getDnsAction := func(svc v1.Service) (*infracommon.DnsSvcAction, error) {
		action := infracommon.DnsSvcAction{}

		if len(svc.Spec.ExternalIPs) > 0 && svc.Spec.ExternalIPs[0] == masterIP {
			log.SpanLog(ctx, log.DebugLevelMexos, "external IP already present in DIND, no patch required", "addr", masterIP)
		} else {
			action.PatchKube = true
			action.PatchIP = masterIP
		}
		if err != nil {
			return nil, err
		}
		action.ExternalIP = externalIP
		// Should only add DNS for external ports
		action.AddDNS = !app.InternalPorts
		return &action, nil
	}
	if err = e.commonPf.CreateAppDNSAndPatchKubeSvc(ctx, client, names, infracommon.NoDnsOverride, getDnsAction); err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "cannot add DNS entries", "error", err)
		return err
	}
	return nil
}

func (e *EdgeboxPlatform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	var err error
	client, err := e.generic.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return err
	}

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	// remove DNS entries if it was added
	if !app.InternalPorts {
		if err = e.commonPf.DeleteAppDNS(ctx, client, names, infracommon.NoDnsOverride); err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "warning, cannot delete DNS record", "error", err)
		}
	}
	if err = e.generic.DeleteAppInst(ctx, clusterInst, app, appInst); err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "warning, cannot delete AppInst", "error", err)
		return err
	}
	return nil
}

func (e *EdgeboxPlatform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("Update not supported for dind")
}

func (e *EdgeboxPlatform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return e.generic.GetAppInstRuntime(ctx, clusterInst, app, appInst)
}

func (e *EdgeboxPlatform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return e.generic.GetContainerCommand(ctx, clusterInst, app, appInst, req)
}

func (e *EdgeboxPlatform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	return e.generic.GetConsoleUrl(ctx, app)
}

func (e *EdgeboxPlatform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return e.generic.SetPowerState(ctx, app, appInst, updateCallback)
}

// GetDINDServiceIP depending on the type of DIND cluster will return either the interface or external address
func (e *EdgeboxPlatform) GetDINDServiceIP(ctx context.Context) (string, error) {
	if e.NetworkScheme == cloudcommon.NetworkSchemePrivateIP {
		return GetLocalAddr()
	}
	return infracommon.GetExternalPublicAddr(ctx)
}

// GetLocalAddr gets the IP address the machine uses for outbound comms
func GetLocalAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

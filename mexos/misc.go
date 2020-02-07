package mexos

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

func PrivateSSHKey() string {
	return MEXDir() + "/id_rsa_mex"
}

func MEXDir() string {
	return os.Getenv("HOME") + "/.mobiledgex"
}

func DefaultKubeconfig() string {
	return os.Getenv("HOME") + "/.kube/config"
}

func CopyFile(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SeedDockerSecret(ctx context.Context, plat platform.Platform, client pc.PlatformClient, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, vaultConfig *vault.Config) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "seed docker secret")

	auth, err := cloudcommon.GetRegistryAuth(ctx, app.ImagePath, vaultConfig)
	if err != nil {
		return err
	}
	if auth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("auth type for %s is not basic auth type", auth.Hostname)
	}
	remoteServer := crmutil.RemoteServerNone

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		_, remoteServer, err = GetMasterNameAndIP(ctx, clusterInst)
		if err != nil {
			return err
		}
	}
	// XXX: not sure writing password to file buys us anything if the
	// echo command is recorded in some history.
	cmd := fmt.Sprintf("echo %s > .docker-pass", auth.Password)
	out, err := crmutil.RunCommand(ctx, plat, client, remoteServer, cmd)
	if err != nil {
		return fmt.Errorf("can't store docker password, %s, %v", out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "stored docker password")
	defer func() {
		cmd := fmt.Sprintf("rm .docker-pass")
		crmutil.RunCommand(ctx, plat, client, remoteServer, cmd)
	}()

	cmd = fmt.Sprintf("cat .docker-pass | docker login -u %s --password-stdin %s ", auth.Username, auth.Hostname)
	out, err = crmutil.RunCommand(ctx, plat, client, remoteServer, cmd)
	if err != nil {
		return fmt.Errorf("can't docker login on rootlb to %s, %s, %v", auth.Hostname, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "docker login ok")
	return nil
}

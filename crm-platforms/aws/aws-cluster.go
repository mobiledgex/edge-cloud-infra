package aws

import (
	"context"
	"fmt"
	//"os"
	"time"

	// sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// AWSLogin logs into Amazon AWS web services
func (a *AWSPlatform) AWSLogin(ctx context.Context) error {
	// log.SpanLog(ctx, log.DebugLevelInfra, "doing AwsLogin", "vault url", a.GetAwsAuthKeyUrl())
	// filename := "/tmp/auth_key.json"
	// err := infracommon.GetVaultDataToFile(a.commonPf.VaultConfig, a.GetAwsAuthKeyUrl(), filename)
	// if err != nil {
	// 	return fmt.Errorf("unable to write auth file %s: %s", filename, err.Error())
	// }
	// defer os.Remove(filename)
	// out, err := sh.Command("aws", "auth", "activate-service-account", "--key-file", filename).CombinedOutput() // What is AWS equivalent?
	// log.SpanLog(ctx, log.DebugLevelInfra, "aws login", "out", string(out), "err", err)
	// if err != nil {
	// 	return err
	// }
	// log.SpanLog(ctx, log.DebugLevelInfra, "AWS login OK")
	return nil
}

func (a *AWSPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	clusterName := clusterInst.Key.ClusterKey.Name
	if err := CreateEKSCluster(clusterName); err != nil {
		return err
	}
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	client, err := a.GetClusterPlatformClient(ctx, clusterInst)
	if err != nil {
		return err
	}
	infracommon.BackupKubeconfig(ctx, client)
	if err = GetEKSCredentials(clusterName); err != nil {
		return err
	}
	kconf := k8smgmt.GetKconfName(clusterInst) //XXX

	log.SpanLog(ctx, log.DebugLevelInfra, "warning, using default config") //XXX
	if err = pc.CopyFile(client, infracommon.DefaultKubeconfig(), kconf); err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created eks", "name", clusterName)
	return nil
}

func (g *AWSPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	return DeleteEKSCluster(clusterInst.Key.ClusterKey.Name)
}

func (g *AWSPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster inst not implemented for AWS")
}

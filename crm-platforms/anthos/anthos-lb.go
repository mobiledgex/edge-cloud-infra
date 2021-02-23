package anthos

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AnthosPlatform) GetSharedLBName(ctx context.Context, key *edgeproto.CloudletKey) string {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSharedLBName", "key", key)
	name := cloudcommon.GetRootLBFQDN(key, a.commonPf.PlatformConfig.AppDNSRoot)
	return name
}

/*
func (a *AnthosPlatform) GetSharedLBNamespace(ctx context.Context, key *edgeproto.CloudletKey) string {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSharedLBNamespace", "key", key)
	name := util.K8SSanitize(key.Name) + "-shared"
	return name
}*/

func (a *AnthosPlatform) SetupLb(ctx context.Context, client ssh.Client, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupLb", "name", name)
	return nil
}

func (a *AnthosPlatform) SetupVirtualCluster(ctx context.Context, client ssh.Client, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupVirtualCluster", "name", name)
	err := a.CreateNamespace(ctx, client, name)

	if err != nil {
		return err
	}
	policyName := name + "-netpol"
	manifest, err := infracommon.CreateK8sNetworkPolicyManifest(ctx, client, policyName, name, a.GetConfigDir())
	if err != nil {
		return err
	}
	return infracommon.ApplyK8sNetworkPolicyManifest(ctx, client, manifest, a.cloudletKubeConfig)
}

func (a *AnthosPlatform) GetKubeConfigForNamespace(namespace string) string {
	return filepath.Dir(a.cloudletKubeConfig) + "/" + namespace + "-kubeconfig"
}

func (a *AnthosPlatform) CreateNamespace(ctx context.Context, client ssh.Client, nameSpace string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNamespace", "lbname", nameSpace)

	cmd := fmt.Sprintf("kubectl create namespace  %s --kubeconfig=%s", nameSpace, a.cloudletKubeConfig)
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "AlreadyExists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "namespace already exists", "out", out)
		} else {
			return fmt.Errorf("Error in creating namespace: %s - %v", out, err)
		}
	}
	// copy the kubeconfig and update the new one with the new namespace
	clustKubeConfig := a.GetKubeConfigForNamespace(nameSpace)
	log.SpanLog(ctx, log.DebugLevelInfra, "create new kubeconfig for lb namespace", "cloudletKubeConfig", a.cloudletKubeConfig, "clustKubeConfig", clustKubeConfig)

	err = pc.CopyFile(client, a.cloudletKubeConfig, clustKubeConfig)
	if err != nil {
		return fmt.Errorf("Failed to create new kubeconfig: %v", err)
	}
	// set the current context
	cmd = fmt.Sprintf("KUBECONFIG=%s kubectl config set-context --current --namespace=%s", clustKubeConfig, nameSpace)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in setting new namespace context: %s - %v", out, err)
	}
	secIps, err := a.GetSecondaryEthInterfaces(ctx, client, a.GetLbEthernetInterface())
	log.SpanLog(ctx, log.DebugLevelInfra, "SECIFS", "secIps", secIps, "err", err)
	return nil
}

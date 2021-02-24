package anthos

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AnthosPlatform) GetNamespaceNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return util.K8SSanitize(fmt.Sprintf("%s-%s", clusterInst.Key.Organization, clusterInst.Key.ClusterKey.Name))
}

func (a *AnthosPlatform) SetupVirtualCluster(ctx context.Context, client ssh.Client, namespace, kubeconfig string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupVirtualCluster", "namespace", namespace)

	err := a.CreateNamespace(ctx, client, namespace, kubeconfig)

	if err != nil {
		return err
	}
	policyName := namespace + "-netpol"
	manifest, err := infracommon.CreateK8sNetworkPolicyManifest(ctx, client, policyName, namespace, a.GetConfigDir())
	if err != nil {
		return err
	}
	return infracommon.ApplyK8sNetworkPolicyManifest(ctx, client, manifest, a.cloudletKubeConfig)
}

/*
func (a *AnthosPlatform) GetKubeConfigForNamespace(namespace string) string {
	return filepath.Dir(a.cloudletKubeConfig) + "/" + namespace + "-kubeconfig"
}*/

func (a *AnthosPlatform) CreateNamespace(ctx context.Context, client ssh.Client, nameSpace, kubeconfig string) error {
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
	log.SpanLog(ctx, log.DebugLevelInfra, "create new kubeconfig for lb namespace", "cloudletKubeConfig", a.cloudletKubeConfig, "clustKubeConfig", kubeconfig)

	err = pc.CopyFile(client, a.cloudletKubeConfig, kubeconfig)
	if err != nil {
		return fmt.Errorf("Failed to create new kubeconfig: %v", err)
	}
	// set the current context
	cmd = fmt.Sprintf("KUBECONFIG=%s kubectl config set-context --current --namespace=%s", kubeconfig, nameSpace)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in setting new namespace context: %s - %v", out, err)
	}
	return nil
}

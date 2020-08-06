package azure

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

//CreateResourceGroup creates azure resource group
func CreateResourceGroup(group, location string) error {
	out, err := sh.Command("az", "group", "create", "-l", location, "-n", group).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

// CreateAKSCluster creates kubernetes cluster on azure
func (a *AzurePlatform) RunClusterCreateCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterCreateCommand", "clusterInst", clusterInst)

	clusterName := a.NameSanitize(clusterInst.Key.ClusterKey.Name)
	rg := a.GetResourceGroupForCluster(clusterInst)
	out, err := sh.Command("az", "aks", "create", "--resource-group", rg,
		"--name", clusterName, "--generate-ssh-keys",
		"--node-vm-size", clusterInst.NodeFlavor,
		"--node-count", clusterInst.NumNodes).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in aks create", "out", out, "err", err)
		return fmt.Errorf("Error in aks create: %s %v", out, err)
	}
	return nil
}

// RunClusterDeleteCommand removes the kubernetes cluster on azure
func (a *AzurePlatform) RunClusterDeleteCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterDeleteCommand", "clusterInst", clusterInst)
	rg := a.GetResourceGroupForCluster(clusterInst)
	out, err := sh.Command("az", "group", "delete", "--name", rg, "--yes", "--no-wait").CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in aks delete", "out", out, "err", err)
		return fmt.Errorf("Error in aks delete: %s %v", out, err)
	}
	return nil
}

//GetCredentials retrieves kubeconfig credentials from azure for the cluster just created
func (a *AzurePlatform) GetCredentials(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCredentials", "clusterInst", clusterInst)
	rg := a.GetResourceGroupForCluster(clusterInst)

	out, err := sh.Command("az", "aks", "get-credentials", "--resource-group", rg, "--name", clusterInst.Key.ClusterKey.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

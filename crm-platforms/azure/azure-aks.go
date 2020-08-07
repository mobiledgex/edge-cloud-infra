package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const NotFound = "could not be found"

//CreateResourceGroup creates azure resource group
func (a *AzurePlatform) CreateResourceGroup(ctx context.Context, group, location string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateResourceGroup", "group", group, "location", location)
	out, err := sh.Command("az", "group", "create", "-l", location, "-n", group).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in resource group create", "out", string(out), "err", err)
		return fmt.Errorf("create resource group failed: %v", err)
	}
	return nil
}

func (a *AzurePlatform) CreateClusterPrerequisites(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	rg := a.GetResourceGroupForCluster(clusterInst)
	err := a.CreateResourceGroup(ctx, rg, a.GetAzureLocation())
	if err != nil {
		return err
	}
	return nil
}

// CreateAKSCluster creates kubernetes cluster on azure
func (a *AzurePlatform) RunClusterCreateCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterCreateCommand", "clusterInst", clusterInst)
	rg := a.GetResourceGroupForCluster(clusterInst)
	clusterName := a.NameSanitize(k8smgmt.GetClusterName(clusterInst))
	numNodes := fmt.Sprintf("%d", clusterInst.NumNodes)
	out, err := sh.Command("az", "aks", "create", "--resource-group", rg,
		"--name", clusterName, "--generate-ssh-keys",
		"--node-vm-size", clusterInst.NodeFlavor,
		"--node-count", numNodes).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in aks create", "out", string(out), "err", err)
		return fmt.Errorf("Error in aks create: %s", err)
	}
	return nil
}

// RunClusterDeleteCommand removes the kubernetes cluster on azure
func (a *AzurePlatform) RunClusterDeleteCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RunClusterDeleteCommand", "clusterInst", clusterInst)
	rg := a.GetResourceGroupForCluster(clusterInst)
	out, err := sh.Command("az", "group", "delete", "--name", rg, "--yes", "--no-wait").CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), NotFound) {
			log.SpanLog(ctx, log.DebugLevelInfra, "Cluster already gone", "out", out, "err", err)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in aks delete", "out", out, "err", err)
		return fmt.Errorf("Error in aks delete: %s %v", out, err)
	}
	return nil
}

//GetCredentials retrieves kubeconfig credentials from azure for the cluster just created
func (a *AzurePlatform) GetCredentials(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCredentials", "clusterInst", clusterInst)
	clusterName := a.NameSanitize(k8smgmt.GetClusterName(clusterInst))
	rg := a.GetResourceGroupForCluster(clusterInst)
	out, err := sh.Command("az", "aks", "get-credentials", "--resource-group", rg, "--name", clusterName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

package aws

import (
	"context"
	"fmt"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// RunClusterCreateCommand creates a kubernetes cluster on AWS
func (a *AWSPlatform) RunClusterCreateCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	// output log messages
	log.DebugLog(log.DebugLevelInfra, "RunClusterCreateCommand", "numNodes:", clusterInst.NumNodes, "NodeFlavor", clusterInst.NodeFlavor)
	clusterName := clusterInst.Key.ClusterKey.Name
	// Can not create a managed cluster if numNodes is 0
	if clusterInst.NumNodes == 0 {
		out, err := sh.Command("eksctl", "create", "cluster", "--name", clusterName, "--node-type", clusterInst.NodeFlavor, "--nodes", fmt.Sprintf("%d", clusterInst.NumNodes)).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s %v", out, err)
		}
	} else {
		out, err := sh.Command("eksctl", "create", "cluster", "--name", clusterName, "--node-type", clusterInst.NodeFlavor, "--nodes", fmt.Sprintf("%d", clusterInst.NumNodes), "--managed").CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s %v", out, err)
		}
	}
	return nil
}

// RunClusterDeleteCommand removes the kubernetes cluster on AWS
func (a *AWSPlatform) RunClusterDeleteCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	clusterName := clusterInst.Key.ClusterKey.Name
	log.DebugLog(log.DebugLevelInfra, "RunClusterDeleteCommand", "clusterName:", clusterName)

	out, err := sh.Command("eksctl", "delete", "cluster", "--name", clusterName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

//GetCredentials retrieves kubeconfig credentials from AWS
// eksctl utils write-kubeconfig myawscluster
// Alternate: aws eks --region region-code update-kubeconfig --name cluster_name
func (a *AWSPlatform) GetCredentials(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	out, err := sh.Command("eksctl", "utils", "write-kubeconfig", clusterInst.Key.ClusterKey.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

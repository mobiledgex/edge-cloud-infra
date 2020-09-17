package infracommon

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

//CopyKubeConfig copies over kubeconfig from the cluster
func CopyKubeConfig(ctx context.Context, rootLBClient ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName, masterIP string) error {
	kconfname := k8smgmt.GetKconfName(clusterInst)
	log.SpanLog(ctx, log.DebugLevelInfra, "attempt to get kubeconfig from k8s master", "masterIP", masterIP, "dest", kconfname)
	client, err := rootLBClient.AddHop(masterIP, 22)
	if err != nil {
		return err
	}

	// fetch kubeconfig from master node
	cmd := "cat ~/.kube/config"
	out, err := client.Output(cmd)
	if err != nil || out == "" {
		return fmt.Errorf("failed to get kubeconfig from master node %s, %s, %v", cmd, out, err)
	}

	// save it in rootLB
	cmd = fmt.Sprintf(`echo -e "%s" > %s`, out, kconfname)
	out, err = rootLBClient.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy kubeconfig via cmd %s, %s, %v", cmd, out, err)
	}

	//TODO generate per proxy password and record in vault
	//port, serr := StartKubectlProxy(mf, rootLB, name, kconfname)
	//if serr != nil {
	//	return serr
	//}
	return nil
}

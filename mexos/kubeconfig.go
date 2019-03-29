package mexos

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8s"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/nanobox-io/golang-ssh"
)

func GetLocalKconfName(clusterInst *edgeproto.ClusterInst) string {
	kconf := fmt.Sprintf("%s/%s", MEXDir(), k8s.GetKconfName(clusterInst))
	return kconf
}

//CopyKubeConfig copies over kubeconfig from the cluster
func CopyKubeConfig(rootLBClient ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName, masterIP string) error {
	kconfname := k8s.GetKconfName(clusterInst)
	log.DebugLog(log.DebugLevelMexos, "attempt to get kubeconfig from k8s master", "masterIP", masterIP, "dest", kconfname)
	cmd := fmt.Sprintf("scp -o %s -o %s -i id_rsa_mex %s@%s:.kube/config %s", sshOpts[0], sshOpts[1], SSHUser, masterIP, kconfname)
	out, err := rootLBClient.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy kubeconfig from %s, %v", out, err)
	}
	cmd = fmt.Sprintf("cat %s", kconfname)
	out, err = rootLBClient.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't cat %s, %s, %v", kconfname, out, err)
	}
	//TODO generate per proxy password and record in vault
	//port, serr := StartKubectlProxy(mf, rootLB, name, kconfname)
	//if serr != nil {
	//	return serr
	//}
	return nil
}

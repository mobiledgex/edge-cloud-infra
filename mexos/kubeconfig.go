package mexos

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetKconf(mf *Manifest, createIfMissing bool) (string, error) {
	name := MEXDir() + "/" + mf.Spec.Key + ".kubeconfig"
	log.DebugLog(log.DebugLevelMexos, "get kubeconfig name", "name", name)

	if createIfMissing {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			// if kubeconfig does not exist, optionally create it.  It is possible it was
			// created on a different container or we had a restart of the container
			log.DebugLog(log.DebugLevelMexos, "creating missing kconf file", "name", name)
			switch mf.Metadata.Operator {
			case "gcp":
				if err = gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
					return "", fmt.Errorf("unable to get GKE credentials %v", err)
				}
				if err = copyFile(defaultKubeconfig(), name); err != nil {
					return "", fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
				}
			case "azure":
				if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Name); err != nil {
					return "", fmt.Errorf("unable to get AKS credentials %v", err)
				}
				if err = copyFile(defaultKubeconfig(), name); err != nil {
					return "", fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
				}
			}
		}
	}
	return name, nil
}

type clusterDetailKc struct {
	CertificateAuthorityData string `json:"certificate-authority-data"`
	Server                   string `json:"server"`
}

type clusterKc struct {
	Name    string          `json:"name"`
	Cluster clusterDetailKc `json:"cluster"`
}

type clusterKcContextDetail struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type clusterKcContext struct {
	Name    string                 `json:"name"`
	Context clusterKcContextDetail `json:"context"`
}

type clusterKcUserDetail struct {
	ClientCertificateData string `json:"client-certificate-data"`
	ClientKeyData         string `json:"client-key-data"`
}

type clusterKcUser struct {
	Name string              `json:"name"`
	User clusterKcUserDetail `json:"user"`
}

type clusterKubeconfig struct {
	APIVersion     string             `json:"apiVersion"`
	Kind           string             `json:"kind"`
	CurrentContext string             `json:"current-context"`
	Users          []clusterKcUser    `json:"users"`
	Clusters       []clusterKc        `json:"clusters"`
	Contexts       []clusterKcContext `json:"contexts"`
	//XXX Missing preferences
}

//CopyKubeConfig copies over kubeconfig from the cluster
func CopyKubeConfig(mf *Manifest, rootLB *MEXRootLB, name string) error {
	log.DebugLog(log.DebugLevelMexos, "copying kubeconfig", "name", name)
	if rootLB == nil {
		return fmt.Errorf("cannot copy kubeconfig, rootLB is null")
	}
	ipaddr, err := FindNodeIP(mf, name)
	if err != nil {
		return err
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("copy kube config, missing external network in platform config")
	}
	client, err := GetSSHClient(mf, rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return fmt.Errorf("can't get ssh client for copying kubeconfig, %v", err)
	}
	kconfname := fmt.Sprintf("%s.kubeconfig", name[strings.LastIndex(name, "-")+1:])
	cmd := fmt.Sprintf("scp -o %s -o %s -i %s root@%s:.kube/config %s", sshOpts[0], sshOpts[1], PrivateSSHKey(), ipaddr, kconfname)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy kubeconfig from %s, %s, %v", name, out, err)
	}
	cmd = fmt.Sprintf("cat %s", kconfname)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't cat %s, %s, %v", kconfname, out, err)
	}
	port, serr := StartKubectlProxy(mf, rootLB, kconfname)
	if serr != nil {
		return serr
	}
	return ProcessKubeconfig(mf, rootLB, name, port, []byte(out))
}

//ProcessKubeconfig validates kubeconfig and saves it and creates a copy for proxy access
func ProcessKubeconfig(mf *Manifest, rootLB *MEXRootLB, name string, port int, dat []byte) error {
	log.DebugLog(log.DebugLevelMexos, "process kubeconfig file", "name", name)
	if rootLB == nil {
		return fmt.Errorf("cannot process kubeconfig, rootLB is null")
	}
	kc := &clusterKubeconfig{}
	err := yaml.Unmarshal(dat, kc)
	if err != nil {
		return fmt.Errorf("can't unmarshal kubeconfig %s, %v", name, err)
	}
	if len(kc.Clusters) < 1 {
		return fmt.Errorf("insufficient clusters info in kubeconfig %s", name)
	}
	//log.Debugln("kubeconfig", kc)
	//TODO: the kconfname should include more details, location, tags, to be distinct from other clusters in other regions
	kconfname := fmt.Sprintf("%s.kubeconfig", name[strings.LastIndex(name, "-")+1:])
	fullpath := MEXDir() + "/" + kconfname
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig %s content,%v", name, err)
	}
	log.DebugLog(log.DebugLevelMexos, "wrote kubeconfig", "file", fullpath)
	kc.Clusters[0].Cluster.Server = fmt.Sprintf("http://%s:%d", rootLB.Name, port)
	dat, err = yaml.Marshal(kc)
	if err != nil {
		return fmt.Errorf("can't marshal kubeconfig proxy edit %s, %v", name, err)
	}
	fullpath = fullpath + "-proxy"
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig proxy %s, %v", fullpath, err)
	}
	log.DebugLog(log.DebugLevelMexos, "kubeconfig-proxy file saved", "file", fullpath)
	return nil
}

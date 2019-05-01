package gcp

import (
	"time"

	"encoding/json"
	"fmt"
	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"io/ioutil"
	"os"
)

// GCPLogin logs into google cloud
func (s *Platform) GCPLogin() error {
	log.DebugLog(log.DebugLevelMexos, "doing GcpLogin", "vault url", s.props.GCPAuthKeyUrl)
	dat, err := mexos.GetVaultData(s.props.GCPAuthKeyUrl)
	if err != nil {
		return err
	}
	vr, err := mexos.GetVaultGenericResponse(dat)
	if err != nil {
		return err
	}
	databytes, err := json.Marshal(vr.Data.Data)
	filename := "/tmp/auth_key.json"
	err = ioutil.WriteFile(filename, databytes, 0644)
	defer os.Remove(filename)
	if err != nil {
		return fmt.Errorf("unable to write auth file %s: %s", filename, err.Error())
	}

	out, err := sh.Command("gcloud", "auth", "activate-service-account", GCPServiceAccount, "--key-file", filename).CombinedOutput()
	log.DebugLog(log.DebugLevelMexos, "gcp login", "out", string(out), "err", err)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "GCP login OK")
	return nil
}

func (s *Platform) CreateCluster(clusterInst *edgeproto.ClusterInst, flavor *edgeproto.ClusterFlavor) error {
	var err error
	project := s.props.Project
	zone := s.props.Zone
	clusterName := clusterInst.Key.ClusterKey.Name

	if err = SetProject(project); err != nil {
		return err
	}
	if err = SetZone(zone); err != nil {
		return err
	}
	if err = CreateGKECluster(clusterName); err != nil {
		return err
	}
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return err
	}
	mexos.BackupKubeconfig(client)
	if err = GetGKECredentials(clusterName); err != nil {
		return err
	}
	kconf := k8smgmt.GetKconfName(clusterInst) //XXX

	log.DebugLog(log.DebugLevelMexos, "warning, using default config") //XXX
	if err = pc.CopyFile(client, mexos.DefaultKubeconfig(), kconf); err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "created gke", "name", clusterName)
	return mexos.CreateDockerRegistrySecret(client, clusterInst)
}

func (s *Platform) DeleteCluster(clusterInst *edgeproto.ClusterInst) error {
	return DeleteGKECluster(clusterInst.Key.ClusterKey.Name)
}

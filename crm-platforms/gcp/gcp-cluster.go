package gcp

import (
	"time"

	"fmt"
	"os"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// GCPLogin logs into google cloud
func (s *Platform) GCPLogin() error {
	log.SpanLog(s.ctx, log.DebugLevelMexos, "doing GcpLogin", "vault url", s.props.GcpAuthKeyUrl)
	filename := "/tmp/auth_key.json"
	err := mexos.GetVaultDataToFile(s.props.GcpAuthKeyUrl, filename)
	if err != nil {
		return fmt.Errorf("unable to write auth file %s: %s", filename, err.Error())
	}
	defer os.Remove(filename)
	out, err := sh.Command("gcloud", "auth", "activate-service-account", GCPServiceAccount, "--key-file", filename).CombinedOutput()
	log.SpanLog(s.ctx, log.DebugLevelMexos, "gcp login", "out", string(out), "err", err)
	if err != nil {
		return err
	}
	log.SpanLog(s.ctx, log.DebugLevelMexos, "GCP login OK")
	return nil
}

func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
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

	log.SpanLog(s.ctx, log.DebugLevelMexos, "warning, using default config") //XXX
	if err = pc.CopyFile(client, mexos.DefaultKubeconfig(), kconf); err != nil {
		return err
	}
	log.SpanLog(s.ctx, log.DebugLevelMexos, "created gke", "name", clusterName)
	return nil
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	return DeleteGKECluster(clusterInst.Key.ClusterKey.Name)
}

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster inst not implemented for GCP")
}

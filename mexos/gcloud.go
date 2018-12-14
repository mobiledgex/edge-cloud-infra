package mexos

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud/log"
)

func gcloudCreateGKE(mf *Manifest) error {
	var err error
	if err = gcloud.SetProject(mf.Metadata.Project); err != nil {
		return err
	}
	if err = gcloud.SetZone(mf.Metadata.Zone); err != nil {
		return err
	}
	if err = gcloud.CreateGKECluster(mf.Metadata.Name); err != nil {
		return err
	}
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	saveKubeconfig()
	if err = gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
		return err
	}
	kconf, err := GetKconf(mf, false)
	if err != nil {
		return fmt.Errorf("cannot get kconf, %v, %v, %v", mf, kconf, err)
	}
	if err = copyFile(defaultKubeconfig(), kconf); err != nil {
		return fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
	}
	log.DebugLog(log.
		DebugLevelMexos, "created gke", "name", mf.Spec.Key)
	return nil
}

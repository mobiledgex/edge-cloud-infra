package aws

import (
	"fmt"

	"github.com/codeskyblue/go-sh"
)

//SetProject sets the project in gcloud config
func SetProject(project string) error {
	return nil
}

//SetZone sets the zone in gcloud config
func SetZone(zone string) error {
	return nil
}

//CreateEKSCluster creates a kubernetes cluster on AWS
func CreateEKSCluster(name string) error {
	out, err := sh.Command("eksctl", "create", "cluster", "--name", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

//GetGKECredentials retrieves kubeconfig credentials from gcloud. Often this retrieves wrong x509 certs. This
//  may require you to use `--insecure-skip-tls-verify=true` to `kubectl`

// aws eks --region region-code update-kubeconfig --name cluster_name

func GetEKSCredentials(name string) error {
	out, err := sh.Command("gcloud", "container", "clusters", "get-credentials", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

//DeleteGKECluster removes kubernetes cluster on gcloud
func DeleteEKSCluster(name string) error {
	out, err := sh.Command("gcloud", "container", "clusters", "delete", "--quiet", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

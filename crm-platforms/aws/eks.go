package aws

import (
	"fmt"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
)

//SetOrganizationUnit sets the OrganizationUnit for AWS
func SetOrganizationUnit(awsOu string) error {
	return nil
}

//SetZone sets the zone in AWS
func SetZone(zone string) error {
	return nil
}

//CreateEKSCluster creates a kubernetes cluster on AWS
func CreateEKSCluster(name string, nodeFlavorName string, numNodes uint32) error {
	// output log messages
	log.DebugLog(log.DebugLevelInfra, "CreateEKSCluster Received", "numNodes:", numNodes)
	log.DebugLog(log.DebugLevelInfra, "CreateEKSCluster Received", "nodeFlavorName", nodeFlavorName)
	out, err := sh.Command("eksctl", "create", "cluster", "--name", name, "--node-type", nodeFlavorName, "--nodes", fmt.Sprintf("%d", numNodes), "--managed").CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}

	return nil
}

//GetEKSCredentials retrieves kubeconfig credentials from AWS
// eksctl utils write-kubeconfig myawscluster
// Alternate: aws eks --region region-code update-kubeconfig --name cluster_name
func GetEKSCredentials(name string) error {
	out, err := sh.Command("eksctl", "utils", "write-kubeconfig", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

//DeleteEKSCluster removes kubernetes cluster on AWS
func DeleteEKSCluster(name string) error {
	out, err := sh.Command("eksctl", "delete", "cluster", "--name", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	return nil
}

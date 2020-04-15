package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// GetMasterNameAndIP gets the name and IP address of the cluster's master node.
func (s *OpenstackPlatform) GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "get master IP", "cluster", clusterInst.Key.ClusterKey.Name)
	srvs, err := s.ListServers(ctx)
	if err != nil {
		return "", "", fmt.Errorf("error getting server list: %v", err)

	}
	namePrefix := ClusterTypeKubernetesMasterLabel
	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		namePrefix = ClusterTypeDockerVMLabel
	}

	nodeNameSuffix := k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	masterName, err := s.FindClusterMaster(ctx, namePrefix, nodeNameSuffix, srvs)
	if err != nil {
		return "", "", fmt.Errorf("%s -- %s, %v", infracommon.ServerDoesNotExistError, nodeNameSuffix, err)
	}
	masterIP, err := s.FindNodeIP(masterName, srvs)
	return masterName, masterIP, err
}

func (o *OpenstackPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("TODO")
}

func (o *OpenstackPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	return o.commonPf.CreateClusterInst(ctx, clusterInst, privacyPolicy, updateCallback, timeout)
}

func (o *OpenstackPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	return o.commonPf.DeleteClusterInst(ctx, clusterInst)
}

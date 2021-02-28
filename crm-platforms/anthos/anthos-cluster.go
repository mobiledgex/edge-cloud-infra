package anthos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (a *AnthosPlatform) GetClusterDir(clusterInst *edgeproto.ClusterInst) string {
	return k8smgmt.GetNormalizedClusterName(clusterInst)
}

func (a *AnthosPlatform) GetNamespaceNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return util.K8SSanitize(fmt.Sprintf("%s-%s", clusterInst.Key.Organization, clusterInst.Key.ClusterKey.Name))
}

func (a *AnthosPlatform) SetupVirtualCluster(ctx context.Context, client ssh.Client, namespace, kubeconfig, dir string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupVirtualCluster", "namespace", namespace)

	err := a.CreateNamespace(ctx, client, namespace, kubeconfig)

	if err != nil {
		return err
	}
	policyName := namespace + "-netpol"
	manifest, err := infracommon.CreateK8sNetworkPolicyManifest(ctx, client, policyName, namespace, dir)
	if err != nil {
		return err
	}
	return infracommon.ApplyK8sNetworkPolicyManifest(ctx, client, manifest, a.cloudletKubeConfig)
}

func (a *AnthosPlatform) CreateNamespace(ctx context.Context, client ssh.Client, nameSpace, kubeconfig string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNamespace", "lbname", nameSpace)

	cmd := fmt.Sprintf("kubectl create namespace  %s --kubeconfig=%s", nameSpace, a.cloudletKubeConfig)
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "AlreadyExists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "namespace already exists", "out", out)
		} else {
			return fmt.Errorf("Error in creating namespace: %s - %v", out, err)
		}
	}
	// copy the kubeconfig and update the new one with the new namespace
	log.SpanLog(ctx, log.DebugLevelInfra, "create new kubeconfig for lb namespace", "cloudletKubeConfig", a.cloudletKubeConfig, "clustKubeConfig", kubeconfig)

	err = pc.CopyFile(client, a.cloudletKubeConfig, kubeconfig)
	if err != nil {
		return fmt.Errorf("Failed to create new kubeconfig: %v", err)
	}
	// set the current context
	cmd = fmt.Sprintf("KUBECONFIG=%s kubectl config set-context --current --namespace=%s", kubeconfig, nameSpace)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in setting new namespace context: %s - %v", out, err)
	}
	return nil
}

func (a *AnthosPlatform) DeleteNamespace(ctx context.Context, client ssh.Client, nameSpace string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteNamespace", "nameSpace", nameSpace)
	cmd := fmt.Sprintf("kubectl delete namespace  %s --kubeconfig=%s", nameSpace, a.cloudletKubeConfig)
	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "not found") {
			return fmt.Errorf("Error in deleting namespace: %s - %v", out, err)
		}
	}
	return nil
}

func (a *AnthosPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateClusterInst")
	client, err := a.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: a.commonPf.PlatformConfig.CloudletKey.String(), Type: "anthoscontrolhost"})
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Virtual Cluster")
	err = a.SetupVirtualCluster(ctx, client, a.GetNamespaceNameForCluster(ctx, clusterInst), k8smgmt.GetKconfName(clusterInst), a.GetClusterDir(clusterInst))
	if err != nil {
		return err
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName := a.GetLbNameForCluster(ctx, clusterInst)
		updateCallback(edgeproto.UpdateTask, "Setting up Dedicated Load Balancer")
		err = a.SetupLb(ctx, client, lbName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AnthosPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateClusterInst todo")
}

func (a *AnthosPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteClusterInst")
	client, err := a.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: a.commonPf.PlatformConfig.CloudletKey.String(), Type: "anthoscontrolhost"})
	if err != nil {
		return err
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		externalDev := a.GetExternalEthernetInterface()
		internalDev := a.GetInternalEthernetInterface()
		rootLBName := a.GetLbNameForCluster(ctx, clusterInst)
		lbinfo, err := a.GetLbInfo(ctx, client, rootLBName)
		if err != nil {
			if strings.Contains(err.Error(), LbInfoDoesNotExist) {
				log.SpanLog(ctx, log.DebugLevelInfra, "lbinfo does not exist")

			} else {
				return err
			}
		} else {
			err := a.RemoveIp(ctx, client, lbinfo.ExternalIpAddr, externalDev)
			if err != nil {
				return err
			}
			err = a.RemoveIp(ctx, client, lbinfo.InternalIpAddr, internalDev)
			if err != nil {
				return err
			}
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			err = a.DeleteLbInfo(ctx, client, rootLBName)
			if err != nil {
				return err
			}
			if err = a.commonPf.DeleteDNSRecords(ctx, rootLBName); err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "failed to delete DNS record", "fqdn", rootLBName, "err", err)
			}
		}
	}
	namespace := a.GetNamespaceNameForCluster(ctx, clusterInst)
	err = a.DeleteNamespace(ctx, client, namespace)
	if err != nil {
		return err
	}
	clusterDir := a.GetClusterDir(clusterInst)
	clusterKubeConf := k8smgmt.GetKconfName(clusterInst)

	err = pc.DeleteFile(client, clusterKubeConf)
	if err != nil {
		// DeleteFile uses -f so an error is really a problem
		return fmt.Errorf("Fail to delete cluster kubeconfig")
	}
	err = pc.DeleteDir(ctx, client, clusterDir, pc.NoSudo)
	if err != nil {
		// DeleteDir uses -rf so an error is really a problem
		return fmt.Errorf("Fail to delete cluster kubeconfig")
	}
	return nil
}

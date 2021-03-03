package baremetal

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

func (b *BareMetalPlatform) GetClusterDir(clusterInst *edgeproto.ClusterInst) string {
	return k8smgmt.GetNormalizedClusterName(clusterInst)
}

func (b *BareMetalPlatform) GetNamespaceNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return util.K8SSanitize(fmt.Sprintf("%s-%s", clusterInst.Key.Organization, clusterInst.Key.ClusterKey.Name))
}

func (b *BareMetalPlatform) SetupVirtualCluster(ctx context.Context, client ssh.Client, namespace, kubeconfig, dir string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupVirtualCluster", "namespace", namespace)

	err := b.CreateNamespace(ctx, client, namespace, kubeconfig)

	if err != nil {
		return err
	}
	policyName := namespace + "-netpol"
	manifest, err := infracommon.CreateK8sNetworkPolicyManifest(ctx, client, policyName, namespace, dir)
	if err != nil {
		return err
	}
	return infracommon.ApplyK8sNetworkPolicyManifest(ctx, client, manifest, b.cloudletKubeConfig)
}

func (b *BareMetalPlatform) CreateNamespace(ctx context.Context, client ssh.Client, nameSpace, kubeconfig string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNamespace", "lbname", nameSpace)

	cmd := fmt.Sprintf("kubectl create namespace  %s --kubeconfig=%s", nameSpace, b.cloudletKubeConfig)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in creating namespace: %s - %v", out, err)
	}
	// copy the kubeconfig add update the new one with the new namespace
	log.SpanLog(ctx, log.DebugLevelInfra, "create new kubeconfig for cluster namespace", "cloudletKubeConfig", b.cloudletKubeConfig, "clustKubeConfig", kubeconfig)

	err = pc.CopyFile(client, b.cloudletKubeConfig, kubeconfig)
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

func (b *BareMetalPlatform) DeleteNamespace(ctx context.Context, client ssh.Client, nameSpace string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteNamespace", "nameSpace", nameSpace)
	cmd := fmt.Sprintf("kubectl delete namespace  %s --kubeconfig=%s", nameSpace, b.cloudletKubeConfig)
	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "not found") {
			return fmt.Errorf("Error in deleting namespace: %s - %v", out, err)
		}
	}
	return nil
}

func (b *BareMetalPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateClusterInst")
	client, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Virtual Cluster")
	err = b.SetupVirtualCluster(ctx, client, b.GetNamespaceNameForCluster(ctx, clusterInst), k8smgmt.GetKconfName(clusterInst), b.GetClusterDir(clusterInst))
	if err != nil {
		return err
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName := b.GetLbNameForCluster(ctx, clusterInst)
		updateCallback(edgeproto.UpdateTask, "Setting up Dedicated Load Balancer")
		err = b.SetupLb(ctx, client, lbName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BareMetalPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateClusterInst not supported")
}

func (b *BareMetalPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteClusterInst")
	client, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: b.commonPf.PlatformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
	if err != nil {
		return err
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		externalDev := b.GetExternalEthernetInterface()
		internalDev := b.GetInternalEthernetInterface()
		rootLBName := b.GetLbNameForCluster(ctx, clusterInst)
		lbinfo, err := b.GetLbInfo(ctx, client, rootLBName)
		if err != nil {
			if strings.Contains(err.Error(), LbInfoDoesNotExist) {
				log.SpanLog(ctx, log.DebugLevelInfra, "lbinfo does not exist")

			} else {
				return err
			}
		} else {
			err := b.RemoveIp(ctx, client, lbinfo.ExternalIpAddr, externalDev)
			if err != nil {
				return err
			}
			err = b.RemoveIp(ctx, client, lbinfo.InternalIpAddr, internalDev)
			if err != nil {
				return err
			}
		}
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			err = b.DeleteLbInfo(ctx, client, rootLBName)
			if err != nil {
				return err
			}
			if err = b.commonPf.DeleteDNSRecords(ctx, rootLBName); err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "failed to delete DNS record", "fqdn", rootLBName, "err", err)
			}
		}
	}
	namespace := b.GetNamespaceNameForCluster(ctx, clusterInst)
	err = b.DeleteNamespace(ctx, client, namespace)
	if err != nil {
		return err
	}
	clusterDir := b.GetClusterDir(clusterInst)
	clusterKubeConf := k8smgmt.GetKconfName(clusterInst)

	err = pc.DeleteFile(client, clusterKubeConf)
	if err != nil {
		// DeleteFile uses -f so an error is really approblem
		return fmt.Errorf("Fail to delete cluster kubeconfig")
	}
	err = pc.DeleteDir(ctx, client, clusterDir, pc.NoSudo)
	if err != nil {
		// DeleteDir uses -rf so an error is really a problem
		return fmt.Errorf("Fail to delete cluster kubeconfig")
	}
	return nil
}

package infracommon

import (
	"context"
	"fmt"
	"net"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type MetalConfigmapParams struct {
	AddressRanges []string
}

var MetalLbConfigMap = `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      {{- range .AddressRanges}}
       - {{.}}
      {{- end}}
`

// GetMetalLbIpRangeFromMasterIp gives an IP range on the same subnet as the master IP
func (ip *InfraProperties) GetMetalLbIpRangeFromMasterIp(ctx context.Context, masterIP string) ([]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetMetalLbIpRangeFromMasterIp", "masterIP", masterIP)
	mip := net.ParseIP(masterIP)
	if mip == nil {
		return nil, fmt.Errorf("unable to parse master ip %s", masterIP)
	}
	start, end, err := ip.GetMetalLbIpRange()
	if err != nil {
		return nil, err
	}
	addr := mip.To4()
	addr[3] = byte(start)
	startAddr := addr.String()
	addr[3] = byte(end)
	endAddr := addr.String()
	return []string{fmt.Sprintf("%s-%s", startAddr, endAddr)}, nil
}

func InstallMetalLb(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InstallMetalLb", "clusterInst", clusterInst)
	kconf := k8smgmt.GetKconfName(clusterInst)
	cmds := []string{
		fmt.Sprintf("kubectl create -f https://raw.githubusercontent.com/metallb/metallb/v0.10.2/manifests/namespace.yaml --kubeconfig=%s", kconf),
		fmt.Sprintf("kubectl create -f https://raw.githubusercontent.com/metallb/metallb/v0.10.2/manifests/metallb.yaml --kubeconfig=%s", kconf),
		// fmt.Sprintf("kubectl apply -f \"https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')&env.NO_MASQ_LOCAL=1\" --kubeconfig=%s", kconf),
	}
	for _, cmd := range cmds {
		log.SpanLog(ctx, log.DebugLevelInfra, "installing metallb", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("failed to run metalLb cmd %s, %s, %v", cmd, out, err)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ok, installed metallb")
	return nil
}

func ConfigureMetalLb(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, addressRanges []string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ConfigureMetalLb", "clusterInst", clusterInst, "addressRanges", addressRanges)
	MetalConfigmapParams := MetalConfigmapParams{
		AddressRanges: addressRanges,
	}
	configBuf, err := ExecTemplate("metalLbConfigMap", MetalLbConfigMap, MetalConfigmapParams)
	if err != nil {
		return err
	}
	dir := k8smgmt.GetNormalizedClusterName(clusterInst)
	err = pc.CreateDir(ctx, client, dir, pc.NoOverwrite)
	if err != nil {
		return err
	}
	fileName := dir + "/metalLbConfigMap.yaml"
	err = pc.WriteFile(client, fileName, configBuf.String(), "configMap", pc.NoSudo)
	if err != nil {
		return fmt.Errorf("WriteTemplateFile failed for metal config map: %s", err)
	}
	kconf := k8smgmt.GetKconfName(clusterInst)
	cmd := fmt.Sprintf("kubectl apply -f %s --kubeconfig=%s", fileName, kconf)
	log.SpanLog(ctx, log.DebugLevelInfra, "installing metallb")
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't add configure metallb %s, %s, %v", cmd, out, err)
	}
	return nil
}

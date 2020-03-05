package main

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func updateCallback(updateType edgeproto.CacheUpdateType, value string) {
	fmt.Printf("UPDATE CALLBACK %d - %s\n", updateType, value)
}

func main() {
	log.SetDebugLevel(log.DebugLevelMexos)
	fmt.Printf("Begin\n")
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	mexos.OpenstackProps.OsExternalNetworkName = "external-network-shared"
	mexos.CloudletInfraCommon.NetworkScheme = "name=mex-k8s-net-1,cidr=10.101.X.0/24"
	mexos.OpenstackProps.OsMexNetwork = "mex-k8s-net-1"

	imageName := "mobiledgex-v3.0.3"
	objName := "jlm-vmapp1"

	var ckey edgeproto.CloudletKey
	ckey.Name = "jlmheattest-cloudlet"
	ckey.OperatorKey.Name = "TDG"
	accessPorts := "tcp:7777"
	vmAppParams, err := mexos.GetVMParams(ctx,
		mexos.UserVMDeployment,
		objName,
		"m4.medium",
		0, //vmspec.ExternalVolumeSize,
		imageName,
		mexos.GetSecurityGroupName(ctx, objName),
		&ckey,
		//mexos.WithPublicKey(app.AuthPublicKey),
		mexos.WithAccessPorts(accessPorts),
		//mexos.WithDeploymentManifest(app.DeploymentManifest),
		//mexos.WithCommand(app.Command),
	//	mexos.WithPrivacyPolicy(privacyPolicy),
	)
	if err != nil {
		fmt.Printf("GetVMParams appvm ERROR %v", err)
		return
	}
	rootLBname := "jlmheattest.mobiledgex.net"

	vmLBParams, err := mexos.GetVMParams(ctx,
		mexos.RootLBVMDeployment,
		rootLBname,
		"m4.small",
		0, // clusterInst.ExternalVolumeSize,
		imageName,
		mexos.GetSecurityGroupName(ctx, rootLBname),
		&ckey,
		mexos.WithComputeAvailabilityZone(mexos.GetCloudletComputeAvailabilityZone()),
		mexos.WithVolumeAvailabilityZone(mexos.GetCloudletVolumeAvailabilityZone()),
	//	mexos.WithPrivacyPolicy(appInst.PrivacyPolicy),
	)

	if err != nil {
		fmt.Printf("GetVMParams rootlb ERROR %v", err)
		return
	}

	err = mexos.HeatCreateAppVMWithRootLB(ctx, rootLBname, imageName, objName, vmAppParams, vmLBParams, updateCallback)
	if err != nil {
		fmt.Printf("HeatCreateAppVMWithRootLB ERROR %v", err)
		return
	}
}

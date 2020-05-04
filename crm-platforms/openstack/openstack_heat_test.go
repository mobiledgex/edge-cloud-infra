package openstack

import (
	"context"
	"testing"

	e2esetup "github.com/mobiledgex/edge-cloud-infra/e2e-tests/e2e-setup"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
)

var subnetName = "subnet-test"
var testGroupName = "openstack-test"

var vms = []*vmlayer.VMRequestSpec{
	{
		Name:                    "rootlb-xyz",
		Type:                    vmlayer.VMTypeRootLB,
		FlavorName:              "m1.medium",
		ImageName:               "mobiledgex-v9.9.9",
		ComputeAvailabilityZone: "nova1",
		ExternalVolumeSize:      100,
		ConnectToExternalNet:    true,
		ConnectToSubnet:         subnetName,
	},
	{
		Name:                    "master-xyz",
		Type:                    vmlayer.VMTypeClusterMaster,
		FlavorName:              "m1.medium",
		ImageName:               "mobiledgex-v9.9.9",
		ComputeAvailabilityZone: "nova1",
		ExternalVolumeSize:      100,
		ConnectToExternalNet:    true,
		ConnectToSubnet:         subnetName,
	},
	{
		Name:                    "node1-xyz",
		Type:                    vmlayer.VMTypeClusterNode,
		FlavorName:              "m1.medium",
		ImageName:               "mobiledgex-v9.9.9",
		ComputeAvailabilityZone: "nova1",
		ConnectToSubnet:         subnetName,
	},
	{
		Name:                    "node2-xyz",
		Type:                    vmlayer.VMTypeClusterNode,
		FlavorName:              "m1.medium",
		ImageName:               "mobiledgex-v9.9.9",
		ComputeAvailabilityZone: "nova1",
		ConnectToSubnet:         subnetName,
	},
}

func TestHeatTemplate(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	log.SetDebugLevel(log.DebugLevelInfra)
	infracommon.SetTestMode(true)

	ctx := log.StartTestSpan(context.Background())
	vaultServer, vaultConfig := vault.DummyServer()
	defer vaultServer.Close()

	pc := pf.PlatformConfig{}

	op := OpenstackPlatform{TestMode: true}
	var vmp = vmlayer.VMPlatform{
		Type:       "openstack",
		VMProvider: &op,
	}
	err := vmp.InitProps(ctx, &pc, vaultConfig)
	log.SpanLog(ctx, log.DebugLevelInfra, "init props done", "err", err)

	require.Nil(t, err)

	vmgp, err := vmp.GetVMGroupOrchestrationParamsFromVMSpec(ctx,
		testGroupName,
		vms,
		vmlayer.WithNewSecurityGroup("testvmgroup-sg"),
		vmlayer.WithAccessPorts("tcp:7777,udp:8888"),
		vmlayer.WithNewSubnet(subnetName),
	)

	log.SpanLog(ctx, log.DebugLevelInfra, "got VM group params", "vmgp", vmgp, "err", err)
	require.Nil(t, err)

	err = op.populateParams(ctx, vmgp, heatTest)
	require.Nil(t, err)

	err = op.createOrUpdateHeatStackFromTemplate(ctx, vmgp, testGroupName, VmGroupTemplate, heatTest, edgeproto.DummyUpdateCallback)
	log.SpanLog(ctx, log.DebugLevelInfra, "created test stack file", "err", err)
	require.Nil(t, err)

	generatedFile := testGroupName + "-heat.yaml"
	expectedResultsFile := testGroupName + "-heat-expected.yaml"
	compareResult := e2esetup.CompareYamlFiles(generatedFile, expectedResultsFile, "heat-test")
	log.SpanLog(ctx, log.DebugLevelInfra, "yaml compare result", "compareResult", compareResult)

	require.Equal(t, compareResult, true)

}

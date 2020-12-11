package vcd

import (
	"context"
	"fmt"

	//"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	//"github.com/mobiledgex/edge-cloud/log"
	//"github.com/mobiledgex/edge-cloud/vault"
	//"strings"
)

// model VcdProps after vsphere to start

// This is now an edgeproto object
var VcdProps = map[string]*edgeproto.PropertyInfo{

	"MEX_ORG": {
		Value: "vcd-org",
	},
	"MEX_CATALOG": {
		Mandatory: false,
	},
	"MEX_EXTERNAL_IP_RANGES": {
		Mandatory: false,
	},
	"MEX_EXTERNAL_NETWORK_MASK": {
		Mandatory: false,
	},
	// We don't get a value for the edgegateway xxx
	"MEX_EXTERNAL_NETWORK_EDGEGATEWAY": {
		Mandatory: false,
	},
	"MEX_INTERNAL_NETWORK_MASK": {
		Value: "24",
	},
	/*
		"MEX_EXT_NETWORK": {
			Mandatory: false,
		},
	*/
	//	"MEX_EXTERNAL_VSWITCH": {
	//		Value: "ExternalVSwitch",
	//	},
}

func (v *VcdPlatform) GetVaultCloudletAccessPath(key *edgeproto.CloudletKey, region, physicalName string) string {

	// vshere:
	// return fmt.Sprintf("/secret/data/%s/cloudlet/vsphere/%s/%s/vcenter.json", region, key.Organization, physicalName)
	return fmt.Sprintf("/secret/data/%s/cloudlet/vcd/%s/%s/vcd.json", region, key.Organization, physicalName)

}

func (v *VcdPlatform) GetVcdVars(ctx context.Context, accessApi platform.AccessApi) error {

	fmt.Printf("\n\nGetVcdVars GetCloudletAccessVars...\n\n")

	vars, err := accessApi.GetCloudletAccessVars(ctx)
	if err != nil {
		fmt.Printf("\nGetCloudletAcessVars error: %s\n", err.Error())
		return err
	}
	v.vcdVars = vars

	if len(vars) == 0 {
		panic("no vars!")
	}

	fmt.Printf("\nGetVcdVars env passed down:\n")
	for k, v := range vars {
		fmt.Printf("\tGetVcdVars:next access var  %s = %s\n", k, v)
	}

	fmt.Printf("\n\nGetVcdVars:\n\tVCD_URL: %s\n\tVCD_USERNAME: %s\n\t: Passwd: %s\n\t Org: %s\n\t",
		v.vcdVars["VCD_URL"],
		v.vcdVars["VCD_USERNAME"],
		v.vcdVars["ORG"],
		v.vcdVars["VCD_PASSWORD"])

	err = v.PopulateOrgLoginCredsFromVault(ctx)
	if err != nil {
		fmt.Printf("\nError from pop creds from vault: %s\n", err.Error())
		return err
	}
	v.vcdVars["VCD_URL"] = v.Creds.Href

	return nil
}

// start fetching access  bits from vault
func (v *VcdPlatform) InitApiAccessProperties(ctx context.Context, accessApi platform.AccessApi, vars map[string]string, stage vmlayer.ProviderInitStage) error {

	fmt.Printf("\nInitApiAccessProperties stage: %s\n", stage)
	err := v.GetVcdVars(ctx, accessApi)
	if err != nil {
		return err
	}
	return nil
}

func (v *VcdPlatform) SetProviderSpecificProps(ctx context.Context) error {
	VcdProps["MEX_ORG"].Value = v.Objs.Org.Org.Name
	VcdProps["MEX_EXTERNAL_NETWORK_MASK"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask
	VcdProps["MEX_CATALOG"].Value = v.Objs.PrimaryCat.Catalog.Name
	VcdProps["MEX_EXTERNAL_NETWORK_EDGEGATEWAY"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway
	VcdProps["MEX_EXTERNAL_IP_RANGES"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress + " - " + v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress

	return nil
}

func (v *VcdPlatform) GetProviderSpecificProps(ctx context.Context) (map[string]*edgeproto.PropertyInfo, error) {
	return VcdProps, nil
}

func (v *VcdPlatform) GetExternalNetmask() string {

	if v.vmProperties.Domain == vmlayer.VMDomainPlatform {
		// check for optional management netmask
		val, _ := v.vmProperties.CommonPf.Properties.GetValue("MEX_MANAGEMENT_EXTERNAL_NETWORK_MASK")
		if val != "" {
			return val
		}
	}
	val, _ := v.vmProperties.CommonPf.Properties.GetValue("MEX_EXTERNAL_NETWORK_MASK")
	return val
}

func (v *VcdPlatform) GetInternalNetmask() string {
	val, _ := v.vmProperties.CommonPf.Properties.GetValue("MEX_INTERNAL_NETWORK_MASK")
	return val
}

/*
func (a *VcdPlatform) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {

	fmt.Printf("\n\nVcdPlatform::GetAccessData vaultConfig: %+v\n\n", vaultConfig)

	log.SpanLog(ctx, log.DebugLevelInfra, "VcdPlatform GetAccessData", "dataType", dataType)
	switch dataType {
	case accessapi.GetCloudletAccessVars:
		vars, err := infracommon.GetEnvVarsFromVault(ctx, vaultConfig, v.G)
		if err != nil {
			return nil, err
		}
		return vars, nil
	}
	return nil, fmt.Errorf("Azure unhandled GetAccessData type %s", dataType)
}
*/

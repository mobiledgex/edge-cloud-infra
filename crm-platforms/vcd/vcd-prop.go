package vcd

import (
	"context"
	"fmt"
	"os"

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

	fmt.Printf("\n\nGetVcdVars:\n\tVCD_IP: %s\n\tVCD_USER: %s\n\tPasswd: %s\n\t Org: %s\n\tVDC_NAME: %s\n",
		v.vcdVars["VCD_IP"],
		v.vcdVars["VCD_USER"],
		v.vcdVars["VCD_PASSWORD"],
		v.vcdVars["VCD_ORG"],
		v.vcdVars["VDC_NAME"])

	err = v.PopulateOrgLoginCredsFromVault(ctx)
	if err != nil {
		fmt.Printf("\nError from pop creds from vault: %s\n", err.Error())
		return err
	}

	v.vcdVars["VCD_URL"] = v.Creds.Href
	fmt.Printf("\n\nGetVcdVars login Creds.Href: %s\n\n", v.Creds.Href)

	return nil
}

func (v *VcdPlatform) GetVCDIP() string {

	return v.vcdVars["VCD_IP"]
}
func (v *VcdPlatform) GetVCDUser() string {

	return v.vcdVars["VCD_USER"]
}
func (v *VcdPlatform) GetVCDPassword() string {

	return v.vcdVars["VCD_PASSWORD"]
}
func (v *VcdPlatform) GetVCD_ORG() string {

	return v.vcdVars["VCD_ORG"]
}
func (v *VcdPlatform) GetVDCName() string {

	return v.vcdVars["VDC_NAME"]
}
func (v *VcdPlatform) GetVCDURL() string {

	return v.vcdVars["VCD_URL"]
}

// Sort out the spelling VCD vs VDC template name in all the secrets. It's offically a vdc template.
func (v *VcdPlatform) GetVDCTemplateName() string {
	if v.TestMode {
		tmplName := os.Getenv("VCDTEMPLATE")
		if tmplName != "" {
			return tmplName
		}

		return os.Getenv("VDCTEMPLATE")
	}
	tmplName := v.vcdVars["VCDTEMPLATE"]
	if tmplName != "" {
		return tmplName
	}

	return v.vcdVars["VDCTEMPLATE"]
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

	// Put template selection bits here
	// can't know this til post discovery, and most we already know XXX
	/*
		VcdProps["MEX_ORG"].Value = v.Objs.Org.Org.Name
		VcdProps["MEX_EXTERNAL_NETWORK_MASK"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask
		VcdProps["MEX_CATALOG"].Value = v.Objs.PrimaryCat.Catalog.Name
		VcdProps["MEX_EXTERNAL_NETWORK_EDGEGATEWAY"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway
		VcdProps["MEX_EXTERNAL_IP_RANGES"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress + " - " + v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress
	*/
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

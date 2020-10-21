package vcd

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"strings"
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
	//	"MEX_EXTERNAL_VSWITCH": {
	//		Value: "ExternalVSwitch",
	//	},
}

func (v *VcdPlatform) GetApiAccessFilename() string {
	fmt.Printf("GetApiAccessFilename-I config_file.yaml \n")

	return "config_file.yaml"

}

func (v *VcdPlatform) GetVcdVars(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config) error {

	//fmt.Printf("GetVcdVars-I-cloudlet %s physicalname: %s using vault config %+v\n", key.Name, physicalName, vaultConfig)

	if vaultConfig == nil || vaultConfig.Addr == "" {
		return fmt.Errorf("vaultAddr is not specified")
	}
	vcpath := vmlayer.GetVaultCloudletAccessPath(key, region, v.GetType(), physicalName, v.GetApiAccessFilename())
	log.SpanLog(ctx, log.DebugLevelInfra, "interning vault", "addr", vaultConfig.Addr, "path", vcpath)
	envData := &infracommon.VaultEnvData{}
	err := vault.GetData(vaultConfig, vcpath, 0, envData)
	if err != nil {
		fmt.Printf("GetVcdVars failed to access vault for key =  %s use env vars instead \n", physicalName)

		if strings.Contains(err.Error(), "no secrets") {
			return fmt.Errorf("Failed to source access variables as '%s/%s' "+
				"does not exist in secure secrets storage (Vault)",
				key.Organization, physicalName)
		}
		return fmt.Errorf("Failed to source access variables from %s, %s: %v", vaultConfig.Addr, vcpath, err)
	}
	v.vcdVars = make(map[string]string, 1)
	for _, envData := range envData.Env {
		v.vcdVars[envData.Name] = envData.Value
	}

	host := v.GetVcdAddress()
	if err != nil {
		return err
	}
	v.vcdVars["VCD_URL"] = host // api endpoint
	v.vcdVars["VCD_USERNAME"] = v.GetVcdUser()
	pass := v.GetVcdPassword()
	v.vcdVars["ORG"] = v.GetVcdOrgName()
	v.vcdVars["VCD_PASSWORD"] = pass
	//	v.vcdVars["VCD_INSECURE"] = v.GetVcdInsecure()  true by default XXX

	return nil
}

// start of fetching access  bits from vault
func (v *VcdPlatform) InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	fmt.Printf("InitApiAccessProperties-TBI\n")
	return nil
}

func (v *VcdPlatform) SetProviderSpecificProps(ctx context.Context) error {
	// What's our vcd Org name?
	VcdProps["MEX_ORG"].Value = v.Objs.Org.Org.Name
	VcdProps["MEX_EXTERNAL_NETWORK_MASK"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Netmask
	VcdProps["MEX_CATALOG"].Value = v.Objs.PrimaryCat.Catalog.Name
	VcdProps["MEX_EXTERNAL_NETWORK_EDGEGATEWAY"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].Gateway
	VcdProps["MEX_EXTERNAL_IP_RANGES"].Value = v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress + " - " + v.Objs.PrimaryNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].EndAddress

	fmt.Printf("\nProviderSPecificProps:\n")
	for name, value := range VcdProps {
		fmt.Printf("\t%s : %s\n", name, value.Value)
	}
	return nil
}

// What are our specific properties?
func (v *VcdPlatform) GetProviderSpecificProps(ctx context.Context, vconf *vault.Config) (map[string]*edgeproto.PropertyInfo, error) {

	fmt.Printf("GetProviderProps returning %+v\n", VcdProps)

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

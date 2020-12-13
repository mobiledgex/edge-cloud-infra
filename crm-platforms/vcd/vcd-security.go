package vcd

import (
	"context"
	"fmt"
	"github.com/mobiledgex/edge-cloud/log"
	"net/url"
	"os"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// vcd security related operations

// physicalname (vault key) not needed when  using insure env vars.
func (v *VcdPlatform) PopulateOrgLoginCredsFromEnv(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateOrgLoginCredsFromEnv")

	creds := VcdConfigParams{
		User:     os.Getenv("VCD_USER"),
		Password: os.Getenv("VCD_PASSWD"),
		Org:      os.Getenv("VCD_ORG"),
		Href:     fmt.Sprintf("https://%s/api", os.Getenv("VCD_IP")),
		VDC:      os.Getenv("VDC_NAME"),
		Insecure: true,
	}
	if creds.User == "" {
		return fmt.Errorf("User not defined")
	}
	if creds.Password == "" {
		return fmt.Errorf("Passwd not defined")
	}
	if creds.Org == "" {
		return fmt.Errorf("Org not defined")
	}
	if creds.Href == "" {
		return fmt.Errorf("Href not defined")
	}
	if creds.VDC == "" {
		return fmt.Errorf("missing VDC name")
	}
	v.Creds = &creds
	return nil
}

func (v *VcdPlatform) GetVcdUser() string {
	return v.Creds.User
}
func (v *VcdPlatform) GetVcdPassword() string {
	return v.Creds.Password
}
func (v *VcdPlatform) GetVcdOrgName() string {
	return v.Creds.Org
}
func (v *VcdPlatform) GetVcdVdcName() string {
	return v.Creds.VDC
}
func (v *VcdPlatform) GetVcdAddress() string {

	return v.Creds.Href
}

func (v *VcdPlatform) PopulateOrgLoginCredsFromVault(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateOrgLoginCredsFromVault")
	creds := VcdConfigParams{
		User:     v.vcdVars["VCD_USER"],
		Password: v.vcdVars["VCD_PASSWORD"],
		Org:      v.vcdVars["VCD_ORG"],
		Href:     fmt.Sprintf("https://%s/api", v.vcdVars["VCD_IP"]),
		VDC:      v.vcdVars["VDC_NAME"],
		Insecure: true,
	}

	if creds.User == "" {
		return fmt.Errorf("User not defined")
	}
	if creds.Password == "" {
		return fmt.Errorf("Passwd not defined")
	}
	if creds.Org == "" {
		return fmt.Errorf("Org not defined")
	}
	if creds.Href == "" {
		return fmt.Errorf("Href not defined")
	}
	if creds.VDC == "" {
		return fmt.Errorf("missing VDC name")
	}
	v.Creds = &creds

	fmt.Printf("\nClient login Creds:\n\tOrg: %s\n\tVDC: %s\n\tVCD_USER: %s\n\tVCD_PASSWORD: %s\n\tHref: %s\n\n",
		creds.Org, creds.VDC, creds.User, creds.Password, creds.Href)

	return nil
}

// PI
func (v *VcdPlatform) PrepareRootLB(ctx context.Context, client ssh.Client, rootLBName string, secGrpName string, privacyPolicy *edgeproto.PrivacyPolicy) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB TBI", "rootLBName", rootLBName)
	// configure iptables based security
	// allow our external vsphere network
	//sshCidrsAllowed := []string{}
	//externalNet, err := v.GetExternalIpNetworkCidr(ctx)
	//if err != nil {
	//		return err
	//	}
	//	sshCidrsAllowed = append(sshCidrsAllowed, externalNet)
	// xxx not clear if we cannot use native security firewalls in the vApp
	// return v.vmProperties.SetupIptablesRulesForRootLB(ctx, client, sshCidrsAllowed, privacyPolicy)
	return nil
}

// IP     tranlate this into our FirewalRule and apply it to the network of this 'serverName'
func (v *VcdPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, serverName, label string, allowedCIDR string, ports []dme.AppPort) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules N TBI", "secGrpName", secGrpName, "allowedCIDR", allowedCIDR, "ports", ports)

	return nil
}

// Refactor, won't be using ssh.Client here, need our VDCClient
func (v *VcdPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, secGrpName, label string, allowedCIDR string, ports []dme.AppPort) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules", "secGrpName", secGrpName, "allowedCIDR", allowedCIDR, "ports", ports)

	return nil
}

// XXX
func setVer(cli *govcd.VCDClient) error {
	if cli != nil {
		cli.Client.APIVersion = "33.0" // vmwarelab is 10.1
	}
	return nil
}

// Login and return connected client
// client.Client = types.Client
func (v *VcdPlatform) GetClient(ctx context.Context, creds *VcdConfigParams, test bool) (client *govcd.VCDClient, err error) {

	if test {
		fmt.Printf("GetClient-I-Test get creds from env\n")
		err := v.PopulateOrgLoginCredsFromEnv(ctx)
		if err != nil {
			fmt.Printf("GetClient err getting creds from env: %s\n", err.Error())
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetClient", "Credentails", creds)
	}

	if v.Client != nil {
		fmt.Printf("\n\nGetClient-W-already have non-nil vcd client for org %s\n\n", v.Creds.Org)
		return v.Client, nil
	}
	u, err := url.ParseRequestURI(creds.Href)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request to org %s at %s err: %s", creds.Org, creds.Href, err)
	}
	vcdClient := govcd.NewVCDClient(*u, creds.Insecure)

	if vcdClient.Client.VCDToken != "" {
		_ = vcdClient.SetToken(creds.Org, govcd.AuthorizationHeader, creds.Token)
	} else {
		_, err := vcdClient.GetAuthResponse(creds.User, creds.Password, creds.Org)
		if err != nil {
			return nil, fmt.Errorf("Unable to login to org %s at %s err: %s", creds.Org, creds.Href, err)
		}
		//creds.Token = resp.Header[govcd.AuthorizationHeader]
	}

	// prefer the highest Api version found on the other end.
	// vCD 10.0 == Api 33
	// vCD 10.1 == Api 34
	// The VMware vcd is 10.1 and highest is 34, but if we change it
	// to say 33, we get ENF (entity not found) for our vdc <sigh>
	// So find another way to set this API version. By default,
	// vcdClient.Client.APIVersion == 31.0 which is a 9.5 version.

	// Ok, checkout api_vcd.go, we'd need to adjust the loginURL
	// to match the version change. NewVCDClient could be used with options
	// but 10.1 supports 31.0, so until we care, we don't.

	/*
		if vcdClient.Client.APIVCDMaxVersionIs(">= 33.0") {
			fmt.Printf("APIVCDMaxVersionIs of >= 33.0 is true")
			vcdClient.Client.APIVersion = "33.0"
		}
		if vcdClient.Client.APIClientVersionIs("= 34.0") {
			fmt.Printf("Talking with vCD v10.1 using API v 34.0\n")
			vcdClient.Client.APIVersion = "34.0"
		}
	*/
	v.Client = vcdClient
	log.SpanLog(ctx, log.DebugLevelInfra, "GetClient connected", "API Version", v.Client.Client.APIVersion)

	return vcdClient, nil

}

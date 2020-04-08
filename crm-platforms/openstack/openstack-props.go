package openstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

// NoConfigExternalRouter is used for the case in which we don't manage the external
// router and don't add ports to it ourself, as happens with Contrail.  The router does exist in
// this case and we use it to route from the LB to the pods
var NoConfigExternalRouter = "NOCONFIG"

// NoExternalRouter means there is no router at all and we connect the LB to the k8s pods on the same subnet
// this may eventually be the default and possibly only option
var NoExternalRouter = "NONE"
var openstackProps = map[string]*infracommon.PropertyInfo{
	"MEX_ROUTER": &infracommon.PropertyInfo{
		Value: NoExternalRouter,
	},
}

func GetVaultCloudletAccessPath(key *edgeproto.CloudletKey, region, physicalName string) string {
	return fmt.Sprintf("/secret/data/%s/cloudlet/openstack/%s/%s/%s", region, key.Organization, physicalName, "openrc.json")
}

func GetCertFilePath(key *edgeproto.CloudletKey) string {
	return fmt.Sprintf("/tmp/%s.%s.cert", key.Name, key.Organization)
}

func (s *OpenstackPlatform) GetOpenRCVars(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config) error {
	if vaultConfig == nil || vaultConfig.Addr == "" {
		return fmt.Errorf("vaultAddr is not specified")
	}
	openRCPath := GetVaultCloudletAccessPath(key, region, physicalName)
	log.SpanLog(ctx, log.DebugLevelMexos, "interning vault", "addr", vaultConfig.Addr, "path", openRCPath)
	envData := &infracommon.VaultEnvData{}
	err := vault.GetData(vaultConfig, openRCPath, 0, envData)
	if err != nil {
		if strings.Contains(err.Error(), "no secrets") {
			return fmt.Errorf("Failed to source access variables as '%s/%s' "+
				"does not exist in secure secrets storage (Vault)",
				key.Organization, physicalName)
		}
		return fmt.Errorf("Failed to source access variables from %s, %s: %v", vaultConfig.Addr, openRCPath, err)
	}
	s.openRCVars = make(map[string]string, 1)
	for _, envData := range envData.Env {
		s.openRCVars[envData.Name] = envData.Value
	}
	if authURL, ok := s.openRCVars["OS_AUTH_URL"]; ok {
		if strings.HasPrefix(authURL, "https") {
			if certData, ok := s.openRCVars["OS_CACERT_DATA"]; ok {
				certFile := GetCertFilePath(key)
				err = ioutil.WriteFile(certFile, []byte(certData), 0644)
				if err != nil {
					return err
				}
				s.openRCVars["OS_CACERT"] = certFile
			}
		}
	}
	return nil
}

func (o *OpenstackPlatform) InitOpenstackProps(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	err := o.GetOpenRCVars(ctx, key, region, physicalName, vaultConfig)
	if err != nil {
		return err
	}
	return nil
}

func (s *OpenstackPlatform) GetCloudletProjectName() string {
	if val, ok := s.openRCVars["OS_PROJECT_NAME"]; ok {
		return val
	}
	return ""
}

//GetCloudletExternalRouter returns default MEX external router name
func (c *OpenstackPlatform) GetCloudletExternalRouter() string {
	return c.commonPf.Properties["MEX_ROUTER"].Value
}

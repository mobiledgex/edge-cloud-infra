package mexos

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
)

//MEXCheckEnvVars sets up environment vars and checks for credentials required for running
func MEXCheckEnvVars(mf *Manifest) error {
	// secrets to be passed via Env var still : MEX_CF_KEY, MEX_CF_USER, MEX_DOCKER_REG_PASS
	// TODO: use `secrets` or `vault`
	for _, evar := range []string{
		"MEX_CF_KEY",
		"MEX_CF_USER",
		"MEX_DOCKER_REG_PASS",
	} {
		if v := mexEnv(mf, evar); v == "" {
			return fmt.Errorf("missing env var %s", evar)
		}
	}
	//original base VM image uses id_rsa_mobiledgex key
	if mexEnv(mf, "MEX_OS_IMAGE") == "mobiledgex-16.04-2" && mexEnv(mf, "MEX_SSH_KEY") != "id_rsa_mobiledgex" {
		return fmt.Errorf("os image %s cannot use key %s", mexEnv(mf, "MEX_OS_IMAGE"), mexEnv(mf, "MEX_SSH_KEY"))
	}
	//packer VM image uses id_rsa_mex key
	if mexEnv(mf, "MEX_OS_IMAGE") == "mobiledgex" && mexEnv(mf, "MEX_SSH_KEY") != "id_rsa_mex" {
		return fmt.Errorf("os image %s cannot use key %s", mexEnv(mf, "MEX_OS_IMAGE"), mexEnv(mf, "MEX_SSH_KEY"))
	}
	//TODO need to allow users to save the environment under platform name inside .mobiledgex or Vault
	return nil
}

func mexEnv(mf *Manifest, name string) string {
	v, ok := mf.Values.VaultEnvMap[name]
	if !ok {
		log.DebugLog(log.DebugLevelMexos, "error, env var not exist in vault", "name", name)
		return ""
	}
	return v
}

package mexos

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
)

var validMEXOSEnv = false

func validateClusterKind(kind string) error {
	log.DebugLog(log.DebugLevelMexos, "cluster kind", "kind", kind)
	for _, k := range []string{"gcp", "azure"} {
		if kind == k {
			return nil
		}
	}
	if strings.HasPrefix(kind, "mex-") {
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "warning, cluster kind, operator has no mex- prefix", "kind", kind)
	return nil
}

func CheckManifest(mf *Manifest) error {
	if mf.APIVersion == "" {
		return fmt.Errorf("mf apiversion not set")
	}
	if mf.APIVersion != APIversion {
		return fmt.Errorf("invalid api version")
	}
	return CheckManifestValues(mf)
}

func CheckManifestValues(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "checking manifest values")
	return checkVals(mf.Values)
}

func checkVals(vals interface{}) error {
	m := reflect.ValueOf(vals)
	//log.DebugLog(log.DebugLevelMexos, "checking", "type", m.Type())
	if strings.HasPrefix(m.Type().String(), "map[") {
		return nil
	}
	for i := 0; i < m.NumField(); i++ {
		n := m.Type().Field(i).Name
		t := m.Type().Field(i).Type
		v := m.Field(i).Interface()
		if t.String() == "string" {
			if v == "" {
				return fmt.Errorf("empty %s", n)
			}
			//log.DebugLog(log.DebugLevelMexos, "ok", "name", n, "type", t, "value", v)
		} else {
			if err := checkVals(m.Field(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCommon(mf *Manifest) error {
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name for the deployment")
	}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty image")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty proxy path")
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing DNS zone, metadata %v", mf.Metadata)
	}
	return nil
}

//ValidateMEXOSEnv makes sure the environment is valid for mexos
func ValidateMEXOSEnv(osEnvValid bool) bool {
	validMEXOSEnv = false
	if !osEnvValid {
		log.DebugLog(log.DebugLevelMexos, "invalid mex env")
		return false
	}
	validMEXOSEnv = true
	log.DebugLog(log.DebugLevelMexos, "ok, valid mex env")
	return validMEXOSEnv
}

func IsValidMEXOSEnv() bool {
	return validMEXOSEnv
}

//ValidateNetSpec parses and validates the netSpec
func ValidateNetSpec(netSpec string) error {
	if netSpec == "" {
		return fmt.Errorf("empty netspec")
	}
	return nil
}

//ValidateTags parses and validates tags
func ValidateTags(tags string) error {
	if tags == "" {
		return fmt.Errorf("empty tags")
	}
	return nil
}

//ValidateTenant parses and validates tenant
func ValidateTenant(tenant string) error {
	if tenant == "" {
		return fmt.Errorf("emtpy tenant")
	}
	return nil
}

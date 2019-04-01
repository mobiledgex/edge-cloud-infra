package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
	// AzureProperties should be moved to edge-cloud-infra
	props edgeproto.AzureProperties
}

func (s *Platform) GetType() string {
	return "azure"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	if err := mexos.InitInfraCommon(); err != nil {
		return err
	}
	s.props.Location = os.Getenv("MEX_AZURE_LOCATION")
	if s.props.Location == "" {
		return fmt.Errorf("Env variable MEX_AZURE_LOCATION not set")
	}
	/** resource group currently derived from cloudletname + cluster name
			s.props.ResourceGroup = os.Getenv("MEX_AZURE_RESOURCE_GROUP")
			if s.props.ResourceGroup == "" {
				return fmt.Errorf("Env variable MEX_AZURE_RESOURCE_GROUP not set")
	                }
	*/
	s.props.UserName = os.Getenv("MEX_AZURE_USER")
	if s.props.UserName == "" {
		return fmt.Errorf("Env variable MEX_AZURE_USER not set, check contents of MEXENV_URL")
	}
	s.props.Password = os.Getenv("MEX_AZURE_PASS")
	if s.props.Password == "" {
		return fmt.Errorf("Env variable MEX_AZURE_PASS not set, check contents of MEXENV_URL")
	}

	// TODO: these should be detected
	mexos.SetAvailableClusterFlavors(mexos.AzureClusterFlavors)
	return nil
}

type AZName struct {
	LocalizedValue string
	Value          string
}

type AZLimit struct {
	CurrentValue string
	Limit        string
	LocalName    string
	Name         AZName
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	log.DebugLog(log.DebugLevelMexos, "GetLimits (Azure)")

	var limits []AZLimit
	out, err := sh.Command("az", "vm", "list-usage", "--location", s.props.Location, sh.Dir("/tmp")).Output()
	if err != nil {
		err = fmt.Errorf("cannot get limits from azure, %v", err)
		return err
	}
	err = json.Unmarshal(out, &limits)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return err
	}
	for _, l := range limits {
		if l.LocalName == "Total Regional vCPUs" {
			vcpus, err := strconv.Atoi(l.Limit)
			if err != nil {
				err = fmt.Errorf("failed to parse azure output, %v", err)
				return err
			}
			info.OsMaxVcores = uint64(vcpus)
			info.OsMaxRam = uint64(4 * vcpus)
			info.OsMaxVolGb = uint64(500 * vcpus)
			break
		}
	}
	return nil
}

func (s *Platform) GetPlatformClient() pc.PlatformClient {
	return &pc.LocalClient{}
}

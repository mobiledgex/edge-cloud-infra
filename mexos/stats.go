package mexos

import (
	"encoding/json"
	"fmt"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

//GetLimits is used to retrieve tenant level platform stats
func GetLimits(info *edgeproto.CloudletInfo) error {
	log.DebugLog(log.DebugLevelMexos, "GetLimits")
	switch GetCloudletKind() {
	case cloudcommon.CloudletKindDIND:
		// todo: we could try to get this from the local machine
		info.OsMaxVcores = 8
		info.OsMaxRam = 16
		info.OsMaxVolGb = 500
	case cloudcommon.CloudletKindAzure:
		/*
			var limits []AZLimit
			out, err := sh.Command("az", "vm", "list-usage", "--location", GetCloudletAzureLocation(), sh.Dir("/tmp")).Output()
			if err != nil {
				err = fmt.Errorf("cannot get limits from azure, %v", err)
				return err
			}
			err = json.Unmarshal(out, &limits)
			if err != nil {
				err = fmt.Errorf("cannot unmarshal, %v", err)
				return err
			}
			log.DebugLog(log.DebugLevelMexos, "get limits azzure", "limits", limits)
			for _, l := range limits {
				if l.LocalName == "Total Regional vCPUs" {
					info.OsMaxVcores = uint64(l.Limit)
				} else if l.Name == "StandardSSDStorageDisks" {
					info.OsMaxVolGb = uint64(l.Limit)
				}
			}
		*/

		/*
		 * Since azure doesn't support custom flavors, it is better to leave the limits
		 * check to azure. We can set our own internal limit to make sure the end-user
		 * doesn't use a lot of it
		 */
		info.OsMaxVcores = 8
		info.OsMaxRam = 16
		info.OsMaxVolGb = 500
	default:
		var limits []OSLimit
		//err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).WriteStdout("os-out.txt")
		out, err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).Output()
		if err != nil {
			err = fmt.Errorf("cannot get limits from openstack, %v", err)
			return err
		}
		err = json.Unmarshal(out, &limits)
		if err != nil {
			err = fmt.Errorf("cannot unmarshal, %v", err)
			return err
		}
		for _, l := range limits {
			if l.Name == "MaxTotalCores" {
				info.OsMaxRam = uint64(l.Value)
			} else if l.Name == "MaxTotalRamSize" {
				info.OsMaxVcores = uint64(l.Value)
			} else if l.Name == "MaxTotalVolumeGigabytes" {
				info.OsMaxVolGb = uint64(l.Value)
			}
		}
	}
	log.DebugLog(log.DebugLevelMexos, "get limits", "limits", info)
	return nil
}

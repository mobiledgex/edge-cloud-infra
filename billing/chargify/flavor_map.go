package chargify

import (
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var dmeApiCode = "dmeapi"

// gets the corresponding compoenent code for the flavor
func getComponentCode(flavor string, cloudlet *edgeproto.CloudletKey, start, end time.Time) string {
	// for now just return flavor, later on we can get more complex with different prices based on cloudlet and peak usage times
	// Handle must start with a letter or number and may only contain lowercase letters, numbers, or the characters ':', '-', or '_'
	// replace .&,!
	return "handle:" + strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(flavor, ".", ":"), "&", ":"), ",", ":"), "!", ":")
}

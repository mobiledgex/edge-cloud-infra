package oscli

import (
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
)

// These are operator specific custom vars

var eMEXExternalRouter = os.Getenv("MEX_EXT_ROUTER")   // mex-k8s-router-1
var eMEXNetwork = os.Getenv("MEX_NETWORK")             //mex-k8s-net-1
var eMEXExternalNetwork = os.Getenv("MEX_EXT_NETWORK") // "external-network-shared"

var defaultMEXNet = "mex-k8s-net-1"
var defaultMEXRouter = "mex-k8s-router-1"
var defaultMEXExternalNetwork = "external-network-shared"
var defaultSecurityRule = "default"

//default net should be xternal net but on some cloudlets it is not set as default
// and there may not be a default!

func init() {
	if eMEXExternalRouter == "" {
		eMEXExternalRouter = defaultMEXRouter
	}

	if eMEXNetwork == "" {
		eMEXNetwork = defaultMEXNet
	}

	if eMEXExternalNetwork == "" {
		eMEXExternalNetwork = defaultMEXExternalNetwork
	}

}

//GetMEXExternalRouter returns default MEX external router name
func GetMEXExternalRouter() string {
	//TODO validate existence and status
	return eMEXExternalRouter
}

//GetMEXExternalNetwork returns default MEX external network name
func GetMEXExternalNetwork() string {
	//TODO validate existence and status
	return eMEXExternalNetwork
}

//GetMEXNetwork returns default MEX network, internal and prepped
func GetMEXNetwork() string {
	//TODO validate existence and status
	return eMEXNetwork
}

//ValidateNetwork makes sure basic network setup is done for MEX
func ValidateNetwork() error {
	nets, err := ListNetworks()
	if err != nil {
		return err
	}

	found := false
	for _, n := range nets {
		if n.Name == eMEXExternalNetwork {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find %s", eMEXExternalNetwork)
	}

	found = false
	for _, n := range nets {
		if n.Name == eMEXNetwork {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find %s", eMEXNetwork)
	}

	routers, err := ListRouters()
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == eMEXExternalRouter {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("router %s not found", eMEXExternalRouter)
	}

	return nil
}

//PrepNetwork validates and does the work needed to ensure MEX network setup
func PrepNetwork() error {
	nets, err := ListNetworks()
	if err != nil {
		return err
	}

	// Not having external network setup by TDG is a hard error.
	// TDG must have setup a network connected to external / internet
	// that is named properly.
	// This is the case at Bonn.
	// The providers are expected to set up one external shared internet
	// routed network with a specific name.

	found := false
	for _, n := range nets {
		if n.Name == eMEXExternalNetwork {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find %s", eMEXExternalNetwork)
	}

	found = false
	for _, n := range nets {
		if n.Name == eMEXNetwork {
			found = true
			break
		}
	}
	if !found {
		// We need at least one network for `mex` clusters
		err = CreateNetwork(eMEXNetwork)
		if err != nil {
			return fmt.Errorf("cannot create %s, %v", eMEXNetwork, err)
		}
	}

	routers, err := ListRouters()
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == eMEXExternalRouter {
			found = true
			break
		}
	}
	if !found {
		// We need at least one router for our `mex` network and external network
		err = CreateRouter(eMEXExternalRouter)
		if err != nil {
			return fmt.Errorf("cannot create %s, %v", eMEXExternalRouter, err)
		}
		err = SetRouter(eMEXExternalRouter, defaultMEXExternalNetwork)
		if err != nil {
			return fmt.Errorf("cannot set default network to router %s, %v", eMEXExternalRouter, err)
		}
	}

	ports := []int{
		18889, //mexosagent HTTP server
		18888, //mexosagent GRPC server
		443,   //mexosagent reverse proxy HTTPS
		8001,  //kubectl proxy
	}

	ruleName := defaultSecurityRule
	for _, p := range ports {
		err := AddSecurityRule(ruleName, p)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "warning, error while adding security rule", "error", err)
		}
	}

	return nil
}

//GetMEXSubnets returns subnets inside MEX Network
func GetMEXSubnets() ([]string, error) {
	nd, err := GetNetworkDetail(eMEXNetwork)
	if err != nil {
		return nil, fmt.Errorf("can't get MEX network detail, %v", err)
	}

	subnets := strings.Split(nd.Subnets, ",")
	if len(subnets) < 1 {
		return nil, fmt.Errorf("can't get a list of subnets for MEX network")
	}

	return subnets, nil
}

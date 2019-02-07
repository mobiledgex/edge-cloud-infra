package mexos

import (
	"fmt"
	"strings"
)

//There needs to be one file like this per provider/operator.  This one is for tdg.

func ValidateNetwork(mf *Manifest) error {
	nets, err := ListNetworks()
	if err != nil {
		return err
	}

	found := false
	for _, n := range nets {
		if n.Name == GetCloudletExternalNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find external network %s", GetCloudletExternalNetwork())
	}

	found = false
	for _, n := range nets {
		if n.Name == GetCloudletNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find network %s", GetCloudletNetwork())
	}

	routers, err := ListRouters()
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == GetCloudletExternalRouter() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("ext router %s not found", GetCloudletExternalRouter())
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
		if n.Name == GetCloudletExternalNetwork() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find ext net %s", GetCloudletExternalNetwork())
	}

	found = false
	for _, n := range nets {
		if n.Name == GetCloudletNetwork() {
			found = true
			break
		}
	}
	if !found {
		// We need at least one network for `mex` clusters
		err = CreateNetwork(GetCloudletNetwork())
		if err != nil {
			return fmt.Errorf("cannot create mex network %s, %v", GetCloudletNetwork(), err)
		}
	}

	routers, err := ListRouters()
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == GetCloudletExternalRouter() {
			found = true
			break
		}
	}
	if !found {
		// We need at least one router for our `mex` network and external network
		err = CreateRouter(GetCloudletExternalRouter())
		if err != nil {
			return fmt.Errorf("cannot create the ext router %s, %v", GetCloudletExternalRouter(), err)
		}
		err = SetRouter(GetCloudletExternalRouter(), GetCloudletExternalNetwork())
		if err != nil {
			return fmt.Errorf("cannot set default network to router %s, %v", GetCloudletExternalRouter(), err)
		}
	}

	return nil
}

//GetCloudletSubnets returns subnets inside MEX Network
func GetCloudletSubnets() ([]string, error) {
	nd, err := GetNetworkDetail(GetCloudletNetwork())
	if err != nil {
		return nil, fmt.Errorf("can't get MEX network detail, %v", err)
	}

	subnets := strings.Split(nd.Subnets, ",")
	if len(subnets) < 1 {
		return nil, fmt.Errorf("can't get a list of subnets for MEX network")
	}

	return subnets, nil
}

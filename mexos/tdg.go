package mexos

import (
	"fmt"
	"strings"
)

//There needs to be one file like this per provider/operator.  This one is for tdg.

func ValidateNetwork(mf *Manifest) error {
	nets, err := ListNetworks(mf)
	if err != nil {
		return err
	}

	found := false
	for _, n := range nets {
		if n.Name == GetMEXExternalNetwork(mf) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find external network %s", GetMEXExternalNetwork(mf))
	}

	found = false
	for _, n := range nets {
		if n.Name == GetMEXNetwork(mf) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find network %s", GetMEXNetwork(mf))
	}

	routers, err := ListRouters(mf)
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == GetMEXExternalRouter(mf) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("ext router %s not found", GetMEXExternalRouter(mf))
	}

	return nil
}

//PrepNetwork validates and does the work needed to ensure MEX network setup
func PrepNetwork(mf *Manifest) error {
	nets, err := ListNetworks(mf)
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
		if n.Name == GetMEXExternalNetwork(mf) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find ext net %s", GetMEXExternalNetwork(mf))
	}

	found = false
	for _, n := range nets {
		if n.Name == GetMEXNetwork(mf) {
			found = true
			break
		}
	}
	if !found {
		// We need at least one network for `mex` clusters
		err = CreateNetwork(mf, GetMEXNetwork(mf))
		if err != nil {
			return fmt.Errorf("cannot create mex network %s, %v", GetMEXNetwork(mf), err)
		}
	}

	routers, err := ListRouters(mf)
	if err != nil {
		return err
	}

	found = false
	for _, r := range routers {
		if r.Name == GetMEXExternalRouter(mf) {
			found = true
			break
		}
	}
	if !found {
		// We need at least one router for our `mex` network and external network
		err = CreateRouter(mf, GetMEXExternalRouter(mf))
		if err != nil {
			return fmt.Errorf("cannot create the ext router %s, %v", GetMEXExternalRouter(mf), err)
		}
		err = SetRouter(mf, GetMEXExternalRouter(mf), GetMEXExternalNetwork(mf))
		if err != nil {
			return fmt.Errorf("cannot set default network to router %s, %v", GetMEXExternalRouter(mf), err)
		}
	}

	return nil
}

//GetMEXSubnets returns subnets inside MEX Network
func GetMEXSubnets(mf *Manifest) ([]string, error) {
	nd, err := GetNetworkDetail(mf, GetMEXNetwork(mf))
	if err != nil {
		return nil, fmt.Errorf("can't get MEX network detail, %v", err)
	}

	subnets := strings.Split(nd.Subnets, ",")
	if len(subnets) < 1 {
		return nil, fmt.Errorf("can't get a list of subnets for MEX network")
	}

	return subnets, nil
}

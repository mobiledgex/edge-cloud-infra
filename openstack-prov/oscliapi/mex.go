package oscli

import (
	"fmt"
	"github.com/rs/xid"
	"net"
	"os"
)

// These are platform specific custom vars

var eMEXLargeImageName = os.Getenv("MEX_LARGE_IMAGE") // "mobiledgex-16.04"
var eMEXLargeFlavor = os.Getenv("MEX_LARGE_FLAVOR")   // "m4.large"
var eMEXUserData = os.Getenv("MEX_USERDATA")          // "/home/bob/userdata.txt"

var defaultImage = "mobiledgex-16.04"
var defaultFlavor = "m4.large"
var defaultUserData = "/home/mobiledgex/userdata.txt"

func init() {
	if eMEXLargeImageName == "" {
		eMEXLargeImageName = defaultImage
	}

	if eMEXLargeFlavor == "" {
		eMEXLargeFlavor = defaultFlavor
	}

	if eMEXUserData == "" {
		eMEXUserData = defaultUserData
	}
}

//CreateLargeMEXVM creates basic KVM for mobiledgex applications
//  with proper initial bootstrap scripts installed on the base image that understands
//  various properties such as role, topology of private net, gateway IP, etc.
// Roles can be any string but special ones are k8s-master and k8s-node.
//  To avoid running bootstrap setup for creating kubernets cluster, set skipk8s to true.
// For more detailed information please read `mobiledgex-init.sh`
func CreateLargeMEXVM(name, image, flavor, netID, userdata, role, skipk8s, k8smaster, privatenet, privaterouter string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}
	if image == "" {
		image = eMEXLargeImageName
	}
	if flavor == "" {
		flavor = eMEXLargeFlavor
	}

	if netID == "" {
		netID = GetMEXExternalNetwork() // XXX default to external net?
	}

	if userdata == "" {
		userdata = eMEXUserData
	}

	opts := &ServerOpt{
		Name:     name,
		Image:    image,
		Flavor:   flavor,
		UserData: userdata,
		NetIDs:   []string{netID},
	}

	eg, err := GetExternalGateway(GetMEXExternalNetwork())
	if err != nil {
		return fmt.Errorf("can't get external gateway for %s, %v", GetMEXExternalNetwork(), err)
	}

	props := []string{}

	props = append(props, "edgeproxy="+eg)
	props = append(props, "role="+role)
	props = append(props, "skipk8s="+skipk8s)
	props = append(props, "k8smaster="+k8smaster)
	props = append(props, "privatenet="+privatenet)
	props = append(props, "privaterouter="+privaterouter)

	opts.Properties = props

	err = CreateServer(opts)
	if err != nil {
		return fmt.Errorf("can't create server %v, %v", opts, err)
	}

	return nil
}

//CreateKVM is easier way to create a MEX app capable KVM
func CreateKVM(role string, node int) error {
	skipk8s := "yes"

	if role == "k8s-master" || role == "k8s-node" {
		skipk8s = "no"
		if node == 0 {
			return fmt.Errorf("can't have node 0")
		}
		if role == "k8s-master" {
			node = 1
		} else if node < 2 {
			return fmt.Errorf("node number has to be >= 2")
		}
	}
	guid := xid.New()
	name := "mex-" + guid.String()

	subnetName := "subnet-" + name

	//err := CreateNetwork("net-"+name)

	subnets, err := GetMEXSubnets()
	if err != nil {
		return fmt.Errorf("can't get MEX subnets, %v", err)
	}

	//TODO: iterate through subnets and make sure to pick the
	//  first subnet marked by a name or range

	netRange, err := GetNextSubnetRange(subnets[0])
	if err != nil {
		return fmt.Errorf("can't get next subnet range on %s, %v", subnets[0], err)
	}
	_, ipv4Net, err := net.ParseCIDR(netRange)
	if err != nil {
		return fmt.Errorf("can't parse %s, %v", netRange, err)
	}
	v4 := ipv4Net.IP.To4()

	externalNet := GetMEXExternalNetwork()
	gateway, err := GetExternalGateway(externalNet)
	if err != nil {
		return fmt.Errorf("can't get external gateway for %s, %v", externalNet, err)
	}

	err = CreateSubnet(netRange, "net-"+name, gateway, "subnet-"+name, false)
	if err != nil {
		return err
	}
	//TODO: consider adding tags to subnet

	// XXX if there was nothing set up on the cloudlet, we may have to ask the operator
	//  the initial external network which has connection to internet.
	//  And add at least one network for MEX use.
	//  And add at least one subnet inside that network.
	//  And add a router for MEX use.
	//  And set that router to external network created by the operator.
	//  And add that router to the subnet.
	// However, we need to normally operate with these things already setup.
	//  So the code here assumes this.

	//err = CreateRouter("router-"+name)
	//err = SetRouter("router-"+name, "net-"+name)

	routerName := GetMEXExternalRouter()

	// We assume router is set to external network. So just attach a new port / subnet to this router
	err = AddRouterSubnet(routerName, subnetName)

	privateNet := netRange

	rd, err := GetRouterDetail(routerName)
	if err != nil {
		return fmt.Errorf("can't get router detail for %s, %v", routerName, err)
	}

	reg, err := GetRouterDetailExternalGateway(rd)
	if err != nil {
		return fmt.Errorf("can't get router detail external gateway, %v", err)
	}
	if len(reg.ExternalFixedIPs) < 1 {
		return fmt.Errorf("can't get external fixed ips list from router detail external gateway")
	}
	fip := reg.ExternalFixedIPs[0]

	privateRouter := fip.IPAddress // router IP for the private network to the external side, which also knows about the private side

	netID := GetMEXExternalNetwork() //do we really want to default to ext?
	masterIP := "unknown-master-IP"
	if skipk8s == "no" {
		if v4[3] != 0 {
			panic("bad v4[3]")
		}
		//last octet should always be zero
		newIP := net.IPv4(v4[0], v4[1], v4[2], byte(node+1))
		//+1 because gatway is at .1
		//master node num is 1
		//so, master node will always have .2
		netID = GetMEXNetwork() + ",v4-fixed-ip=" + newIP.String()
		newIP = net.IPv4(v4[0], v4[1], v4[2], 2)
		masterIP = newIP.String()
	}
	err = CreateLargeMEXVM(name,
		eMEXLargeImageName,
		eMEXLargeFlavor,
		netID,
		eMEXUserData,
		role,
		skipk8s,
		masterIP,
		privateNet,
		privateRouter,
	)

	if err != nil {
		return err
	}

	return nil
}

package oscli

import (
	"fmt"
	"github.com/bobbae/q"
	"github.com/rs/xid"
	"net"
	"os"
)

// These are platform specific custom vars

var eMEXLargeImageName = os.Getenv("MEX_LARGE_IMAGE") // "mobiledgex-16.04"
var eMEXLargeFlavor = os.Getenv("MEX_LARGE_FLAVOR")   // "m4.large"
var eMEXUserData = os.Getenv("MEX_USERDATA")          // "/home/bob/userdata.txt"

var defaultImage = "mobiledgex-16.04-1"
var defaultFlavor = "m4.large"

func init() {
	if eMEXLargeImageName == "" {
		eMEXLargeImageName = defaultImage
	}

	if eMEXLargeFlavor == "" {
		eMEXLargeFlavor = defaultFlavor
	}

	if eMEXUserData == "" {
		hm := os.Getenv("HOME")
		eMEXUserData = hm + "/userdata.txt"
	}
}

//CreateLargeMEXVM creates basic KVM for mobiledgex applications
//  with proper initial bootstrap scripts installed on the base image that understands
//  various properties such as role, topology of private net, gateway IP, etc.
// Roles can be any string but special ones are k8s-master and k8s-node.
//  To avoid running bootstrap setup for creating kubernets cluster, set skipk8s to true.
// For more detailed information please read `mobiledgex-init.sh`
func CreateLargeMEXVM(name, image, flavor, netID, userdata, role, edgeproxy, skipk8s, k8smaster, privatenet, privaterouter string) error {
	if name == "" {
		return fmt.Errorf("name required")
	}

	if netID == "" {
		return fmt.Errorf("net-id required")
	}

	if image == "" {
		image = eMEXLargeImageName
	}
	if flavor == "" {
		flavor = eMEXLargeFlavor
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

	props := []string{}

	//edgeproxy should be pointing to external gateway IP when running a agent-proxy node.
	//  agent proxy node has direct connection to external network. The gateway of that
	//  network is edgeproxy setting.
	//edgeproxy should be pointing to internal gateway IP when running on private network.
	//  Typically like 10.101.101.1

	props = append(props, "edgeproxy="+edgeproxy)
	props = append(props, "role="+role)
	props = append(props, "skipk8s="+skipk8s)
	props = append(props, "k8smaster="+k8smaster)

	//privatenet, privaterouter are used when in agent-proxy mode. It deals with external
	//  and internet network.  Normal k8s nodes do not look at these.
	//privaterouter should be pointing to the router instance's external network address
	//  which is reachable from internal network.
	props = append(props, "privatenet="+privatenet)
	props = append(props, "privaterouter="+privaterouter)

	opts.Properties = props

	err := CreateServer(opts)
	if err != nil {
		return fmt.Errorf("can't create server %v, %v", opts, err)
	}

	return nil
}

//CreateKVM is easier way to create a MEX app capable KVM
//  role can be k8s-master, k8s-node, or something else
//  node can be 1 for k8s-master, >1 for k8s-nodes; if not using k8s, it can be >0
func CreateKVM(role string, node int) error {
	guid := xid.New()
	name := "mex-" + guid.String()

	skipk8s := "yes"

	mexRouter := GetMEXExternalRouter()

	subnets, err := GetMEXSubnets()
	if err != nil {
		return fmt.Errorf("can't get MEX subnets, %v", err)
	}

	sd, err := GetSubnetDetail(subnets[0])
	if err != nil {
		return fmt.Errorf("cannot get details for subnet %s, %v", subnets[0], err)
	}
	netID := GetMEXExternalNetwork() //do we really want to default to ext?
	masterIP := "unknown-master-IP"
	pRouterIP := "unknown-prouter-ip"
	edgeProxy := "unknown-edge-proxy"

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
		// XXX if there was nothing set up on the cloudlet, we may have to ask the operator
		//  the initial external network which has connection to internet.
		//  And add at least one network for MEX use.
		//  And add at least one subnet inside that network.
		//  And add a router for MEX use.
		//  And set that router to external network created by the operator.
		//  And add that router to the subnet.
		// However, we need to normally operate with these things already setup.
		//  So the code here assumes this.

		// XXX do not create a new network, use existing MEXNet
		//nn := "net-" + name
		//err := CreateNetwork(nn)
		//if err != nil {
		//	return fmt.Errorf("can't create network %s, %v", nn, err)
		//}

		//err = CreateRouter("router-"+name)
		//err = SetRouter("router-"+name, "net-"+name)

		// We assume router is set to external network. So just attach a new port / subnet to this router

		//TODO: iterate through `subnets` and make sure to pick the
		//  first subnet marked by a name or range

		nextCIDR, err := GetNextSubnetRange(subnets[0])
		if err != nil {
			return fmt.Errorf("can't get next subnet range on %s, %v", subnets[0], err)
		}

		_, ipv4Net, err := net.ParseCIDR(nextCIDR)
		if err != nil {
			return fmt.Errorf("can't parse %s, %v", nextCIDR, err)
		}
		v4 := ipv4Net.IP.To4()

		//last octet should always be zero
		if v4[3] != 0 {
			panic("bad v4[3]")
		}

		//X.X.X.1
		gatewayIP := net.IPv4(v4[0], v4[1], v4[2], byte(1))

		sn := "subnet-" + name
		edgeProxy = gatewayIP.String()

		if role == "k8s-master" {
			err = CreateSubnet(nextCIDR, GetMEXNetwork(), edgeProxy, sn, false)
			if err != nil {
				return err
			}

			//TODO: consider adding tags to subnet

			err = AddRouterSubnet(mexRouter, sn)
			if err != nil {
				return fmt.Errorf("cannot add router %s to subnet %s, %v", mexRouter, sn, err)
			}
		}

		newIP := net.IPv4(v4[0], v4[1], v4[2], byte(node+1))
		//+1 because gatway is at .1
		//master node num is 1
		//so, master node will always have .2
		netID = GetMEXNetwork() + ",v4-fixed-ip=" + newIP.String()
		q.Q(netID)
		newIP = net.IPv4(v4[0], v4[1], v4[2], 2)
		masterIP = newIP.String()
	} else {
		edgeProxy, err = GetExternalGateway(netID)
		if err != nil {
			return fmt.Errorf("can't get external gateway for %s, %v", netID, err)
		}

		rd, err := GetRouterDetail(mexRouter)
		if err != nil {
			return fmt.Errorf("can't get router detail for %s, %v", mexRouter, err)
		}

		reg, err := GetRouterDetailExternalGateway(rd)
		if err != nil {
			return fmt.Errorf("can't get router detail external gateway, %v", err)
		}
		if len(reg.ExternalFixedIPs) < 1 {
			return fmt.Errorf("can't get external fixed ips list from router detail external gateway")
		}
		fip := reg.ExternalFixedIPs[0]

		// router IP for the private network to the external side, which
		//  also knows about the private side. Only needed for agent gw node.
		pRouterIP = fip.IPAddress
	}

	err = CreateLargeMEXVM(name,
		eMEXLargeImageName,
		eMEXLargeFlavor,
		netID, // either external-net or internal-net,v4-fixed-ip=X.X.X.X
		eMEXUserData,
		role, // k8s-master,k8s-node,something else
		edgeProxy,
		skipk8s,  // if yes, skip
		masterIP, // relevant when forming k8s cluster
		sd.CIDR,  // first priv subnet CIDR. XXX need agent API to add more subnets
		pRouterIP,
	)

	if err != nil {
		return err
	}

	return nil
}

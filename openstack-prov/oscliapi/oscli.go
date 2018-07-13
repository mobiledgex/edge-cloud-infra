package oscli

import (
	"encoding/json"
	"fmt"
	"github.com/codeskyblue/go-sh"
)

// There are issues with x509 certfication and token retrieval when using ../api
// with certain Openstack cloudlets. The issues are handled correctly by
// openstack CLI python code. But not in the gophercloud library.
// So the existence of this code is for: (1) avoid all the issues that are
// time-wasting and go directly to something that always works,
// (2) show clear examples, as a tutorial, to those who are not familiar
// with openstack.

// Limit name,value pairs of Openstack tenant level limits. The only platform
// level stats available to us reliably at some cloudlets.
type Limit struct {
	Name  string
	Value int
}

// Server is output of 'openstack server list'.  In Openstack, 'server' means
//  an instance of KVM.
type Server struct {
	Status, Name, Image, ID, Flavor, Networks string
}

// ServerOpt is used to specify options when creating servers.
type ServerOpt struct {
	AvailabilityZone    string //XXX not used yet
	Name, Image, Flavor string
	UserData            string
	NetIDs              []string
	Properties          []string
}

// ServerDetail  is used with output of 'openstack server show' to list
//   gory details per server.
type ServerDetail struct {
	TaskState        string `json:"OS-EXT-STS:task_state"`
	Addresses        string `json:"addresses"`
	Image            string `json:"image"`
	VMState          string `json:"OS-EXT-STS:vm_state"`
	LaunchedAt       string `json:"OS-SRV-USG:launched_at"`
	Flavor           string `json:"flavor"`
	ID               string `json:"id"`
	SecurityGroups   string `json:"security_groups"`
	VolumesAttached  string `json:"volumes_attached"`
	UserID           string `json:"user_id"`
	DiskConfig       string `json:"OS-DCF:diskConfig"`
	AccessIPv4       string `json:"accessIPv4"`
	AccessIPv6       string `json:"accessIPv6"`
	Progress         int    `json:"progress"`
	PowerState       string `json:"OS-EXT-STS:power_state"`
	ProjectID        string `json:"project_id"`
	ConfigDrive      string `json:"config_drive"`
	Status           string `json:"status"`
	Updated          string `json:"updated"`
	HostID           string `json:"hostId"`
	TerminatedAt     string `json:"OS-SRV-USG:terminated_at"`
	KeyName          string `json:"key_name"`
	AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`
	Name             string `json:"name"`
	Created          string `json:"created"`
	Properties       string `json:"properties"`
}

// Image is used with 'openstack image list'
type Image struct {
	Status, ID, Name string
}

// Network is used with 'openstack network list'
//   A network in openstack is created as needed. Some are
//   created by the provider and has extra features. Such
//   as external connectivity.
type Network struct {
	Subnets, ID, Name string
}

// Flavor is used with 'openstack flavor list'
type Flavor struct {
	Name, ID                    string
	RAM, Ephemeral, VCPUs, Disk int
}

// Subnet is used with 'openstack subnet list'
//  In openstack, each 'network' contains 'subnets'.
//  Networks have higher abstractions. The subnets
//  actually contain network ranges, etc.
type Subnet struct {
	Name, ID, Network, Subnet string
}

//Router is used with 'openstack router' commands.
// In openstack  router is virtually created to connect different subnets
//  and possible to external networks.
type Router struct {
	Name, ID, Status, State, HA, Project, Distributed string
}

// "Is Public" field  is missing in 'Router'

//GetLimits is used to retrieve tenant level platform stats
func GetLimits() ([]Limit, error) {
	//err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).WriteStdout("os-out.txt")
	out, err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).Output()
	if err != nil {
		err = fmt.Errorf("cannot get limits from openstack, %v", err)
		return nil, err
	}

	var limits []Limit
	err = json.Unmarshal(out, &limits)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	return limits, nil
}

//ListServers returns list of servers, KVM instances, running on the system
func ListServers() ([]Server, error) {
	out, err := sh.Command("openstack", "server", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("cannot get server list, %v", err)
		return nil, err
	}

	var servers []Server
	err = json.Unmarshal(out, &servers)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	return servers, nil
}

//ListImages lists avilable images in glance
func ListImages() ([]Image, error) {
	out, err := sh.Command("openstack", "image", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("cannot get image list, %v", err)
		return nil, err
	}

	var images []Image
	err = json.Unmarshal(out, &images)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	return images, nil
}

//ListNetworks lists networks known to the platform. Some created by the operator, some by users.
func ListNetworks() ([]Network, error) {
	out, err := sh.Command("openstack", "network", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("cannot get network list, %v", err)
		return nil, err
	}

	var networks []Network
	err = json.Unmarshal(out, &networks)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	return networks, nil
}

//ListFlavors lists flavors known to the platform. These vary. On Bonn cloudlet the are m4. prefixed.
func ListFlavors() ([]Flavor, error) {
	out, err := sh.Command("openstack", "flavor", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("cannot get flavor list, %v", err)
		return nil, err
	}

	var flavors []Flavor
	err = json.Unmarshal(out, &flavors)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	return flavors, nil
}

//CreateServer instantiates a new server instance, which is a KVM instance based on a qcow2 image from glance
func CreateServer(opts *ServerOpt) error {
	args := []string{
		"server", "create",
		"--image", opts.Image, "--flavor", opts.Flavor,
		"--user-data", opts.UserData,
	}

	for _, p := range opts.Properties {
		args = append(args, "--property", p)
		// `p` should be like: "key=value"
	}

	for _, n := range opts.NetIDs {
		args = append(args, "--nic", "net-id="+n)
		// `n` should be like: "public,v4-fixed-ip=172.24.4.201"
	}

	args = append(args, opts.Name)

	//TODO additional args

	iargs := make([]interface{}, len(args))
	for i, v := range args {
		iargs[i] = v
	}

	out, err := sh.Command("openstack", iargs...).Output()
	if err != nil {
		err = fmt.Errorf("cannot create server, %v, '%s'", err, out)
		return err
	}

	return nil
}

// GetServerDetails returns details of the KVM instance
func GetServerDetails(name string) (*ServerDetail, error) {
	out, err := sh.Command("openstack", "server", "show", "-f", "json", name).Output()
	if err != nil {
		err = fmt.Errorf("can't show server %s, %s, %v", name, out, err)
		return nil, err
	}

	srvDetail := &ServerDetail{}

	//fmt.Printf("%s\n", out)
	err = json.Unmarshal(out, srvDetail)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}

	return srvDetail, nil
}

//DeleteServer destroys a KVM instance
//  sometimes it is not possible to destroy. Like most things in Openstack, try again.
func DeleteServer(id string) error {
	out, err := sh.Command("openstack", "server", "delete", id).Output()
	if err != nil {
		err = fmt.Errorf("can't delete server %s, %s, %v", id, out, err)
		return err
	}
	return nil
}

// CreateNetwork creates a network with a name.
func CreateNetwork(name string) error {
	out, err := sh.Command("openstack", "network", "create", name).Output()
	if err != nil {
		err = fmt.Errorf("can't create network %s, %s, %v", name, out, err)
		return err
	}
	return nil
}

//DeleteNetwork destroys a named network
//  Sometimes it will fail. Openstack will refuse if there are resources attached.
func DeleteNetwork(name string) error {
	out, err := sh.Command("openstack", "network", "delete", name).Output()
	if err != nil {
		err = fmt.Errorf("can't delete network %s, %s, %v", name, out, err)
		return err
	}
	return nil
}

//CreateSubnet creates a subnet within a network. A subnet is assigned ranges. Optionally DHCP can be enabled.
func CreateSubnet(netRange, networkName, gatewayAddr, subnetName string, dhcpEnable bool) error {
	dhcpFlag := ""
	if dhcpEnable == true {
		dhcpFlag = "--dhcp"
	} else {
		dhcpFlag = "--no-dhcp"
	}

	out, err := sh.Command("openstack", "subnet", "create",
		"--subnet-range", netRange, // e.g. 10.101.101.0/24
		"--network", networkName, // mex-k8s-net-1
		dhcpFlag,
		"--gateway", gatewayAddr, // e.g. 10.101.101.1
		subnetName).Output() // e.g. mex-k8s-subnet-1
	if err != nil {
		err = fmt.Errorf("can't create subnet %s, %s, %v", subnetName, out, err)
		return err
	}

	return nil
}

//DeleteSubnet deletes the subnet. If this fails, remove any attached resources, like router, and try again.
func DeleteSubnet(subnetName string) error {
	out, err := sh.Command("openstack", "subnet", "delete", subnetName).Output()
	if err != nil {
		err = fmt.Errorf("can't delete subnet %s, %s, %v", subnetName, out, err)
		return err
	}
	return nil
}

//CreateRouter creates new router. A router can be attached to network and subnets.
func CreateRouter(routerName string) error {
	out, err := sh.Command("openstack", "router", "create", routerName).Output()
	if err != nil {
		err = fmt.Errorf("can't create router %s, %s, %v", routerName, out, err)
		return err
	}
	return nil
}

//DeleteRouter removes the named router. The router needs to not be in use at the time of deletion.
func DeleteRouter(routerName string) error {
	out, err := sh.Command("openstack", "router", "delete", routerName).Output()
	if err != nil {
		err = fmt.Errorf("can't delete router %s, %s, %v", routerName, out, err)
		return err
	}

	return nil
}

//SetRouter assigns the router to a particular network. The network needs to be attached to
// a real external network. This is intended only for routing to external network for now. No internal routers.
// Sometimes, oftentimes, it will fail if the network is not external.
func SetRouter(routerName, networkName string) error {
	out, err := sh.Command("openstack", "router", "set", routerName, "--external-gateway", networkName).Output()
	if err != nil {
		err = fmt.Errorf("can't set router %s to %s, %s, %v", routerName, networkName, out, err)
		return err
	}
	return nil
}

//AddRouterSubnet will connect subnet to another network, possibly external, via a router
func AddRouterSubnet(routerName, subnetName string) error {
	out, err := sh.Command("openstack", "router", "add", "subnet", routerName, subnetName).Output()
	if err != nil {
		err = fmt.Errorf("can't add router %s to subnet %s, %s, %v", routerName, subnetName, out, err)
		return err
	}
	return nil
}

//RemoveRouterSubnet is useful to remove the router from the subnet before deletion. Otherwise subnet cannot
//  be deleted.
func RemoveRouterSubnet(routerName, subnetName string) error {
	out, err := sh.Command("openstack", "router", "remove", "subnet", routerName, subnetName).Output()
	if err != nil {
		err = fmt.Errorf("can't remove router %s from subnet %s, %s, %v", routerName, subnetName, out, err)
		return err
	}
	return nil
}

//ListSubnets returns a list of subnets available
func ListSubnets() ([]Subnet, error) {
	out, err := sh.Command("openstack", "subnet", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("can't get a list of subnets, %v", err)
		return nil, err
	}

	subnets := []Subnet{}
	err = json.Unmarshal(out, &subnets)
	if err != nil {
		err = fmt.Errorf("can't unmarshal subnets, %v", err)
		return nil, err
	}
	return subnets, nil

}

//ListRouters returns a list of routers available
func ListRouters() ([]Router, error) {
	out, err := sh.Command("openstack", "router", "list", "-f", "json").Output()
	if err != nil {
		err = fmt.Errorf("can't get a list of routers, %s, %v", out, err)
		return nil, err
	}

	routers := []Router{}

	err = json.Unmarshal(out, &routers)
	if err != nil {
		err = fmt.Errorf("can't unmarshal routers, %v", err)
		return nil, err
	}

	return routers, nil
}

//CreateImage snapshots running service into a qcow2 image
func CreateImage(serverName, imageName string) error {
	out, err := sh.Command("openstack", "image", "create", serverName, "--name", imageName).Output()
	if err != nil {
		err = fmt.Errorf("can't create image from %s into %s, %s, %v", serverName, imageName, out, err)
		return err
	}

	return nil
}

//SaveImage takes the image name available in glance, as a result of for example the above create image.
// It will then save that into a local file. The image transfer happens from glance into your own laptop
// or whatever.
// This can take a while, transferring all the data.
func SaveImage(saveName, imageName string) error {
	out, err := sh.Command("openstack", "image", "save", "--file", saveName, imageName).Output()
	if err != nil {
		err = fmt.Errorf("can't save image from %s to file %s, %s, %v", imageName, saveName, out, err)
		return err
	}
	return nil
}

//DeleteImage deletes the named image from glance. Sometimes backing store is still busy and
// will refuse to honor the request. Like most things in Openstack, wait for a while and try
// again.
func DeleteImage(imageName string) error {
	out, err := sh.Command("openstack", "image", "delete", imageName).Output()
	if err != nil {
		err = fmt.Errorf("can't delete image %s, %s, %v", imageName, out, err)
		return err
	}
	return nil
}

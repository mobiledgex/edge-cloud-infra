package oscli

import (
	"fmt"
	"testing"

	log "github.com/bobbae/logrus"
)

// The order of tests are important here. Also 'source openrc' or the equivalent to set up
// the OS_XXX env variables before running this.
// There is no mock for this for now.
// For CICD, either create a real live test cluster for that use, or
// write a bunch of fake mock returns -- will not do this here.
// During development, each of the tests were run individually against a live cluster.

var testServerName = "go-test1"
var testImageName = "mobiledgex-16.04"
var testFlavor = "m4.small"
var testUserData = "userdata.txt"
var testNetwork = "test-external-network-shared"
var testSubnet = "test-internal-network"
var testRange = "10.102.102.0/24"
var testGateway = "10.102.102.1"
var testRouter = "test-router-1"

func TestInit(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestGetLimits(t *testing.T) {
	out, err := GetLimits()
	if err != nil {
		t.Errorf("cannot GetLimits, %v", err)
	} else {
		fmt.Println(out)
	}
}

func TestListServers(t *testing.T) {
	out, err := ListServers()
	if err != nil {
		t.Errorf("cannot ListServers, %v", err)
	} else {
		fmt.Println(out)
	}
}

func TestListImages(t *testing.T) {
	out, err := ListImages()
	if err != nil {
		t.Errorf("cannot ListImages, %v", err)
	} else {
		fmt.Println(out)
	}
}

func TestListNetworks(t *testing.T) {
	out, err := ListNetworks()
	if err != nil {
		t.Errorf("cannot ListNetworks, %v", err)
	} else {
		fmt.Println(out)
	}
}

func TestListFlavors(t *testing.T) {
	out, err := ListFlavors()
	if err != nil {
		t.Errorf("cannot ListFlavors, %v", err)
	} else {
		fmt.Println(out)
	}
}

func TestCreateServer(t *testing.T) {

	opts := &ServerOpt{
		Name:       testServerName,
		Image:      testImageName,
		Flavor:     testFlavor,
		UserData:   testUserData,
		NetIDs:     []string{testNetwork},
		Properties: []string{},
	}
	err := CreateServer(opts)
	if err != nil {
		t.Errorf("%v", err)
	} else {
		fmt.Println("created", opts)
	}
}

func TestGetServerDetails(t *testing.T) {
	sd, err := GetServerDetails(testServerName)
	if err != nil {
		t.Errorf("server show err, %v", err)
		return
	}
	fmt.Println("server", sd)
}

func TestDeleteServer(t *testing.T) {
	sd, err := GetServerDetails(testServerName)
	if err != nil {
		t.Errorf("cannot show server %s, %v", testServerName, err)
	}

	if sd.Name != testServerName { //Xxx never happen
		t.Errorf("name mismatch")
	}

	log.Debugln("delete server %s %s", sd.Name, sd.ID)

	err = DeleteServer(sd.ID)
	if err != nil {
		t.Errorf("cannot delete server %s %s, %v", sd.Name, sd.ID, err)
	} else {
		fmt.Println("server", sd.Name, "deleted ok")
	}
}

func TestCreateNetwork(t *testing.T) {
	err := CreateNetwork(testNetwork)
	if err != nil {
		t.Errorf("cannot create network, %v", err)
		return
	}
	fmt.Println("network", testNetwork, "created")
}

func TestCreateSubnet(t *testing.T) {
	err := CreateSubnet(testRange, testNetwork, testGateway, testSubnet, false)
	if err != nil {
		t.Errorf("cannot create subnet, %v", err)
		return
	}
	fmt.Println("created subnet ", testSubnet)
}

func TestCreateRouter(t *testing.T) {
	err := CreateRouter(testRouter)
	if err != nil {
		t.Errorf("can't create router, %v", err)
		return
	}
	fmt.Println("created router ", testRouter)
}

func TestSetRouter(t *testing.T) {
	err := SetRouter(testRouter, testNetwork)
	if err != nil {
		fmt.Printf("can't set router, %v", err)
		fmt.Printf("not an error.")
		//because testNetwork is not a real external network
		return
	}
	fmt.Printf("set router %s in net %s\n", testRouter, testNetwork)
}

func TestAddRouterSubnet(t *testing.T) {
	err := AddRouterSubnet(testRouter, testSubnet)
	if err != nil {
		t.Errorf("can't add router to subnet, %v", err)
		return
	}
	fmt.Printf("added router %s to subnet %s\n", testRouter, testSubnet)
}

func TestListSubnets(t *testing.T) {
	subnets, err := ListSubnets()
	if err != nil {
		t.Errorf("can't list subnets, %v", err)
		return
	}
	fmt.Printf("subnets %v\n", subnets)
}

func TestListRouters(t *testing.T) {
	routers, err := ListRouters()
	if err != nil {
		t.Errorf("can't list routers, %v", err)
		return
	}
	fmt.Printf("routers %v\n", routers)
}

// The order of the following sequence of tests are particularly important

func TestRemoveRouterSubnet(t *testing.T) {
	err := RemoveRouterSubnet(testRouter, testSubnet)
	if err != nil {
		t.Errorf("can't remove router from subnet, %v", err)
		return
	}
	fmt.Printf("removed router %s from subnet %s\n", testRouter, testSubnet)
}

func TestDeleteRouter(t *testing.T) {
	err := DeleteRouter(testRouter)
	if err != nil {
		t.Errorf("can't delete router, %v", err)
		return
	}
	fmt.Println("deleted router ", testRouter)
}
func TestDeleteSubnet(t *testing.T) {
	err := DeleteSubnet(testSubnet)
	if err != nil {
		t.Errorf("cannot delete subnet , %v", err)
		return
	}
	fmt.Println("deleted subnet s", testSubnet)
}

func TestDeleteNetwork(t *testing.T) {
	err := DeleteNetwork(testNetwork)
	if err != nil {
		t.Errorf("cannot delete network, %v", err)
		return
	}
	fmt.Println("deleted network ", testNetwork)
}

// The orders are not preserved here. The server should be running.
// But it won't be when we run tests serially.

var testImage = "test-image-1"

func TestCreateImage(t *testing.T) {
	err := CreateImage(testServerName, testImage)
	if err != nil {
		t.Errorf("cannot create image , %v", err)
		return
	}
	fmt.Println("created image", testImage)
}

var saveImageFile = "test-save-image.qcow2" // will be created locally. Potentially very large.

func TestSaveImage(t *testing.T) {
	//This can take a while
	err := SaveImage(saveImageFile, testImage)
	if err != nil {
		t.Errorf("cannot save image , %v", err)
		return
	}
	fmt.Println("saved image", testImage)
}

func TestDeleteImage(t *testing.T) {
	err := DeleteImage(testImage)
	if err != nil {
		t.Errorf("cannot delete image , %v", err)
		return
	}
	fmt.Println("deleted image", testImage)
}

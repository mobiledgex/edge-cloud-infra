package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

const VpcDoesNotExistError string = "vpc does not exist"
const SubnetDoesNotExistError string = "subnet does not exist"
const SecGrpDoesNotExistError string = "security group does not exist"
const GatewayDoesNotExistError string = "gateway does not exist"
const ResourceAlreadyAssociatedError string = "Resource.AlreadyAssociated"
const GroupAlreadyExistsError string = "InvalidGroup.Duplicate"
const RuleAlreadyExistsError string = "InvalidPermission.Duplicate"
const RouteTableDoesNotExistError string = "route table does not exist"
const ElasticIpDoesNotExistError string = "elastic ip does not exist"
const ImageDoesNotExistError string = "image does not exist"
const ImageNotAvailableError string = "image is not available"

var orchVmLock sync.Mutex

type RouteTableSearchType string

const SearchForMainRouteTable RouteTableSearchType = "main"
const SearchForRouteTableByName RouteTableSearchType = "name"

// when MainRouteTable is used, the route table is not specified and defaults to the main RT
const MainRouteTable string = "mainRouteTable"

const ArnAccountIdIdx = 4

type AwsIamUser struct {
	UserId string
	Arn    string
}

type AwsIamUserResult struct {
	User AwsIamUser
}

type AwsEc2Tag struct {
	Key   string
	Value string
}

type AwsEc2SecGrp struct {
	GroupName string
	GroupId   string
	VpcId     string
}

type AwsEc2SecGrpList struct {
	SecurityGroups []AwsEc2SecGrp
}

type AwsEc2RouteTableAssociation struct {
	Main                    bool
	RouteTableAssociationId string
}
type AwsEc2RouteTable struct {
	RouteTableId string
	VpcId        string
	Associations []AwsEc2RouteTableAssociation
	Tags         []AwsEc2Tag
}
type AwsEc2RouteTableList struct {
	RouteTables []AwsEc2RouteTable
}

type AwsEc2RouteTableCreateResult struct {
	RouteTable AwsEc2RouteTable
}

type AwsEc2Gateway struct {
	InternetGatewayId string
	Tags              []AwsEc2Tag
}

type AwsEc2GatewayList struct {
	InternetGateways []AwsEc2Gateway
}

type AwsEc2NatGateway struct {
	NatGatewayId string
	State        string
	VpcId        string
	SubnetId     string
	Tags         []AwsEc2Tag
}

type AwsEc2NatGatewayList struct {
	NatGateways []AwsEc2NatGateway
}

type AwsEc2NatGatewayCreateResult struct {
	NatGateway AwsEc2NatGateway
}

type AwsEc2Address struct {
	PublicIp           string
	AllocationId       string
	Domain             string
	NetworkBorderGroup string
	PublicIpv4Pool     string
}
type AwsEc2AddressList struct {
	Addresses []AwsEc2Address
}

type AwsEc2Subnet struct {
	CidrBlock string
	State     string
	SubnetId  string
	VpcId     string
	Name      string
	Tags      []AwsEc2Tag
}

type AwsEc2SubnetList struct {
	Subnets []AwsEc2Subnet
}

type AwsEc2Vpc struct {
	CidrBlock string
	VpcId     string
	Tags      []AwsEc2Tag
}

type AwsEc2VpcCreateResult struct {
	Vpc AwsEc2Vpc
}

type AwsEc2SubnetCreateResult struct {
	Subnet AwsEc2Subnet
}

type AwsEc2VpcList struct {
	Vpcs []AwsEc2Vpc
}

type AwsEc2State struct {
	Code int
	Name string
}

type AwsEc2NetworkInterfaceCreateSpec struct {
	AssociatePublicIpAddress bool     `json:"AssociatePublicIpAddress,omitempty"`
	SubnetId                 string   `json:"SubnetId,omitempty"`
	PrivateIpAddress         string   `json:"PrivateIpAddress,omitempty"`
	Groups                   []string `json:"Groups,omitempty"`
	DeviceIndex              int      `json:"DeviceIndex"`
}

type AwsEc2IpAddrPublicIpAssociation struct {
	PublicIp string
}
type AwsEc2IpAddress struct {
	PrivateIpAddress string
	Association      AwsEc2IpAddrPublicIpAssociation
}
type AwsEc2NetworkInterface struct {
	VpcId              string
	SubnetId           string
	MacAddress         string
	NetworkInterfaceId string
	PrivateIpAddresses []AwsEc2IpAddress
}
type AwsEc2NetworkInterfaceCreateResult struct {
	NetworkInterface AwsEc2NetworkInterface
}

type AwsEc2Ebs struct {
	DeleteOnTermination bool
	Status              string `json:"Status,omitempty"`
}
type AwsEc2BlockDeviceMapping struct {
	DeviceName string
	Ebs        AwsEc2Ebs
}

type AwsEc2Image struct {
	ImageId string
	OwnerId string
	State   string
	Name    string
}
type AwsEc2ImageList struct {
	Images []AwsEc2Image
}

type AwsEc2Instance struct {
	ImageId           string
	InstanceId        string
	NetworkInterfaces []AwsEc2NetworkInterface
	Tags              []AwsEc2Tag
	State             AwsEc2State
}

type AwsEc2Reservation struct {
	Instances []AwsEc2Instance
}

type AwsEc2Instances struct {
	Reservations []AwsEc2Reservation
}

type VmGroupResources struct {
	vpcId         string
	secGrpMap     map[string]*AwsEc2SecGrp
	subnetMap     map[string]*AwsEc2Subnet
	imageNameToId map[string]string
}

func (a *AWSPlatform) GetServerDetail(ctx context.Context, vmname string) (*vmlayer.ServerDetail, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetServerDetail", "vmname", vmname)

	var sd vmlayer.ServerDetail
	var ec2insts AwsEc2Instances
	snMap, err := a.GetSubnets(ctx)
	if err != nil {
		return nil, err
	}
	out, err := a.TimedAwsCommand(ctx,
		"aws", "ec2",
		"describe-instances",
		"--region", a.GetAwsRegion(),
		"--filters", fmt.Sprintf("Name=tag-value,Values=%s", vmname))
	if err != nil {
		return nil, fmt.Errorf("Error in describe-instances: %v", err)
	}
	err = json.Unmarshal(out, &ec2insts)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-instances unmarshal fail", "vmname", vmname, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	for _, res := range ec2insts.Reservations {
		for _, inst := range res.Instances {
			log.SpanLog(ctx, log.DebugLevelInfra, "found server", "vmname", vmname, "state", inst.State)

			switch inst.State.Name {
			case "terminated":
				// ec2 stay visible in terminated state for a while but they do not really exist
				continue
			case "running":
				sd.Status = vmlayer.ServerActive
			case "stopped":
				fallthrough
			case "stopping":
				fallthrough
			case "pending":
				fallthrough
			case "shutting-down":
				sd.Status = vmlayer.ServerShutoff
			default:
				return nil, fmt.Errorf("unexpected server state: %s server: %s", inst.State.Name, vmname)
			}
			sd.Name = vmname
			sd.ID = inst.InstanceId
			for _, netif := range inst.NetworkInterfaces {
				var sip vmlayer.ServerIP

				log.SpanLog(ctx, log.DebugLevelInfra, "found network interface", "vmname", vmname, "netif", netif)

				if len(netif.PrivateIpAddresses) != 1 {
					log.SpanLog(ctx, log.DebugLevelInfra, "unexpected number of private ips", "netif.PrivateIpAddresses", netif.PrivateIpAddresses)
					continue
				}
				sip.InternalAddr = netif.PrivateIpAddresses[0].PrivateIpAddress
				if netif.PrivateIpAddresses[0].Association.PublicIp != "" {
					sip.ExternalAddr = netif.PrivateIpAddresses[0].Association.PublicIp
					sip.ExternalAddrIsFloating = true
				} else {
					sip.ExternalAddr = sip.InternalAddr
				}
				for sname, sn := range snMap {
					if sn.SubnetId == netif.SubnetId {
						sip.Network = sname
						break
					}
				}
				if sip.Network == "" {
					log.SpanLog(ctx, log.DebugLevelInfra, "Could not find subnet for network interface", "netif.SubnetId", netif.SubnetId, "snMap", snMap)
					return nil, fmt.Errorf("Could not find subnet for network interface subnetid: %s", netif.SubnetId)
				}
				sip.PortName = vmlayer.GetPortName(vmname, sip.Network)
				sip.MacAddress = netif.MacAddress
				sd.Addresses = append(sd.Addresses, sip)

			}
			log.SpanLog(ctx, log.DebugLevelInfra, "active server", "vmname", vmname, "state", inst.State, "sd", sd)
			return &sd, nil
		}
	}
	return &sd, fmt.Errorf(vmlayer.ServerDoesNotExistError)
}

func (a *AWSPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer", "serverName", serverName, "subnetName", subnetName, "portName", portName, "ipaddr", ipaddr)

	sn, err := a.GetSubnet(ctx, subnetName)
	if err != nil {
		return err
	}
	sd, err := a.GetServerDetail(ctx, serverName)
	if err != nil {
		return err
	}

	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return err
	}

	secGrpName := vmlayer.GetServerSecurityGroupName(serverName)
	sgrp, err := a.GetSecurityGroup(ctx, secGrpName, vpc.VpcId)
	if err != nil {
		return err
	}
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-network-interface",
		"--subnet-id", sn.SubnetId,
		"--description", "port "+portName,
		"--private-ip-address", ipaddr,
		"--groups", sgrp.GroupId,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "create-network-interface result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("AttachPortToServer create interface failed: %s - %v", string(out), err)
	}
	var createdIf AwsEc2NetworkInterfaceCreateResult
	err = json.Unmarshal(out, &createdIf)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws attach-network-interface unmarshal fail", "out", string(out), "err", err)
		return fmt.Errorf("cannot unmarshal, %v", err)
	}
	deviceIndex := len(sd.Addresses)
	log.SpanLog(ctx, log.DebugLevelInfra, "created interface", "interface", createdIf)

	// Attach the interface
	out, err = a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"attach-network-interface",
		"--instance-id", sd.ID,
		"--network-interface-id", createdIf.NetworkInterface.NetworkInterfaceId,
		"--device-index", fmt.Sprintf("%d", deviceIndex),
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "attach-network-interface result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("AttachPortToServer attach interface failed: %s - %v", string(out), err)
	}

	// Disable SourceDestCheck to allow NAT
	out, err = a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"modify-network-interface-attribute",
		"--no-source-dest-check",
		"--network-interface-id", createdIf.NetworkInterface.NetworkInterfaceId,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "modify-network-interface-attribute result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("AttachPortToServer modify interface failed: %s - %v", string(out), err)
	}

	return nil
}

func (a *AWSPlatform) CheckServerReady(ctx context.Context, client ssh.Client, serverName string) error {
	// no special checks to be done
	return nil
}

func (a *AWSPlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams, groupPorts []vmlayer.PortOrchestrationParams, resources *VmGroupResources) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM", "vm", vm)

	udFileName := "/var/tmp/" + vm.Name + "-userdata.txt"
	udFile, err := os.Create(udFileName)
	defer udFile.Close()
	defer os.Remove(udFileName)
	_, err = udFile.WriteString(vm.UserData)
	if err != nil {
		return fmt.Errorf("Unable to write userdata file for vm: %s - %v", vm.Name, err)
	}

	if len(vm.Ports) == 0 {
		return fmt.Errorf("No ports specified in VM: %s", vm.Name)
	}
	extNet := a.VMProperties.GetCloudletExternalNetwork()
	tagspec := fmt.Sprintf("ResourceType=instance,Tags=[{Key=Name,Value=%s}]", vm.Name)
	var networkInterfaces []AwsEc2NetworkInterfaceCreateSpec
	for i, p := range vm.Ports {
		var ni AwsEc2NetworkInterfaceCreateSpec
		ni.DeviceIndex = i
		snName := ""
		if p.NetworkId == extNet {
			snName = p.NetworkId
			ni.AssociatePublicIpAddress = true
		} else {
			// for internal interface allow masquerading
			snName = p.SubnetId
			for _, f := range vm.FixedIPs {
				if f.Subnet.Name == snName {
					ni.PrivateIpAddress = f.Address
				}
			}
		}
		snId, ok := resources.subnetMap[snName]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfra, "subnet not in map", "snId", snId, "subnets", resources.subnetMap)

			return fmt.Errorf("Could not find subnet: %s", snName)
		}
		ni.SubnetId = snId.SubnetId
		for _, gp := range groupPorts {
			if gp.Name == p.Name {
				for _, s := range gp.SecurityGroups {
					sg, ok := resources.secGrpMap[s.Name]
					if !ok {
						return fmt.Errorf("Cannot find EC2 security group: %s in vpc: %s", s.Name, resources.vpcId)
					}
					ni.Groups = append(ni.Groups, sg.GroupId)
				}
			}

		}
		networkInterfaces = append(networkInterfaces, ni)
	}
	niParms, err := json.Marshal(networkInterfaces)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to marshal network interfaces", "networkInterfaces", networkInterfaces, "err", err)
		return fmt.Errorf("Failed to marshal network interfaces: %v", err)
	}
	var blockDevices []AwsEc2BlockDeviceMapping
	for _, v := range vm.Volumes {
		blockDevice := AwsEc2BlockDeviceMapping{
			DeviceName: v.DeviceName,
			Ebs: AwsEc2Ebs{
				DeleteOnTermination: true,
			},
		}
		blockDevices = append(blockDevices, blockDevice)
	}
	ebsParams, err := json.Marshal(blockDevices)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to marshal ebs block devices", "blockDevices", blockDevices, "err", err)
		return fmt.Errorf("Failed to marshal ec2 ebs disks: %v", err)
	}

	imgId, ok := resources.imageNameToId[vm.ImageName]
	if !ok {
		// should not happen, we should have failed earlier
		return fmt.Errorf("Image not found: %s", imgId)
	}
	createArgs := []string{
		"ec2",
		"run-instances",
		"--image-id", imgId,
		"--count", fmt.Sprintf("%d", 1),
		"--instance-type", vm.FlavorName,
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec,
		"--user-data", "file://" + udFileName,
		"--network-interfaces", string(niParms),
		"--block-device-mappings", string(ebsParams),
	}
	out, err := a.TimedAwsCommand(ctx, "aws", createArgs...)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("run-instances error: %s - %v", string(out), err)
	}
	return nil
}

// meta data needs to have an extra layer "meta" for vsphere
func awsMetaDataFormatter(instring string) string {
	indented := ""
	for _, v := range strings.Split(instring, "\n") {
		indented += strings.Repeat(" ", 4) + v + "\n"
	}
	withMeta := fmt.Sprintf("meta:\n%s", indented)
	return base64.StdEncoding.EncodeToString([]byte(withMeta))
}

// meta data needs to have an extra layer "meta" for vsphere
func awsUserDataFormatter(instring string) string {
	// aws ec2 needs to leave as raw text
	return instring
}

func (a *AWSPlatform) populateOrchestrationParams(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "populateOrchestrationParams")

	usedCidrs := make(map[string]string)
	if !vmgp.SkipInfraSpecificCheck {
		subs, err := a.GetSubnets(ctx)
		if err != nil {
			return nil
		}
		for _, s := range subs {
			usedCidrs[s.CidrBlock] = s.Name
		}
	}
	masterIP := ""
	currentSubnetName := ""
	if action != vmlayer.ActionCreate {
		currentSubnetName = vmlayer.MexSubnetPrefix + vmgp.GroupName
	}

	// find an available subnet or the current subnet for update and delete
	for i, s := range vmgp.Subnets {
		if s.CIDR != vmlayer.NextAvailableResource || vmgp.SkipInfraSpecificCheck {
			// no need to compute the CIDR
			continue
		}
		found := false
		for octet := 0; octet <= 255; octet++ {
			subnet := fmt.Sprintf("%s.%s.%d.%d/%s", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 0, vmgp.Netspec.NetmaskBits)
			// either look for an unused one (create) or the current one (update)
			newSubnet := action == vmlayer.ActionCreate
			if (newSubnet && usedCidrs[subnet] == "") || (!newSubnet && usedCidrs[subnet] == currentSubnetName) {
				found = true
				vmgp.Subnets[i].CIDR = subnet
				vmgp.Subnets[i].GatewayIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, AwsGwOctet)
				vmgp.Subnets[i].NodeIPPrefix = fmt.Sprintf("%s.%s.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet)
				masterIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 10)
				break
			}
		}
		if !found {
			return fmt.Errorf("cannot find subnet cidr")
		}
	}

	metaDir := "/mnt/mobiledgex-config/openstack/latest/"
	for vmidx, vm := range vmgp.VMs {
		metaData := vmlayer.GetVMMetaData(vm.Role, masterIP, awsMetaDataFormatter)
		vm.CloudConfigParams.ExtraBootCommands = append(vm.CloudConfigParams.ExtraBootCommands, "mkdir -p "+metaDir)
		vm.CloudConfigParams.ExtraBootCommands = append(vm.CloudConfigParams.ExtraBootCommands,
			fmt.Sprintf("echo %s |base64 -d|python3 -c \"import sys, yaml, json; json.dump(yaml.load(sys.stdin), sys.stdout)\" > "+metaDir+"meta_data.json", metaData))
		userdata, err := vmlayer.GetVMUserData(vm.Name, vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, &vm.CloudConfigParams, awsUserDataFormatter)
		if err != nil {
			return err
		}
		vmgp.VMs[vmidx].UserData = userdata

		if len(vm.Volumes) == 0 {
			vol := vmlayer.VolumeOrchestrationParams{
				DeviceName:         "/dev/sda1",
				Name:               "root disk",
				Size:               vmgp.VMs[vmidx].Disk,
				AttachExternalDisk: false,
			}
			vmgp.VMs[vmidx].Volumes = append(vmgp.VMs[vmidx].Volumes, vol)
		}

		for f, fip := range vm.FixedIPs {
			if fip.Address == vmlayer.NextAvailableResource && fip.LastIPOctet != 0 {
				log.SpanLog(ctx, log.DebugLevelInfra, "updating fixed ip", "fixedip", fip)
				for _, s := range vmgp.Subnets {
					if s.Name == fip.Subnet.Name {
						addr := fmt.Sprintf("%s.%d", s.NodeIPPrefix, fip.LastIPOctet)
						log.SpanLog(ctx, log.DebugLevelInfra, "populating fixed ip based on subnet", "addr", addr, "subnet", s)
						vmgp.VMs[vmidx].FixedIPs[f].Address = addr
						break
					}
				}
			}
		}

	}
	return nil
}

func (a *AWSPlatform) GetImageId(ctx context.Context, imageName, accountId string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetImageId", "imageName", imageName, "accountId", accountId)

	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-images",
		"--region", a.GetAwsRegion(),
		"--owners", accountId)

	log.SpanLog(ctx, log.DebugLevelInfra, "describe-images result", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("GetImageId failed: %s - %v", string(out), err)
	}

	var images AwsEc2ImageList
	err = json.Unmarshal(out, &images)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-images unmarshal fail", "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return "", err
	}
	for _, img := range images.Images {
		if img.Name == imageName {
			if img.State != "available" {
				return "", fmt.Errorf("%s:%s", ImageNotAvailableError, img.State)
			}
			return img.ImageId, nil
		}
	}
	return "", fmt.Errorf(ImageDoesNotExistError)
}

// GetIamAccountId gets the account Id for the logged in user
func (a *AWSPlatform) GetIamAccountId(ctx context.Context) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetIamAccountId")

	out, err := a.TimedAwsCommand(ctx, "aws",
		"iam",
		"get-user")

	log.SpanLog(ctx, log.DebugLevelInfra, "get-user result", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("GetIamAccountId failed: %s - %v", string(out), err)
	}
	var iamResult AwsIamUserResult
	err = json.Unmarshal(out, &iamResult)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws get-user unmarshal fail", "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return "", err
	}
	arns := strings.Split(iamResult.User.Arn, ":")
	if len(arns) <= ArnAccountIdIdx {
		log.SpanLog(ctx, log.DebugLevelInfra, "Wrong number of fields in ARN", "iamResult.User.Arn", iamResult.User.Arn)
		return "", fmt.Errorf("Cannot parse IAM ARN: %s", iamResult.User.Arn)
	}
	return arns[ArnAccountIdIdx], nil
}

// createVmGroupResources creates subnets, secgrps ahead of vms.  returns secGrpMap, subnetMap, vpcid, err
func (a *AWSPlatform) createVmGroupResources(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (*VmGroupResources, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "createVmGroupResources", "vmgp", vmgp)

	var resources VmGroupResources
	// lock to reserve subnets.  AWS is very fast on create so this is probably ok, but
	// should be revisited
	orchVmLock.Lock()
	defer orchVmLock.Unlock()
	err := a.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)
	if err != nil {
		return nil, err
	}
	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return nil, err
	}
	resources.vpcId = vpc.VpcId

	mexNet := a.VMProperties.GetCloudletMexNetwork()
	internalRouteTableId, err := a.GetRouteTableId(ctx, vpc.VpcId, SearchForRouteTableByName, mexNet)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "Params after populate", "vmgp", vmgp)

	secGrpMap, err := a.GetSecurityGroups(ctx, vpc.VpcId)
	if err != nil {
		return nil, err
	}

	for _, sg := range vmgp.SecurityGroups {
		_, ok := secGrpMap[sg.Name]
		if !ok {
			newgrp, err := a.CreateSecurityGroup(ctx, sg.Name, vpc.VpcId, "security group for VM group "+vmgp.GroupName)
			if err != nil {
				if strings.Contains(err.Error(), GroupAlreadyExistsError) {
					log.SpanLog(ctx, log.DebugLevelInfra, "security group already exists", "vmgp", vmgp)
				}
			} else {
				return nil, err
			}
			secGrpMap[sg.Name] = newgrp
		}
	}
	for _, sn := range vmgp.Subnets {
		routeTableId := MainRouteTable
		if sn.NetworkName == mexNet {
			routeTableId = internalRouteTableId
		}
		_, err := a.CreateSubnet(ctx, sn.Name, sn.CIDR, routeTableId)
		if err != nil {
			return nil, err
		}
	}
	resources.secGrpMap = secGrpMap

	snMap, err := a.GetSubnets(ctx)
	if err != nil {
		return nil, err
	}
	resources.subnetMap = snMap

	// populate image map
	resources.imageNameToId = make(map[string]string)
	for _, vm := range vmgp.VMs {
		_, ok := resources.imageNameToId[vm.ImageName]
		if !ok {
			imgId, err := a.GetImageId(ctx, vm.ImageName, a.IamAccountId)
			if err != nil {
				return nil, err
			}
			resources.imageNameToId[vm.ImageName] = imgId
		}
	}

	return &resources, nil
}

func (a *AWSPlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs", "vmgp", vmgp)
	resources, err := a.createVmGroupResources(ctx, vmgp, updateCallback)
	if err != nil {
		return err
	}
	for _, vm := range vmgp.VMs {
		err := a.CreateVM(ctx, &vm, vmgp.Ports, resources)
		if err != nil {
			return err
		}
	}
	return nil
}
func (o *AWSPlatform) UpdateVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("UpdateVMs not implemented")
}

func (o *AWSPlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	return fmt.Errorf("DeleteVMs not implemented")
}

func (s *AWSPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName string, portName string) error {
	return fmt.Errorf("DetachPortFromServer not implemented")
}

func (a *AWSPlatform) GetInternalPortPolicy() vmlayer.InternalPortAttachPolicy {
	return vmlayer.AttachPortAfterCreate
}

func (a *AWSPlatform) GetVMStats(ctx context.Context, key *edgeproto.AppInstKey) (*vmlayer.VMMetrics, error) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "GetVMStats not supported")
	return &vmlayer.VMMetrics{}, nil
}

func (a *AWSPlatform) SetPowerState(ctx context.Context, serverName, serverAction string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetPowerState not supported")
	return nil
}

func (a *AWSPlatform) GetType() string {
	return "awsvm"
}

func (a *AWSPlatform) GetVpcName() string {
	return a.NameSanitize(a.VMProperties.CommonPf.PlatformConfig.CloudletKey.Name)
}

// CreateVPC returns the vpcid
func (a *AWSPlatform) CreateVPC(ctx context.Context, name string, cidr string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVPC", "name", name, "cidr", cidr)
	vpc, err := a.GetVPC(ctx, name)
	if err == nil {
		// VPC already exists
		return vpc.VpcId, nil
	}
	if !strings.Contains(err.Error(), VpcDoesNotExistError) {
		// unexpected error
		return "", err
	}
	tagspec := fmt.Sprintf("ResourceType=vpc,Tags=[{Key=Name,Value=%s}]", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-vpc",
		"--cidr-block", cidr,
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec)
	log.SpanLog(ctx, log.DebugLevelInfra, "create-vpc result", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("CreateVPC failed: %s - %v", string(out), err)
	}
	// the create-vpc command returns a json of the vpc
	var createdVpc AwsEc2VpcCreateResult
	err = json.Unmarshal(out, &createdVpc)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-vpc unmarshal fail", "name", name, "out", string(out), "err", err)
		return "", fmt.Errorf("cannot unmarshal, %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created vpc", "vpc", createdVpc)
	if createdVpc.Vpc.VpcId == "" {
		return "", fmt.Errorf("no VPCID in VPC %s", name)
	}
	return createdVpc.Vpc.VpcId, err
}

// CreateSubnet returns subnetId, error
func (a *AWSPlatform) CreateSubnet(ctx context.Context, name string, cidr string, routeTableId string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateSubnet", "name", name, "routeTableId", routeTableId)
	tagspec := fmt.Sprintf("ResourceType=subnet,Tags=[{Key=Name,Value=%s}]", name)

	sn, err := a.GetSubnet(ctx, name)
	if err == nil {
		// already exists
		return sn.SubnetId, nil
	}
	if !strings.Contains(err.Error(), SubnetDoesNotExistError) {
		return "", err
	}
	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return "", err
	}
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-subnet",
		"--vpc-id", vpc.VpcId,
		"--region", a.GetAwsRegion(),
		"--cidr-block", cidr,
		"--tag-specifications", tagspec)

	log.SpanLog(ctx, log.DebugLevelInfra, "create-subnet result", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("Error in creating subnet: %s - %v", string(out), err)
	}
	var createdSn AwsEc2SubnetCreateResult
	err = json.Unmarshal(out, &createdSn)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-subnet unmarshal fail", "out", string(out), "err", err)
		return "", fmt.Errorf("cannot unmarshal, %v", err)
	}
	if routeTableId != MainRouteTable {
		// associate the non default route table
		out, err := a.TimedAwsCommand(ctx, "aws",
			"ec2",
			"associate-route-table",
			"--route-table-id", routeTableId,
			"--subnet-id", createdSn.Subnet.SubnetId)

		log.SpanLog(ctx, log.DebugLevelInfra, "associate-route-table result", "out", string(out), "err", err)
		if err != nil {
			return "", fmt.Errorf("Error in associating route table: %s - %v", string(out), err)
		}
	}

	return createdSn.Subnet.SubnetId, nil
}

func (a *AWSPlatform) GetGateway(ctx context.Context, name string) (*AwsEc2Gateway, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetGateway", "name", name)
	filter := fmt.Sprintf("Name=tag-value,Values=%s", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-internet-gateways",
		"--filters", filter,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-internet-gateways result", "out", string(out), "err", err)
	if err != nil {
		return nil, fmt.Errorf("GetGateway failed: %s - %v", string(out), err)
	}
	var gwList AwsEc2GatewayList
	err = json.Unmarshal(out, &gwList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-internet-gateways unmarshal fail", "name", name, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	if len(gwList.InternetGateways) == 0 {
		return nil, fmt.Errorf(GatewayDoesNotExistError + ":" + name)
	}
	// there is nothing to prevent creating 2 GWs with the same name tag, but it indicates
	// an error for us.
	if len(gwList.InternetGateways) > 2 {
		return nil, fmt.Errorf("more than one subnet matching name tag: %s - numsubnets: %d", name, len(gwList.InternetGateways))
	}
	return &gwList.InternetGateways[0], nil
}

// CreateInternalRouteTable returns routeTableId, error
func (a *AWSPlatform) CreateInternalRouteTable(ctx context.Context, vpcId, natGwId, name string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateInternalRouteTable", "name", name)
	rt, err := a.GetRouteTableId(ctx, vpcId, SearchForRouteTableByName, name)
	if err == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "RouteTable already exists")
		return rt, nil
	}
	if err != nil {
		if !strings.Contains(err.Error(), RouteTableDoesNotExistError) {
			return "", err
		}
	}

	tagspec := fmt.Sprintf("ResourceType=route-table,Tags=[{Key=Name,Value=%s}]", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-route-table",
		"--vpc-id", vpcId,
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec)

	log.SpanLog(ctx, log.DebugLevelInfra, "create-route-table", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("Error in creating route table: %s - %v", string(out), err)
	}

	// the create-route-table command returns a json of the rt
	var createdRt AwsEc2RouteTableCreateResult
	err = json.Unmarshal(out, &createdRt)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-route-table unmarshal fail", "name", name, "out", string(out), "err", err)
		return "", fmt.Errorf("cannot unmarshal, %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created route table", "rt", createdRt)

	// now add the natgw as the default route
	out, err = a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-route",
		"--route-table-id", createdRt.RouteTable.RouteTableId,
		"--nat-gateway-id", natGwId,
		"--destination-cidr-block", "0.0.0.0/0",
		"--region", a.GetAwsRegion())

	log.SpanLog(ctx, log.DebugLevelInfra, "create-route", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("Error in creating route : %s - %v", string(out), err)
	}
	return createdRt.RouteTable.RouteTableId, nil
}

func (a *AWSPlatform) GetNatGateway(ctx context.Context, name string) (*AwsEc2NatGateway, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNatGateway", "name", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-nat-gateways",
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-nat-gateways result", "out", string(out), "err", err)
	if err != nil {
		return nil, fmt.Errorf("GetNatGateway failed: %s - %v", string(out), err)
	}
	var ngwList AwsEc2NatGatewayList
	err = json.Unmarshal(out, &ngwList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-nat-gateways unmarshal fail", "name", name, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	if len(ngwList.NatGateways) == 0 {
		return nil, fmt.Errorf(GatewayDoesNotExistError + ":" + name)
	}
	// there is nothing to prevent creating 2 GWs with the same name tag, but it indicates
	// an error for us.
	if len(ngwList.NatGateways) > 2 {
		return nil, fmt.Errorf("more than one subnet matching name tag: %s - numsubnets: %d", name, len(ngwList.NatGateways))
	}
	return &ngwList.NatGateways[0], nil
}

func (a *AWSPlatform) CreateGateway(ctx context.Context, vpcName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateGateway", "vpcName", vpcName)
	tagspec := fmt.Sprintf("ResourceType=internet-gateway,Tags=[{Key=Name,Value=%s}]", vpcName)

	_, err := a.GetGateway(ctx, vpcName)
	if err == nil {
		// already exists
		return err
	}
	if !strings.Contains(err.Error(), GatewayDoesNotExistError) {
		return err
	}
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-internet-gateway",
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec)

	log.SpanLog(ctx, log.DebugLevelInfra, "create-internet-gateway result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("Error in creating gateway: %s - %v", string(out), err)
	}
	return nil
}

// CreateNatGateway returns natGatewayId, error
func (a *AWSPlatform) CreateNatGateway(ctx context.Context, subnetId, elasticIpId, vpcName string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNatGateway", "subnetId", subnetId, "vpcName", vpcName)
	tagspec := fmt.Sprintf("ResourceType=natgateway,Tags=[{Key=Name,Value=%s}]", vpcName)

	ng, err := a.GetNatGateway(ctx, vpcName)
	if err == nil {
		// already exists
		return ng.NatGatewayId, nil
	}
	if !strings.Contains(err.Error(), GatewayDoesNotExistError) {
		return "", err
	}
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-nat-gateway",
		"--subnet-id", subnetId,
		"--allocation-id", elasticIpId,
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec)

	log.SpanLog(ctx, log.DebugLevelInfra, "create-nat-gateway result", "out", string(out), "err", err)
	if err != nil {
		return "", fmt.Errorf("Error in creating gateway: %s - %v", string(out), err)
	}
	var createdNg AwsEc2NatGatewayCreateResult
	err = json.Unmarshal(out, &createdNg)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-nat-gateway unmarshal fail", "out", string(out), "err", err)
		return "", fmt.Errorf("cannot unmarshal, %v", err)
	}
	return createdNg.NatGateway.NatGatewayId, nil
}

func (a *AWSPlatform) CreateGatewayDefaultRoute(ctx context.Context, vpcName, vpcId string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateGatewayDefaultRoute", "vpcName", vpcName, "vpcId", vpcId)

	gw, err := a.GetGateway(ctx, vpcName)
	if err != nil {
		return err
	}
	// attach the GW to the VPC
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"attach-internet-gateway",
		"--region", a.GetAwsRegion(),
		"--internet-gateway-id", gw.InternetGatewayId,
		"--vpc-id", vpcId)

	log.SpanLog(ctx, log.DebugLevelInfra, "attach-internet-gateway", "out", string(out), "err", err)
	if err != nil {
		if strings.Contains(string(out), ResourceAlreadyAssociatedError) {
			log.SpanLog(ctx, log.DebugLevelInfra, "gateway already attached")
		} else {
			return fmt.Errorf("Error in attach-internet-gateway: %s - %v", string(out), err)
		}
	}

	rtid, err := a.GetRouteTableId(ctx, vpcId, SearchForMainRouteTable, "")
	if err != nil {
		return err
	}
	out, err = a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-route",
		"--region", a.GetAwsRegion(),
		"--gateway-id", gw.InternetGatewayId,
		"--destination-cidr-block", "0.0.0.0/0",
		"--route-table-id", rtid)

	log.SpanLog(ctx, log.DebugLevelInfra, "create-route result", "out", string(out), "err", err)
	if err != nil {
		return fmt.Errorf("Error in create-route: %s - %v", out, err)
	}
	return nil
}

func (a *AWSPlatform) GetElasticIP(ctx context.Context, name, vpcId string) (string, error) {
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-addresses",
		"--region", a.GetAwsRegion())

	log.SpanLog(ctx, log.DebugLevelInfra, "describe-addresses", "out", string(out), "err", err)

	var addresses AwsEc2AddressList
	err = json.Unmarshal(out, &addresses)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-addresses unmarshal fail", "name", name, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return "", err
	}
	if len(addresses.Addresses) == 0 {
		return "", fmt.Errorf(ElasticIpDoesNotExistError + ":" + name)
	}
	return addresses.Addresses[0].AllocationId, nil
}

func (a *AWSPlatform) CreateSecurityGroup(ctx context.Context, name, vpcId, description string) (*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateSecurityGroup", "name", name, "vpcId", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"create-security-group",
		"--region", a.GetAwsRegion(),
		"--group-name", name,
		"--vpc-id", vpcId,
		"--description", description)
	if err != nil {
		return nil, fmt.Errorf("Error in create-security-group: %s - %v", string(out), err)
	}
	var sg AwsEc2SecGrp
	err = json.Unmarshal(out, &sg)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws create-security-group unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	sg.GroupName = name
	sg.VpcId = vpcId
	return &sg, nil
}

func (a *AWSPlatform) GetSecurityGroup(ctx context.Context, name string, vpcId string) (*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecurityGroup", "name", name, "vpcId", vpcId)

	grpMap, err := a.GetSecurityGroups(ctx, vpcId)
	if err != nil {
		return nil, err
	}
	grp, ok := grpMap[name]
	if !ok {
		return nil, fmt.Errorf(SecGrpDoesNotExistError)
	}
	return grp, nil
}

func (a *AWSPlatform) GetSecurityGroups(ctx context.Context, vpcId string) (map[string]*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecurityGroups", "vpcId", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-security-groups",
		"--region", a.GetAwsRegion())
	if err != nil {
		return nil, fmt.Errorf("error in describe-security-groups: %s - %v", string(out), err)
	}
	var sgList AwsEc2SecGrpList
	err = json.Unmarshal(out, &sgList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-security-groups unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	var sgMap = make(map[string]*AwsEc2SecGrp)
	for i, sg := range sgList.SecurityGroups {
		if sg.VpcId == vpcId {
			sgMap[sg.GroupName] = &sgList.SecurityGroups[i]
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "found security groups", "sgMap", sgMap)
	return sgMap, nil

}

func (a *AWSPlatform) GetRouteTableId(ctx context.Context, vpcId string, searchType RouteTableSearchType, name string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetRouteTableId", "vpcId", vpcId, "searchType", searchType, "name", name)
	filter := fmt.Sprintf("Name=vpc-id,Values=%s", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-route-tables",
		"--filters", filter,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-route-tables result", "out", string(out), "err", err)

	if err != nil {
		return "", fmt.Errorf("GetRouteTableId failed: %s - %v", string(out), err)
	}

	var rtList AwsEc2RouteTableList
	err = json.Unmarshal(out, &rtList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-route-tables unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return "", err
	}
	for i, rt := range rtList.RouteTables {
		if searchType == SearchForRouteTableByName {
			for _, tag := range rt.Tags {
				if tag.Value == name {
					return rtList.RouteTables[i].RouteTableId, nil
				}
			}
		} else if searchType == SearchForMainRouteTable {
			for _, a := range rt.Associations {
				if a.Main {
					return rtList.RouteTables[i].RouteTableId, nil
				}
			}
		} else {
			return "", fmt.Errorf("Must search route table either by main or name")
		}
	}
	return "", fmt.Errorf(RouteTableDoesNotExistError)
}

func (a *AWSPlatform) GetVPC(ctx context.Context, name string) (*AwsEc2Vpc, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetVPCs", "name", name)
	filter := fmt.Sprintf("Name=tag-value,Values=%s", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-vpcs",
		"--filters", filter,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-vpcs result", "out", string(out), "err", err)
	if err != nil {
		return nil, fmt.Errorf("GetVPC failed: %s - %v", string(out), err)
	}
	var vpclist AwsEc2VpcList
	err = json.Unmarshal(out, &vpclist)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-vpcs unmarshal fail", "name", name, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	if len(vpclist.Vpcs) == 0 {
		return nil, fmt.Errorf(VpcDoesNotExistError + ":" + name)
	}
	// there is nothing to prevent creating 2 VPCs with the same name tag, but it indicates
	// an error for us.
	if len(vpclist.Vpcs) > 2 {
		return nil, fmt.Errorf("more than one VPC matching name tag: %s - numvpcs: %d", name, len(vpclist.Vpcs))
	}
	return &vpclist.Vpcs[0], nil
}

func (a *AWSPlatform) GetSubnets(ctx context.Context) (map[string]*AwsEc2Subnet, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSubnets")
	snMap := make(map[string]*AwsEc2Subnet)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-subnets",
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-subnets result", "out", string(out), "err", err)
	if err != nil {
		return nil, fmt.Errorf("GetSubnets failed: %s - %v", string(out), err)
	}
	var subnetList AwsEc2SubnetList
	err = json.Unmarshal(out, &subnetList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-subnets unmarshal fail", "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	for i, s := range subnetList.Subnets {
		subnetName := ""
		for _, t := range s.Tags {
			if t.Key == "Name" {
				subnetName = t.Value
			}
		}
		if subnetName != "" {
			subnetList.Subnets[i].Name = subnetName
			snMap[subnetName] = &subnetList.Subnets[i]
		}
	}
	return snMap, nil
}

func (a *AWSPlatform) GetSubnet(ctx context.Context, name string) (*AwsEc2Subnet, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSubnet", "name", name)
	filter := fmt.Sprintf("Name=tag-value,Values=%s", name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-subnets",
		"--filters", filter,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-subnets result", "out", string(out), "err", err)
	if err != nil {
		return nil, fmt.Errorf("GetSubnet failed: %s - %v", string(out), err)
	}
	var subnetList AwsEc2SubnetList
	err = json.Unmarshal(out, &subnetList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-subnets unmarshal fail", "name", name, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return nil, err
	}
	if len(subnetList.Subnets) == 0 {
		return nil, fmt.Errorf(SubnetDoesNotExistError + ":" + name)
	}
	// there is nothing to prevent creating 2 VPCs with the same name tag, but it indicates
	// an error for us.
	if len(subnetList.Subnets) > 2 {
		return nil, fmt.Errorf("more than one subnet matching name tag: %s - numsubnets: %d", name, len(subnetList.Subnets))
	}
	subnetList.Subnets[0].Name = name
	return &subnetList.Subnets[0], nil
}

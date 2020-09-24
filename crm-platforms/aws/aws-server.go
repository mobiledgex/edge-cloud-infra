package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

const VpcDoesNotExistError string = "vpc does not exist"
const SubnetDoesNotExistError string = "subnet does not exist"
const GatewayDoesNotExistError string = "gateway does not exist"
const ResourceAlreadyAssociated string = "Resource.AlreadyAssociated"
const GroupAlreadyExists string = "InvalidGroup.Duplicate"

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

type AwsEc2RouteTable struct {
	RouteTableId string
	VpcId        string
}
type AwsEc2RouteTableList struct {
	RouteTables []AwsEc2RouteTable
}

type AwsEc2Gateway struct {
	InternetGatewayId string
	Tags              []AwsEc2Tag
}

type AwsEc2GatewayList struct {
	InternetGateways []AwsEc2Gateway
}

type AwsEc2Subnet struct {
	CidrBlock string
	State     string
	SubnetId  string
	VpcId     string
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

type AwsEc2VpcList struct {
	Vpcs []AwsEc2Vpc
}

type AwsEc2State struct {
	Code int
	Name string
}

type AwsEc2NetworkInterfaceCreateSpec struct {
	AssociatePublicIpAddress bool     `json:"AssociatePublicIpAddress"`
	SubnetId                 string   `json:"SubnetId,omitempty"`
	PrivateIpAddress         string   `json:"PrivateIpAddress,omitempty"`
	Groups                   []string `json:"Groups,omitempty"`
	DeviceIndex              int      `json:"DeviceIndex"`
}

type AwsEc2Ebs struct {
	DeleteOnTermination bool
	Status              string `json:"Status,omitempty"`
}
type AwsEc2BlockDeviceMapping struct {
	DeviceName string
	Ebs        AwsEc2Ebs
}

type AwsEc2Instance struct {
	ImageId          string
	PrivateIpAddress string
	PublicIpAddress  string
	Tags             []AwsEc2Tag
	State            AwsEc2State
}

type AwsEc2Reservation struct {
	Instances []AwsEc2Instance
}

type AwsEc2Instances struct {
	Reservations []AwsEc2Reservation
}

func (a *AWSPlatform) GetServerDetail(ctx context.Context, vmname string) (*vmlayer.ServerDetail, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetServerDetail", "vmname", vmname)

	var sd vmlayer.ServerDetail
	var ec2insts AwsEc2Instances
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

			if inst.PublicIpAddress != "" {
				var sip vmlayer.ServerIP
				sip.ExternalAddr = inst.PublicIpAddress
				sip.InternalAddr = inst.PrivateIpAddress
				sip.ExternalAddrIsFloating = true
				sip.Network = a.VMProperties.GetCloudletExternalNetwork()
				sd.Addresses = append(sd.Addresses, sip)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "active server", "vmname", vmname, "state", inst.State, "sd", sd)
			return &sd, nil
		}
	}
	return &sd, fmt.Errorf(vmlayer.ServerDoesNotExistError)
}

func (a *AWSPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer not supported")
	return nil
}

func (a *AWSPlatform) CheckServerReady(ctx context.Context, client ssh.Client, serverName string) error {
	// no special checks to be done
	return nil
}

func (a *AWSPlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams, groupPorts []vmlayer.PortOrchestrationParams, awsSecGrps map[string]*AwsEc2SecGrp, vpcid string) error {
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
	subnet, err := a.GetSubnet(ctx, vm.Ports[0].NetworkId)
	if err != nil {
		return err
	}

	tagspec := fmt.Sprintf("ResourceType=instance,Tags=[{Key=Name,Value=%s}]", vm.Name)
	var networkInterfaces []AwsEc2NetworkInterfaceCreateSpec
	for i, p := range vm.Ports {
		var ni AwsEc2NetworkInterfaceCreateSpec
		ni.DeviceIndex = i
		if p.NetworkId == a.VMProperties.GetCloudletExternalNetwork() {
			ni.AssociatePublicIpAddress = true
			ni.SubnetId = p.SubnetId
		} else {
			ni.SubnetId = p.SubnetId
		}
		for _, gp := range groupPorts {
			if gp.Name == p.Name {
				for _, s := range gp.SecurityGroups {
					sg, ok := awsSecGrps[s.Name]
					if !ok {
						return fmt.Errorf("Cannot find EC2 security group: %s in vpc: %s", s.Name, vpcid)
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

	createArgs := []string{
		"ec2",
		"run-instances",
		"--image-id", vm.ImageName,
		"--count", fmt.Sprintf("%d", 1),
		"--instance-type", vm.FlavorName,
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec,
		"--subnet-id", subnet.SubnetId,
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

	metaDir := "/mnt/mobiledgex-config/openstack/latest/"
	for vmidx, vm := range vmgp.VMs {
		masterIp := ""

		metaData := vmlayer.GetVMMetaData(vm.Role, masterIp, awsMetaDataFormatter)
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
	}

	return nil
}

func (a *AWSPlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs", "vmgp", vmgp)
	err := a.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)
	if err != nil {
		return err
	}
	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Params after populate", "vmgp", vmgp)

	secGrpMap, err := a.GetSecurityGroups(ctx, vpc.VpcId)
	if err != nil {
		return err
	}
	for _, s := range vmgp.SecurityGroups {
		_, ok := secGrpMap[s.Name]
		if !ok {
			newgrp, err := a.CreateSecurityGroup(ctx, s.Name, vpc.VpcId, "security group for VM group "+vmgp.GroupName)
			if err != nil && !strings.Contains(err.Error(), GroupAlreadyExists) {
				return err
			}
			secGrpMap[s.Name] = newgrp
		}
	}
	for _, vm := range vmgp.VMs {
		err := a.CreateVM(ctx, &vm, vmgp.Ports, secGrpMap, vpc.VpcId)
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
	return vmlayer.AttachPortDuringCreate
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

func (a *AWSPlatform) CreateSubnet(ctx context.Context, name string, cidr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateSubnet", "name", name)
	tagspec := fmt.Sprintf("ResourceType=subnet,Tags=[{Key=Name,Value=%s}]", name)

	_, err := a.GetSubnet(ctx, name)
	if err == nil {
		// already exists
		return err
	}
	if !strings.Contains(err.Error(), SubnetDoesNotExistError) {
		return err
	}
	vpc, err := a.GetVPC(ctx, a.GetVpcName())
	if err != nil {
		return err
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
		return fmt.Errorf("Error in creating subnet: %s - %v", string(out), err)
	}
	return nil
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
		if strings.Contains(string(out), ResourceAlreadyAssociated) {
			log.SpanLog(ctx, log.DebugLevelInfra, "gateway already attached")
		} else {
			return fmt.Errorf("Error in attach-internet-gateway: %s - %v", string(out), err)
		}
	}

	rtid, err := a.GetMainRouteTableForVpcId(ctx, vpcId)
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

func (a *AWSPlatform) GetSecurityGroups(ctx context.Context, vpcId string) (map[string]*AwsEc2SecGrp, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecurityGroup", "vpcId", vpcId)
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

func (a *AWSPlatform) GetMainRouteTableForVpcId(ctx context.Context, vpcId string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetMainRouteTableForVpcId", "vpcId", vpcId)
	filter := fmt.Sprintf("Name=vpc-id,Values=%s,association.main", vpcId)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"describe-route-tables",
		"--filters", filter,
		"--region", a.GetAwsRegion())
	log.SpanLog(ctx, log.DebugLevelInfra, "describe-route-tables result", "out", string(out), "err", err)

	if err != nil {
		return "", fmt.Errorf("GetMainRouteTableForVpcId failed: %s - %v", string(out), err)
	}

	var rtList AwsEc2RouteTableList
	err = json.Unmarshal(out, &rtList)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "aws describe-route-tables unmarshal fail", "vpcId", vpcId, "out", string(out), "err", err)
		err = fmt.Errorf("cannot unmarshal, %v", err)
		return "", err
	}
	if len(rtList.RouteTables) != 1 {
		return "", fmt.Errorf("Expected to find one main route table for VPC: %s, found: %d", vpcId, len(rtList.RouteTables))
	}
	return rtList.RouteTables[0].RouteTableId, nil
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
	return &subnetList.Subnets[0], nil
}

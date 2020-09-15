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

type AwsEc2Tag struct {
	Key   string
	Value string
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

type AwsEc2VpcList struct {
	Vpcs []AwsEc2Vpc
}

type AwsEc2State struct {
	Code int
	Name string
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

func (a *AWSPlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams, vpcid string) error {
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
	}
	if vm.Ports[0].NetworkId == a.VMProperties.GetCloudletExternalNetwork() {
		createArgs = append(createArgs, "--associate-public-ip-address")
	}
	out, err := a.TimedAwsCommand(ctx, "aws", createArgs...)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM result", "out", string(out), "err", err)
	return err
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
		vm.UserDataParams.ExtraBootCommands = append(vm.UserDataParams.ExtraBootCommands, "mkdir -p "+metaDir)
		vm.UserDataParams.ExtraBootCommands = append(vm.UserDataParams.ExtraBootCommands,
			fmt.Sprintf("echo %s |base64 -d|python3 -c \"import sys, yaml, json; json.dump(yaml.load(sys.stdin), sys.stdout)\" > "+metaDir+"meta_data.json", metaData))
		userdata, err := vmlayer.GetVMUserData(vm.Name, vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, &vm.UserDataParams, awsUserDataFormatter)
		if err != nil {
			return err
		}
		vmgp.VMs[vmidx].UserData = userdata
	}
	return nil
}

func (a *AWSPlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs", "vmgp", vmgp)
	err := a.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)
	if err != nil {
		return err
	}
	vpc, err := a.GetVPC(ctx, a.GetVpcName(ctx))
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Params after populate", "vmgp", vmgp)

	for _, vm := range vmgp.VMs {
		err := a.CreateVM(ctx, &vm, vpc.VpcId)
		if err != nil {
			// TOTO CLEANUP
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

func (a *AWSPlatform) GetVpcName(ctx context.Context) string {
	return a.NameSanitize(a.VMProperties.CommonPf.PlatformConfig.CloudletKey.Organization + "-" + a.VMProperties.CommonPf.PlatformConfig.CloudletKey.Name)
}

func (a *AWSPlatform) CreateVPC(ctx context.Context, name string, cidr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVPC", "name", name, "cidr", cidr)
	_, err := a.GetVPC(ctx, name)
	if err == nil {
		// VPC already exists
		return nil
	}
	if !strings.Contains(err.Error(), VpcDoesNotExistError) {
		// unexpected error
		return err
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
		return fmt.Errorf("CreateVPC failed: %s - %v", string(out), err)
	}
	return nil
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
	vpc, err := a.GetVPC(ctx, a.GetVpcName(ctx))
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

func (a *AWSPlatform) CreateGateway(ctx context.Context, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateGateway", "name", name)
	tagspec := fmt.Sprintf("ResourceType=internet-gateway,Tags=[{Key=Name,Value=%s}]", name)

	_, err := a.GetGateway(ctx, name)
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

func (a *AWSPlatform) CreateGatewayDefaultRoute(ctx context.Context, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateGatewayDefaultRoute", "name", name)

	vpc, err := a.GetVPC(ctx, name)
	if err != nil {
		return err
	}
	gw, err := a.GetGateway(ctx, name)
	if err != nil {
		return err
	}
	// attach the GW to the VPC
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"attach-internet-gateway",
		"--region", a.GetAwsRegion(),
		"--internet-gateway-id", gw.InternetGatewayId,
		"--vpc-id", vpc.VpcId)

	log.SpanLog(ctx, log.DebugLevelInfra, "attach-internet-gateway", "out", string(out), "err", err)
	if err != nil {
		if strings.Contains(string(out), ResourceAlreadyAssociated) {
			log.SpanLog(ctx, log.DebugLevelInfra, "gateway already attached")
		} else {
			return fmt.Errorf("Error in attach-internet-gateway: %s - %v", string(out), err)
		}
	}

	rtid, err := a.GetMainRouteTableForVpcId(ctx, vpc.VpcId)
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

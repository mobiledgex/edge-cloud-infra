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

type AwsEc2Tag struct {
	Key   string
	Value string
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

func (a *AWSPlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM", "vm", vm)

	udFileName := "/var/tmp/" + vm.Name + "-userdata.txt"
	udFile, err := os.Create(udFileName)
	defer udFile.Close()
	_, err = udFile.WriteString(vm.UserData)
	if err != nil {
		return fmt.Errorf("Unable to write userdata file for vm: %s - %v", vm.Name, err)
	}

	tagspec := fmt.Sprintf("ResourceType=instance,Tags=[{Key=Name,Value=%s}]", vm.Name)
	out, err := a.TimedAwsCommand(ctx, "aws",
		"ec2",
		"run-instances",
		"--image-id", vm.ImageName,
		"--count", fmt.Sprintf("%d", 1),
		"--instance-type", vm.FlavorName,
		"--security-group-ids", "mex-scgrp", //TODO
		"--region", a.GetAwsRegion(),
		"--tag-specifications", tagspec,
		"--user-data", "file://"+udFileName)
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
	log.SpanLog(ctx, log.DebugLevelInfra, "Params after populate", "vmgp", vmgp)

	for _, vm := range vmgp.VMs {
		err := a.CreateVM(ctx, &vm)
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

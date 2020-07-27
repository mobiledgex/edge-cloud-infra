package vmpool

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func (s *VMPoolPlatform) GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error) {
	var flavors []*edgeproto.FlavorInfo
	if s.caches == nil || s.caches.VMPool == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	accessIP := ""
	// find one of the VM with external IP
	// we'll use this VM to access VMs with just internal network access
	for _, poolVM := range s.caches.VMPool.Vms {
		if poolVM.NetInfo.ExternalIp != "" {
			accessIP = poolVM.NetInfo.ExternalIp
			break
		}
	}
	if accessIP == "" {
		return nil, fmt.Errorf("atleast one VM should have access to external network")
	}
	accessClient, err := s.VMProperties.GetSSHClientFromIPAddr(ctx, accessIP)
	if err != nil {
		return nil, fmt.Errorf("can't get ssh client for %s, %v", accessIP, err)
	}

	flavorMap := make(map[string]string)
	for _, vm := range s.caches.VMPool.Vms {
		var client ssh.Client
		if vm.NetInfo.ExternalIp != "" {
			client, err = s.VMProperties.GetSSHClientFromIPAddr(ctx, vm.NetInfo.ExternalIp)
			if err != nil {
				return nil, fmt.Errorf("failed to verify vm %s, can't get ssh client for %s, %v", vm.Name, vm.NetInfo.ExternalIp, err)
			}
		} else if vm.NetInfo.InternalIp != "" {

			client, err = accessClient.AddHop(vm.NetInfo.InternalIp, 22)
			if err != nil {
				return nil, fmt.Errorf("failed to verify vm %s, can't get ssh client for %s, %v", vm.Name, vm.NetInfo.InternalIp, err)
			}
		} else {
			return nil, fmt.Errorf("VM %s is missing network info", vm.Name)
		}
		out, err := client.Output("sudo bash /etc/mobiledgex/get-flavor.sh")
		if err != nil {
			return nil, fmt.Errorf("failed to get flavor info for %s: %s - %v", vm.Name, out, err)
		}
		flavorMap[out] = vm.Name
	}

	count := 1
	for fID, vmName := range flavorMap {
		parts := strings.Split(fID, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid flavor info for %s: %s", vmName, fID)
		}

		memMb, err := strconv.Atoi(parts[0])
		if err != nil || memMb <= 0 {
			return nil, fmt.Errorf("invalid memory info %s for %s: %v", parts[0], vmName, err)
		}

		vcpus, err := strconv.Atoi(parts[1])
		if err != nil || vcpus <= 0 {
			return nil, fmt.Errorf("invalid vcpu info %s for %s: %v", parts[1], vmName, err)
		}

		diskGb, err := strconv.Atoi(parts[2])
		if err != nil || diskGb <= 0 {
			return nil, fmt.Errorf("invalid disk info %s for %s: %v", parts[2], vmName, err)
		}

		var flavInfo edgeproto.FlavorInfo
		flavInfo.Name = fmt.Sprintf("flavor%d", count)
		flavInfo.Ram = uint64(memMb)
		flavInfo.Vcpus = uint64(vcpus)
		flavInfo.Disk = uint64(diskGb)
		flavors = append(flavors, &flavInfo)
	}

	return flavors, nil
}

func (s *VMPoolPlatform) SetPowerState(ctx context.Context, serverName, serverAction string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetPowerState not supported")
	return nil
}

func (s *VMPoolPlatform) GetCloudletImageSuffix(ctx context.Context) string {
	return ".qcow2"
}

func (s *VMPoolPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddCloudletImageIfNotPresent not supported")
	imgPath := vmlayer.GetCloudletVMImagePath(imgPathPrefix, imgVersion, s.GetCloudletImageSuffix(ctx))

	// Fetch platform base image name
	pfImageName, err := cloudcommon.GetFileName(imgPath)
	if err != nil {
		return "", err
	}
	return pfImageName, nil
}

func (s *VMPoolPlatform) DeleteImage(ctx context.Context, folder, imageName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteImage not supported")
	return nil
}

func (s *VMPoolPlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer not supported")
	return nil
}

func (s *VMPoolPlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName string, portName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer not supported")
	return nil
}

func (s *VMPoolPlatform) GetNetworkList(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

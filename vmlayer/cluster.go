package vmlayer

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

//ClusterNodeFlavor contains details of flavor for the node
type ClusterNodeFlavor struct {
	Type string
	Name string
}

//ClusterFlavor contains definitions of cluster flavor
type ClusterFlavor struct {
	Kind           string
	Name           string
	PlatformFlavor string
	Status         string
	NumNodes       int
	MaxNodes       int
	NumMasterNodes int
	NetworkSpec    string
	StorageSpec    string
	NodeFlavor     ClusterNodeFlavor
	Topology       string
}

func GetClusterName(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
}

func GetClusterSubnetName(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return "mex-k8s-subnet-" + GetClusterName(ctx, clusterInst)
}

func GetClusterMasterName(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	namePrefix := ClusterTypeKubernetesMasterLabel
	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		namePrefix = ClusterTypeDockerVMLabel
	}
	return namePrefix + "-" + GetClusterName(ctx, clusterInst)
}

func GetClusterNodeName(ctx context.Context, clusterInst *edgeproto.ClusterInst, nodeNum uint32) string {
	return ClusterNodePrefix(nodeNum) + "-" + GetClusterName(ctx, clusterInst)
}

func (v *VMPlatform) GetDockerNodeName(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	return "docker-node" + "-" + GetClusterName(ctx, clusterInst)
}

func ClusterNodePrefix(num uint32) string {
	return fmt.Sprintf("%s%d", cloudcommon.MexNodePrefix, num)
}

func ParseClusterNodePrefix(name string) (bool, uint32) {
	reg := regexp.MustCompile("^" + cloudcommon.MexNodePrefix + "(\\d+).*")
	matches := reg.FindSubmatch([]byte(name))
	if matches == nil || len(matches) < 2 {
		return false, 0
	}
	num, _ := strconv.Atoi(string(matches[1]))
	return true, uint32(num)
}

func (v *VMPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	lbName := v.VMProperties.GetRootLBNameForCluster(ctx, clusterInst)
	client, err := v.GetClusterPlatformClient(ctx, clusterInst)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "verify if cloudlet base image exists")
	imgName, err := v.VMProvider.AddCloudletImageIfNotPresent(ctx, v.VMProperties.CommonPf.PlatformConfig.CloudletVMImagePath, v.VMProperties.CommonPf.PlatformConfig.VMImageVersion, updateCallback)
	if err != nil {
		log.InfoLog("error with cloudlet base image", "imgName", imgName, "error", err)
		return err
	}
	return v.updateClusterInternal(ctx, client, lbName, imgName, clusterInst, privacyPolicy, updateCallback)
}

func (v *VMPlatform) updateClusterInternal(ctx context.Context, client ssh.Client, rootLBName, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) (reterr error) {
	updateCallback(edgeproto.UpdateTask, "Updating Cluster Resources")

	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		// if removing nodes, need to tell kubernetes that nodes are
		// going away forever so that tolerating pods can be migrated
		// off immediately.
		kconfName := k8smgmt.GetKconfName(clusterInst)
		cmd := fmt.Sprintf("KUBECONFIG=%s kubectl get nodes --no-headers -o custom-columns=Name:.metadata.name", kconfName)
		out, err := client.Output(cmd)
		if err != nil {
			return err
		}
		allnodes := strings.Split(strings.TrimSpace(out), "\n")
		toRemove := []string{}
		numMaster := uint32(0)
		numNodes := uint32(0)
		for _, n := range allnodes {
			if !strings.HasPrefix(n, cloudcommon.MexNodePrefix) {
				// skip master
				numMaster++
				continue
			}
			ok, num := ParseClusterNodePrefix(n)
			if !ok {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to parse node name, ignoring", "name", n)
				continue
			}
			numNodes++
			// heat will remove the higher-numbered nodes
			if num > clusterInst.NumNodes {
				toRemove = append(toRemove, n)
			}
		}
		if len(toRemove) > 0 {
			log.SpanLog(ctx, log.DebugLevelInfra, "delete nodes", "toRemove", toRemove)
			err = k8smgmt.DeleteNodes(ctx, client, kconfName, toRemove)
			if err != nil {
				return err
			}
		}
		if numMaster == clusterInst.NumMasters && numNodes == clusterInst.NumNodes {
			// nothing changing
			log.SpanLog(ctx, log.DebugLevelInfra, "no change in nodes", "ClusterInst", clusterInst.Key, "nummaster", numMaster, "numnodes", numNodes)
			return nil
		}
	}
	_, err := v.CreateOrUpdateVMsForCluster(ctx, imgName, clusterInst, privacyPolicy, ActionUpdate, updateCallback)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Waiting for Cluster to Update")
	//todo: calculate timeouts instead of hardcoded value

	return v.waitClusterReady(ctx, clusterInst, rootLBName, updateCallback, time.Minute*15)

}

//DeleteCluster deletes kubernetes cluster
func (v *VMPlatform) deleteCluster(ctx context.Context, rootLBName string, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting kubernetes cluster", "clusterInst", clusterInst)
	name := GetClusterName(ctx, clusterInst)

	dedicatedRootLB := clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED
	client, err := v.GetClusterPlatformClient(ctx, clusterInst)
	if err != nil {
		if strings.Contains(err.Error(), ServerDoesNotExistError) {
			log.SpanLog(ctx, log.DebugLevelInfra, "Dedicated RootLB is gone, allow stack delete to proceed")
		} else {
			return err
		}
	}
	if !dedicatedRootLB {
		clusterSnName := GetClusterSubnetName(ctx, clusterInst)
		ip, err := v.GetIPFromServerName(ctx, clusterSnName, clusterSnName, rootLBName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to get ips from server, proceed with VM deletion", "err", err)
		} else {
			detachPort := v.VMProvider.GetInternalPortPolicy() == AttachPortAfterCreate
			err = v.DetachAndDisableRootLBInterface(ctx, client, rootLBName, detachPort, clusterSnName, GetPortName(rootLBName, clusterSnName), ip.InternalAddr)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to detach rootLB interface, proceed with VM deletion", "err", err)
			}
		}
	}
	err = v.VMProvider.DeleteVMs(ctx, name)
	if err != nil {
		return err
	}

	if dedicatedRootLB {
		proxy.RemoveDedicatedCluster(ctx, clusterInst.Key.ClusterKey.Name)
		DeleteRootLB(rootLBName)
	}
	return nil
}

func (v *VMPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	lbName := v.VMProperties.GetRootLBNameForCluster(ctx, clusterInst)
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateClusterInst", "clusterInst", clusterInst, "lbName", lbName)

	//find the flavor and check the disk size
	for _, flavor := range v.FlavorList {
		if flavor.Name == clusterInst.NodeFlavor && flavor.Disk < MINIMUM_DISK_SIZE && clusterInst.ExternalVolumeSize < MINIMUM_DISK_SIZE {
			log.SpanLog(ctx, log.DebugLevelInfra, "flavor disk size too small", "flavor", flavor, "ExternalVolumeSize", clusterInst.ExternalVolumeSize)
			return fmt.Errorf("Insufficient disk size, please specify a flavor with at least %dgb", MINIMUM_DISK_SIZE)
		}
	}

	//adjust the timeout just a bit to give some buffer for the API exchange and also sleep loops
	timeout -= time.Minute

	log.SpanLog(ctx, log.DebugLevelInfra, "verify if cloudlet base image exists")
	imgName, err := v.VMProvider.AddCloudletImageIfNotPresent(ctx, v.VMProperties.CommonPf.PlatformConfig.CloudletVMImagePath, v.VMProperties.CommonPf.PlatformConfig.VMImageVersion, updateCallback)
	if err != nil {
		log.InfoLog("error with cloudlet base image", "imgName", imgName, "error", err)
		return err
	}
	return v.createClusterInternal(ctx, lbName, imgName, clusterInst, privacyPolicy, updateCallback, timeout)
}

func (v *VMPlatform) createClusterInternal(ctx context.Context, rootLBName string, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) (reterr error) {
	// clean-up func
	defer func() {
		if reterr == nil {
			return
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "error in CreateCluster", "err", reterr)
		if v.VMProperties.CommonPf.GetCleanupOnFailure(ctx) {
			log.SpanLog(ctx, log.DebugLevelInfra, "cleaning up cluster resources after cluster fail, set envvar CLEANUP_ON_FAILURE to 'no' to avoid this")
			delerr := v.deleteCluster(ctx, rootLBName, clusterInst)
			if delerr != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "fail to cleanup cluster")
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "skipping cleanup on failure")
		}
	}()

	start := time.Now()
	log.SpanLog(ctx, log.DebugLevelInfra, "creating cluster instance", "clusterInst", clusterInst, "timeout", timeout)

	var err error
	if clusterInst.AvailabilityZone == "" {
		//use the cloudlet default AZ if it exists
		clusterInst.AvailabilityZone = v.VMProperties.GetCloudletComputeAvailabilityZone()
	}
	vmgp, err := v.CreateOrUpdateVMsForCluster(ctx, imgName, clusterInst, privacyPolicy, ActionCreate, updateCallback)
	if err != nil {
		return fmt.Errorf("Cluster VM create Failed: %v", err)
	}

	client, err := v.GetClusterPlatformClient(ctx, clusterInst)
	if err != nil {
		return fmt.Errorf("can't get rootLB client, %v", err)
	}

	if v.VMProperties.GetCloudletExternalRouter() == NoExternalRouter && clusterInst.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		log.SpanLog(ctx, log.DebugLevelInfra, "Need to attach internal interface on rootlb", "IpAccess", clusterInst.IpAccess)

		// after vm creation, the orchestrator will update some fields in the group params including gateway IP.
		// this IP is used on the rootLB to server as the GW for this new subnet
		subnetName := GetClusterSubnetName(ctx, clusterInst)
		gw, err := v.GetSubnetGatewayFromVMGroupParms(ctx, subnetName, vmgp)
		if err != nil {
			return err
		}

		attachPort := true
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED && v.VMProvider.GetInternalPortPolicy() == AttachPortDuringCreate {
			attachPort = false
		}
		err = v.AttachAndEnableRootLBInterface(ctx, client, rootLBName, attachPort, subnetName, GetPortName(rootLBName, subnetName), gw)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "AttachAndEnableRootLBInterface failed", "err", err)
			return err
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "External router in use, no internal interface for rootlb")
	}

	// the root LB was created as part of cluster creation, but it needs to be prepped and
	// mex agent started
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		log.SpanLog(ctx, log.DebugLevelInfra, "new dedicated rootLB", "IpAccess", clusterInst.IpAccess)
		_, err := v.NewRootLB(ctx, rootLBName)
		if err != nil {
			// likely already exists which means something went really wrong
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Setting Up Root LB")
		err = v.SetupRootLB(ctx, rootLBName, &clusterInst.Key.CloudletKey, updateCallback)
		if err != nil {
			return err
		}
	}

	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		elapsed := time.Since(start)
		// subtract elapsed time from total time to get remaining time
		timeout -= elapsed
		updateCallback(edgeproto.UpdateTask, "Waiting for Cluster to Initialize")
		err := v.waitClusterReady(ctx, clusterInst, rootLBName, updateCallback, timeout)
		if err != nil {
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Creating config map")
		if err := infracommon.CreateClusterConfigMap(ctx, client, clusterInst); err != nil {
			return err
		}
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		proxy.NewDedicatedCluster(ctx, clusterInst.Key.ClusterKey.Name, client)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created kubernetes cluster")
	return nil
}

func (v *VMPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	lbName := v.VMProperties.GetRootLBNameForCluster(ctx, clusterInst)
	return v.deleteCluster(ctx, lbName, clusterInst)
}

func (v *VMPlatform) waitClusterReady(ctx context.Context, clusterInst *edgeproto.ClusterInst, rootLBName string, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	start := time.Now()
	masterName := ""
	masterIP := ""
	var currReadyCount uint32
	log.SpanLog(ctx, log.DebugLevelInfra, "waitClusterReady", "cluster", clusterInst.Key, "timeout", timeout)

	for {
		if masterIP == "" {
			mip, err := v.GetIPFromServerName(ctx, v.VMProperties.GetCloudletMexNetwork(), GetClusterSubnetName(ctx, clusterInst), GetClusterMasterName(ctx, clusterInst))
			if err == nil {
				masterIP = mip.ExternalAddr
				updateCallback(edgeproto.UpdateStep, "Checking Master for Available Nodes")
			}
		}
		if masterIP == "" {
			log.SpanLog(ctx, log.DebugLevelInfra, "master IP not available yet")
		} else {
			ready, readyCount, err := v.isClusterReady(ctx, clusterInst, masterName, masterIP, rootLBName, updateCallback)
			if readyCount != currReadyCount {
				numNodes := readyCount - 1
				updateCallback(edgeproto.UpdateStep, fmt.Sprintf("%d of %d nodes active", numNodes, clusterInst.NumNodes))
			}
			currReadyCount = readyCount
			if err != nil {
				return err
			}
			if ready {
				log.SpanLog(ctx, log.DebugLevelInfra, "kubernetes cluster ready")
				return nil
			}
			if time.Since(start) > timeout {
				return fmt.Errorf("cluster not ready (yet)")
			}
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "waiting for kubernetes cluster to be ready...")
		time.Sleep(30 * time.Second)
	}
}

//IsClusterReady checks to see if cluster is read, i.e. rootLB is running and active.  returns ready,nodecount, error
func (v *VMPlatform) isClusterReady(ctx context.Context, clusterInst *edgeproto.ClusterInst, masterName, masterIP string, rootLBName string, updateCallback edgeproto.CacheUpdateCallback) (bool, uint32, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "checking if cluster is ready")

	// some commands are run on the rootlb and some on the master directly, so we use separate clients
	rootLBClient, err := v.GetClusterPlatformClient(ctx, clusterInst)
	if err != nil {
		return false, 0, fmt.Errorf("can't get rootlb ssh client for cluster ready check, %v", err)
	}
	// masterClient is to run commands on the master
	masterClient, err := rootLBClient.AddHop(masterIP, 22)
	if err != nil {
		return false, 0, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "checking master k8s node for available nodes", "ipaddr", masterIP)
	cmd := "kubectl get nodes"
	out, err := masterClient.Output(cmd)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error checking for kubernetes nodes", "out", out, "err", err)
		return false, 0, nil //This is intentional
	}
	//                   node       state               role     age     version
	nodeMatchPattern := "(\\S+)\\s+(Ready|NotReady)\\s+(\\S+)\\s+\\S+\\s+\\S+"
	reg, err := regexp.Compile(nodeMatchPattern)
	if err != nil {
		// this is a bug if the regex does not compile
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to compile regex", "pattern", nodeMatchPattern)
		return false, 0, fmt.Errorf("Internal Error compiling regex for k8s node")
	}
	masterString := ""
	lines := strings.Split(out, "\n")
	var readyCount uint32
	var notReadyCount uint32
	for _, l := range lines {
		if reg.MatchString(l) {
			matches := reg.FindStringSubmatch(l)
			nodename := matches[1]
			state := matches[2]
			role := matches[3]

			if role == "master" {
				masterString = nodename
			}
			if state == "Ready" {
				readyCount++
			} else {
				notReadyCount++
			}
		}
	}
	if readyCount < (clusterInst.NumNodes + clusterInst.NumMasters) {
		log.SpanLog(ctx, log.DebugLevelInfra, "kubernetes cluster not ready", "readyCount", readyCount, "notReadyCount", notReadyCount)
		return false, 0, nil
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "cluster nodes ready", "numnodes", clusterInst.NumNodes, "nummasters", clusterInst.NumMasters, "readyCount", readyCount, "notReadyCount", notReadyCount)

	if err := infracommon.CopyKubeConfig(ctx, rootLBClient, clusterInst, rootLBName, masterIP); err != nil {
		return false, 0, fmt.Errorf("kubeconfig copy failed, %v", err)
	}
	if clusterInst.NumNodes == 0 {
		// k8s nodes are limited to MaxK8sNodeNameLen chars
		//remove the taint from the master if there are no nodes. This has potential side effects if the cluster
		// becomes very busy but is useful for testing and PoC type clusters.
		// TODO: if the cluster is subsequently increased in size do we need to add the taint?
		//For now leaving that alone since an increased cluster size means we needed more capacity.
		log.SpanLog(ctx, log.DebugLevelInfra, "removing NoSchedule taint from master", "master", masterString)
		cmd := fmt.Sprintf("kubectl taint nodes %s node-role.kubernetes.io/master:NoSchedule-", masterString)

		out, err := masterClient.Output(cmd)
		if err != nil {
			if strings.Contains(out, "not found") {
				log.SpanLog(ctx, log.DebugLevelInfra, "master taint already gone")
			} else {
				log.InfoLog("error removing master taint", "out", out, "err", err)
				return false, 0, fmt.Errorf("Cannot remove NoSchedule taint from master, %v", err)
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "cluster ready.")
	return true, readyCount, nil
}

func (v *VMPlatform) getVMRequestSpecForDockerCluster(ctx context.Context, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, action ActionType, updateCallback edgeproto.CacheUpdateCallback) ([]*VMRequestSpec, string, string, error) {
	newSubnet := ""
	var vms []*VMRequestSpec
	dockerVmConnectExternal := false
	var dockerVmType VMType
	var dockerVMName string
	var newSubnetName string
	var newSecgrpName string

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		// dedicated access means the docker VM acts as its own rootLB
		dockerVmConnectExternal = true
		dockerVmType = VMTypeRootLB
		dockerVMName = v.VMProperties.GetRootLBNameForCluster(ctx, clusterInst)
		newSecgrpName = v.GetServerSecurityGroupName(dockerVMName)
	} else {
		// shared access means docker vm goes on its own subnet which is connected
		// via shared rootlb
		newSubnetName = GetClusterSubnetName(ctx, clusterInst)
		dockerVmType = VMTypeClusterNode
		dockerVMName = v.GetDockerNodeName(ctx, clusterInst)

		if v.VMProperties.GetCloudletExternalRouter() == NoExternalRouter {
			// If no router in use, create ports on the existing shared rootLB
			rootlb, err := v.GetVMSpecForRootLBPorts(ctx, v.VMProperties.sharedRootLBName, newSubnet)
			if err != nil {
				return vms, newSubnetName, newSecgrpName, err
			}
			vms = append(vms, rootlb)
		}
	}
	dockervm, err := v.GetVMRequestSpec(
		ctx,
		dockerVmType,
		dockerVMName,
		clusterInst.NodeFlavor,
		imgName,
		dockerVmConnectExternal,
		WithExternalVolume(clusterInst.ExternalVolumeSize),
		WithSubnetConnection(newSubnetName),
	)
	if err != nil {
		return vms, newSubnetName, newSecgrpName, err
	}
	vms = append(vms, dockervm)
	return vms, newSubnetName, newSecgrpName, nil
}

func (v *VMPlatform) CreateOrUpdateVMsForCluster(ctx context.Context, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, action ActionType, updateCallback edgeproto.CacheUpdateCallback) (*VMGroupOrchestrationParams, error) {
	log.SpanLog(ctx, log.DebugLevelInfo, "CreateVMsForCluster", "clusterInst", clusterInst)

	var vms []*VMRequestSpec
	var err error
	vmgroupName := GetClusterName(ctx, clusterInst)
	var newSubnetName string
	var newSecgrpName string

	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		vms, newSubnetName, newSecgrpName, err = v.getVMRequestSpecForDockerCluster(ctx, imgName, clusterInst, privacyPolicy, action, updateCallback)
		if err != nil {
			return nil, err
		}
	} else {
		newSubnetName = GetClusterSubnetName(ctx, clusterInst)
		var rootlb *VMRequestSpec
		var err error
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			// dedicated for docker means the docker VM acts as its own rootLB
			rootlb, err = v.GetVMSpecForRootLB(ctx, v.VMProperties.GetRootLBNameForCluster(ctx, clusterInst), newSubnetName, updateCallback)
			if err != nil {
				return nil, err
			}
			vms = append(vms, rootlb)
			newSecgrpName = v.GetServerSecurityGroupName(rootlb.Name)
		} else if v.VMProperties.GetCloudletExternalRouter() == NoExternalRouter {
			// If no router in use, create ports on the existing shared rootLB
			rootlb, err = v.GetVMSpecForRootLBPorts(ctx, v.VMProperties.sharedRootLBName, newSubnetName)
			if err != nil {
				return nil, err
			}
			vms = append(vms, rootlb)
			// docker goes into a new subnet, the rootlb will be connected to it later
			newSubnetName = GetClusterSubnetName(ctx, clusterInst)
		}

		masterFlavor := clusterInst.MasterNodeFlavor
		if masterFlavor == "" {
			masterFlavor = clusterInst.NodeFlavor
		}
		master, err := v.GetVMRequestSpec(ctx,
			VMTypeClusterMaster,
			GetClusterMasterName(ctx, clusterInst),
			masterFlavor,
			v.VMProperties.GetCloudletOSImage(),
			false, //connect external
			WithSharedVolume(clusterInst.SharedVolumeSize),
			WithExternalVolume(clusterInst.ExternalVolumeSize),
			WithSubnetConnection(newSubnetName),
		)
		if err != nil {
			return nil, err
		}
		vms = append(vms, master)

		for nn := uint32(1); nn <= clusterInst.NumNodes; nn++ {
			node, err := v.GetVMRequestSpec(ctx,
				VMTypeClusterNode,
				GetClusterNodeName(ctx, clusterInst, nn),
				clusterInst.NodeFlavor,
				v.VMProperties.GetCloudletOSImage(),
				false, //connect external
				WithExternalVolume(clusterInst.ExternalVolumeSize),
				WithSubnetConnection(newSubnetName),
			)
			if err != nil {
				return nil, err
			}
			vms = append(vms, node)
		}
	}

	//	return v.GetVMGroupOrchestrationParamsFromVMSpec(ctx, name, vms, WithNewSubnet(subnetname))
	if action == ActionCreate {
		return v.CreateVMsFromVMSpec(ctx, vmgroupName, vms, updateCallback, WithNewSubnet(newSubnetName), WithPrivacyPolicy(privacyPolicy), WithNewSecurityGroup(newSecgrpName))
	} else if action == ActionUpdate {
		return v.UpdateVMsFromVMSpec(ctx, vmgroupName, vms, updateCallback, WithNewSubnet(newSubnetName), WithPrivacyPolicy(privacyPolicy), WithNewSecurityGroup(newSecgrpName))
	} else {
		return nil, fmt.Errorf("unexpected action: %s", action)
	}
}

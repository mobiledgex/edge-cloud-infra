package infracommon

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"
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
func (c *CommonPlatform) GetRootLBNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	lbName := c.SharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(c.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return lbName
}

func (c *CommonPlatform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	lbName := c.GetRootLBNameForCluster(ctx, clusterInst)
	client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "verify if cloudlet base image exists")
	imgName, err := c.infraProvider.AddCloudletImageIfNotPresent(ctx, c.PlatformConfig.CloudletVMImagePath, c.PlatformConfig.VMImageVersion, updateCallback)
	if err != nil {
		log.InfoLog("error with cloudlet base image", "imgName", imgName, "error", err)
		return err
	}
	return c.updateClusterInternal(ctx, client, lbName, imgName, clusterInst, privacyPolicy, updateCallback)
}

func (c *CommonPlatform) updateClusterInternal(ctx context.Context, client ssh.Client, rootLBName, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) (reterr error) {
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
				log.SpanLog(ctx, log.DebugLevelMexos, "unable to parse node name, ignoring", "name", n)
				continue
			}
			numNodes++
			// heat will remove the higher-numbered nodes
			if num > clusterInst.NumNodes {
				toRemove = append(toRemove, n)
			}
		}
		if len(toRemove) > 0 {
			log.SpanLog(ctx, log.DebugLevelMexos, "delete nodes", "toRemove", toRemove)
			err = k8smgmt.DeleteNodes(ctx, client, kconfName, toRemove)
			if err != nil {
				return err
			}
		}
		if numMaster == clusterInst.NumMasters && numNodes == clusterInst.NumNodes {
			// nothing changing
			log.SpanLog(ctx, log.DebugLevelMexos, "no change in nodes", "ClusterInst", clusterInst.Key, "nummaster", numMaster, "numnodes", numNodes)
			return nil
		}
	}

	dedicatedRootLB := clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED

	err := c.infraProvider.UpdateClusterVMs(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Waiting for Cluster to Update")
	//todo: calculate timeouts instead of hardcoded value
	return c.waitClusterReady(ctx, clusterInst, rootLBName, updateCallback, time.Minute*15)
}

//DeleteCluster deletes kubernetes cluster
func (c *CommonPlatform) deleteCluster(ctx context.Context, rootLBName string, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "deleting kubernetes cluster", "clusterInst", clusterInst)

	dedicatedRootLB := clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED
	client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		if strings.Contains(err.Error(), "No server with a name or ID") {
			log.SpanLog(ctx, log.DebugLevelMexos, "Dedicated RootLB is gone, allow stack delete to proceed")
		} else {
			return err
		}
	}
	err = c.infraProvider.DeleteClusterResources(ctx, client, clusterInst)
	if err != nil {
		return err
	}
	if dedicatedRootLB {
		proxy.RemoveDedicatedCluster(ctx, clusterInst.Key.ClusterKey.Name)
		DeleteRootLB(rootLBName)
	}
	return nil
}

func (c *CommonPlatform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	lbName := c.GetRootLBNameForCluster(ctx, clusterInst)
	log.SpanLog(ctx, log.DebugLevelMexos, "OpenStack CreateClusterInst", "clusterInst", clusterInst, "lbName", lbName)

	//find the flavor and check the disk size
	for _, flavor := range c.FlavorList {
		if flavor.Name == clusterInst.NodeFlavor && flavor.Disk < MINIMUM_DISK_SIZE && clusterInst.ExternalVolumeSize < MINIMUM_DISK_SIZE {
			log.SpanLog(ctx, log.DebugLevelMexos, "flavor disk size too small", "flavor", flavor, "ExternalVolumeSize", clusterInst.ExternalVolumeSize)
			return fmt.Errorf("Insufficient disk size, please specify a flavor with at least %dgb", MINIMUM_DISK_SIZE)
		}
	}

	//adjust the timeout just a bit to give some buffer for the API exchange and also sleep loops
	timeout -= time.Minute

	log.SpanLog(ctx, log.DebugLevelMexos, "verify if cloudlet base image exists")
	imgName, err := c.infraProvider.AddCloudletImageIfNotPresent(ctx, c.PlatformConfig.CloudletVMImagePath, c.PlatformConfig.VMImageVersion, updateCallback)
	if err != nil {
		log.InfoLog("error with cloudlet base image", "imgName", imgName, "error", err)
		return err
	}
	return c.createClusterInternal(ctx, lbName, imgName, clusterInst, privacyPolicy, updateCallback, timeout)
}

func (c *CommonPlatform) createClusterInternal(ctx context.Context, rootLBName string, imgName string, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) (reterr error) {
	// clean-up func
	defer func() {
		if reterr == nil {
			return
		}

		log.SpanLog(ctx, log.DebugLevelMexos, "error in CreateCluster", "err", reterr)
		if c.GetCleanupOnFailure(ctx) {
			log.SpanLog(ctx, log.DebugLevelMexos, "cleaning up cluster resources after cluster fail, set envvar CLEANUP_ON_FAILURE to 'no' to avoid this")
			delerr := c.deleteCluster(ctx, rootLBName, clusterInst, updateCallback)
			if delerr != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "fail to cleanup cluster")
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "skipping cleanup on failure")
		}
	}()

	start := time.Now()
	log.SpanLog(ctx, log.DebugLevelMexos, "creating cluster instance", "clusterInst", clusterInst, "timeout", timeout)

	dedicatedRootLB := false
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		dedicatedRootLB = true
	}

	var err error
	if clusterInst.AvailabilityZone == "" {
		//use the cloudlet default AZ if it exists
		clusterInst.AvailabilityZone = c.GetCloudletComputeAvailabilityZone()
	}

	vmspec := vmspec.VMCreationSpec{
		FlavorName:         clusterInst.NodeFlavor,
		ExternalVolumeSize: clusterInst.ExternalVolumeSize,
		AvailabilityZone:   clusterInst.AvailabilityZone,
		ImageName:          clusterInst.ImageName,
		PrivacyPolicy:      privacyPolicy,
	}

	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		if dedicatedRootLB {
			// in the dedicated case for docker, the RootLB and the docker worker are the same
			updateCallback(edgeproto.UpdateTask, "Creating Dedicated VM for Docker")
			err = c.infraProvider.CreateRootLBVM(ctx, rootLBName, k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key), imgName, &vmspec, &clusterInst.Key.CloudletKey, updateCallback)
		} else {
			updateCallback(edgeproto.UpdateTask, "Creating single-node cluster for docker using shared RootLB")
			err = c.infraProvider.CreateClusterVMs(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
		}
	} else {
		err = c.infraProvider.CreateClusterVMs(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
	}
	if err != nil {
		return err
	}
	// the root LB was created as part of cluster creation, but it needs to be prepped and
	// mex agent started
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		log.SpanLog(ctx, log.DebugLevelMexos, "need dedicated rootLB", "IpAccess", clusterInst.IpAccess)
		_, err := c.NewRootLB(ctx, rootLBName)
		if err != nil {
			// likely already exists which means something went really wrong
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Setting Up Root LB")
		err = c.SetupRootLB(ctx, rootLBName, &vmspec, &clusterInst.Key.CloudletKey, "", "", updateCallback)
		if err != nil {
			return err
		}
	}
	client, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return fmt.Errorf("can't get rootLB client, %v", err)
	}
	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		elapsed := time.Since(start)
		// subtract elapsed time from total time to get remaining time
		timeout -= elapsed
		updateCallback(edgeproto.UpdateTask, "Waiting for Cluster to Initialize")
		err := c.waitClusterReady(ctx, clusterInst, rootLBName, updateCallback, timeout)
		if err != nil {
			return err
		}
		updateCallback(edgeproto.UpdateTask, "Creating config map")
		if err := CreateClusterConfigMap(ctx, client, clusterInst); err != nil {
			return err
		}
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		proxy.NewDedicatedCluster(ctx, clusterInst.Key.ClusterKey.Name, client)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "created kubernetes cluster")
	return nil
}

func (c *CommonPlatform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	lbName := c.GetRootLBNameForCluster(ctx, clusterInst)
	return c.deleteCluster(ctx, lbName, clusterInst, updateCallback)
}

func (c *CommonPlatform) waitClusterReady(ctx context.Context, clusterInst *edgeproto.ClusterInst, rootLBName string, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	start := time.Now()
	masterName := ""
	masterIP := ""
	var currReadyCount uint32
	log.SpanLog(ctx, log.DebugLevelMexos, "waitClusterReady", "cluster", clusterInst.Key, "timeout", timeout)

	for {
		if masterIP == "" {
			masterName, masterIP, _ = c.infraProvider.GetClusterMasterNameAndIP(ctx, clusterInst)
			if masterIP != "" {
				updateCallback(edgeproto.UpdateStep, "Checking Master for Available Nodes")
			}
		}
		if masterIP == "" {
			log.SpanLog(ctx, log.DebugLevelMexos, "master IP not available yet")
		} else {
			ready, readyCount, err := c.isClusterReady(ctx, clusterInst, masterName, masterIP, rootLBName, updateCallback)
			if readyCount != currReadyCount {
				numNodes := readyCount - 1
				updateCallback(edgeproto.UpdateStep, fmt.Sprintf("%d of %d nodes active", numNodes, clusterInst.NumNodes))
			}
			currReadyCount = readyCount
			if err != nil {
				return err
			}
			if ready {
				log.SpanLog(ctx, log.DebugLevelMexos, "kubernetes cluster ready")
				return nil
			}
			if time.Since(start) > timeout {
				return fmt.Errorf("cluster not ready (yet)")
			}
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "waiting for kubernetes cluster to be ready...")
		time.Sleep(30 * time.Second)
	}
}

//IsClusterReady checks to see if cluster is read, i.e. rootLB is running and active.  returns ready,nodecount, error
func (c *CommonPlatform) isClusterReady(ctx context.Context, clusterInst *edgeproto.ClusterInst, masterName, masterIP string, rootLBName string, updateCallback edgeproto.CacheUpdateCallback) (bool, uint32, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "checking if cluster is ready")

	// some commands are run on the rootlb and some on the master directly, so we use separate clients
	rootLBClient, err := c.infraProvider.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return false, 0, fmt.Errorf("can't get rootlb ssh client for cluster ready check, %v", err)
	}
	// masterClient is to run commands on the master
	masterClient, err := rootLBClient.AddHop(masterIP, 22)
	if err != nil {
		return false, 0, err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "checking master k8s node for available nodes", "ipaddr", masterIP)
	cmd := "kubectl get nodes"
	out, err := masterClient.Output(cmd)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "error checking for kubernetes nodes", "out", out, "err", err)
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
		log.SpanLog(ctx, log.DebugLevelMexos, "kubernetes cluster not ready", "readyCount", readyCount, "notReadyCount", notReadyCount)
		return false, 0, nil
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "cluster nodes ready", "numnodes", clusterInst.NumNodes, "nummasters", clusterInst.NumMasters, "readyCount", readyCount, "notReadyCount", notReadyCount)

	if err := CopyKubeConfig(ctx, rootLBClient, clusterInst, rootLBName, masterIP); err != nil {
		return false, 0, fmt.Errorf("kubeconfig copy failed, %v", err)
	}
	if clusterInst.NumNodes == 0 {
		// k8s nodes are limited to MaxK8sNodeNameLen chars
		//remove the taint from the master if there are no nodes. This has potential side effects if the cluster
		// becomes very busy but is useful for testing and PoC type clusters.
		// TODO: if the cluster is subsequently increased in size do we need to add the taint?
		//For now leaving that alone since an increased cluster size means we needed more capacity.
		log.SpanLog(ctx, log.DebugLevelMexos, "removing NoSchedule taint from master", "master", masterString)
		cmd := fmt.Sprintf("kubectl taint nodes %s node-role.kubernetes.io/master:NoSchedule-", masterString)

		out, err := masterClient.Output(cmd)
		if err != nil {
			if strings.Contains(out, "not found") {
				log.SpanLog(ctx, log.DebugLevelMexos, "master taint already gone")
			} else {
				log.InfoLog("error removing master taint", "out", out, "err", err)
				return false, 0, fmt.Errorf("Cannot remove NoSchedule taint from master, %v", err)
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "cluster ready.")
	return true, readyCount, nil
}

// GetCleanupOnFailure should be true unless we want to debug the failure,
// in which case this env var can be set to no.  We could consider making
// this configurable at the controller but really is only needed for debugging.
func (c *CommonPlatform) GetCleanupOnFailure(ctx context.Context) bool {
	cleanup := c.Properties["CLEANUP_ON_FAILURE"].Value
	log.SpanLog(ctx, log.DebugLevelMexos, "GetCleanupOnFailure", "cleanup", cleanup)
	cleanup = strings.ToLower(cleanup)
	cleanup = strings.ReplaceAll(cleanup, "'", "")
	if cleanup == "no" || cleanup == "false" {
		return false
	}
	return true
}

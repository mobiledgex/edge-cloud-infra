package vmlayer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type VMType string

const (
	VMTypeAppVM     VMType = "appvm"
	VMTypeRootLB    VMType = "rootlb"
	VMTypePlatform  VMType = "platform"
	VMTypeK8sMaster VMType = "k8s-master"
	VMTypeK8sNode   VMType = "k8s-node"
)

var ClusterTypeKubernetesMasterLabel = "mex-k8s-master"
var ClusterTypeDockerVMLabel = "mex-docker-vm"

type SkipK8sChoice string

const (
	SkipK8sNo  SkipK8sChoice = "no"
	SkipK8sYes SkipK8sChoice = "yes"
)

type VMRole string

var RoleAgent VMRole = "mex-agent-node"
var RoleMaster VMRole = "k8s-master"
var RoleNode VMRole = "k8s-node"
var RoleUser VMRole = "user"

var MasterIPNone = "NONE"

// NextAvailableResource means the orchestration code needs to find an available
// resource of the given type as the calling code won't know what is free
var NextAvailableResource = "NextAvailable"

// ResourceReference identifies a resource that is referenced by another resource. The
// Preexisting flag indicates whether the resource is already present or is being created
// as part of this operation.  How the resource is referred to during the orchestration process
// may be different for preexisting vs new resources.
type ResourceReference struct {
	Name        string
	Preexisting bool
}

func NewResourceReference(name string, preexisting bool) ResourceReference {
	// we may want to compute an id here
	return ResourceReference{Name: name, Preexisting: preexisting}
}

type SubnetParams struct {
	Name         string
	CIDR         string
	NodeIPPrefix string
	GatewayIP    string
	DNSServers   []string
	DHCPEnabled  string
}

type FixedIPParams struct {
	LastIPOctet uint32
	Address     string
	Subnet      ResourceReference
}

type PortParams struct {
	Name           string
	NetworkName    string
	VnicType       string
	FixedIPs       []FixedIPParams
	SecurityGroups []ResourceReference
}

type FloatingIPParams struct {
	Name         string
	Port         ResourceReference
	FloatingIpId ResourceReference
}

type RouterInterfaceParams struct {
	RouterName string
	RouterPort string
}

type SecurityGroupParams struct {
	Name             string
	AccessPorts      []util.PortSpec
	EgressRestricted bool
	EgressRules      []edgeproto.OutboundSecurityRule
}

type SecgrpParamsOp func(vmp *SecurityGroupParams) error

func secGrpWithEgressRules(rules []edgeproto.OutboundSecurityRule) SecgrpParamsOp {
	return func(sp *SecurityGroupParams) error {
		sp.EgressRules = rules
		sp.EgressRestricted = true
		return nil
	}
}

func secGrpWithAccessPorts(accessPorts string) SecgrpParamsOp {
	return func(sgp *SecurityGroupParams) error {
		if accessPorts == "" {
			return nil
		}
		parsedAccessPorts, err := util.ParsePorts(accessPorts)
		if err != nil {
			return err
		}
		for _, port := range parsedAccessPorts {
			endPort, err := strconv.ParseInt(port.EndPort, 10, 32)
			if err != nil {
				return err
			}
			if endPort == 0 {
				port.EndPort = port.Port
			}
			sgp.AccessPorts = append(sgp.AccessPorts, port)
		}
		return nil
	}
}

func GetSecGrpParams(name string, opts ...SecgrpParamsOp) (*SecurityGroupParams, error) {
	var sgp SecurityGroupParams
	sgp.Name = name
	for _, op := range opts {
		if err := op(&sgp); err != nil {
			return nil, err
		}
	}
	return &sgp, nil
}

type VolumeParams struct {
	Name             string
	ImageName        string
	Size             uint64
	AvailabilityZone string
	DeviceName       string
}
type VolumeParamsOp func(vmp *VolumeParams) error

// VMRequestSpec has the infromation which the caller needs
// to provide when creating a VM.
type VMRequestSpec struct {
	Name                    string
	Type                    VMType
	FlavorName              string
	ImageName               string
	ComputeAvailabilityZone string
	AuthPublicKey           string
	ExternalVolumeSize      uint64
	SharedVolumeSize        uint64
	DeploymentManifest      string
	Command                 string
	ConnectToExternalNet    bool
	ConnectToInternalNet    bool
}

type VMReqOp func(vmp *VMRequestSpec) error

func WithPublicKey(authPublicKey string) VMReqOp {
	return func(vmo *VMRequestSpec) error {
		if authPublicKey == "" {
			return nil
		}
		convKey, err := util.ConvertPEMtoOpenSSH(authPublicKey)
		if err != nil {
			return err
		}
		vmo.AuthPublicKey = convKey
		return nil
	}
}

func WithDeploymentManifest(deploymentManifest string) VMReqOp {
	return func(vrs *VMRequestSpec) error {
		vrs.DeploymentManifest = deploymentManifest
		return nil
	}
}
func WithCommand(command string) VMReqOp {
	return func(vrs *VMRequestSpec) error {
		vrs.Command = command
		return nil
	}
}
func WithComputeAvailabilityZone(zone string) VMReqOp {
	return func(vrs *VMRequestSpec) error {
		vrs.ComputeAvailabilityZone = zone
		return nil
	}
}
func WithExternalVolume(size uint64) VMReqOp {
	return func(s *VMRequestSpec) error {
		s.ExternalVolumeSize = size
		return nil
	}
}
func WithSharedVolume(size uint64) VMReqOp {
	return func(s *VMRequestSpec) error {
		s.SharedVolumeSize = size
		return nil
	}
}

// VMParams contains all details  that are needed by the orchestator
type VMParams struct {
	Name                    string
	Role                    VMRole
	ImageName               string
	FlavorName              string
	ComputeAvailabilityZone string
	UserData                string
	MetaData                string
	SharedVolume            bool
	DNSServers              string
	AuthPublicKey           string
	DeploymentManifest      string
	Command                 string
	Volumes                 []VolumeParams
	Ports                   []ResourceReference
}

// VMGroupParams contains all the details used by the orchestator to create a set of associated VMs
type VMGroupParams struct {
	GroupName        string
	Subnets          []SubnetParams
	Ports            []PortParams
	RouterInterfaces []RouterInterfaceParams
	VMs              []VMParams
	FloatingIPs      []FloatingIPParams
	SecurityGroups   []SecurityGroupParams
	Netspec          *NetSpecInfo
}

var VmCloudConfig = `#cloud-config
bootcmd:
 - echo MOBILEDGEX CLOUD CONFIG START
 - echo 'APT::Periodic::Enable "0";' > /etc/apt/apt.conf.d/10cloudinit-disable
 - apt-get -y purge update-notifier-common ubuntu-release-upgrader-core landscape-common unattended-upgrades
 - echo "Removed APT and Ubuntu extra packages" | systemd-cat
ssh_authorized_keys:
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrHlOJOJUqvd4nEOXQbdL8ODKzWaUxKVY94pF7J3diTxgZ1NTvS6omqOjRS3loiU7TOlQQU4cKnRRnmJW8QQQZSOMIGNrMMInGaEYsdm6+tr1k4DDfoOrkGMj3X/I2zXZ3U+pDPearVFbczCByPU0dqs16TWikxDoCCxJRGeeUl7duzD9a65bI8Jl+zpfQV+I7OPa81P5/fw15lTzT4+F9MhhOUVJ4PFfD+d6/BLnlUfZ94nZlvSYnT+GoZ8xTAstM7+6pvvvHtaHoV4YqRf5CelbWAQ162XNa9/pW5v/RKDrt203/JEk3e70tzx9KAfSw2vuO1QepkCZAdM9rQoCd ubuntu@registry
chpasswd: { expire: False }
ssh_pwauth: False
timezone: UTC
runcmd:
 - echo MOBILEDGEX doing ifconfig
 - ifconfig -a`

// vmCloudConfigShareMount is appended optionally to vmCloudConfig.   It assumes
// the end of vmCloudConfig is runcmd
var VmCloudConfigShareMount = `
- chown nobody:nogroup /share
- chmod 777 /share 
- echo "/share *(rw,sync,no_subtree_check,no_root_squash)" >> /etc/exports
- exportfs -a
- echo "showing exported filesystems"
- exportfs
disk_setup:
  /dev/vdb:
    table_type: 'gpt'
    overwrite: true
    layout: true
fs_setup:
- label: share_fs
 filesystem: 'ext4'
 device: /dev/vdb
 partition: auto
 overwrite: true
mounts:
- [ "/dev/vdb1", "/share" ]`

// VMGroupRequestSpec is used to specify a set of VMs to be created.  It is used as input to create VMGroupParams
type VMGroupRequestSpec struct {
	GroupName     string
	VMs           []*VMRequestSpec
	NewSubnetName string
	AccessPorts   string
	PrivacyPolicy *edgeproto.PrivacyPolicy
}

type VMGroupReqOp func(vmp *VMGroupRequestSpec) error

func WithPrivacyPolicy(pp *edgeproto.PrivacyPolicy) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.PrivacyPolicy = pp
		return nil
	}
}
func WithAccessPorts(ap string) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.AccessPorts = ap
		return nil
	}
}
func WithNewSubnet(sn string) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.NewSubnetName = sn
		return nil
	}
}

func (v *VMPlatform) GetVMRequestSpec(ctx context.Context, vmtype VMType, serverName, flavorName string, imageName string, connectExternal, connectInternal bool, opts ...VMReqOp) (*VMRequestSpec, error) {
	var vrs VMRequestSpec
	for _, op := range opts {
		if err := op(&vrs); err != nil {
			return nil, err
		}
	}
	vrs.Name = serverName
	vrs.FlavorName = flavorName
	vrs.ImageName = imageName
	vrs.ConnectToExternalNet = connectExternal
	vrs.ConnectToInternalNet = connectInternal
	return &vrs, nil
}

func (v *VMPlatform) GetVMGroupRequestSpec(ctx context.Context, name string, vms []*VMRequestSpec, opts ...VMGroupReqOp) (*VMGroupRequestSpec, error) {
	var vmgrs VMGroupRequestSpec
	vmgrs.GroupName = name
	vmgrs.VMs = vms
	for _, op := range opts {
		if err := op(&vmgrs); err != nil {
			return nil, err
		}
	}
	return &vmgrs, nil
}

func (v *VMPlatform) GetVMGroupParamsFromVMSpec(ctx context.Context, name string, vms []*VMRequestSpec, opts ...VMGroupReqOp) (*VMGroupParams, error) {
	vmgp, err := v.GetVMGroupRequestSpec(ctx, name, vms, opts...)
	if err != nil {
		return nil, err
	}
	return v.GetVMGroupParamsFromGroupSpec(ctx, vmgp)
}

func (v *VMPlatform) GetVMGroupParamsFromGroupSpec(ctx context.Context, spec *VMGroupRequestSpec) (*VMGroupParams, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetVMGroupParams", "spec", spec)

	vmgp := VMGroupParams{GroupName: spec.GroupName}
	internalNetName := v.GetCloudletMexNetwork()
	externalNetName := v.GetCloudletExternalNetwork()

	// DNS is applied either at the subnet or VM level
	cloudflareDns := []string{"1.1.1.1", "1.0.0.1"}
	vmDns := ""
	subnetDns := []string{}
	cloudletSecGrpID, err := v.vmProvider.GetResourceID(ctx, ResourceTypeSecurityGroup, v.GetCloudletSecurityGroupName())
	internalSecgrpID := ""
	if err != nil {
		return nil, err
	}
	if v.GetSubnetDNS() == NoSubnetDNS {
		// Contrail workaround, see EDGECLOUD-2420 for details
		vmDns = strings.Join(cloudflareDns, " ")
	} else {
		subnetDns = cloudflareDns
	}

	vmgp.Netspec, err = ParseNetSpec(ctx, v.GetCloudletNetworkScheme())
	if err != nil {
		return nil, err
	}

	rtrInUse := false
	rtr := v.GetCloudletExternalRouter()
	if rtr == NoConfigExternalRouter {
		log.SpanLog(ctx, log.DebugLevelInfra, "NoConfigExternalRouter in use")
	} else if rtr == NoExternalRouter {
		log.SpanLog(ctx, log.DebugLevelInfra, "NoExternalRouter in use ")
	} else {
		log.SpanLog(ctx, log.DebugLevelMexos, "External router in use")
		internalSecgrpID = cloudletSecGrpID
		rtrInUse = true
		return nil, fmt.Errorf("TODO: Router interface not yet implemented")
		//  next need to create router interfaces
	}

	newSecGrpName := spec.GroupName + "-sg"
	externalSecGrp, err := GetSecGrpParams(newSecGrpName, secGrpWithAccessPorts(spec.AccessPorts), secGrpWithEgressRules(spec.PrivacyPolicy.OutboundSecurityRules))
	if err != nil {
		return nil, err
	}
	vmgp.SecurityGroups = append(vmgp.SecurityGroups, *externalSecGrp)

	internalSubnet := ""
	if err != nil {
		return nil, err
	}
	if spec.NewSubnetName != "" {
		newSubnet := SubnetParams{
			Name:        spec.NewSubnetName,
			CIDR:        NextAvailableResource,
			DHCPEnabled: "no",
			DNSServers:  subnetDns,
		}
		vmgp.Subnets = append(vmgp.Subnets, newSubnet)
	}

	var internalPortNextOctet uint32 = 101
	for _, vm := range spec.VMs {
		log.SpanLog(ctx, log.DebugLevelInfra, "Defining VM", "vm", vm)

		var role VMRole
		var newPorts []PortParams
		internalPortName := fmt.Sprintf("%s-%s-port", vm.Name, internalNetName)
		externalPortName := fmt.Sprintf("%s-port", vm.Name, externalNetName)

		switch vm.Type {
		case VMTypePlatform:
			fallthrough
		case VMTypeRootLB:
			role = RoleAgent
			// if the router is used we don't create an internal port for rootlb
			if vm.ConnectToInternalNet && !rtrInUse {
				// no router means rootlb must be connected to other VMs directly
				internalPort := PortParams{
					Name:        internalPortName,
					NetworkName: internalNetName,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs:    []FixedIPParams{{Address: NextAvailableResource, LastIPOctet: 1, Subnet: NewResourceReference(internalSubnet, false)}},
				}
				newPorts = append(newPorts, internalPort)
			}

		case VMTypeAppVM:
			role = RoleUser
			if vm.ConnectToInternalNet {
				// connect via internal network to LB
				internalPort := PortParams{
					Name:        internalPortName,
					NetworkName: internalNetName,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs:    []FixedIPParams{{Address: NextAvailableResource, LastIPOctet: internalPortNextOctet, Subnet: NewResourceReference(internalSubnet, false)}},
				}
				internalPortNextOctet++
				newPorts = append(newPorts, internalPort)
			}

		case VMTypeK8sMaster:
			role = RoleMaster
			if vm.ConnectToInternalNet {
				// connect via internal network to LB
				internalPort := PortParams{
					Name:        internalPortName,
					NetworkName: internalNetName,
					FixedIPs:    []FixedIPParams{{Address: NextAvailableResource, LastIPOctet: 10, Subnet: NewResourceReference(internalSubnet, false)}},
				}
				newPorts = append(newPorts, internalPort)
			} else {
				fmt.Errorf("k8s master not specified to be connected to internal network")
			}
		case VMTypeK8sNode:
			role = RoleNode
			if vm.ConnectToInternalNet {
				// connect via internal network to LB
				internalPort := PortParams{
					Name:        internalPortName,
					NetworkName: internalNetName,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs:    []FixedIPParams{{Address: NextAvailableResource, LastIPOctet: internalPortNextOctet, Subnet: NewResourceReference(internalSubnet, false)}},
				}
				internalPortNextOctet++
				newPorts = append(newPorts, internalPort)
			} else {
				fmt.Errorf("k8s node not specified to be connected to internal network")
			}
		}
		// ports contains only internal ports at this point. Optionally add the internal
		// security group which is used when we have a router
		if internalSecgrpID != "" {
			for i, _ := range newPorts {
				sec := NewResourceReference(internalSecgrpID, false)
				newPorts[i].SecurityGroups = append(newPorts[i].SecurityGroups, sec)
			}
		}

		if vm.ConnectToExternalNet {
			var externalport PortParams
			if vmgp.Netspec.FloatingIPNet != "" {
				externalport = PortParams{
					Name:        externalPortName,
					NetworkName: vmgp.Netspec.FloatingIPNet,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs:    []FixedIPParams{{Subnet: NewResourceReference(vmgp.Netspec.FloatingIPSubnet, false)}},
				}
			} else {
				externalport = PortParams{
					Name:        externalPortName,
					NetworkName: externalNetName,
					VnicType:    vmgp.Netspec.VnicType,
				}
				externalport.SecurityGroups = []ResourceReference{
					NewResourceReference(newSecGrpName, false),
					NewResourceReference(cloudletSecGrpID, true),
				}
				newPorts = append(newPorts, externalport)
			}
		}
		newVM := VMParams{
			Name:                    vm.Name,
			Role:                    role,
			DNSServers:              vmDns,
			ImageName:               vm.ImageName,
			FlavorName:              vm.FlavorName,
			DeploymentManifest:      vm.DeploymentManifest,
			Command:                 vm.Command,
			ComputeAvailabilityZone: vm.ComputeAvailabilityZone,
		}
		if vm.ExternalVolumeSize > 0 {
			externalVolume := VolumeParams{
				Name:       vm.Name + "-volume",
				Size:       vm.ExternalVolumeSize,
				DeviceName: "vda",
			}
			newVM.Volumes = append(newVM.Volumes, externalVolume)
		}
		if vm.SharedVolumeSize > 0 {
			sharedVolume := VolumeParams{
				Name:       vm.Name + "-shared-volume",
				Size:       vm.ExternalVolumeSize,
				DeviceName: "vdb",
			}
			newVM.Volumes = append(newVM.Volumes, sharedVolume)
			newVM.SharedVolume = true
		}
		for _, p := range newPorts {
			newVM.Ports = append(newVM.Ports, NewResourceReference(p.Name, false))
		}
		vmgp.VMs = append(vmgp.VMs, newVM)
		vmgp.Ports = append(vmgp.Ports, newPorts...)
	}

	return &vmgp, nil
}

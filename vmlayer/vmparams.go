package vmlayer

//
// This file contains the functionality needed to input data into the VMProvider orchestrator.   There are 2 categories of structs:
// 1) Request Specs.  These contain high level info used by client code to request the creation of VMs and Groups of VMs
// 2) Orchestration Params.   These contain detailed level info used by the orchestrator to instantiate all the resources related to creating VMs,
//    including Subnets, Ports, Security Groups, etc.  Orchestration Params are derived by code here from Request Specs

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/chefmgmt"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type VMType string

const (
	VMTypeAppVM                 VMType = "appvm"
	VMTypeRootLB                VMType = "rootlb"
	VMTypePlatform              VMType = "platform"
	VMTypePlatformClusterMaster VMType = "platform-cluster-master"
	VMTypePlatformClusterNode   VMType = "platform-cluster-node"
	VMTypeClusterMaster         VMType = "cluster-master"
	VMTypeClusterNode           VMType = "cluster-node"
)

type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
	ActionSync   ActionType = "sync"
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
var RoleVMApplication VMRole = "vmapp"

// NextAvailableResource means the orchestration code needs to find an available
// resource of the given type as the calling code won't know what is free
var NextAvailableResource = "NextAvailable"

// ResourceReference identifies a resource that is referenced by another resource. The
// Preexisting flag indicates whether the resource is already present or is being created
// as part of this operation.  How the resource is referred to during the orchestration process
// may be different for preexisting vs new resources.
type ResourceReference struct {
	Name        string
	Id          string
	Preexisting bool
}

// PortResourceReference needs also a network id
type PortResourceReference struct {
	Name        string
	Id          string
	NetworkId   string
	SubnetId    string
	Preexisting bool
}

func GetPortName(vmname, netname string) string {
	return fmt.Sprintf("%s-%s-port", vmname, netname)
}

func NewResourceReference(name string, id string, preexisting bool) ResourceReference {
	return ResourceReference{Name: name, Id: id, Preexisting: preexisting}
}

func NewPortResourceReference(name string, id string, netid, subnetid string, preexisting bool) PortResourceReference {
	return PortResourceReference{Name: name, Id: id, NetworkId: netid, SubnetId: subnetid, Preexisting: preexisting}
}

// VMRequestSpec has the infromation which the caller needs to provide when creating a VM.
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
	CreatePortsOnly         bool
	ConnectToSubnet         string
	ChefParams              *chefmgmt.VMChefParams
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
func WithSubnetConnection(subnetName string) VMReqOp {
	return func(s *VMRequestSpec) error {
		s.ConnectToSubnet = subnetName
		return nil
	}
}
func WithCreatePortsOnly(portsonly bool) VMReqOp {
	return func(s *VMRequestSpec) error {
		s.CreatePortsOnly = portsonly
		return nil
	}
}
func WithChefParams(chefParams *chefmgmt.VMChefParams) VMReqOp {
	return func(s *VMRequestSpec) error {
		s.ChefParams = chefParams
		return nil
	}
}

// VMGroupRequestSpec is used to specify a set of VMs to be created.  It is used as input to create VMGroupOrchestrationParams
type VMGroupRequestSpec struct {
	GroupName              string
	VMs                    []*VMRequestSpec
	NewSubnetName          string
	NewSecgrpName          string
	AccessPorts            string
	PrivacyPolicy          *edgeproto.PrivacyPolicy
	SkipDefaultSecGrp      bool
	SkipSubnetGateway      bool
	SkipInfraSpecificCheck bool
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
func WithNewSecurityGroup(sg string) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.NewSecgrpName = sg
		return nil
	}
}
func WithSkipDefaultSecGrp(skip bool) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.SkipDefaultSecGrp = skip
		return nil
	}
}
func WithSkipSubnetGateway(skip bool) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.SkipSubnetGateway = skip
		return nil
	}
}
func WithSkipInfraSpecificCheck(skip bool) VMGroupReqOp {
	return func(s *VMGroupRequestSpec) error {
		s.SkipInfraSpecificCheck = skip
		return nil
	}
}

type SubnetOrchestrationParams struct {
	Id           string
	Name         string
	CIDR         string
	NodeIPPrefix string
	GatewayIP    string
	DNSServers   []string
	DHCPEnabled  string
	Vlan         uint32
	SkipGateway  bool
}

type FixedIPOrchestrationParams struct {
	LastIPOctet uint32
	Address     string
	Mask        string
	Subnet      ResourceReference
}

type PortOrchestrationParams struct {
	Name           string
	Id             string
	SubnetId       string
	NetworkName    string
	NetworkId      string
	VnicType       string
	SkipAttachVM   bool
	FixedIPs       []FixedIPOrchestrationParams
	SecurityGroups []ResourceReference
}

type FloatingIPOrchestrationParams struct {
	Name         string
	Port         ResourceReference
	FloatingIpId ResourceReference
}

type RouterInterfaceOrchestrationParams struct {
	RouterName string
	RouterPort ResourceReference
}

type SecurityGroupOrchestrationParams struct {
	Name             string
	AccessPorts      []util.PortSpec
	EgressRestricted bool
	EgressRules      []edgeproto.OutboundSecurityRule
}

type SecgrpParamsOp func(vmp *SecurityGroupOrchestrationParams) error

func secGrpWithEgressRules(rules []edgeproto.OutboundSecurityRule) SecgrpParamsOp {
	return func(sp *SecurityGroupOrchestrationParams) error {
		sp.EgressRules = rules
		if len(rules) > 0 {
			sp.EgressRestricted = true
		}
		return nil
	}
}

func secGrpWithAccessPorts(accessPorts string) SecgrpParamsOp {
	return func(sgp *SecurityGroupOrchestrationParams) error {
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

func GetSecGrpParams(name string, opts ...SecgrpParamsOp) (*SecurityGroupOrchestrationParams, error) {
	var sgp SecurityGroupOrchestrationParams
	sgp.Name = name
	for _, op := range opts {
		if err := op(&sgp); err != nil {
			return nil, err
		}
	}
	return &sgp, nil
}

type VolumeOrchestrationParams struct {
	Name             string
	ImageName        string
	Size             uint64
	AvailabilityZone string
	DeviceName       string
}
type VolumeOrchestrationParamsOp func(vmp *VolumeOrchestrationParams) error

type TagOrchestrationParams struct {
	Id       string
	Name     string
	Category string
}

// VMOrchestrationParams contains all details  that are needed by the orchestator
type VMOrchestrationParams struct {
	Id                      string
	Name                    string
	Role                    VMRole
	ImageName               string
	HostName                string
	DomainName              string
	FlavorName              string
	Vcpus                   uint64
	Ram                     uint64
	Disk                    uint64
	ComputeAvailabilityZone string
	UserData                string
	MetaData                string
	SharedVolume            bool
	DNSServers              string
	AuthPublicKey           string
	DeploymentManifest      string
	Command                 string
	Volumes                 []VolumeOrchestrationParams
	Ports                   []PortResourceReference      // depending on the orchestrator, IPs may be assigned to ports or
	FixedIPs                []FixedIPOrchestrationParams // to VMs directly
	ExternalGateway         string
	Tags                    string
	CustomizeGuest          bool
	ChefParams              *chefmgmt.VMChefParams
}

var (
	ChefClientKeyType     = true
	ChefValidationKeyType = false
)

func (v *VMPlatform) GetChefClientName(name string) string {
	// Prefix with region name
	return v.VMProperties.GetRegion() + "-" + name
}

func (v *VMPlatform) GetVMChefParams(nodeName, clientKey string, runList []string, attributes map[string]interface{}) *chefmgmt.VMChefParams {
	chefServerPath := v.VMProperties.GetChefServerPath()

	chefRunList := []string{"role[base]"}
	if runList != nil {
		chefRunList = append(chefRunList, runList...)
	}

	return &chefmgmt.VMChefParams{
		NodeName:   nodeName,
		ServerPath: chefServerPath,
		ClientKey:  clientKey,
		RunList:    chefRunList,
		Attributes: attributes,
	}
}

// VMGroupOrchestrationParams contains all the details used by the orchestator to create a set of associated VMs
type VMGroupOrchestrationParams struct {
	GroupName              string
	Subnets                []SubnetOrchestrationParams
	Ports                  []PortOrchestrationParams
	RouterInterfaces       []RouterInterfaceOrchestrationParams
	VMs                    []VMOrchestrationParams
	FloatingIPs            []FloatingIPOrchestrationParams
	SecurityGroups         []SecurityGroupOrchestrationParams
	Netspec                *NetSpecInfo
	Tags                   []TagOrchestrationParams
	SkipInfraSpecificCheck bool
	SkipSubnetGateway      bool
}

func (v *VMPlatform) GetVMRequestSpec(ctx context.Context, vmtype VMType, serverName, flavorName string, imageName string, connectExternal bool, opts ...VMReqOp) (*VMRequestSpec, error) {
	var vrs VMRequestSpec
	for _, op := range opts {
		if err := op(&vrs); err != nil {
			return nil, err
		}
	}
	vrs.Name = serverName
	vrs.Type = vmtype
	vrs.FlavorName = flavorName
	vrs.ImageName = imageName
	vrs.ConnectToExternalNet = connectExternal
	return &vrs, nil
}

func (v *VMPlatform) getVMGroupRequestSpec(ctx context.Context, name string, vms []*VMRequestSpec, opts ...VMGroupReqOp) (*VMGroupRequestSpec, error) {
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

func (v *VMPlatform) GetVMGroupOrchestrationParamsFromVMSpec(ctx context.Context, name string, vms []*VMRequestSpec, opts ...VMGroupReqOp) (*VMGroupOrchestrationParams, error) {
	vmgp, err := v.getVMGroupRequestSpec(ctx, name, vms, opts...)
	if err != nil {
		return nil, err
	}
	return v.getVMGroupOrchestrationParamsFromGroupSpec(ctx, vmgp)
}

func (v *VMPlatform) getVMGroupOrchestrationParamsFromGroupSpec(ctx context.Context, spec *VMGroupRequestSpec) (*VMGroupOrchestrationParams, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetVMGroupOrchestrationParams", "spec", spec)

	vmgp := VMGroupOrchestrationParams{GroupName: spec.GroupName}
	internalNetName := v.VMProperties.GetCloudletMexNetwork()
	internalNetId := v.VMProvider.NameSanitize(internalNetName)
	externalNetName := v.VMProperties.GetCloudletExternalNetwork()

	// DNS is applied either at the subnet or VM level
	cloudflareDns := []string{"1.1.1.1", "1.0.0.1"}
	vmDns := ""
	subnetDns := []string{}
	var err error
	cloudletSecGrpID := ""
	if !spec.SkipDefaultSecGrp {
		cloudletSecGrpID, err = v.VMProvider.GetResourceID(ctx, ResourceTypeSecurityGroup, v.VMProperties.GetCloudletSecurityGroupName())
	}
	internalSecgrpID := ""
	internalSecgrpPreexisting := false

	if err != nil {
		return nil, err
	}
	if v.VMProperties.GetSubnetDNS() == NoSubnetDNS {
		// Contrail workaround, see EDGECLOUD-2420 for details
		vmDns = strings.Join(cloudflareDns, " ")
	} else {
		subnetDns = cloudflareDns
	}

	vmgp.Netspec, err = ParseNetSpec(ctx, v.VMProperties.GetCloudletNetworkScheme())
	if err != nil {
		return nil, err
	}
	if spec.SkipInfraSpecificCheck {
		vmgp.SkipInfraSpecificCheck = true
	}

	rtrInUse := false
	rtr := v.VMProperties.GetCloudletExternalRouter()
	if rtr == NoConfigExternalRouter {
		log.SpanLog(ctx, log.DebugLevelInfra, "NoConfigExternalRouter in use")
	} else if rtr == NoExternalRouter {
		log.SpanLog(ctx, log.DebugLevelInfra, "NoExternalRouter in use ")
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "External router in use")
		if spec.NewSubnetName != "" {
			if !spec.SkipDefaultSecGrp {
				log.SpanLog(ctx, log.DebugLevelInfra, "SkipDefaultSecGrp flag set")
				internalSecgrpID = cloudletSecGrpID
				internalSecgrpPreexisting = true
			}
			rtrInUse = true
			routerPortName := spec.NewSubnetName + "-rtr-port"
			routerPort := PortOrchestrationParams{
				Name:        routerPortName,
				Id:          v.VMProvider.IdSanitize(routerPortName),
				NetworkName: internalNetName,
				NetworkId:   v.VMProvider.IdSanitize(internalNetName),
				SubnetId:    v.VMProvider.IdSanitize(spec.NewSubnetName),
				FixedIPs: []FixedIPOrchestrationParams{
					{
						Address:     NextAvailableResource,
						LastIPOctet: 1,
						Subnet:      NewResourceReference(spec.NewSubnetName, spec.NewSubnetName, false),
					},
				},
			}
			vmgp.Ports = append(vmgp.Ports, routerPort)
			newRouterIf := RouterInterfaceOrchestrationParams{
				RouterName: v.VMProperties.GetCloudletExternalRouter(),
				RouterPort: NewResourceReference(routerPortName, routerPortName, false),
			}
			vmgp.RouterInterfaces = append(vmgp.RouterInterfaces, newRouterIf)
		}
	}

	var egressRules []edgeproto.OutboundSecurityRule
	if spec.PrivacyPolicy != nil {
		egressRules = spec.PrivacyPolicy.OutboundSecurityRules
	}
	if spec.NewSecgrpName != "" {
		externalSecGrp, err := GetSecGrpParams(spec.NewSecgrpName, secGrpWithAccessPorts(spec.AccessPorts), secGrpWithEgressRules(egressRules))
		if err != nil {
			return nil, err
		}
		vmgp.SecurityGroups = append(vmgp.SecurityGroups, *externalSecGrp)
	}

	if err != nil {
		return nil, err
	}
	if spec.NewSubnetName != "" {
		newSubnet := SubnetOrchestrationParams{
			Name:        spec.NewSubnetName,
			Id:          v.VMProvider.IdSanitize(spec.NewSubnetName),
			CIDR:        NextAvailableResource,
			DHCPEnabled: "no",
			DNSServers:  subnetDns,
		}
		if spec.SkipSubnetGateway {
			newSubnet.SkipGateway = true
		}
		vmgp.Subnets = append(vmgp.Subnets, newSubnet)
	}

	var internalPortNextOctet uint32 = 101
	for _, vm := range spec.VMs {
		log.SpanLog(ctx, log.DebugLevelInfra, "Defining VM", "vm", vm)
		var role VMRole
		var newPorts []PortOrchestrationParams
		internalPortName := GetPortName(vm.Name, vm.ConnectToSubnet)
		externalPortName := GetPortName(vm.Name, externalNetName)

		connectToPreexistingSubnet := false
		if vm.ConnectToSubnet != "" && spec.NewSubnetName != vm.ConnectToSubnet {
			// we have specified a subnet to connect to which is not one we are creating
			// It therefore has to be a preexisting subnet
			connectToPreexistingSubnet = true
		}

		switch vm.Type {
		case VMTypePlatform:
			fallthrough
		case VMTypeRootLB:
			role = RoleAgent
			// do not attach the port to the VM if the policy is to do it after creation
			skipAttachVM := true
			internalPortSubnet := ""
			if v.VMProvider.GetInternalPortPolicy() == AttachPortDuringCreate {
				skipAttachVM = false
				internalPortSubnet = v.VMProvider.NameSanitize(spec.NewSubnetName)
			}
			// if the router is used we don't create an internal port for rootlb
			if vm.ConnectToSubnet != "" && !rtrInUse {
				// no router means rootlb must be connected to other VMs directly
				internalPort := PortOrchestrationParams{
					Name:        internalPortName,
					Id:          v.VMProvider.NameSanitize(internalPortName),
					NetworkName: internalNetName,
					NetworkId:   internalNetId,
					SubnetId:    internalPortSubnet,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs: []FixedIPOrchestrationParams{
						{
							Address:     NextAvailableResource,
							LastIPOctet: 1,
							Subnet:      NewResourceReference(vm.ConnectToSubnet, vm.ConnectToSubnet, connectToPreexistingSubnet),
						},
					},
					SkipAttachVM: skipAttachVM, //rootlb internal ports are attached in a separate step
				}
				newPorts = append(newPorts, internalPort)
			}

		case VMTypeAppVM:
			role = RoleVMApplication
			if vm.ConnectToSubnet != "" {
				// connect via internal network to LB
				internalPort := PortOrchestrationParams{
					Name:        internalPortName,
					Id:          v.VMProvider.NameSanitize(internalPortName),
					SubnetId:    v.VMProvider.NameSanitize(spec.NewSubnetName),
					NetworkName: internalNetName,
					NetworkId:   internalNetId,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs: []FixedIPOrchestrationParams{
						{Address: NextAvailableResource,
							LastIPOctet: internalPortNextOctet,
							Subnet:      NewResourceReference(vm.ConnectToSubnet, vm.ConnectToSubnet, connectToPreexistingSubnet),
						},
					},
				}
				internalPortNextOctet++
				newPorts = append(newPorts, internalPort)
			}

		case VMTypeClusterMaster:
			role = RoleMaster
			if vm.ConnectToSubnet != "" {
				// connect via internal network to LB
				internalPort := PortOrchestrationParams{
					Name:        internalPortName,
					Id:          v.VMProvider.NameSanitize(internalPortName),
					SubnetId:    v.VMProvider.NameSanitize(spec.NewSubnetName),
					NetworkId:   internalNetId,
					NetworkName: internalNetName,
					FixedIPs: []FixedIPOrchestrationParams{
						{Address: NextAvailableResource,
							LastIPOctet: 10,
							Subnet:      NewResourceReference(vm.ConnectToSubnet, vm.ConnectToSubnet, connectToPreexistingSubnet),
						},
					},
				}
				newPorts = append(newPorts, internalPort)
			} else {
				return nil, fmt.Errorf("k8s master not specified to be connected to internal network")
			}
		case VMTypeClusterNode:
			role = RoleNode
			if vm.ConnectToSubnet != "" {
				// connect via internal network to LB
				internalPort := PortOrchestrationParams{
					Name:        internalPortName,
					Id:          v.VMProvider.IdSanitize(internalPortName),
					SubnetId:    v.VMProvider.NameSanitize(spec.NewSubnetName),
					NetworkName: internalNetName,
					NetworkId:   internalNetId,
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs: []FixedIPOrchestrationParams{
						{Address: NextAvailableResource,
							LastIPOctet: internalPortNextOctet,
							Subnet:      NewResourceReference(vm.ConnectToSubnet, vm.ConnectToSubnet, connectToPreexistingSubnet),
						},
					},
				}
				internalPortNextOctet++
				newPorts = append(newPorts, internalPort)
			} else {
				return nil, fmt.Errorf("k8s node not specified to be connected to internal network")
			}
		default:
			return nil, fmt.Errorf("unexpected VM type: %s", vm.Type)
		}
		// ports contains only internal ports at this point. Optionally add the internal
		// security group which is used when we have a router
		if internalSecgrpID != "" {
			for i := range newPorts {
				sec := NewResourceReference(internalSecgrpID, internalSecgrpID, internalSecgrpPreexisting)
				newPorts[i].SecurityGroups = append(newPorts[i].SecurityGroups, sec)
			}
		}

		if vm.ConnectToExternalNet {
			if spec.NewSecgrpName == "" {
				return nil, fmt.Errorf("external network specified with no security group: %s", vm.Name)
			}
			var externalport PortOrchestrationParams
			if vmgp.Netspec.FloatingIPNet != "" {
				externalport = PortOrchestrationParams{
					Name:        externalPortName,
					Id:          v.VMProvider.NameSanitize(externalPortName),
					NetworkName: vmgp.Netspec.FloatingIPNet,
					NetworkId:   v.VMProvider.NameSanitize(vmgp.Netspec.FloatingIPNet),
					VnicType:    vmgp.Netspec.VnicType,
					FixedIPs: []FixedIPOrchestrationParams{
						{
							Subnet: NewResourceReference(vmgp.Netspec.FloatingIPSubnet, vmgp.Netspec.FloatingIPSubnet, false),
						},
					},
				}
			} else {
				externalport = PortOrchestrationParams{
					Name:        externalPortName,
					Id:          v.VMProvider.IdSanitize(externalPortName),
					NetworkName: externalNetName,
					NetworkId:   v.VMProvider.IdSanitize(externalNetName),
					VnicType:    vmgp.Netspec.VnicType,
				}

				externalport.SecurityGroups = []ResourceReference{
					NewResourceReference(spec.NewSecgrpName, spec.NewSecgrpName, false),
				}
				if !spec.SkipDefaultSecGrp {
					externalport.SecurityGroups = append(externalport.SecurityGroups, NewResourceReference(cloudletSecGrpID, cloudletSecGrpID, true))
				}
				newPorts = append(newPorts, externalport)
			}
		}
		if !vm.CreatePortsOnly {
			log.SpanLog(ctx, log.DebugLevelInfra, "Defining new VM orch param", "vm.Name", vm.Name, "ports", newPorts)
			newVM := VMOrchestrationParams{
				Name:                    v.VMProvider.NameSanitize(vm.Name),
				Id:                      v.VMProvider.IdSanitize(vm.Name),
				Role:                    role,
				DNSServers:              vmDns,
				ImageName:               vm.ImageName,
				FlavorName:              vm.FlavorName,
				HostName:                util.DNSSanitize(strings.Split(vm.Name, ".")[0]),
				DomainName:              v.VMProperties.CommonPf.GetCloudletDNSZone(),
				DeploymentManifest:      vm.DeploymentManifest,
				Command:                 vm.Command,
				ComputeAvailabilityZone: vm.ComputeAvailabilityZone,
				ChefParams:              vm.ChefParams,
			}
			if vm.ExternalVolumeSize > 0 {
				externalVolume := VolumeOrchestrationParams{
					Name:       vm.Name + "-volume",
					Size:       vm.ExternalVolumeSize,
					ImageName:  vm.ImageName,
					DeviceName: "vda",
				}
				newVM.ImageName = ""
				newVM.Volumes = append(newVM.Volumes, externalVolume)
			}
			if vm.SharedVolumeSize > 0 {
				sharedVolume := VolumeOrchestrationParams{
					Name:       vm.Name + "-shared-volume",
					Size:       vm.SharedVolumeSize,
					ImageName:  vm.ImageName,
					DeviceName: "vdb",
				}
				newVM.Volumes = append(newVM.Volumes, sharedVolume)
				newVM.SharedVolume = true
			}
			for _, p := range newPorts {
				if !p.SkipAttachVM {
					newVM.Ports = append(newVM.Ports, NewPortResourceReference(p.Name, p.Id, p.NetworkId, p.SubnetId, false))
					newVM.FixedIPs = append(newVM.FixedIPs, p.FixedIPs...)
				}
			}
			vmgp.VMs = append(vmgp.VMs, newVM)
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "Preexisting vm not added to group params", "vm.Name", vm.Name, "ports", newPorts)
		}
		vmgp.Ports = append(vmgp.Ports, newPorts...)
	}

	return &vmgp, nil
}

// OrchestrateVMsFromVMSpec calls the provider function to do the orchestation of the VMs.  It returns the updated VM group spec
func (v *VMPlatform) OrchestrateVMsFromVMSpec(ctx context.Context, name string, vms []*VMRequestSpec, action ActionType, updateCallback edgeproto.CacheUpdateCallback, opts ...VMGroupReqOp) (*VMGroupOrchestrationParams, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "OrchestrateVMsFromVMSpec", "name", name)
	chefClient := v.VMProperties.GetChefClient()
	if chefClient == nil {
		return nil, fmt.Errorf("Chef client is not initialzied")
	}
	gp, err := v.GetVMGroupOrchestrationParamsFromVMSpec(ctx, name, vms, opts...)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetVMGroupOrchestrationParamsFromVMSpec failed", "error", err)
		return gp, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created vm group spec", "gp", gp)
	switch action {
	case ActionCreate:
		for _, vm := range vms {
			if vm.CreatePortsOnly || vm.Type == VMTypeAppVM {
				continue
			}
			if vm.ChefParams == nil {
				return gp, fmt.Errorf("chef params doesn't exist for %s", vm.Name)
			}
			clientKey, err := chefmgmt.ChefClientCreate(ctx, chefClient, vm.ChefParams)
			if err != nil {
				return gp, err
			}
			vm.ChefParams.ClientKey = clientKey
		}
		err = v.VMProvider.CreateVMs(ctx, gp, updateCallback)
	case ActionUpdate:
		err = v.VMProvider.UpdateVMs(ctx, gp, updateCallback)
	case ActionSync:
		err = v.VMProvider.SyncVMs(ctx, gp, updateCallback)
	}
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error while orchestrating vms", "name", name, "action", action, "err", err)
		return gp, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "VM action done", "action", action)
	return gp, nil
}

func (v *VMPlatform) GetSubnetGatewayFromVMGroupParms(ctx context.Context, subnetName string, vmgp *VMGroupOrchestrationParams) (string, error) {
	for _, s := range vmgp.Subnets {
		if s.Name == subnetName {
			return s.GatewayIP, nil
		}
	}
	return "", fmt.Errorf("Subnet: %s not found in vm group params", subnetName)
}

package mexos

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type VMParams struct {
	VMName              string
	Flavor              string
	ImageName           string
	SecurityGroup       string
	NetworkName         string
	SubnetName          string
	VnicType            string
	MEXRouterIP         string
	GatewayIP           string
	FloatingIPAddressID string
	AuthPublicKey       template.HTML // Must be of this type to skip HTML escaping
	AccessPorts         []util.PortSpec
	DeploymentManifest  template.HTML
	Command             template.HTML
	IsRootLB            bool
}

type DeploymentType string

const (
	RootLBVMDeployment DeploymentType = "mexrootlb"
	UserVMDeployment   DeploymentType = "mexuservm"
)

var heatStackLock sync.Mutex
var heatCreate string = "CREATE"
var heatUpdate string = "UPDATE"
var heatDelete string = "DELETE"
var vmCloudConfig = `#cloud-config
bootcmd:
 - echo MOBILEDGEX CLOUD CONFIG START
 - echo 'APT::Periodic::Enable "0";' > /etc/apt/apt.conf.d/10cloudinit-disable
 - apt-get -y purge update-notifier-common ubuntu-release-upgrader-core landscape-common unattended-upgrades
 - echo "Removed APT and Ubuntu extra packages" | systemd-cat
ssh_authorized_keys:
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDZiZ16uwmHOuafD6a9AmZ5kYF9LqtfyrUOIVMF1eoRJCMALQrWbzNz/NOnqi5h5dwhPn+49oWMU16BKDkEgDik2jgNUOSZ69oZM4/ovPsB8yL55qdNBTx32kov5O8NkSwMEDter2mAPi9czCEv18MRC1qkiZCUxmfFs0BBgXtNfE42Utr97YcKFtvutLDGA1hoFVjon0Yk7wSMNZfwkBznVoShRISCzMvG5uVtf6miJwIIA9+SiwA/aa2OjCRQaiPCKJrPzHMcuLg4oZcs0ltd1CaIVLtMGaqpEoIvDumXEpuk0TSBJwxWUDAEgO5ILmVxi2fSLKa0yuLala6bcfwJ stack@sv1.mobiledgex.com
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrHlOJOJUqvd4nEOXQbdL8ODKzWaUxKVY94pF7J3diTxgZ1NTvS6omqOjRS3loiU7TOlQQU4cKnRRnmJW8QQQZSOMIGNrMMInGaEYsdm6+tr1k4DDfoOrkGMj3X/I2zXZ3U+pDPearVFbczCByPU0dqs16TWikxDoCCxJRGeeUl7duzD9a65bI8Jl+zpfQV+I7OPa81P5/fw15lTzT4+F9MhhOUVJ4PFfD+d6/BLnlUfZ94nZlvSYnT+GoZ8xTAstM7+6pvvvHtaHoV4YqRf5CelbWAQ162XNa9/pW5v/RKDrt203/JEk3e70tzx9KAfSw2vuO1QepkCZAdM9rQoCd ubuntu@registry
chpasswd: { expire: False }
ssh_pwauth: False
timezone: UTC
runcmd:
 - [ echo, MOBILEDGEX, doing, ifconfig ]
 - [ ifconfig, -a ]`

// This is the resources part of a template for a VM. It is for use within another template
// the parameters under VMP can come from either a standalone struture (VM Create) or a cluster (for rootLB)
var vmTemplateResources = `
   {{.VMName}}:
      type: OS::Nova::Server
      properties:
         name: 
            {{.VMName}}
         image: {{.ImageName}}
        {{if not .FloatingIPAddressID}}
         security_groups:
          - {{.SecurityGroup}}
         {{if .AccessPorts}} - { get_resource: vm_security_group } {{- end}}
        {{- end}}
         flavor: {{.Flavor}}
        {{if .AuthPublicKey}} key_name: { get_resource: ssh_key_pair } {{- end}}
         config_drive: true       
         user_data_format: RAW
         networks:
        {{if .FloatingIPAddressID}}
          - port: { get_resource: vm-port }
        {{else}}
          - network: {{.NetworkName}}
        {{- end}}
        {{if .IsRootLB}}
         metadata:
            skipk8s: yes
            role: mex-agent-node 
            edgeproxy: {{.GatewayIP}}
            mex-flavor: {{.Flavor}}
           {{if .MEXRouterIP}}
            privaterouter: {{.MEXRouterIP}}
           {{- end}}
         user_data: |
` + reindent(vmCloudConfig, 12) + `
        {{- end}}
        {{if .DeploymentManifest}}
         user_data: |
{{ Indent .DeploymentManifest 13 }}
        {{- end}}
        {{if .Command}}
         user_data: |
            #cloud-config
            runcmd:
             - {{.Command}}
        {{- end}}
  {{if .AuthPublicKey}}
   ssh_key_pair:
       type: OS::Nova::KeyPair
       properties:
          name: {{.VMName}}-ssh-keypair
          public_key: "{{.AuthPublicKey}}"
  {{- end}}
  {{if .FloatingIPAddressID}}
   vm-port:
       type: OS::Neutron::Port
       properties:
           name: {{.VMName}}-port
          {{if .VnicType}}
           binding:vnic_type: {{.VnicType}}
          {{- end}}
           network_id: {{.NetworkName}}
           fixed_ips: 
            - subnet_id: {{.SubnetName}}
           security_groups:
            - {{$.SecurityGroup}}
           {{if .AccessPorts}} - { get_resource: vm_security_group } {{- end}}
   floatingip:
       type: OS::Neutron::FloatingIPAssociation
       properties:
          floatingip_id: {{.FloatingIPAddressID}}
          port_id: { get_resource: vm-port }
  {{- end}}
  {{if .AccessPorts}}
   vm_security_group:
       type: OS::Neutron::SecurityGroup
       properties:
          name: {{.VMName}}-sg
          rules: [
             {{range .AccessPorts}}
              {remote_ip_prefix: 0.0.0.0/0,
               protocol: {{.Proto}},
               port_range_min: {{.Port}},
               port_range_max: {{.Port}}},
             {{end}}
          ]
  {{- end}}`

var VmTemplate = `
heat_template_version: 2016-10-14
description: Create a VM
resources:
` + vmTemplateResources

// ClusterNode is a k8s node
type ClusterNode struct {
	NodeName string
	NodeIP   string
}

// ClusterParams has the info needed to populate the heat template
type ClusterParams struct {
	NodeFlavor     string
	ImageName      string
	SecurityGroup  string
	MEXRouterName  string
	MEXNetworkName string
	VnicType       string
	ClusterName    string
	CIDR           string
	GatewayIP      string
	MasterIP       string
	DNSServers     []string
	Nodes          []ClusterNode
	*VMParams      //rootlb
}

var k8sClusterTemplate = `
heat_template_version: 2016-10-14
description: Create a cluster

resources:
   k8s-subnet:
      type: OS::Neutron::Subnet
      properties:
        cidr: {{.CIDR}}
        network: mex-k8s-net-1
        gateway_ip: {{.GatewayIP}}
        enable_dhcp: false
        dns_nameservers:
       {{range .DNSServers}}
         - {{.}}
       {{end}}
        name: 
           mex-k8s-subnet-{{.ClusterName}}
  {{if .MEXRouterName}}
   router-port:
       type: OS::Neutron::Port
       properties:
          name: mex-k8s-subnet-{{.ClusterName}}
          network_id: mex-k8s-net-1
          fixed_ips:
          - subnet: { get_resource: k8s-subnet}
            ip_address: {{.GatewayIP}}
          security_groups:
           - {{$.SecurityGroup}}

   router-interface:
      type: OS::Neutron::RouterInterface
      properties:
         router:  {{.MEXRouterName}}
         port: { get_resource: router-port }
  {{- end}}
   k8s-master-port:
      type: OS::Neutron::Port
      properties:
         name: k8s-master-port
        {{if .VnicType}}
         binding:vnic_type: {{.VnicType}}
        {{- end}}
         network_id: {{.MEXNetworkName}}
         fixed_ips:
          - subnet: { get_resource: k8s-subnet} 
            ip_address: {{.MasterIP}}
         security_groups:
          - {{$.SecurityGroup}}

   k8s_master:
      type: OS::Nova::Server
      properties:
         name: 
            mex-k8s-master-{{.ClusterName}}
         image: {{.ImageName}}
         flavor: {{.NodeFlavor}}
         config_drive: true
         user_data_format: RAW
         user_data: |
` + reindent(vmCloudConfig, 12) + `
         networks:
          - port: { get_resource: k8s-master-port }
         metadata:
            skipk8s: no
            role: k8s-master 
            edgeproxy: {{.GatewayIP}}
            mex-flavor: {{.NodeFlavor}}
            k8smaster: {{.MasterIP}}

  {{range .Nodes}}
   {{.NodeName}}-port:
      type: OS::Neutron::Port
      properties:
          name: mex-k8s-{{.NodeName}}-port-{{$.ClusterName}}
         {{if $.VnicType}}
          binding:vnic_type: {{$.VnicType}}
         {{- end}}
          network_id: {{$.MEXNetworkName}}
          fixed_ips:
          - subnet: { get_resource: k8s-subnet}
            ip_address: {{.NodeIP}}
          security_groups:
           - {{$.SecurityGroup}}

   {{.NodeName}}:
      type: OS::Nova::Server
      depends_on: k8s_master
      properties:
         name: {{.NodeName}}-{{$.ClusterName}}
         image: {{$.ImageName}}
         flavor: {{$.NodeFlavor}}
         config_drive: true
         user_data_format: RAW
         user_data: |
` + reindent(vmCloudConfig, 12) + `
         networks:
          - port: { get_resource: {{.NodeName}}-port } 
         metadata:
            skipk8s: no 
            role: k8s-node
            edgeproxy: {{$.GatewayIP}}
            mex-flavor: {{$.NodeFlavor}}
            k8smaster: {{$.MasterIP}}
  {{end}}
`

func reindent(str string, indent int) string {
	out := ""
	for _, v := range strings.Split(str, "\n") {
		out += strings.Repeat(" ", indent) + v + "\n"
	}
	return strings.TrimSuffix(out, "\n")
}

func WriteTemplateFile(filename string, buf *bytes.Buffer) error {
	outFile, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to write heat template %s: %s", filename, err.Error())
	}
	_, err = outFile.WriteString(buf.String())

	if err != nil {
		outFile.Close()
		os.Remove(filename)
		return fmt.Errorf("unable to write heat template file %s: %s", filename, err.Error())
	}
	outFile.Sync()
	outFile.Close()
	return nil
}

func waitForStack(ctx context.Context, stackname string, action string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "waiting for stack", "name", stackname, "action", action)
	start := time.Now()
	for {
		time.Sleep(10 * time.Second)
		hd, err := getHeatStackDetail(ctx, stackname)
		if action == heatDelete && hd == nil {
			// it's gone
			return nil
		}
		if err != nil {
			return err
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "Got Heat Stack detail", "detail", hd)
		updateCallback(edgeproto.UpdateStep, fmt.Sprintf("Heat Stack Status: %s", hd.StackStatus))

		switch hd.StackStatus {
		case action + "_COMPLETE":
			log.SpanLog(ctx, log.DebugLevelMexos, "Heat Stack succeeded", "action", action, "stackName", stackname)
			return nil
		case action + "_IN_PROGRESS":
			elapsed := time.Since(start)
			if elapsed >= (time.Minute * 20) {
				// this should not happen and indicates the stack is stuck somehow
				log.InfoLog("Heat stack taking too long", "status", hd.StackStatus, "elasped time", elapsed)
				return fmt.Errorf("Heat stack taking too long")
			}
			continue
		case action + "_FAILED":
			log.InfoLog("Heat Stack failed", "action", action, "stackName", stackname)
			return fmt.Errorf("Heat Stack failed: %s", hd.StackStatusReason)
		default:
			log.InfoLog("Unexpected Heat Stack status", "status", stackname)
			return fmt.Errorf("Stack create unexpected status: %s", hd.StackStatus)
		}
	}
}

func GetVMParams(ctx context.Context, depType DeploymentType, serverName, flavor, imageName, authPublicKey, accessPorts, deploymentManifest, command string, ni *NetSpecInfo) (*VMParams, error) {
	var vmp VMParams
	var err error
	vmp.VMName = serverName
	vmp.Flavor = flavor
	vmp.ImageName = imageName
	vmp.SecurityGroup = GetCloudletSecurityGroup()
	if depType == RootLBVMDeployment {
		vmp.GatewayIP, err = GetExternalGateway(ctx, GetCloudletExternalNetwork())
		if err != nil {
			return nil, err
		}
		vmp.MEXRouterIP, err = GetMexRouterIP(ctx)
		if err != nil {
			return nil, err
		}
		vmp.IsRootLB = true
	}
	if authPublicKey != "" {
		convKey, err := util.ConvertPEMtoOpenSSH(authPublicKey)
		if err != nil {
			return nil, err
		}
		vmp.AuthPublicKey = template.HTML(convKey)
	}
	if accessPorts != "" {
		vmp.AccessPorts, err = util.ParsePorts(accessPorts)
		if err != nil {
			return nil, err
		}
	}
	if deploymentManifest != "" {
		vmp.DeploymentManifest = template.HTML(deploymentManifest)
	}
	if command != "" {
		vmp.Command = template.HTML(command)
	}
	if ni != nil && ni.FloatingIPNet != "" {

		fips, err := ListFloatingIPs(ctx)
		for _, f := range fips {
			if f.Port == "" && f.FloatingIPAddress != "" {
				vmp.FloatingIPAddressID = f.ID
			}
		}
		if vmp.FloatingIPAddressID == "" {
			return nil, fmt.Errorf("Unable to allocate a floating IP")
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to list floating IPs %v", err)
		}
		vmp.NetworkName = ni.FloatingIPNet
		vmp.SubnetName = ni.FloatingIPSubnet
	} else {
		vmp.NetworkName = GetCloudletExternalNetwork()
	}
	if ni != nil {
		vmp.VnicType = ni.VnicType
	}
	return &vmp, nil
}

func createOrUpdateHeatStackFromTemplate(ctx context.Context, templateData interface{}, stackName string, templateString string, action string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "createHeatStackFromTemplate", "stackName", stackName)

	var buf bytes.Buffer
	updateCallback(edgeproto.UpdateTask, "Creating Heat Stack for "+stackName)

	funcMap := template.FuncMap{
		"Indent": func(values ...interface{}) template.HTML {
			s := values[0].(template.HTML)
			l := 4
			if len(values) > 1 {
				l = values[1].(int)
			}
			var newStr []string
			for _, v := range strings.Split(string(s), "\n") {
				nV := fmt.Sprintf("%s%s", strings.Repeat(" ", l), v)
				newStr = append(newStr, nV)
			}
			return template.HTML(strings.Join(newStr, "\n"))
		},
	}

	tmpl, err := template.New(stackName).Funcs(funcMap).Parse(templateString)
	if err != nil {
		return err
	}
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		return err
	}
	filename := stackName + "-heat.yaml"
	err = WriteTemplateFile(filename, &buf)
	if err != nil {
		return err
	}
	if action == heatCreate {
		err = createHeatStack(ctx, filename, stackName)
	} else {
		err = updateHeatStack(ctx, filename, stackName)
	}
	if err != nil {
		return err
	}
	err = waitForStack(ctx, stackName, action, updateCallback)
	return err
}

// UpdateHeatStackFromTemplate fills the template from templateData and creates the stack
func UpdateHeatStackFromTemplate(ctx context.Context, templateData interface{}, stackName, templateString string, updateCallback edgeproto.CacheUpdateCallback) error {
	return createOrUpdateHeatStackFromTemplate(ctx, templateData, stackName, templateString, heatUpdate, updateCallback)
}

// CreateHeatStackFromTemplate fills the template from templateData and creates the stack
func CreateHeatStackFromTemplate(ctx context.Context, templateData interface{}, stackName, templateString string, updateCallback edgeproto.CacheUpdateCallback) error {
	return createOrUpdateHeatStackFromTemplate(ctx, templateData, stackName, templateString, heatCreate, updateCallback)
}

// HeatDeleteStack deletes the VM resources
func HeatDeleteStack(ctx context.Context, stackName string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "deleting heat stack for stack", "stackName", stackName)
	deleteHeatStack(ctx, stackName)
	return waitForStack(ctx, stackName, heatDelete, edgeproto.DummyUpdateCallback)
}

//GetClusterParams fills template parameters for the cluster.  A non blank rootLBName will add a rootlb VM
func getClusterParams(ctx context.Context, clusterInst *edgeproto.ClusterInst, rootLBName string, action string) (*ClusterParams, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "getClusterParams", "cluster", clusterInst, "action", action)

	var cp ClusterParams
	var err error
	ni, err := ParseNetSpec(ctx, GetCloudletNetworkScheme())
	if err != nil {
		return nil, err
	}
	cp.VnicType = ni.VnicType
	if rootLBName != "" {
		cp.VMParams, err = GetVMParams(ctx,
			RootLBVMDeployment,
			rootLBName,
			clusterInst.NodeFlavor,
			GetCloudletOSImage(),
			"", // AuthPublicKey
			"", // AccessPorts
			"", // DeploymentManifest
			"", // Command
			ni,
		)
		if err != nil {
			return nil, fmt.Errorf("Unable to get rootlb params: %v", err)
		}
	}
	cp.ClusterName = k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	cp.MEXRouterName = GetCloudletExternalRouter()
	cp.MEXNetworkName = GetCloudletMexNetwork()
	cp.ImageName = GetCloudletOSImage()
	cp.SecurityGroup = GetCloudletSecurityGroup()
	usedCidrs := make(map[string]string)

	currentSubnetName := ""
	found := false
	if action == heatUpdate {
		currentSubnetName = "mex-k8s-subnet-" + cp.ClusterName
	}
	sns, snserr := ListSubnets(ctx, ni.Name)
	if snserr != nil {
		return nil, fmt.Errorf("can't get list of subnets for %s, %v", ni.Name, snserr)
	}
	for _, s := range sns {
		usedCidrs[s.Subnet] = s.Name
	}

	nodeIPPrefix := ""

	//find an available subnet
	for i := 0; i <= 255; i++ {
		subnet := fmt.Sprintf("%s.%s.%d.%d/%s", ni.Octets[0], ni.Octets[1], i, 0, ni.NetmaskBits)
		// either look for an unused one (create) or the current one (update)
		if (action == heatCreate && usedCidrs[subnet] == "") || (action == heatUpdate && usedCidrs[subnet] == currentSubnetName) {
			found = true
			cp.CIDR = subnet
			cp.GatewayIP = fmt.Sprintf("%s.%s.%d.%d", ni.Octets[0], ni.Octets[1], i, 1)
			cp.MasterIP = fmt.Sprintf("%s.%s.%d.%d", ni.Octets[0], ni.Octets[1], i, 10)
			nodeIPPrefix = fmt.Sprintf("%s.%s.%d", ni.Octets[0], ni.Octets[1], i)
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("cannot find free subnet cidr")
	}

	if clusterInst.NodeFlavor == "" {
		return nil, fmt.Errorf("Node Flavor is not set")
	}
	cp.NodeFlavor = clusterInst.NodeFlavor
	for i := uint32(1); i <= clusterInst.NumNodes; i++ {
		nn := fmt.Sprintf("mex-k8s-node-%d", i)
		nip := fmt.Sprintf("%s.%d", nodeIPPrefix, i+100)
		cn := ClusterNode{NodeName: nn, NodeIP: nip}
		cp.Nodes = append(cp.Nodes, cn)
	}
	// cloudflare primary and backup
	cp.DNSServers = []string{"1.1.1.1", "1.0.0.1"}
	return &cp, nil
}

// HeatCreateRootLBVM creates a roobLB VM
func HeatCreateRootLBVM(ctx context.Context, serverName string, stackName string, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "HeatCreateRootLBVM", "serverName", serverName, "stackName", stackName, "flavor", flavor)
	ni, err := ParseNetSpec(ctx, GetCloudletNetworkScheme())
	if err != nil {
		return err
	}
	// lock here to avoid getting the same floating IP; we need to lock until the stack is done
	// Floating IPs are allocated both by VM and cluster creation
	// TODO: floating IP lock should apply to developer app VMs also
	if ni.FloatingIPNet != "" {
		heatStackLock.Lock()
		defer heatStackLock.Unlock()
	}
	vmp, err := GetVMParams(ctx,
		RootLBVMDeployment,
		serverName,
		flavor,
		GetCloudletOSImage(),
		"", // AuthPublicKey
		"", // AccessPorts
		"", // DeploymentManifest
		"", // Command
		ni,
	)
	if err != nil {
		return fmt.Errorf("Unable to get VM params: %v", err)
	}
	return CreateHeatStackFromTemplate(ctx, vmp, stackName, VmTemplate, updateCallback)
}

// HeatCreateClusterKubernetes creates a k8s cluster which may optionally include a dedicated root LB
func HeatCreateClusterKubernetes(ctx context.Context, clusterInst *edgeproto.ClusterInst, dedicatedRootLBName string, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelMexos, "HeatCreateClusterKubernetes", "clusterInst", clusterInst)
	// It is problematic to create 2 clusters at the exact same time because we will look for available subnet CIDRS when
	// defining the template.  If 2 start at once they may end up trying to create the same subnet and one will fail.
	// So we will do this one at a time.   It will slightly slow down the creation of the second cluster, but the heat
	// stack create time is relatively quick compared to the k8s startup which can be done in parallel
	// Floating IPs can also be allocated within the stack and need to be locked as well.
	heatStackLock.Lock()
	defer heatStackLock.Unlock()

	cp, err := getClusterParams(ctx, clusterInst, dedicatedRootLBName, heatCreate)
	if err != nil {
		return err
	}

	templateString := k8sClusterTemplate
	//append the VM resources for the rootLB is specified
	if dedicatedRootLBName != "" {
		templateString += vmTemplateResources
	}
	err = CreateHeatStackFromTemplate(ctx, cp, cp.ClusterName, templateString, updateCallback)
	return err
}

// HeatUpdateClusterKubernetes creates a k8s cluster which may optionally include a dedicated root LB
func HeatUpdateClusterKubernetes(ctx context.Context, clusterInst *edgeproto.ClusterInst, dedicatedRootLBName string, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelMexos, "HeatUpdateClusterKubernetes", "clusterInst", clusterInst)

	cp, err := getClusterParams(ctx, clusterInst, dedicatedRootLBName, heatUpdate)
	if err != nil {
		return err
	}

	templateString := k8sClusterTemplate
	//append the VM resources for the rootLB is specified
	if dedicatedRootLBName != "" {
		templateString += vmTemplateResources
	}
	err = UpdateHeatStackFromTemplate(ctx, cp, cp.ClusterName, templateString, updateCallback)
	return err
}

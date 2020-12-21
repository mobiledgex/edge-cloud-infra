package vcd

import (
	"context"
	"fmt"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Vcd support objects
// Note regarding govcd SDK:
// Most all types in govcd are arranged as a containerized type in govcd that compose Client and
// specific methods, with the business end of the type in types.go. example found in vdc.go:
//
//   type Vdc struct {
//  	  Vdc    *types.Vdc
//	  client *Client
//   }
//
// The method calls are accessed via the "outer" govcd.Vdc object utilizing the client object, and operate on
// the 'inner' types.Vdc object.
//
var vcdProviderVersion = "-0.1-alpha"

type VcdPlatform struct {
	vmProperties *vmlayer.VMProperties
	vcdVars      map[string]string
	caches       *platform.Caches
	Creds        *VcdConfigParams
	Client       *govcd.VCDClient
	Objs         VcdObjects
	TestMode     bool
	Verbose      bool
}

type VcdConfigParams struct {
	User     string
	Password string
	Org      string
	Href     string
	VDC      string
	Insecure bool
	Token    string
}

type VdcMap map[string]*govcd.Vdc

type CatContainer struct {
	OrgCat    *govcd.Catalog // contains types.Catalog
	CatRec    *types.CatalogRecord
	MediaRecs []*types.MediaRecordType // Media found in OrgCat
}

type NetMap map[string]*govcd.OrgVDCNetwork
type CatMap map[string]CatContainer

type VApp struct {
	VApp *govcd.VApp
	VMs  VMMap
}

type VAppsMap map[string]*VApp

type VMMap map[string]*govcd.VM // alt VMRecord?
type VAppTmplMap map[string]*govcd.VAppTemplate
type TmplVMsMap map[string]*types.QueryResultVMRecordType
type MediaMap map[string]*govcd.Media

// A VM element of a clusterInst in some Cloudlet
type ClusterVm struct {
	vmName          string
	vmRole          string
	vmType          string
	vmFlavor        string
	vmMeta          []string
	vmIPs           edgeproto.IpAddr // ExternalIp , InternalIp
	vmParentCluster string
	vm              *govcd.VM
}

// IPaddr + vm attributes per cluster
type VMIPsMap map[string]ClusterVm

// A map key'ed by CIDR whose value is another map of all VMs in the
// Cluster represented by this CIDR, this key'ed by IP addr
// This set of vms under this CIDR represent a single cluster
type Cluster struct {
	Name string
	VMs  VMIPsMap
}

type CidrMap map[string]*Cluster
type CloudVMsMap map[string]*govcd.VM

// One cloudlet per vdc instance
type MexCloudlet struct {
	ParentVdc    *govcd.Vdc
	CloudVapp    *govcd.VApp
	CloudletName string
	ExtNet       *govcd.OrgVDCNetwork // The external network shared by all agent nodes of cluster in cloudlet
	ExtIp        string
	Clusters     CidrMap     // Clusters are keyed by their internal net CIDR
	ExtVMMap     CloudVMsMap // keyed by exteral net ip
	// federation partner TBI (single remote org/vdc:  a pair wise assocication)

}

type VcdObjects struct {
	Org       *govcd.Org
	Vdc       *govcd.Vdc // VdcMap
	Nets      NetMap
	Cats      CatMap
	VApps     VAppsMap
	VAppTmpls VAppTmplMap
	// while we'll discover all external networks
	// avaliable to our vdc, we'll only utilize the first we find as
	// v.Objs.PrimaryNet. May be overriden using vcd.Vars
	PrimaryNet  *govcd.OrgVDCNetwork
	PrimaryCat  *govcd.Catalog
	VMs         VMMap
	DeployedVMs VMMap
	TemplateVMs TmplVMsMap
	EdgeGateway govcd.EdgeGateway
	Media       MediaMap
	Cloudlet    *MexCloudlet
	Template    *govcd.VAppTemplate // tmp xxx debug
}

func (v *VcdPlatform) GetType() string {
	return "vcd"
}

func (v *VcdPlatform) InitProvider(ctx context.Context, caches *platform.Caches, stage vmlayer.ProviderInitStage, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "InitProvider for Vcd 1", "stage", stage)
	v.Verbose = v.SetVcdVerbose()
	v.InitData(ctx, caches)
	v.Objs.Nets = make(map[string]*govcd.OrgVDCNetwork)
	v.Objs.Cats = make(map[string]CatContainer)
	v.Objs.VApps = make(map[string]*VApp)
	v.Objs.VMs = make(map[string]*govcd.VM)
	v.Objs.VAppTmpls = make(map[string]*govcd.VAppTemplate)
	v.Objs.TemplateVMs = make(map[string]*types.QueryResultVMRecordType)
	v.Objs.Media = make(MediaMap)

	if v.Client == nil {
		client, err := v.GetClient(ctx, v.Creds)
		if err != nil {
			return fmt.Errorf("InitProvider Unable to create Vcd Client: %s\n", err.Error())
		}
		v.Client = client
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "Discover resources for", "Org", v.Creds.Org)
	err := v.ImportDataFromInfra(ctx)
	if err != nil {
		return fmt.Errorf("ImportDataFromInfra failed: %s", err.Error())
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "InitProvider Discovery Complete", "stage", stage)

	err = v.SetProviderSpecificProps(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *VcdPlatform) InitData(ctx context.Context, caches *platform.Caches) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitData caches set")
	v.caches = caches
}

func (v *VcdPlatform) ImportDataFromInfra(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "ImportDataFromInfra")
	if v.Client == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Obtain vcd client")
		client, err := v.GetClient(ctx, v.Creds)
		if err != nil {
			return fmt.Errorf("Unable to create Vcd Client %s\n", err.Error())
		}
		v.Client = client
	}
	err := v.GetPlatformResources(ctx)
	if err != nil {
		return fmt.Errorf("Error retrieving Platform Resources: %s", err.Error())
	}
	return nil
}

func (v *VcdPlatform) GetPlatformResourceInfo(ctx context.Context) (*vmlayer.PlatformResources, error) {

	var resources *vmlayer.PlatformResources
	log.SpanLog(ctx, log.DebugLevelInfra, "GetPlatformResourceInfo N ")

	resources.CollectTime, _ = gogotypes.TimestampProto(time.Now())

	org, err := v.GetOrg(ctx, v.Client, v.Creds.Org)
	if err != nil {
		return nil, err
	}
	v.Objs.Org = org
	vdc, err := v.GetVdc(ctx, v.Creds.VDC)
	if err != nil {
		return nil, err
	}
	v.Objs.Vdc = vdc

	c_capacity := vdc.Vdc.ComputeCapacity
	for _, cap := range c_capacity {
		resources.VCpuMax = uint64(cap.CPU.Limit)
		resources.VCpuUsed = uint64(cap.CPU.Used)
		resources.MemMax = uint64(cap.Memory.Limit)
		resources.MemUsed = uint64(cap.Memory.Used)
	}
	return resources, nil
}

func (v *VcdPlatform) GetResourceID(ctx context.Context, resourceType vmlayer.ResourceType, resourceName string) (string, error) {

	// VM, Subnet and SecGrp are the current potential values of Type
	// The only one we have so far is VMs, (subnets soon, and secGrps eventually)
	if resourceType == vmlayer.ResourceTypeVM {
		vm, err := v.FindVM(ctx, resourceName)
		if err != nil {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		return vm.VM.ID, nil
	} else if resourceType == vmlayer.ResourceTypeSecurityGroup {
		// Get the security Group ID for default XXX
		return "1234", nil
	}
	return "", fmt.Errorf("GetResourceID not implemented for resource type: %s name %s", resourceType, resourceName)
}

func (v VcdPlatform) CheckServerReady(ctx context.Context, client ssh.Client, serverName string) error {

	// ServerName here is really the external ip address.
	// Worker nodes are never checked?
	// Run our Cluster Cider map looking for this IP
	vmName := ""
	if v.Objs.Cloudlet != nil {

		for addr, vm := range v.Objs.Cloudlet.ExtVMMap {
			if serverName == addr {
				vmName = vm.VM.Name
				break
			}
		}
		detail, err := v.GetServerDetail(ctx, vmName)
		if err != nil {
			return err
		}
		if detail.Status == vmlayer.ServerActive {
			out, err := client.Output("systemctl status mobiledgex.service")
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckServerReady Mobiledgex service status", "serverName", serverName, "out", out, "err", err)
			return nil
		} else {
			return fmt.Errorf("Server %s status: %s", serverName, detail.Status)
		}
	}
	return nil
}

func (v *VcdPlatform) GetOrg(ctx context.Context, cli *govcd.VCDClient, orgName string) (*govcd.Org, error) {

	org, err := cli.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("GetOrgByName error %s", err.Error())
	}
	return org, nil
}

func (v *VcdPlatform) GetVdc(ctx context.Context, vdcName string) (*govcd.Vdc, error) {

	vdc, err := v.Objs.Org.GetVDCByName(vdcName, true)
	if err != nil {
		return nil, err
	}
	return vdc, err

}

// return everything else,
func (v *VcdPlatform) GetPlatformResources(ctx context.Context) error {

	var err error
	if v.Objs.Org == nil {
		v.Objs.Org, err = v.GetOrg(ctx, v.Client, v.Creds.Org)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Org Not Found", "Org", v.Creds.Org)
			return fmt.Errorf("Unable to fetch Org %s err: %s", v.Creds.Org, err.Error())
		}

	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Org  as", "Vdc", v.Objs.Org.Org.Name)
	// Get resources under our vdc, give a chance to override.
	// the target vdc via env/property
	vdcName := v.GetPrimaryVdc()
	if vdcName == "" {
		vdcName = v.GetVDCName()
	}
	if v.Objs.Vdc == nil {
		v.Objs.Vdc, err = v.GetVdc(ctx, vdcName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Vdc Not Found", "Vdc", vdcName)
			return err
		}
	}
	vdc := v.Objs.Vdc
	log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Vdc as", "Vdc", vdcName)

	primNet := ""
	if v.TestMode {
		primNet = os.Getenv("MEX_EXT_NETWORK")
	} else {
		primNet = v.GetMexExtNetwork() // ["MEX_EXT_NETWORK"]
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Collecting resources", "vdc", vdc.Vdc.Name)
	// dumpVdcResourceEntities(vdc, 1)
	// fill our maps with bits from our virtual data center object
	nets := vdc.Vdc.AvailableNetworks
	for _, net := range nets {
		for _, ref := range net.Network {
			orgvdcnet, err := vdc.GetOrgVdcNetworkByName(ref.Name, false)
			if err != nil {
				// optional mark as failed and move on? XXX
				return fmt.Errorf("GetOrgVdcNetworkByName %s failed err:%s", ref.Name, err.Error())
			}
			v.Objs.Nets[ref.Name] = orgvdcnet
			if ref.Name == primNet {
				log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Primary", "network", orgvdcnet.OrgVDCNetwork.Name)
				v.Objs.PrimaryNet = orgvdcnet
			}
			v.Objs.Nets[ref.Name] = orgvdcnet
			log.SpanLog(ctx, log.DebugLevelInfra, "Discover:", "network", orgvdcnet.OrgVDCNetwork.Name)
		}
	}
	// cats map
	//
	catalog := &govcd.Catalog{}
	catalogRecords, err := v.Objs.Org.QueryCatalogList()
	if err != nil {
		return err
	} else {
		// Query all Org cats returns a types.CatalogRecord, we want both representations of a catalog
		for n, cat := range catalogRecords {
			orgcat, err := v.Objs.Org.GetCatalogByName(cat.Name, true)
			if err != nil {
				return fmt.Errorf("No org cat for CatRec %s", cat.Name)
			}
			v.Objs.Cats[cat.Name] = CatContainer{
				CatRec: cat,
				OrgCat: orgcat,
			}

			if n == 0 { // xxx first cat found, typically only one but... xxx
				log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Primary", "catalog", orgcat.Catalog.Name)
				v.Objs.PrimaryCat = orgcat
			}
			if len(catalogRecords) > 1 && n == 0 { // j
				log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Multiple catalogs found, add non Primary as ", "catalog", catalog.Catalog.Name)
			}
		}
	}

	// Vapps map
	// Alt. client.QueryVappList: returns a list o all VApps in all the orgainzations available to the caller
	// (returns []*types.QueryResultVAppRecordType, error) So, we'll have to turn around and get the govcd.VApp objects
	//
	// This should be a rtn given res.Type xxx
	for _, r := range vdc.Vdc.ResourceEntities {
		for _, res := range r.ResourceEntity {
			log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Next VDC  Resource:", "Type", res.Type, "Name", res.Name, "Href", "res.HREF")

			if res.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vapp, err := vdc.GetVAppByName(res.Name, false)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover: error fetching Vdc VApp", "Name", res.Name)
				} else {
					a := VApp{
						VApp: vapp,
					}
					v.Objs.VApps[res.Name] = &a
					mdata, err := vapp.GetMetadata()
					if err != nil {
						return err
					}
					for _, data := range mdata.MetadataEntry {
						if data.Key == "CloudletName" {
							log.SpanLog(ctx, log.DebugLevelInfra, "Discover: existing vapp marked cloudlet", "Name", res.Name, "metavalue", data.TypedValue.Value)
							extAddr, err := v.GetExtAddrOfVapp(ctx, vapp, v.Objs.PrimaryNet.OrgVDCNetwork.Name)
							if err != nil {
								if v.Verbose {
									log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Error getting ext addr", "cloudlet", res.Name)
								}
								return err
							}
							v.Objs.Cloudlet = &MexCloudlet{
								ParentVdc: vdc,
								CloudVapp: vapp,
								ExtNet:    v.Objs.PrimaryNet,
								ExtIp:     extAddr,
								Clusters:  make(CidrMap),
								ExtVMMap:  make(CloudVMsMap),
							}
							// getbyname the first vm here
							vm, err := vapp.GetVMByName(vapp.VApp.Children.VM[0].Name, false)
							v.Objs.Cloudlet.ExtVMMap[extAddr] = vm
						}
					}
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover: VApp", "Name", res.Name)
					// now collect any VMs in this Vapp
					if vapp.VApp.Children != nil {
						if v.Verbose {
							log.SpanLog(ctx, log.DebugLevelInfra, "Discover: VApp", vapp.VApp.Name, "child vms", len(vapp.VApp.Children.VM))
						}
						for _, child := range vapp.VApp.Children.VM {
							vm, err := vapp.GetVMByName(child.Name, true)
							if err != nil {
								continue
							} else {
								if v.Verbose {
									log.SpanLog(ctx, log.DebugLevelInfra, "Discover: Adding VM", "Name", vm.VM.Name)
								}
								v.Objs.VMs[vm.VM.Name] = vm
							}
						}
					}
				}
			} else if res.Type == "application/vnd.vmware.vcloud.vms+xml" {
				// VMs
				if v.Verbose {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover Vdc VM resource", "VmName", res.Name, "Href", res.HREF)
				}
				vm, err := v.Client.Client.GetVMByHref(res.HREF)
				if err != nil {
					return err
				} else {
					v.Objs.VMs[res.Name] = vm
				}

				// Templates
			} else if res.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
				if v.Verbose {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover Vdc  resource template", "Name", res.Name, "Href", res.HREF)
				}

				tmpl, err := v.Objs.PrimaryCat.GetVappTemplateByHref(res.HREF)
				if err != nil {
					continue
				} else {
					v.Objs.VAppTmpls[res.Name] = tmpl
				}
				// Media
			} else if res.Type == "application/vnd.vmware.vcloud.media+xml" {
				media, err := v.Objs.PrimaryCat.GetMediaByHref(res.HREF)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover error retrive meida %s from catalog\n", "Catalog", res.Name, "error", err)
				}
				v.Objs.Media[res.Name] = media

			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "Unhandled resource type ignored", "name", res.Name, "Type", res.Type)
			}
		}
	}

	// These are not retreivable by GetVMByHref, These mime types will be templates.
	templateVmQueryRecs, err := v.Client.Client.QueryVmList(types.VmQueryFilterOnlyTemplates)
	for _, qr := range templateVmQueryRecs {
		v.Objs.TemplateVMs[qr.Name] = qr
		log.SpanLog(ctx, log.DebugLevelInfra, "Discover: template", "Name", qr.Name, "type", qr.Type, "Href", qr.HREF)

		targetTemplateName := v.GetVDCTemplateName()
		if qr.Name == targetTemplateName {
			log.SpanLog(ctx, log.DebugLevelInfra, "Discover found VDCTEMPLATE", "Name", qr.Name)
			tmpl, err := v.Objs.PrimaryCat.GetVappTemplateByHref(qr.HREF)
			if err != nil {
				if v.Verbose {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover err getting template by href", "error", err)
				}
			} else {
				// debug issue template not found by at least one test vdc
				v.Objs.Template = tmpl
			}

			tmpls, err := v.GetAllVdcTemplates(ctx)
			if err != nil {
				if v.Verbose {
					log.SpanLog(ctx, log.DebugLevelInfra, "Discover err GetAllVdcTemplates", "error", err)
				}
			} else {
				for _, tmpl := range tmpls {
					if tmpl.VAppTemplate.Name == targetTemplateName {
						v.Objs.Template = tmpl

					}
					v.Objs.VAppTmpls[tmpl.VAppTemplate.Name] = tmpl
					if v.Verbose {
						log.SpanLog(ctx, log.DebugLevelInfra, "Discover add", "template", tmpl.VAppTemplate.Name)
					}
				}
			}
		}
	}
	return nil
}

func (v *VcdPlatform) GetConsoleUrl(ctx context.Context, serverName string) (string, error) {

	return v.Creds.Href, nil
}

func (v *VcdPlatform) ImportImage(ctx context.Context, folder, imageFile string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportImage", "imageFile", imageFile, "folder", folder)
	// first delete anything that may be there for this image
	v.DeleteImage(ctx, folder, imageFile)
	// .ova's are the unit of upload to our catalog
	cat := v.Objs.PrimaryCat
	// ovaFile, itemName, description, uploadPieceSize xxx is folder appropriate for itemName?
	cat.UploadOvf(imageFile, folder+"-tmpl", "mex base iamge", 4*1024)
	return nil
}

func (v *VcdPlatform) DeleteImage(ctx context.Context, folder, image string) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteImage", "image", image)
	// Fetch the folder-tmpl item and call item.Delete()
	// TBI
	return nil
}

// Assuming Vcd is more similar to Vsphere rather than vmPool or OpenStack
// alphanumeric plus -_. first char must be alpha, <= 255 chars.
func (v *VcdPlatform) NameSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "",
		",", "",
		"/", "_",
		"!", "")
	str := r.Replace(name)
	if str == "" {
		return str
	}
	if !unicode.IsLetter(rune(str[0])) {
		// first character must be alpha
		str = "a" + str
	}
	if len(str) > 255 {
		str = str[:254]
	}
	return str
}

// IdSanitize is NameSanitize plus removing "."
func (v *VcdPlatform) IdSanitize(name string) string {
	str := v.NameSanitize(name)
	str = strings.ReplaceAll(str, ".", "-")
	return str
}

func (v *VcdPlatform) GetServerDetail(ctx context.Context, vappName string) (*vmlayer.ServerDetail, error) {
	var vm *govcd.VM
	serverName := vappName
	vappName = serverName + "-vapp"

	log.SpanLog(ctx, log.DebugLevelInfra, "GetServerDetail", "vmname", vappName)

	detail := vmlayer.ServerDetail{}
	vapp, err := v.FindVApp(ctx, vappName)

	if err != nil {
		// Not a vapp, vm?
		vm, err := v.FindVM(ctx, serverName)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "serverName not found ", "serverName", serverName)
			return nil, fmt.Errorf("Server Not found")
		}
		vmStatus, err := vm.GetStatus()
		if err != nil {
			return nil, err
		}
		// replace with block while not status...xxx
		detail.Name = vm.VM.Name
		detail.ID = vm.VM.ID

		if vmStatus == "POWERED_ON" {
			detail.Status = vmlayer.ServerActive
		} else if vmStatus == "POWERED_OFF" {
			detail.Status = vmlayer.ServerShutoff
		} else {
			detail.Status = vmStatus
		}

		addresses, _, err := v.GetVMAddresses(ctx, vm)
		if err != nil {
			return nil, err
		}
		detail.Addresses = addresses
		// Ok, so the govcd.VM has a vm.GetStatus returning a string, while the vm.VM has a int field status (resource status)
		return &detail, nil

	} else { // do the cloudlet
		if vapp.VApp.Children == nil {
			return nil, fmt.Errorf("VApp %s has no vms\n", vappName)
			// this is so wrong, the VApp state would be "RESOLVED" here.
		}
		vmname := vapp.VApp.Children.VM[0].Name
		vm, err = v.FindVMInVApp(ctx, vmname, *vapp)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Vm not found in vapp", "vmName", vmname)
			return &detail, err
		}
		vmStatus, err := vm.GetStatus()
		if err != nil {
			return nil, err
		}
		detail.Name = vm.VM.Name
		detail.ID = vm.VM.ID
		if vmStatus == "POWERED_ON" {
			detail.Status = vmlayer.ServerActive
		} else if vmStatus == "POWERED_OFF" {
			detail.Status = vmlayer.ServerShutoff
		} else {
			detail.Status = vmStatus
		}
		addresses, ip, err := v.GetVMAddresses(ctx, vm)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "GetAddresses failed for vm", "vmName", detail.Name, "ip", ip)
			return &detail, err
		}
		detail.Addresses = addresses
	}
	return &detail, nil
}

// Given a vappName, does it exist in our vdc?
func (v *VcdPlatform) FindVdcVapp(ctx context.Context, vappName string) (*govcd.VApp, error) {
	vdc := v.Objs.Vdc
	vappRefs := vdc.GetVappList()
	for _, ref := range vappRefs {
		vapp, err := v.Objs.Vdc.GetVAppByName(ref.Name, false)
		if err != nil {
			continue
		}
		if ref.Name == vapp.VApp.Name {
			vapp, err := vdc.FindVAppByName(vappName)
			if err != nil {
				continue
			}
			return &vapp, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

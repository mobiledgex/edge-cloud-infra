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
	// Properties common to all VM providers
	vmProperties *vmlayer.VMProperties
	vcdVars      map[string]string
	TestMode     bool
	caches       *platform.Caches
	Creds        *VcdConfigParams
	Client       *govcd.VCDClient
	Objs         VcdObjects
}

type VcdConfigParams struct {
	User     string
	Password string
	Org      string
	Href     string
	VDCName  string // []string xxx multiple vdc/org potential
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
type VmNet struct {
	vmName string
	vmRole string
	vmMeta []string
	vm     *govcd.VM
}

// IPaddr + vm attributes per cluster
type VMIPsMap map[string]VmNet

// A map key'ed by CIDR whose value is another map of all VMs in the
// Cluster represented by this CIDR, this key'ed by IP addr
// This set of vms under this CIDR represent a cluster

type CidrMap map[string]VMIPsMap

// One cloudlet per vdc instance
type MexCloudlet struct {
	ParentVdc    *govcd.Vdc
	CloudVapp    *govcd.VApp
	CloudletName string
	Clusters     CidrMap // Clusters are keyed by their internal net CIDR
	// federation partner TBI (single remote org/vdc:  a pair wise assocication)

}

// cloudletName
type VdcCloudlets map[string]*MexCloudlet

type VcdObjects struct {
	Org       *govcd.Org
	Vdcs      VdcMap
	Nets      NetMap
	Cats      CatMap
	VApps     VAppsMap
	VAppTmpls VAppTmplMap
	// while we'll discover all external networks
	// avaliable to our vdc, we'll only utilize the first we find as
	// v.Objs.PrimaryNet
	PrimaryVdc  *govcd.Vdc
	PrimaryNet  *govcd.OrgVDCNetwork
	PrimaryCat  *govcd.Catalog
	VMs         VMMap
	DeployedVMs VMMap
	TemplateVMs TmplVMsMap
	EdgeGateway govcd.EdgeGateway
	Media       MediaMap
	Cloudlets   VdcCloudlets
}

func (v *VcdPlatform) GetType() string {
	return "vcd"
}

func (v *VcdPlatform) InitProvider(ctx context.Context, caches *platform.Caches, stage vmlayer.ProviderInitStage, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "InitProvider for Vcd 1", "stage", stage)
	v.InitData(ctx, caches)
	// XXX read env creds for now, vault soon
	v.PopulateOrgLoginCredsFromEnv(ctx, "mex-cldlet1") // need to move to first physicalname reference (vault key lookup not env)
	//v.initDebug(o.VMProperties.CommonPf.PlatformConfig.NodeMgr) // XXX needed now?

	// make our object maps
	v.Objs.Vdcs = make(VdcMap)
	v.Objs.Nets = make(map[string]*govcd.OrgVDCNetwork)
	v.Objs.Cats = make(map[string]CatContainer)
	v.Objs.VApps = make(map[string]*VApp)
	v.Objs.VMs = make(map[string]*govcd.VM)
	v.Objs.VAppTmpls = make(map[string]*govcd.VAppTemplate)
	v.Objs.TemplateVMs = make(map[string]*types.QueryResultVMRecordType)
	v.Objs.Media = make(MediaMap)
	v.Objs.Cloudlets = make(VdcCloudlets)

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
		fmt.Printf("Error from SetProviderSpecificProps: %s\n", err.Error())
		return err
	}
	return nil
}

func (v *VcdPlatform) InitData(ctx context.Context, caches *platform.Caches) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitProvider::SetCaches 2")
	v.caches = caches
}

func (v *VcdPlatform) ImportDataFromInfra(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "ImportDataFromInfra N")
	if v.Client == nil {
		fmt.Printf("\n\nImportDataFromInfra-I-v.Client nil, login first time\n\n")
		client, err := v.GetClient(ctx, v.Creds)
		if err != nil {
			return fmt.Errorf("Unable to create Vcd Client %s\n", err.Error())
		}
		v.Client = client
	}

	err := v.GetComputeResources(ctx)
	if err != nil {
		return fmt.Errorf("Error retrieving Compute Resources: %s", err.Error())
	}

	err = v.GetPlatformResources(ctx)
	if err != nil {
		return fmt.Errorf("Error retrieving Platform  Resources: %s", err.Error())
	}
	return nil
}

func (v *VcdPlatform) GetPlatformResourceInfo(ctx context.Context) (*vmlayer.PlatformResources, error) {

	var resources *vmlayer.PlatformResources
	log.SpanLog(ctx, log.DebugLevelInfra, "GetPlatformResourceInfo N ")

	resources.CollectTime, _ = gogotypes.TimestampProto(time.Now())

	org, err := v.GetOrg(ctx, v.Client, v.Creds.Org)
	if err != nil {
		fmt.Printf("\n\nGetPlatformResourceInfo-E-GetOrg %s returns %s\n", v.Creds.Org, err.Error())
		return nil, err
	}

	vdc, err := v.GetVdc(ctx, v.Client, org, v.Creds.VDCName) // change not to need VDC apriori dig out of org
	if err != nil {
		fmt.Printf("\n\nGetPlatformResourceInfo-E-GetOrg returns %s\n", err.Error())
		return nil, err
	}

	c_capacity := vdc.Vdc.ComputeCapacity
	fmt.Printf("\n\nGetPlatformResourceInfo Vdc.ComputeCapacity : len %d  %+v\n\n", len(c_capacity), c_capacity)
	for _, cap := range c_capacity {

		// so we get vdc from our Org with refresh true.
		resources.VCpuMax = uint64(cap.CPU.Limit)
		resources.VCpuUsed = uint64(cap.CPU.Used)
		resources.MemMax = uint64(cap.Memory.Limit)
		resources.MemUsed = uint64(cap.Memory.Used)
	}
	/*
	   type ResourceEntities struct {
	   	ResourceEntity []*ResourceReference `xml:"ResourceEntity,omitempty"`
	   }
	      need to dig out how much disk we can allocate
	*/

	// sets PrimaryNet also
	err = v.GetPlatformResources(ctx)
	if err != nil {
		fmt.Printf("\nGetPltformREsourceInfo-I-failed error: %s\n", err.Error())
		return nil, nil
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

	detail, err := v.GetServerDetail(ctx, serverName)
	if err != nil {
		fmt.Printf("CheckServerReady-E-from GetServerDetail: %s\n", err.Error())
		return err
	}
	if detail.Status == vmlayer.ServerActive {
		out, err := client.Output("systemctl status mobiledgex.service")
		log.SpanLog(ctx, log.DebugLevelInfra, "CheckServerReady Mobiledgex service status", "serverName", serverName, "out", out, "err", err)
		return nil
	} else {
		fmt.Printf("\nCheckServerReady-E-detail.Status := %s (not ServerActive) \n\n", detail.Status)
		return fmt.Errorf("Server %s status: %s", serverName, detail.Status)
	}
}

var qcowConvertTimeout = 5 * time.Minute

func (v *VcdPlatform) GetOrg(ctx context.Context, cli *govcd.VCDClient, orgName string) (*govcd.Org, error) {

	org, err := cli.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("GetOrgByName error %s", err.Error())
	}
	// now, org.Org is it? where org is from system, and Org is from types ?  Right?
	fmt.Printf("GetOrgByName returns org org.Org.HREF: %s\n", org.Org.HREF)

	return org, nil
}

func (v *VcdPlatform) GetVdc(ctx context.Context, cli *govcd.VCDClient, org *govcd.Org, vdcName string) (*govcd.Vdc, error) {

	vdc, err := org.GetVDCByName(vdcName, true)
	if err != nil {
		fmt.Printf("Unable to retrieve vdc by name err: %s\n", err.Error())
	}
	return vdc, err

}

// return cpu mem disk quota (discard current usage)
func (v *VcdPlatform) GetComputeResources(ctx context.Context) error {
	// if we have yet to fetch our org and vdc object do it now, we should have a non-nil client
	// withwhich to make the query. We know we have v.Client as we're called from Import Data from infra
	var err error

	if v.Objs.Org == nil {
		v.Objs.Org, err = v.GetOrg(ctx, v.Client, v.Creds.Org)
		if err != nil {
			return fmt.Errorf("Unable to fetch Org %s err: %s", v.Creds.Org, err.Error())
		}

	}
	return nil
}

// return everything else,
func (v *VcdPlatform) GetPlatformResources(ctx context.Context) error {

	var err error
	if v.Objs.Org == nil {
		//fmt.Printf("\n\n GetComputeResources N Initial Fetch Org\n\n")
		v.Objs.Org, err = v.GetOrg(ctx, v.Client, v.Creds.Org)
		if err != nil {
			return fmt.Errorf("Unable to fetch Org %s err: %s", v.Creds.Org, err.Error())
		}

	}
	// We need all Org vdcs and their resources and add each
	// If we have only tenant privs, we'll not be able to retrieve the adminOrg to find the constituent vdcs...XXX
	// So we'd need to get Vdc name from config and use the single VDC we do have access to.
	primVdc := os.Getenv("PRIMARY_VDC")
	if len(v.Objs.Vdcs) == 0 {
		// look for mime type "application/vnd.vmware.vcloud.vdc+xml"
		adminOrg, err := govcd.GetAdminOrgByName(v.Client, v.Creds.Org)
		if err != nil {
			return fmt.Errorf("Unable to fetch adminOrg by name %s err: %s", v.Creds.Org, err.Error())
		}

		vdcList := adminOrg.AdminOrg.Vdcs
		for _, vdcRef := range vdcList.Vdcs {
			vdc, err := v.Objs.Org.GetVdcByName(vdcRef.Name)
			if err != nil {
				fmt.Printf("\n\nFailed to fetch Org.Vdc name: %s err: %s\n", vdcRef.Name, err.Error())
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "add Org.", "Vdc", vdcRef.Name)
				v.Objs.Vdcs[vdcRef.Name] = &vdc
				if vdcRef.Name == primVdc {
					fmt.Printf("\n\tDiscover-I-setting PrimaryVdc to %s\n", vdcRef.Name)
					v.Objs.PrimaryVdc = &vdc
				}
			}
		}
	}

	// we may have a perference of what vdc to use as our Primary Vdc
	// Check if we have a clue mex-qe test vdc is one that may be set.

	primNet := os.Getenv("MEX_EXT_NETWORK")
	for _, vdc := range v.Objs.Vdcs {
		fmt.Printf("Discover: Collecting resources of vdc: %s\n", vdc.Vdc.Name)
		// dumpVdcResourceEntities(vdc, 1)
		// fill our maps with bits from our virtual data center object
		nets := vdc.Vdc.AvailableNetworks
		for _, net := range nets {
			for n, ref := range net.Network {
				orgvdcnet, err := vdc.GetOrgVdcNetworkByName(ref.Name, false)
				if err != nil {
					// optional mark as failed and move on? XXX
					return fmt.Errorf("GetOrgVdcNetworkByName %s failed err:%s", ref.Name, err.Error())
				}
				v.Objs.Nets[ref.Name] = orgvdcnet
				if ref.Name == primNet {
					fmt.Printf("\nDiscover-I-PrimaryNet = %s n=%d \n", orgvdcnet.OrgVDCNetwork.Name, n)
					log.SpanLog(ctx, log.DebugLevelInfra, "Primary", "network", orgvdcnet.OrgVDCNetwork.Name)
					v.Objs.PrimaryNet = orgvdcnet
				} else {
					fmt.Printf("\nDiscover VDCOrgNetwork %s\n", orgvdcnet.OrgVDCNetwork.Name)
				}
			}
		}
		// cats map
		//
		catalog := &govcd.Catalog{}
		catalogRecords, err := v.Objs.Org.QueryCatalogList()
		if err != nil {
			//fmt.Printf("QueryCatalogList-E-returns : %s ignoring\n", err.Error())
			//spanlog
			// ignor  e
		} else {
			// Query all Org cats returns a types.CatalogRecord, we want both  representations of a catalog
			for n, cat := range catalogRecords {
				orgcat, err := v.Objs.Org.GetCatalogByName(cat.Name, true)
				if err != nil {
					fmt.Printf("GetPlatformResource-E-catRecord Name finds no govcd Catalog by name %s\n", cat.Name)
					return fmt.Errorf("No org cat for CatRec %s", cat.Name)
				}
				v.Objs.Cats[cat.Name] = CatContainer{
					CatRec: cat,
					OrgCat: orgcat,
				}
				if n == 0 {
					fmt.Printf("GetPlatformResources-I-PrimaryCat set as %s\n", orgcat.Catalog.Name)
					v.Objs.PrimaryCat = orgcat
					fmt.Printf("\nvalidate: GetPlatformResources: v.Objs.PrimaryCat.Catalog.Name: %s\n", v.Objs.PrimaryCat.Catalog.Name)

				}
				if len(catalogRecords) > 1 && n == 0 { // j
					log.SpanLog(ctx, log.DebugLevelInfra, "Multiple catalogs found, using Primary as ", "catalog", catalog.Catalog.Name)

				}

			}
		}

		// Vapps map
		// Alt. client.QueryVappList: returns a list o all VApps in all the orgainzations available to the caller
		// (returns []*types.QueryResultVAppRecordType, error) So, we'll have to turn around and get the govcd.VApp objects
		//
		// This should be a rtn given res.Type
		for _, r := range vdc.Vdc.ResourceEntities {
			for _, res := range r.ResourceEntity {

				fmt.Printf("Discover-I-Next VDC  Resource:\n\tType \t%s\n\tName\t%s\n\tHREF: %s\n",
					res.Type, res.Name, res.HREF)

				if res.Type == "application/vnd.vmware.vcloud.vApp+xml" {
					vapp, err := vdc.GetVAppByName(res.Name, true)
					if err != nil {
						fmt.Printf("\n Error GetVAppbyName for %s err: %s\n", res.Name, err.Error())
						// spanlog
					} else {
						a := VApp{
							VApp: vapp,
						}
						v.Objs.VApps[res.Name] = &a
						fmt.Printf("Discover: Added VApp %s to vapps map\n", res.Name)
						// now collect any VMs in this Vapp
						if vapp.VApp.Children != nil {
							fmt.Printf("Vapp %s has %d child VMs\n", vapp.VApp.Name, len(vapp.VApp.Children.VM))
							for _, child := range vapp.VApp.Children.VM {
								vm, err := vapp.GetVMByName(child.Name, true)
								if err != nil {
									fmt.Printf("error GetByName for %s skipping err: %s \n", child.Name, err.Error())
									continue
								} else {
									fmt.Printf("\tAdding vapp vm %s\n", vm.VM.Name)
									v.Objs.VMs[vm.VM.Name] = vm
								}
							}
						}
					}
					// VMs
				} else if res.Type == "application/vnd.vmware.vcloud.vms+xml" {

					fmt.Printf("\n########## Discover-I-found Vdc resource VmName: %s VmHref %s\n", res.Name, res.HREF)

					vm, err := v.Client.Client.GetVMByHref(res.HREF)
					if err != nil {
						fmt.Printf("Disover-I-GetVappTemplateyByHref: %s\n", err.Error())
					} else {
						fmt.Printf("\tAdding vm named: %s\n", vm.VM.Name)
						v.Objs.VMs[res.Name] = vm
					}

					// So that typically fails, but with our lastest pull we have new per version calls:
					// Using these (vm.go) you pass both the client and href into the call

					// Templates
				} else if res.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
					tmpl, err := v.Objs.PrimaryCat.GetVappTemplateByHref(res.HREF)
					if err != nil {
						continue
					} else {
						fmt.Printf("\tAdding template %s to local cache\n", tmpl.VAppTemplate.Name)
						v.Objs.VAppTmpls[res.Name] = tmpl
					}
					// Media
				} else if res.Type == "application/vnd.vmware.vcloud.media+xml" {
					media, err := v.Objs.PrimaryCat.GetMediaByHref(res.HREF)
					if err != nil {
						fmt.Printf("Discover-E-retrive meida %s from catalog %s\n", res.Name, v.Objs.PrimaryCat.Catalog.Name)
					}
					fmt.Printf("\tAdding media %s to local meida cache\n", res.Name)
					v.Objs.Media[res.Name] = media

				} else {
					fmt.Printf("Unhandled resource type %s name: %s  ignored\n", res.Type, res.Name)
				}
			}
		}
	}
	// These are not retreivable by GetVMByHref, These mime types will be templates.
	// The Vapp VMs are caputure above
	templateVmQueryRecs, err := v.Client.Client.QueryVmList(types.VmQueryFilterOnlyTemplates)
	for _, qr := range templateVmQueryRecs {
		v.Objs.TemplateVMs[qr.Name] = qr
		//fmt.Printf("\nDiscover: found Template VM named %s type %s HREF: %s\n",
		//	qr.Name, qr.Type, qr.HREF)
	}
	return nil
}

// GetClient in vcd-security for whatever reason

// orignally sourced from vault using physical name from CreateCloudlet as key
// temp, use env vars
func (v *VcdPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {

	return v.Creds.Href, nil

}

func (v *VcdPlatform) GetConsoleUrl(ctx context.Context, serverName string) (string, error) {
	fmt.Printf("GetConsoleUrl  TBI\n")
	return "", nil
}

func (v *VcdPlatform) ImportImage(ctx context.Context, folder, imageFile string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportImage", "imageFile", imageFile, "folder", folder)
	// first delete anything that may be there for this image
	v.DeleteImage(ctx, folder, imageFile)
	// .ova's are the unit of upload to our catalog
	cat := v.Objs.PrimaryCat
	// ovaFile, itemName, description, uploadPieceSize xxx is folder appropriate for itemName?
	// Likely want -tmpl append xxx
	cat.UploadOvf(imageFile, folder+"-tmpl", "mex base iamge", 1024)

	return nil
}

func (v *VcdPlatform) DeleteImage(ctx context.Context, folder, image string) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteImage", "image", image)
	// Fetch the folder-tmpl item and call item.Delete()
	fmt.Printf("DeleteImage-TBI\n")
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

/*
vmlayer.server.go
type ServerDetail struct {
	Addresses []ServerIP
	ID        string
	Name      string
	Status    string
}
*/

// Currently, the external-network-name is set in env var. We _could_ just add
// meta data to our primary external net that caches this, and give it back if needed.
func (v *VcdPlatform) GetServerDetail(ctx context.Context, vappName string) (*vmlayer.ServerDetail, error) {
	var vm *govcd.VM
	serverName := vappName
	vappName = serverName /*+ "-vapp" xxx 11/04 mfw nneded? */

	log.SpanLog(ctx, log.DebugLevelInfra, "PI GetServerDetail 4", "vmname", vappName)

	fmt.Printf("GetServerDetail-I-asking for server name %s vm or vapp?\n", vappName)

	detail := vmlayer.ServerDetail{}
	vapp, err := v.FindVApp(ctx, vappName)

	if err != nil {
		fmt.Printf("No Vapp found for Vapp name: %s\n", vappName)
		// not found? return "Server does not exist"
		// will this trigger us to create a new VM
		return &detail, fmt.Errorf(vmlayer.ServerDoesNotExistError)
	} else {
		if vapp.VApp.Children == nil {
			return nil, fmt.Errorf("VApp %s has no vms\n", vappName)
			// this is so wrong, the VApp state would be "RESOLVED" here.
		}
		vmname := vapp.VApp.Children.VM[0].Name
		vm, err = v.FindVMInVApp(ctx, vmname, *vapp)
		if err != nil {
			fmt.Printf("Not finding %s in our local cache\n", vmname)

		}

		vmStatus, err := vm.GetStatus()
		if err != nil {
			fmt.Printf("vm.GetStatus failed %s\n", err.Error())
			return nil, err
		} else {

			fmt.Printf("GetServerDetail-I-vm %s has status %s\n", vm.VM.Name, vmStatus)
		}
		if vmStatus != "POWERED_ON" {
			for i := 0; i < 10; i++ {
				time.Sleep(time.Second * 10)
				vmStatus, err = vm.GetStatus()
				fmt.Printf("\tGetServerDetail-I-status : %s\n", vmStatus)
				if vmStatus == "POWERED_ON" {
					fmt.Printf("GetServerDetail-I-server %s now state %s\n", vmname, vmStatus)
					break
				}
			}
		}
		if vmStatus != "POWERED_ON" {
			fmt.Printf("GetServerDetail-E-Timeout vm %s state: %s\n", vm.VM.Name, vmStatus)
			return nil, fmt.Errorf("error waiting for VM to come ready")
		}
		// we may need to wait a tiny bit more for DHCP to catch  up
		// fill in ServerDetail from our vm
		detail.Name = vm.VM.Name
		detail.ID = vm.VM.ID
		detail.Status = "ACTIVE"
		addresses, ip, err := v.GetVMAddresses(ctx, vm)
		if err != nil {
			fmt.Printf("Error fetchng addresses from vm %s err: %s\n", vm.VM.Name, err.Error())
		}
		// We will use DHCP so if we don't have it yet... This is some waitForDHCP call somewhere I've seen XXX
		fmt.Printf("GetServerDetail-I-retrieved server %s IP as %s\n", vm.VM.Name, ip)
		detail.Addresses = addresses
		// Ok, so the govcd.VM has a vm.GetStatus returning a string, while the vm.VM has a int field status (resource status)
		if vm.VM.Status == 8 {
			//detail.Status = "Resolved Powered Off"
			fmt.Printf("And the vm.VM.status value : %d\n", vm.VM.Status)

		}

	}
	fmt.Printf("\nGetServerDetail-I-found existing VApp %s xlate to vmlayer.ServerDetail vm: %+v\n", vappName, vm)
	return &detail, nil
}

// Return the cloudlet this name references
// We're looking for the existing vapp(cloudlet) that this GroupName specifies it's to be created on.
// Don't return vapp, it's in the cloudlet object
func (v *VcdPlatform) FindCloudletForCluster(GroupName string) (*MexCloudlet, *govcd.VApp, error) {

	targetVapp := &govcd.VApp{}
	fmt.Printf("FindVappForCloudlet-I-looking for vappName who cloudlet name is contained within : %s\n", GroupName)

	// First check to see if it's a duplicate create cloudlet call
	for name, vdcCloudlet := range v.Objs.Cloudlets {
		if GroupName == name {
			// Test creating cloudlet twice in a row returns this error XXX
			fmt.Printf("Ok, we have the cloudlet in question, and we're being asked to create it again return exists \n")
			return nil, nil, fmt.Errorf("Cloudlet %s already exists\n", name)
		}

		// We've stored the CloudletName in our vdcCloudlet object, The vapp name stripped of it's operator and mex.net bits.

		fmt.Printf("FindVappForCloudlet: is %s found in %s?\n", vdcCloudlet.CloudletName, GroupName)
		if strings.Contains(GroupName, vdcCloudlet.CloudletName) {
			fmt.Printf("\nFindVappForCloudlet CreateVMs Selecting existing\n\tClouldlet %s\n\t vapp  %s\n\tvdc: %s\n for adding vms in %s\n",
				vdcCloudlet.CloudletName,
				vdcCloudlet.CloudVapp.VApp.Name,
				vdcCloudlet.ParentVdc.Vdc.Name,
				GroupName)

			targetVapp = vdcCloudlet.CloudVapp
			return vdcCloudlet, targetVapp /*vdcCloudlet.ParentVdc,*/, nil

		} else {
			fmt.Printf("\tSkipped vapp: %s \n", vdcCloudlet.CloudVapp.VApp.Name)
			continue
		}
	}
	return nil, nil, fmt.Errorf("Not found")
}

// Given a vappName, does it exist in any vdcs?
func (v *VcdPlatform) FindVdcVapp(ctx context.Context, vappName string) (*govcd.Vdc, *govcd.VApp, error) {

	for _, vdc := range v.Objs.Vdcs {
		vappRefs := vdc.GetVappList()
		for _, ref := range vappRefs {
			vapp, err := vdc.GetVAppByName(ref.Name, false)
			if err != nil {
				fmt.Printf("\nFindVdcVapp-I-GetVAppByName %s of vdc %s failed: %s\n", ref.Name, vdc.Vdc.Name, err.Error())
				continue
			}
			if ref.Name == vapp.VApp.Name {
				vapp, err := vdc.FindVAppByName(vappName)
				if err != nil {
					continue
				}
				return vdc, &vapp, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("Not found")
}

func (v *VcdPlatform) FindVdcParent(ctx context.Context, vapp *govcd.VApp) (*govcd.Vdc, error) {
	for _, vdc := range v.Objs.Vdcs {
		vappRefs := vdc.GetVappList()
		for _, ref := range vappRefs {
			if ref.Name == vapp.VApp.Name {
				return vdc, nil
			}
		}
	}
	return nil, fmt.Errorf("Not found")
}

// We only allow one cloudlet per vdc. This returns the first vdc found that curently has no
// cloudlet VApp, (or none in the powered on state?)
func (v *VcdPlatform) GetNextAvailableVdc(ctx context.Context) (*govcd.Vdc, error) {

	fmt.Printf("GetNextAvailableVdc-I-have %d vdcs in org %s\n", len(v.Objs.Vdcs), v.Objs.Org.Org.Name)
	for vdcName, vdc := range v.Objs.Vdcs {
		fmt.Printf("\tSearching vdc %s for any vapps\n", vdcName)

		vappRefs := vdc.GetVappList()
		if len(vappRefs) == 0 {
			fmt.Printf("\n\nGetNextAvailableVdc-I-vdc %s has no VApps currently use it for new cloudlet\n", vdcName)
			return vdc, nil
		}
		fmt.Printf("\tvdc %s has %d vapps we consider\n", vdcName, len(vappRefs))
		var available = true
		for n, ref := range vappRefs {
			fmt.Printf("\t%d : name: %s\n\t\t HREF: %s\n\t\tType: %s\n\t\t Status: %s \n",
				n, ref.Name, ref.HREF, ref.Type, ref.Status)

			// Minor confusing detail, we find 4 vapps in our test results, even though
			// in the console, there are only 2 VApps defined. Why? Because all VMs have a parent VApp.
			// So if you have a VM defined that is currently not part of any VApp, a parent VApp is created
			// for it.
			//
			// Here's two example ref bits, the first is a valid VApp:
			// 0 : name: test-vcd2-1
			//     HREF: https://10.70.2.71/api/vApp/vapp-3c94c6cf-1ee1-4487-a82e-6169466daa8c
			//     Type: application/vnd.vmware.vcloud.vApp+xml
			//
			// Which is correct, the vm inside this is vcd2-vm
			// But this one:
			//
			// 1 : name: QA-webvm-45d13550-7afa-4621-b002-774dde5f5795
			//     HREF: https://10.70.2.71/api/vApp/vapp-078e80ca-5cc3-4e09-9acc-829eb5873871
			//     Type: application/vnd.vmware.vcloud.vApp+xml
			//
			// This is a manufactured VApp that containes the "standalone" vm named: QA-webvm
			// that is not part of any VApp currently.

			// this returns vms vapp, err := vdc.GetVAppByHref(ref.HREF)
			vapp, err := vdc.GetVAppByNameOrId(ref.ID, true)
			if err != nil {
				fmt.Printf("\tGetVAppByHREF failed vdc: %s err: %s\n", vdcName, err.Error())
				continue
			}

			// Get all VApps in this vdc and check their state
			// For now, since we're sharing this vdc, skip looking at any vapp with "qa" in its name.
			// look for deployed vapps, if any to to the next one.
			// ok, here's a thought: If these are really vms, then we should be able to find it's parent
			// No, this doesn't compile since it thinks this is a vapp

			if strings.Contains(ref.Name, "qa") || strings.Contains(ref.Name, "QA") {
				fmt.Printf("GetNextAvailableVdc-I-skipping qa vapp %s\n", vdcName)
				continue
			}
			// For now (dev) allow other cloudlets on this vdc as long as they're other than Powered On
			status, err := vapp.GetStatus()
			if status == "POWERED_ON" {
				fmt.Printf("\tvapp %s has status %s\n", vapp.VApp.Name, status)
				available = false
				break
			} else {
				continue
			}

		}
		if available {
			fmt.Printf("\treturning vdc: %s\n", vdc.Vdc.Name)
			return vdc, nil
		}
	}
	return nil, fmt.Errorf("No available Vdc for new Cloudlet")
}

func (v *VcdPlatform) FindVdc(ctx context.Context, vdcName string) (*govcd.Vdc, error) {

	for name, vdc := range v.Objs.Vdcs {
		if name == vdcName {
			return vdc, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

func (v *VcdPlatform) GetVdcNames(ctx context.Context) ([]string, error) {
	vdcs := []string{}
	adminOrg, err := govcd.GetAdminOrgByName(v.Client, v.Creds.Org)
	if err != nil {
		return vdcs, fmt.Errorf("Unable to fetch adminOrg by name %s err: %s", v.Creds.Org, err.Error())
	}

	vdcList := adminOrg.AdminOrg.Vdcs
	for _, vdcRef := range vdcList.Vdcs {
		vdcs = append(vdcs, vdcRef.Name)
	}
	return vdcs, nil
}

// Given our scheme for networks 10.101.X.0/24 return the next available Isolated network CIDR
func (v *VcdPlatform) GetNextInternalNet(ctx context.Context, cloudlet *MexCloudlet) (string, error) {
	var MAX_CIDRS = 10 // implies a limit MAX_CIDRS  clusters per Cloudlet. XXX
	// run our current cloudlet.IsoNetMap and either return the first hole in X space,
	// or add a new X at the end if no holes.
	numCloudlets := len(v.Objs.Cloudlets)
	if numCloudlets == 0 {
		// In case we are creating platform VM,
		return "10.101.1.0/24", nil
	}
	if cloudlet == nil {
		return "", fmt.Errorf("Invaild argument")
	}
	fmt.Printf("GetNext-I-have %d cloudlets\n", numCloudlets)
	// we wish to avoid the zero value (why?)
	next := 0
	for n := 1; n < MAX_CIDRS; n++ {
		taddr := fmt.Sprintf("%s.%s.%d.%s", "10", "101", n, "0/24")
		//fmt.Printf("Testing %s for existance\n", taddr)

		for addr, mexCloud := range cloudlet.Clusters {
			fmt.Printf("\tCloudlet addr : %s\n", addr)
			// if our map has this cidr continue
			//x, err := v.ThirdOctet(ctx, addr)
			if len(cloudlet.Clusters[taddr]) == 0 { //  == nil {
				// use this one

				fmt.Printf("\n\t addr %s is unused return it\n\n", addr)
				return addr, nil
			} else {
				fmt.Printf("cloudlet with addr %s non nil as: %+v\n", addr, mexCloud)
			}
		}
	}
	// Reached the end, add a new one
	next++
	addr := fmt.Sprintf("%s.%s.%d.%s", "10", "101", next, "0/24")
	cloudlet.Clusters[addr] = VMIPsMap{}
	return addr, nil
}

package mexos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type templateFill struct {
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone              string
	ImageFlavor, Location, RootLB, Resource, ResourceKind, ResourceGroup string
	StorageSpec, NetworkScheme, MasterFlavor, Topology                   string
	NodeFlavor, Operator, Key, Image, Options                            string
	ImageType, AppURI, ProxyPath, AgentImage                             string
	ExternalNetwork, Project, AppTemplate                                string
	ExternalRouter, Flags, IPAccess                                      string
	NumMasters, NumNodes                                                 int
	ConfigDetailDeployment, ConfigDetailResources, ConfigDetailManifest  string
	Command                                                              []string
}

var yamlMEXCluster = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  project: {{.Project}}
  dnszone: {{.DNSZone}}
  resourcegroup: {{.ResourceGroup}}
spec:
  flags: {{.Flags}}
  flavor: {{.Flavor}}
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  rootlb: {{.RootLB}}
  networkscheme: {{.NetworkScheme}}
`

var yamlMEXFlavor = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
spec:
  flags: {{.Flags}}
  flavor: {{.Name}}
  flavordetail: 
     name: {{.Name}}
     nodes: {{.NumNodes}}
     masters: {{.NumMasters}}
     networkscheme: {{.NetworkScheme}}
     masterflavor: {{.MasterFlavor}}
     nodeflavor: {{.NodeFlavor}}
     storagescheme: {{.StorageSpec}}
     topology: {{.Topology}}
`

var yamlMEXPlatform = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  openrc: ~/.mobiledgex/openrc
  dnszone: {{.DNSZone}}
  operator: {{.Operator}}
spec:
  flags: {{.Flags}}
  flavor: {{.Flavor}}
  rootlb: {{.RootLB}}
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  externalnetwork: {{.ExternalNetwork}}
  networkscheme: {{.NetworkScheme}}
  externalrouter: {{.ExternalRouter}}
  options: {{.Options}}
  agent: 
    image: {{.AgentImage}}
    status: active
`

var yamlMEXApp = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  dnszone: {{.DNSZone}}
config:
  kind:
  source:
  detail:
    resources: {{.ConfigDetailResources}}
    deployment: {{.ConfigDetailDeployment}}
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  image: {{.Image}}
  imagetype: {{.ImageType}}
  imageflavor: {{.ImageFlavor}}
  proxypath: {{.ProxyPath}}
  flavor: {{.Flavor}}
  uri: {{.AppURI}}
  ipaccess: {{.IPAccess}}
  networkscheme: {{.NetworkScheme}}
  kubemanifesttemplate: {{.AppTemplate}}
  command:
{{- range .Command}}
  - {{.}}
{{- end}}
`

func fillPlatformTemplateCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) (*Manifest, error) {
	log.DebugLog(log.DebugLevelMexos, "fill template cloudletkeystr", "cloudletkeystr", cloudletKeyStr)
	clk := edgeproto.CloudletKey{}
	err := json.Unmarshal([]byte(cloudletKeyStr), &clk)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal json cloudletkey %s, %v", cloudletKeyStr, err)
	}
	log.DebugLog(log.DebugLevelMexos, "unmarshalled cloudletkeystr", "cloudletkey", clk)
	if clk.Name == "" || clk.OperatorKey.Name == "" {
		log.DebugLog(log.DebugLevelMexos, "will not fill template with invalid cloudletkeystr", "cloudletkeystr", cloudletKeyStr)
		return nil, fmt.Errorf("invalid cloudletkeystr %s", cloudletKeyStr)
	}

	log.DebugLog(log.DebugLevelMexos, "using external network", "extNet", GetMEXExternalNetwork(rootLB.PlatConf))

	data := templateFill{
		ResourceKind:    "platform",
		Resource:        util.K8SSanitize(clk.OperatorKey.Name),
		Name:            clk.Name,
		Tags:            clk.Name + "-tag",
		Key:             clk.Name + "-" + util.K8SSanitize(clk.OperatorKey.Name),
		Flavor:          "x1.medium",
		Operator:        util.K8SSanitize(clk.OperatorKey.Name),
		Location:        "bonn",
		Region:          "eu-central-1",
		Zone:            "eu-central-1c",
		RootLB:          rootLB.Name,
		AgentImage:      "registry.mobiledgex.net:5000/mobiledgex/mexosagent",
		Kind:            "mex-platform",
		ExternalNetwork: GetMEXExternalNetwork(rootLB.PlatConf),
		NetworkScheme:   "priv-subnet,mex-k8s-net-1,10.101.X.0/24",
		DNSZone:         "mobiledgex.net",
		ExternalRouter:  "mex-k8s-router-1",
		Options:         "dhcp",
	}
	mf, err := templateUnmarshal(&data, yamlMEXPlatform)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func fillAppTemplate(rootLB *MEXRootLB, appInst *edgeproto.AppInst, app *edgeproto.App, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	var data templateFill
	var err error
	var mf *Manifest
	log.DebugLog(log.DebugLevelMexos, "fill app template", "appinst", appInst, "clusterInst", clusterInst)
	imageType, ok := edgeproto.ImageType_name[int32(app.ImageType)]
	if !ok {
		return nil, fmt.Errorf("cannot find imagetype in map")
	}
	if clusterInst.Flavor.Name == "" {
		return nil, fmt.Errorf("will not fill app template, invalid cluster flavor name")
	}
	if verr := validateClusterKind(clusterInst.Key.CloudletKey.OperatorKey.Name); verr != nil {
		return nil, verr
	}
	if appInst.Key.AppKey.Name == "" {
		return nil, fmt.Errorf("will not fill app template, invalid appkey name")
	}
	ipAccess, ok := edgeproto.IpAccess_name[int32(appInst.IpAccess)]
	if !ok {
		return nil, fmt.Errorf("cannot find ipaccess in map")
	}
	if len(appInst.Key.AppKey.Name) < 3 {
		log.DebugLog(log.DebugLevelMexos, "warning, very short appkey name", "name", appInst.Key.AppKey.Name)
	}
	config, err := cloudcommon.ParseAppConfig(app.Config)
	if err != nil {
		return nil, fmt.Errorf("error parsing appinst config %s, %v", app.Config, err)
	}
	log.DebugLog(log.DebugLevelMexos, "appinst config", "config", config)
	appDeploymentType := app.Deployment
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelMexos, "app deploying", "imageType", imageType, "deploymentType", appDeploymentType)
	if !cloudcommon.IsValidDeploymentForImage(app.ImageType, appDeploymentType) {
		return nil, fmt.Errorf("deployment is not valid for image type")
	}
	vp := &rootLB.PlatConf.Values
	switch appDeploymentType {
	case cloudcommon.AppDeploymentTypeKubernetes:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               util.K8SSanitize(appInst.Key.AppKey.Name),
			Kind:                   vp.Application.Kind, //"kubernetes",
			Name:                   util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:                   util.K8SSanitize(appInst.Key.AppKey.Name) + "-kubernetes-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			DNSZone:                vp.Network.DNSZone, // "mobiledgex.net",
			Operator:               util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageType:              imageType,
			ImageFlavor:            appInst.Flavor.Name,
			ProxyPath:              util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:                 appInst.Uri,
			IPAccess:               ipAccess,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
			Command:                strings.Split(app.Command, " "),
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeDocker:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               util.K8SSanitize(appInst.Key.AppKey.Name),
			Kind:                   vp.Application.Kind, //"docker"
			Name:                   util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:                   util.K8SSanitize(appInst.Key.AppKey.Name) + "-docker-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			DNSZone:                vp.Network.DNSZone, //"mobiledgex.net",
			Operator:               util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageType:              imageType,
			ImageFlavor:            appInst.Flavor.Name,
			ProxyPath:              util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:                 appInst.Uri,
			IPAccess:               ipAccess,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
			Command:                strings.Split(app.Command, " "),
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeKVM:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               appInst.Key.AppKey.Name,
			Kind:                   vp.Application.Kind, //"kvm",
			Name:                   appInst.Key.AppKey.Name,
			Tags:                   appInst.Key.AppKey.Name + "-qcow-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 appInst.Key.AppKey.Name + "-tenant",
			Operator:               clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageFlavor:            appInst.Flavor.Name,
			DNSZone:                vp.Network.DNSZone, //"mobiledgex.net",
			ImageType:              imageType,
			ProxyPath:              appInst.Key.AppKey.Name,
			AppURI:                 appInst.Uri,
			NetworkScheme:          vp.Network.Scheme, //XXX "external-ip," + GetMEXExternalNetwork(rootLB.PlatConf),
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeHelm:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               appInst.Key.AppKey.Name,
			Kind:                   vp.Application.Kind, //"helm",
			Name:                   appInst.Key.AppKey.Name,
			Tags:                   appInst.Key.AppKey.Name + "-helm-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 appInst.Key.AppKey.Name + "-tenant",
			Operator:               clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageFlavor:            appInst.Flavor.Name,
			DNSZone:                vp.Network.DNSZone, //"mobiledgex.net",
			ImageType:              imageType,
			ProxyPath:              appInst.Key.AppKey.Name,
			AppURI:                 appInst.Uri,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown image type %s", imageType)
	}
	mf.Config.ConfigDetail.Manifest = app.DeploymentManifest
	log.DebugLog(log.DebugLevelMexos, "filled app manifest")
	err = addPorts(mf, appInst)
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelMexos, "added port to app manifest")
	return mf, nil
}

func templateUnmarshal(data *templateFill, yamltext string) (*Manifest, error) {
	//log.DebugLog(log.DebugLevelMexos, "template unmarshal", "yamltext", string, "data", data)
	tmpl, err := template.New("mex").Parse(yamltext)
	if err != nil {
		return nil, fmt.Errorf("can't create template for, %v", err)
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, data)
	if err != nil {
		//log.DebugLog(log.DebugLevelMexos, "template data", "data", data)
		return nil, fmt.Errorf("can't execute template, %v", err)
	}
	mf := &Manifest{}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error yaml unmarshal, templated data", "templated buffer data", outbuffer.String())
		return nil, fmt.Errorf("can't unmarshal templated data, %v", err)
	}
	return mf, nil
}

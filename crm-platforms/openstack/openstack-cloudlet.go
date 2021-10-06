package openstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type OpenstackResources struct {
	InstancesUsed   uint64
	SecGrpsUsed     uint64
	FloatingIPsUsed uint64
}

func (o *OpenstackPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Saving cloudlet access vars to vault", "cloudletName", cloudlet.Key.Name)
	openrcData, ok := accessVarsIn["OPENRC_DATA"]
	if !ok {
		return fmt.Errorf("Invalid accessvars, missing OPENRC_DATA")
	}
	out := strings.Split(openrcData, "\n")
	if len(out) <= 1 {
		return fmt.Errorf("Invalid accessvars, as OPENRC_DATA is invalid: %v", out)
	}
	accessVars := make(map[string]string)
	for _, v := range out {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out1 := strings.Split(v, "=")
		if len(out1) != 2 {
			return fmt.Errorf("Invalid separator for key-value pair: %v", out1)
		}
		key := strings.TrimSpace(out1[0])
		value := strings.TrimSpace(out1[1])
		origVal := value
		value, err := strconv.Unquote(value)
		if err != nil {
			// Unquote didn't find quotes or had some other complaint so use the original value
			value = origVal
		}
		if value == "" || key == "" {
			continue
		}
		if !strings.HasPrefix(key, "OS_") {
			return fmt.Errorf("Invalid accessvars: %s, must start with 'OS_' prefix", key)
		}
		accessVars[key] = value
	}
	authURL, ok := accessVars["OS_AUTH_URL"]
	if !ok {
		return fmt.Errorf("Invalid accessvars, missing OS_AUTH_URL")
	}
	if strings.HasPrefix(authURL, "https") {
		certData, ok := accessVarsIn["CACERT_DATA"]
		if !ok {
			return fmt.Errorf("Invalid accessvars, missing CACERT_DATA")
		}
		certFile := vmlayer.GetCertFilePath(&cloudlet.Key)
		err := ioutil.WriteFile(certFile, []byte(certData), 0644)
		if err != nil {
			return err
		}
		accessVars["OS_CACERT"] = certFile
		accessVars["OS_CACERT_DATA"] = certData
	}
	updateCallback(edgeproto.UpdateTask, "Saving access vars to secure secrets storage (Vault)")
	var varList infracommon.VaultEnvData
	for key, value := range accessVars {
		if key == "OS_CACERT" {
			continue
		}
		varList.Env = append(varList.Env, infracommon.EnvData{
			Name:  key,
			Value: value,
		})
	}
	data := map[string]interface{}{
		"data": varList,
	}

	path := o.GetVaultCloudletAccessPath(&cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName)
	err := infracommon.PutDataToVault(vaultConfig, path, data)
	if err != nil {
		updateCallback(edgeproto.UpdateTask, "Failed to save access vars to vault")
		log.SpanLog(ctx, log.DebugLevelInfra, err.Error(), "cloudletName", cloudlet.Key.Name)
		return fmt.Errorf("Failed to save access vars to vault: %v", err)
	}
	return nil
}

func (o *OpenstackPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {
	osAuthUrl := o.openRCVars["OS_AUTH_URL"]
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr", "authUrl", osAuthUrl)

	if osAuthUrl == "" {
		return "", fmt.Errorf("unable to find OS_AUTH_URL")
	}
	return osAuthUrl, nil
}

func (o *OpenstackPlatform) GetSessionTokens(ctx context.Context, vaultConfig *vault.Config, account string) (map[string]string, error) {
	return nil, fmt.Errorf("GetSessionTokens not supported in OpenStack")
}

func (o *OpenstackPlatform) GetCloudletManifest(ctx context.Context, name string, cloudletImagePath string, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest", "name", name, "VMGroupOrchestrationParams", vmgp)
	var manifest infracommon.CloudletManifest

	o.InitResourceReservations(ctx)
	resources, err := o.populateParams(ctx, vmgp, heatCreate)
	if err != nil {
		return "", err
	}
	err = o.ReleaseReservations(ctx, resources)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "ReleaseReservations error", "err", err)
	}

	if len(vmgp.VMs) == 0 {
		return "", fmt.Errorf("No VMs in orchestation spec")
	}

	// generate the heat template
	buf, err := infracommon.ExecTemplate(name, VmGroupTemplate, vmgp)
	if err != nil {
		return "", err
	}
	templateText := buf.String()

	// download instructions and link
	manifest.AddItem("Download the MobiledgeX bootstrap VM image (please use your console credentials) from the link", infracommon.ManifestTypeURL, infracommon.ManifestSubTypeNone, cloudletImagePath)

	// image create
	title := "Execute the following command to upload the image to your glance store"
	content := fmt.Sprintf("openstack image create %s --disk-format qcow2 --container-format bare --shared --file %s.qcow2", vmgp.VMs[0].ImageName, vmgp.VMs[0].ImageName)
	manifest.AddItem(title, infracommon.ManifestTypeCommand, infracommon.ManifestSubTypeNone, content)

	// heat template download
	manifest.AddItem("Download the manifest template", infracommon.ManifestTypeCode, infracommon.ManifestSubTypeYaml, templateText)

	// heat create commands
	stackName := vmgp.GroupName
	stackCmd := fmt.Sprintf("openstack stack create -t %s.yml", vmgp.GroupName)
	stackParams := []string{}
	for _, fIP := range vmgp.FloatingIPs {
		stackParams = append(stackParams, fmt.Sprintf("--parameter %s=<FLOATING_IP_ID>", fIP.ParamName))
	}
	if len(stackParams) > 0 {
		stackCmd += fmt.Sprintf(" %s", strings.Join(stackParams, " "))
	}
	stackCmd += fmt.Sprintf(" %s", stackName)
	manifest.AddItem("Execute the following command to use manifest to setup the cloudlet", infracommon.ManifestTypeCommand, infracommon.ManifestSubTypeNone, stackCmd)
	return manifest.ToString()
}

func (o *OpenstackPlatform) GetCloudletInfraResourcesInfo(ctx context.Context) ([]edgeproto.InfraResource, error) {

	ramUsed := uint64(0)
	ramMax := uint64(999999)
	vcpusUsed := uint64(0)
	vcpusMax := uint64(200)
	instancesUsed := uint64(0)
	instancesMax := uint64(2000)
	fipsUsed := uint64(0)
	fipsMax := uint64(20)

	// Get external IP usage
	pfRes := vmlayer.PlatformResources{}

	resInfo := []edgeproto.InfraResource{
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceRamMb,
			Value:         ramUsed,
			InfraMaxValue: ramMax,
			Units:         cloudcommon.ResourceRamUnits,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceVcpus,
			Value:         vcpusUsed,
			InfraMaxValue: vcpusMax,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceExternalIPs,
			Value:         pfRes.Ipv4Used,
			InfraMaxValue: pfRes.Ipv4Max,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceInstances,
			Value:         instancesUsed,
			InfraMaxValue: instancesMax,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceFloatingIPs,
			Value:         fipsUsed,
			InfraMaxValue: fipsMax,
		},
	}
	return resInfo, nil
}

func (o *OpenstackPlatform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{
		Properties: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceInstances,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceInstances],
			},
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceFloatingIPs,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceFloatingIPs],
			},
		},
	}, nil
}

func getOpenstackResources(cloudlet *edgeproto.Cloudlet, resources []edgeproto.VMResource) *OpenstackResources {
	floatingIp := false
	if val, ok := cloudlet.EnvVar["MEX_NETWORK_SCHEME"]; ok {
		if strings.Contains(val, "floatingip") {
			floatingIp = true
		}
	}
	var oRes OpenstackResources
	for _, vmRes := range resources {
		// Number of Instances = Number of resources
		oRes.InstancesUsed += 1
		if floatingIp && vmRes.Type == cloudcommon.VMTypeRootLB ||
			(vmRes.Type == cloudcommon.VMTypeAppVM && vmRes.AppAccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT) {
			// Number of floating IPs = NetworkScheme==FloatingIP && Number of external facing resources
			oRes.FloatingIPsUsed += 1
		}
	}
	return &oRes
}

// called by controller, make sure it doesn't make any calls to infra API
func (o *OpenstackPlatform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource {
	// resource name -> resource units
	cloudletRes := map[string]string{
		cloudcommon.ResourceInstances:   "",
		cloudcommon.ResourceFloatingIPs: "",
	}
	resInfo := make(map[string]edgeproto.InfraResource)
	for resName, resUnits := range cloudletRes {
		resMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			resMax = infraRes.InfraMaxValue
		}
		resInfo[resName] = edgeproto.InfraResource{
			Name:          resName,
			InfraMaxValue: resMax,
			Units:         resUnits,
		}
	}

	oRes := getOpenstackResources(cloudlet, vmResources)
	outInfo, ok := resInfo[cloudcommon.ResourceInstances]
	if ok {
		outInfo.Value += oRes.InstancesUsed
		resInfo[cloudcommon.ResourceInstances] = outInfo
	}
	outInfo, ok = resInfo[cloudcommon.ResourceFloatingIPs]
	if ok {
		outInfo.Value += oRes.FloatingIPsUsed
		resInfo[cloudcommon.ResourceFloatingIPs] = outInfo
	}
	return resInfo
}

func (o *OpenstackPlatform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	oRes := getOpenstackResources(cloudlet, resources)

	resMetric.AddIntVal(cloudcommon.ResourceMetricInstances, oRes.InstancesUsed)
	resMetric.AddIntVal(cloudcommon.ResourceMetricFloatingIPs, oRes.FloatingIPsUsed)
	return nil
}

func (p *OpenstackPlatform) InternalCloudletUpdatedCallback(ctx context.Context, old *edgeproto.CloudletInternal, new *edgeproto.CloudletInternal) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InternalCloudletUpdatedCallback")
}

func (p *OpenstackPlatform) GetGPUSetupStage(ctx context.Context) vmlayer.GPUSetupStage {
	return vmlayer.ClusterInstStage
}

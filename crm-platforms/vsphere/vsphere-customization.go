package vsphere

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/govmomi/vim25/types"
)

func (v *VSpherePlatform) GetCustomizationSpec(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams) (string, error) {
	utc := true
	if len(vmgp.VMs) != 1 {
		return "", fmt.Errorf("unexpected number of vms in spec :%d", len(vmgp.VMs))
	}
	custSpec := &types.CustomizationSpecItem{
		Info: types.CustomizationSpecInfo{
			Name: vmgp.GroupName + "-cust-spec",
			Type: "Linux",
		},
		Spec: types.CustomizationSpec{
			NicSettingMap: []types.CustomizationAdapterMapping{
				{
					Adapter: types.CustomizationIPSettings{
						Ip: &types.CustomizationUnknownIpGenerator{},
					},
				},
			},
			Identity: &types.CustomizationLinuxPrep{
				HostName: &types.CustomizationFixedName{
					Name: vmgp.VMs[0].HostName,
				},
				Domain:     "mobiledgex.net",
				TimeZone:   "Etc/UTC",
				HwClockUTC: &utc,
				ScriptText: "#!/bin/sh\n" +
					"vmtoolsd  --cmd \"info-set guestinfo.metadata.encoding base64\"\n" +
					"vmtoolsd  --cmd \"info-set guestinfo.metadata " + vmgp.VMs[0].MetaData + "\"\n" +
					"vmtoolsd  --cmd \"info-set guestinfo.userdata.encoding base64\"\n" +
					"vmtoolsd  --cmd \"info-set guestinfo.userdata " + vmgp.VMs[0].UserData + "\"\n",
			},
			GlobalIPSettings: types.CustomizationGlobalIPSettings{
				DnsSuffixList: []string{v.vmProperties.CommonPf.GetCloudletDNSZone()},
				DnsServerList: vmlayer.CloudflareDns,
			},
		},
	}
	out, err := xml.MarshalIndent(custSpec, "  ", "    ")
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCustomizationSpec", "cspec", custSpec)
	return string(out), err
}

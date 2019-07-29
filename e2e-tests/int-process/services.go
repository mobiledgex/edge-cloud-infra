package intprocess

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

func getShepherdProc(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (*Shepherd, []process.StartOp, error) {
	opts := []process.StartOp{}

	cloudletKeyStr, err := json.Marshal(cloudlet.Key)
	if err != nil {
		return nil, opts, fmt.Errorf("unable to marshal cloudlet key")
	}

	/*
	   - shepherd
	   - "--influxAddr"
	   - "https://{{ influxdb_dns }}.{{ cloudflare_zone }}:8086"
	   # notifyAddr is localhost because use use host networking for docker containers
	   - "--notifyAddrs"
	   - "localhost:{{ notify_port }}"
	   - "--tls"
	   - "/root/tls/mex-server.crt"
	   - "--platform"
	   - "{{ item.platform | default('openstack') }}"
	   - "--vaultAddr"
	   - "{{ vault_vm_hostname }}:{{ vault_port }}"
	   - "--cloudletKey"
	   - "{\\\"operator_key\\\":{\\\"name\\\":\\\"{{ item.operator_key | mandatory }}\\\"},\\\"name\\\":\\\"{{ item.cloudlet_name }}\\\"}"
	   - "--region"
	   - "{{ item.controller_region }}"
	   - "-d"
	   - "api,notify,mexos,metrics"

	*/
	envVars := make(map[string]string)
	notifyAddr := ""
	tlsCertFile := ""
	vaultAddr := ""
	if pfConfig != nil {
		// Same role-id/secret-id as CRM
		envVars["VAULT_ROLE_ID"] = pfConfig.CrmRoleId
		envVars["VAULT_SECRET_ID"] = pfConfig.CrmSecretId
		notifyAddr = cloudlet.NotifySrvAddr
		tlsCertFile = pfConfig.TlsCertFile
		vaultAddr = pfConfig.VaultAddr
	}

	opts = append(opts, process.WithDebug("api,mexos,notify"))

	return &Shepherd{
		NotifyAddrs: notifyAddr,
		CloudletKey: string(cloudletKeyStr),
		Platform:    cloudlet.PlatformType.String(),
		Common: process.Common{
			Hostname: cloudlet.Key.Name,
			EnvVars:  envVars,
		},
		TLS: process.TLSCerts{
			ServerCert: tlsCertFile,
		},
		VaultAddr:    vaultAddr,
		PhysicalName: cloudlet.PhysicalName,
	}, opts, nil
}

func GetShepherdCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	ShepherdProc, opts, err := getShepherdProc(cloudlet, pfConfig)
	if err != nil {
		return "", nil, err
	}

	return ShepherdProc.String(opts...), &ShepherdProc.Common.EnvVars, nil
}

func StartShepherdService(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) error {
	ShepherdProc, opts, err := getShepherdProc(cloudlet, pfConfig)
	if err != nil {
		return err
	}

	err = ShepherdProc.StartLocal("/tmp/"+cloudlet.Key.Name+"shepherd.log", opts...)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "started "+ShepherdProc.GetExeName())

	return nil
}

func StopShepherdService(cloudlet *edgeproto.Cloudlet) error {
	ShepherdProc, _, err := getShepherdProc(cloudlet, nil)
	if err != nil {
		return err
	}
	// max wait time for process to go down gracefully, after which it is killed forcefully
	maxwait := 5 * time.Second

	c := make(chan string)
	go process.KillProcessesByName(ShepherdProc.GetExeName(), maxwait, ShepherdProc.LookupArgs(), c)

	log.DebugLog(log.DebugLevelMexos, "stopped Shepherdserver", "msg", <-c)
	return nil
}

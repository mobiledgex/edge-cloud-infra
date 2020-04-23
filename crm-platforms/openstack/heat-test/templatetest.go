package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

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

func main() {
	fmt.Printf("begin\n")
	stackName := "test"
	var buf bytes.Buffer

	funcMap := template.FuncMap{
		"Indent": func(values ...interface{}) string {
			s := values[0].(string)
			l := 4
			if len(values) > 1 {
				l = values[1].(int)
			}
			var newStr []string
			for _, v := range strings.Split(string(s), "\n") {
				nV := fmt.Sprintf("%s%s", strings.Repeat(" ", l), v)
				newStr = append(newStr, nV)
			}
			return strings.Join(newStr, "\n")
		},
	}

	tmpl, err := template.New(stackName).Funcs(funcMap).Parse(openstack.VmGroupTemplate)
	if err != nil {
		// this is a bug
		fmt.Printf("template new failed %v", err)
		return
	}

	newSecGrp := "test-secgrp"
	cloudletSecGrp := "default"

	newSubnet := "test-new-subnet"
	rootLbExtPortName := "rootlb-external-port"
	rootLbIntPortName := "rootlb-internal-port"

	var vmgp vmlayer.VMGroupOrchestrationParams

	rootlbExternalPort := vmlayer.PortOrchestrationParams{
		Name:        rootLbExtPortName,
		NetworkName: "external-network-shared",
		SecurityGroups: []vmlayer.ResourceReference{
			vmlayer.NewResourceReference(cloudletSecGrp, true),
			vmlayer.NewResourceReference(newSecGrp, false),
		},
	}

	rootlbInternalPort := vmlayer.PortOrchestrationParams{
		Name:        rootLbIntPortName,
		NetworkName: "mex-k8s-net-1",
		FixedIPs: []vmlayer.FixedIPOrchestrationParams{
			{Address: "10.101.99.1", Subnet: vmlayer.ResourceReference{Name: newSubnet, Preexisting: false}},
		},
		SecurityGroups: []vmlayer.ResourceReference{
			{Name: newSecGrp, Preexisting: false},
		},
	}

	rootlb := vmlayer.VMOrchestrationParams{
		Name:       "rootlb1",
		ImageName:  "mobiledgex-v3.1.0",
		FlavorName: "m4.medium",
		Role:       vmlayer.RoleAgent,
		Ports: []vmlayer.ResourceReference{
			vmlayer.NewResourceReference(rootLbExtPortName, false),
			vmlayer.NewResourceReference(rootLbIntPortName, false),
		},
	}

	// fip doesn't usually go on an external network this is for test
	fipid := "4eedb4bd-b5cb-4738-8b4f-11254e181b8b"

	fip := vmlayer.FloatingIPOrchestrationParams{
		Name:         "fip-test1",
		FloatingIpId: vmlayer.NewResourceReference(fipid, false),
		Port:         vmlayer.NewResourceReference(rootLbIntPortName, false),
	}

	vmgp.FloatingIPs = append(vmgp.FloatingIPs, fip)

	secGrp := vmlayer.SecurityGroupOrchestrationParams{
		Name: newSecGrp,
		AccessPorts: []util.PortSpec{
			{
				Port:    "8443",
				EndPort: "8443",
				Proto:   "TCP",
			},
			{
				Port:    "9000",
				EndPort: "9002",
				Proto:   "UDP",
			},
		},
		EgressRestricted: true,
		EgressRules: []edgeproto.OutboundSecurityRule{
			{
				Protocol:     "TCP",
				PortRangeMin: 4000,
				PortRangeMax: 4002,
				RemoteCidr:   "47.0.0.0/8",
			},
		},
	}
	vmgp.SecurityGroups = append(vmgp.SecurityGroups, secGrp)

	subNet := vmlayer.SubnetOrchestrationParams{
		Name:        newSubnet,
		CIDR:        "10.101.99.0/24",
		DHCPEnabled: "no",
		DNSServers:  []string{"1.1.1.1", "1.0.0.1"},
	}
	vmgp.Subnets = append(vmgp.Subnets, subNet)
	vmgp.VMs = append(vmgp.VMs, rootlb)
	vmgp.Ports = append(vmgp.Ports, rootlbExternalPort)
	vmgp.Ports = append(vmgp.Ports, rootlbInternalPort)

	masterport := vmlayer.PortOrchestrationParams{
		Name:        "master-port",
		NetworkName: "mex-k8s-net-1",
		FixedIPs: []vmlayer.FixedIPOrchestrationParams{
			{Address: "10.101.99.10", Subnet: vmlayer.ResourceReference{Name: newSubnet, Preexisting: false}},
		},
		SecurityGroups: []vmlayer.ResourceReference{
			{Name: newSecGrp, Preexisting: false},
		},
	}
	vmgp.Ports = append(vmgp.Ports, masterport)

	master := vmlayer.VMOrchestrationParams{
		Name:       "master",
		Role:       vmlayer.RoleMaster,
		ImageName:  "mobiledgex-v3.1.0",
		FlavorName: "m4.medium",
	}
	master.Ports = append(master.Ports, vmlayer.ResourceReference{Name: "master-port", Preexisting: false})
	vmgp.VMs = append(vmgp.VMs, master)

	err = tmpl.Execute(&buf, vmgp)
	if err != nil {
		fmt.Printf("Template Execute Failed: %v", err)
		return
	}
	unescaped := html.UnescapeString(buf.String())
	var unescapedBuf bytes.Buffer
	unescapedBuf.WriteString(unescaped)

	filename := stackName + "-heat.yaml"
	err = WriteTemplateFile(filename, &unescapedBuf)
	if err != nil {
		fmt.Printf("WriteTemplateFile failed: %v", err)
		return
	}
	fmt.Printf("Created file : %s\n", filename)
}

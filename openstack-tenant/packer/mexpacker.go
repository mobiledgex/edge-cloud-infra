package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/codeskyblue/go-sh"
	"github.com/nanobox-io/golang-ssh"
)

var sshOpts = []string{"StrictHostKeyChecking=no", "UserKnownHostsFile=/dev/null"}

type item struct {
	ID, Name string
}
type param struct {
	SrcImageID, NetworkID string
}

var packerJSON = `
{
    "builders": [{
	"type": "openstack",
	"flavor": "m4.medium",
	"ssh_username": "ubuntu",
	"region": "RegionOne",
	"image_name": "mobiledgex",
	"source_image": "{{.SrcImageID}}",
	"networks": "{{.NetworkID}}",
	"security_groups": ["default" ]
    }],
    "provisioners": [{
	"type": "shell",
	"script": "setup.sh"
    }]
}
`

func main() {
	if os.Getenv("OS_PASSWORD") == "" {
		fmt.Println("missing openstack env var")
		os.Exit(1)
	}
	fn := "xenial-server-cloudimg-amd64-disk1.img"
	ens := "external-network-shared"
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		fmt.Println("file", fn, "does not exist, download from registry")
		md := os.Getenv("MEX_DIR")
		if md == "" {
			md = os.Getenv("HOME") + "/.mobiledgex"
		}
		sk := os.Getenv("MEX_SSH_KEY")
		if sk == "" {
			fmt.Println("env var MEX_SSH_KEY does not exist")
			os.Exit(1)
		}
		un := os.Getenv("MEX_REGISTRY_USER")
		if un == "" {
			un = "mobiledgex"
		}
		kp := fmt.Sprintf("%s/%s", md, sk)
		auth := ssh.Auth{Keys: []string{kp}}
		ad := os.Getenv("MEX_REGISTRY")
		if ad == "" {
			ad = "registry.mobiledgex.net"
		}
		fmt.Println("using registry", ad, "key", kp, "user", un)
		client, err := ssh.NewNativeClient(un, ad, "SSH-2.0-mobiledgex-ssh-client-1.0", 22, &auth, nil)
		if err != nil {
			fmt.Println("cannot get ssh client", err)
			os.Exit(1)
		}
		cmd := fmt.Sprintf("scp -o %s -o %s -i %s %s@%s:files-repo/mobiledgex/%s .", sshOpts[0], sshOpts[1], kp, un, ad, fn)
		fmt.Println(cmd)
		out, err := client.Output(cmd)
		if err != nil {
			fmt.Println("cannot download", fn, err, out)
			os.Exit(1)
		}
		fmt.Println("downloaded ok", fn)
	}
	in := "mobiledgex" // XXX
	out, err := sh.Command("openstack", "image", "create", "--disk-format", "qcow2", "--file", fn, in).Output()
	if err != nil {
		fmt.Println("cannot create glance image for", in, fn, out)
		os.Exit(1)
	}
	var assetList []item
	paramMap := make(map[string]string)
	assets := []struct {
		asset string
		name  string
	}{
		{"image", fn},
		{"network", ens},
	}
	for _, a := range assets {
		out, err = sh.Command("openstack", a.asset, "list", "-f", "json").Output()
		if err != nil {
			fmt.Println("cannot list ", a.asset, err)
			os.Exit(1)
		}
		err = json.Unmarshal([]byte(out), &assetList)
		if err != nil {
			fmt.Println("cannot unmarshal", a.asset, "list", err)
			os.Exit(1)
		}
		found := false
		for _, i := range assetList {
			if i.Name == a.name {
				found = true
				paramMap[a.asset] = i.ID
			}
		}
		if !found {
			fmt.Println("no", a.asset, "match", fn)
			os.Exit(1)
		}
	}
	for _, n := range []string{"image", "network"} {
		if _, ok := paramMap[n]; !ok {
			fmt.Println("no", n, "in map", paramMap)
			os.Exit(1)
		}
	}
	params := &param{SrcImageID: paramMap["image"], NetworkID: paramMap["network"]}
	tmpl, err := template.New("packer").Parse(packerJSON)
	if err != nil {
		fmt.Println("cannot create packer templ", err)
		os.Exit(1)
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, params)
	if err != nil {
		fmt.Println("cannot execute packer tmpl", err)
		os.Exit(1)
	}
	fmt.Println(outbuffer.Bytes())
	// out, err = sh.Command("PACKER_LOG=1", "packer", "build", "packer_template.mobiledgex.json").Output()
	// if err != nil {
	// 	fmt.Println("cannot run packer", err)
	// 	os.Exit(1)
	// }
	os.Exit(0)
}

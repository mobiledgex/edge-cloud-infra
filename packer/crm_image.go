package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
)

type CRMParams struct {
	SrcImg, ImgName, NetworkID string
}

const (
	SrcImgURL    = "http://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img"
	SrcImg       = "xenial-server"
	SrcImgCksum  = "4deafac2469ff170385e5a88464ac4e0"
	Network      = "external-network-shared"
	SecGroup     = "default"
	PhysCloudlet = "frankfurt"
	VaultAddr    = "vault.mobiledgex.net"
	CRMImgVers   = "v1.0"
	CRMImgName   = "mobiledgex-crm-" + CRMImgVers
)

var packerJSON = `
{
    "builders": [{
        "type": "openstack",
        "flavor": "m4.medium",
        "ssh_username": "ubuntu",
        "image_name": "{{.ImgName}}",
        "source_image_name": "{{.SrcImg}}",
        "networks": "{{.NetworkID}}",
        "security_groups": ["default" ]
    }],
    "provisioners": [{
        "type": "shell",
        "script": "setup-scripts/setup.sh"
    }]
}
`

func main() {
	fmt.Println("Init Openstack Props")
	err := mexos.InitOpenstackProps("tdg", PhysCloudlet, VaultAddr)
	if err != nil {
		fmt.Printf("Unable to source OpenRC file, %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Check if image exists")
	imageDetail, err := mexos.GetImageDetail(SrcImg)
	if err == nil && imageDetail.Status != "active" {
		fmt.Printf("image %s is not active", SrcImg)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("image %s not found, adding cloud image to glance", SrcImg)
		err = mexos.CreateImageFromUrl(SrcImg, SrcImgURL, SrcImgCksum)
		if err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}
	}
	fmt.Println("Get Openstack network details")
	nd, err := mexos.GetNetworkDetail(Network)
	if err != nil {
		fmt.Printf("can't get details for external network %s, %v", Network, err)
		os.Exit(1)
	}
	if nd.Status != "ACTIVE" {
		fmt.Printf("external network %s is not in ACTIVE state", Network)
		os.Exit(1)
	}

	params := &CRMParams{
		SrcImg:    SrcImg,
		ImgName:   CRMImgName,
		NetworkID: nd.ID,
	}
	tmpl, err := template.New("packer").Parse(packerJSON)
	if err != nil {
		fmt.Println("cannot create packer templ", err)
		os.Exit(1)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		fmt.Println("cannot execute packer tmpl", err)
		os.Exit(1)
	}
	fileName := "crm_template.json"
	err = mexos.WriteTemplateFile(fileName, &buf)

	fmt.Println("Execute Packer")
	os.Setenv("PACKER_LOG", "1")
	_, err = sh.Command("packer", "build", fileName).Output()
	if err != nil {
		fmt.Println("cannot run packer", err)
		os.Exit(1)
	}
	fmt.Println("packer run ok")

	_, err = mexos.GetImageDetail(CRMImgName)
	if err != nil {
		fmt.Println("image does not exist in Openstack", err)
		os.Exit(1)
	}
	savePath := "/tmp/" + CRMImgName + ".qcow2"
	err = mexos.SaveImage(savePath, CRMImgName)
	if err != nil {
		fmt.Println("failed saving image", err)
		os.Exit(1)
	}
	fmt.Println("Image successfully built and saved at " + savePath)
}

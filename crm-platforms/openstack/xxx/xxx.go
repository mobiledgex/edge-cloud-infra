package main

import (
	"encoding/json"
	"fmt"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
)

func main() {

	net := "bogus"
	out, err := sh.Command("openstack", "floating", "ip", "list", "--network", net, "-f", "json").CombinedOutput()
	fmt.Printf("OUT %s ERR %v\n", string(out), err)
	var fips []openstack.OSFloatingIP
	err = json.Unmarshal(out, &fips)
	fmt.Printf("FIPS %+v, ERR2 %v\n", fips, err)

}

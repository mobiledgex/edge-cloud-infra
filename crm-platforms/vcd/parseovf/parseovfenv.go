package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Parse the result of vcd's ProductSection ovfenv produced via
// vmtoolsd --cmd "info-get guestinfo.ovfenv"
// and just write to stdout as mobiledgex-init will
// send to output to the existing /mnt/mobiledgex/openstack/latest/meta-data.json file
//
func getValue(line string) string {
	value := strings.SplitAfter(line, "value=")
	if len(value) > 0 {
		val := value[1]
		v := strings.Split(val, "/")
		s := v[0]
		if len(s) > 1 {
			s = s[1 : len(s)-1]
			return s
		} else {
			return ""
		}
	}
	return ""
}

// XXX This still uses the Contains operation, which can match unwanted tokens
// Revisit
func main() {
	// default read from /var/log but use -ovf foo.txt to override
	ovfile := flag.String("ovf", "/var/log/userdata.log", "default ovfenv file")
	flag.Parse()
	// combined metada params and network_params?
	// This appears to be the current extent of options that
	// might be encountered today xxx
	envars := map[string]string{
		"ROLE":           "",
		"SKIPK8S":        "",
		"K8SMASTER":      "",
		"HOSTNAME":       "",
		"UPDATE":         "",
		"SKIPINIT":       "",
		"INTERFACE":      "",
		"UPDATEHOSTNAME": "",
		"IPADDR":         "",
		"NETMASK":        "",
		"NETTYPE":        "",
	}
	type EnvProperty struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	}
	type PropertySection struct {
		Properties []EnvProperty `xml:"Property"`
	}
	type Environment struct {
		Property *PropertySection `xml:"PropertySection"`
	}
	info := Environment{}

	ovf, err := os.OpenFile(*ovfile, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", *ovfile, err.Error())
		return
	}
	defer ovf.Close()

	data, err := ioutil.ReadFile(*ovfile)
	if err != nil {
		fmt.Printf("ReadFile failed: %s\n", err.Error())
		os.Exit(-1)
	}

	err = xml.Unmarshal(data, &info)
	if err != nil {
		fmt.Printf("Error unmarshal: %s\n", err.Error())
		os.Exit(-1)
	}
	vars := make(map[string]interface{})
	for _, p := range info.Property.Properties {
		_, ok := envars[strings.ToUpper(p.Key)]
		if ok {
			vars[p.Key] = p.Value
		}
	}
	metadata := map[string]interface{}{
		"meta": vars,
	}
	out, err := json.Marshal(metadata)
	if err != nil {
		fmt.Printf("Error marshalling meta data %s\n", err.Error())
		os.Exit(-1)
	}
	fmt.Printf("%s", string(out))
}

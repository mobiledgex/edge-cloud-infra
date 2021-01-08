package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	//	"strconv"
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
		"MASTERADDR":     "",
		"HOSTNAME":       "",
		"UPDATE":         "",
		"SKIPINIT":       "",
		"INTERFACE":      "",
		"UPDATEHOSTNAME": "",
		"IPADDR":         "",
		"NETMASK":        "",
		"NETTYPE":        "",
	}

	ovf, err := os.OpenFile(*ovfile, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", *ovfile, err.Error())
		return
	}
	defer ovf.Close()

	scanner := bufio.NewScanner(ovf)
	if err != nil {
		fmt.Printf("could not create scanner! :%s\n", err.Error())
		os.Exit(1)
	}
	vars := make(map[string]interface{}) // Jon alt.
	for scanner.Scan() {
		nextline := scanner.Text()
		for key, _ := range envars {
			if strings.Contains(nextline, key) {
				value := getValue(nextline)
				vars[strings.ToLower(key)] = value // Jon's alt. add to vars map
			}
		}
	}
	metadata := map[string]interface{}{ // Jons impl
		"meta": vars, // jon
	}
	out, err := json.Marshal(metadata)
	if err != nil {
		fmt.Printf("Error marsheling meta data %s\n", err.Error())
		os.Exit(-1)
	}
	fmt.Printf("%s", string(out))
}

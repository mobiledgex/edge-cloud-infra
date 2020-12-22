package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Parse the result of vcd's ProductSection ovfenv produced via
// vmtoolsd --cmd "info-get guestinfo.ovfenv"
// and just write to stdout as mobiledgex-init will
// send to output to the existing /mnt/mobiledgex/openstack/latest/meta-data.json file
//
func getValue(line string) string {
	value := strings.SplitAfter(line, "value=")
	val := value[1]
	v := strings.Split(val, "/")
	s := v[0]
	s = s[1 : len(s)-1]
	return s
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
		return
	}

	// we want to end up with a json fragment looking something like:
	//{"meta": {"skipk8s": true, "role": "mex-agent-node", "k8smaster": "10.103.0.10"}}
	outline := []string{}
	outline = append(outline, "{") // why are we losing this first brace in the join?
	meta := "meta"
	v := strconv.Quote(meta)
	outline = append(outline, v)
	outline = append(outline, ": {")

	for scanner.Scan() {
		nextline := scanner.Text()
		for key, _ := range envars {
			if strings.Contains(nextline, key) {
				value := getValue(nextline)
				// quote values other than bools
				if value != "true" && value != "false" {
					value = strconv.Quote(value)
				}
				key := strings.ToLower(key)
				key = strconv.Quote(key)
				val := fmt.Sprintf("%s : %s, ", key, value)
				outline = append(outline, val)
			}
		}
	}

	s := strings.Join(outline[:], "")
	idx := strings.LastIndex(s, ",")
	s = s[1:idx]
	s = s + "}}"
	s = fmt.Sprintf("%s %s\n", "{", s)
	fmt.Printf("%s", s)
}

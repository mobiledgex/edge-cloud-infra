package anthos

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type SecondaryEthInterface struct {
	IpAddress        string
	SecInterfaceName string
}

func (a *AnthosPlatform) GetSecondaryEthInterfaces(ctx context.Context, client ssh.Client, devname string) ([]SecondaryEthInterface, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSecondaryEthInterfaces", "devname", devname)
	cmd := fmt.Sprintf("ip address show %s", devname)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("Error in finding secondary interfaces: %s - %v", out, err)
	}
	var secIfs []SecondaryEthInterface
	lines := strings.Split(out, "\n")
	ifPattern := "inet (\\d+\\.\\d+\\.\\d+\\.\\d+)/\\d\\d.* secondary (\\w+:\\d+)"
	ifReg := regexp.MustCompile(ifPattern)
	for _, line := range lines {
		if ifReg.MatchString(line) {
			matches := ifReg.FindStringSubmatch(line)
			ip := matches[1]
			ifname := matches[2]
			secIf := SecondaryEthInterface{
				IpAddress:        ip,
				SecInterfaceName: ifname,
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "GetSecondaryEthInterfaces found interface", "secIf", secIf)
			secIfs = append(secIfs, secIf)
		}
	}
	return secIfs, nil
}

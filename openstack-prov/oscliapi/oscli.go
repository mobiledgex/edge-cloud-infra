package oscli

import (
	"encoding/json"
	"github.com/codeskyblue/go-sh"

	log "gitlab.com/bobbae/logrus"
)

type Limit struct {
	Name  string
	Value int
}

func GetLimits() ([]Limit, error) {
	//err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).WriteStdout("os-out.txt")
	out, err := sh.Command("openstack", "limits", "show", "--absolute", "-f", "json", sh.Dir("/tmp")).Output()
	if err != nil {
		log.Debugf("cannot get limits from openstack, %v", err)
		return nil, err
	}

	log.Debugln(out)

	var limits []Limit
	err = json.Unmarshal(out, &limits)
	log.Debugln(limits)
	if err != nil {
		log.Debugf("cannot unmarshal, %v", err)
		return nil, err
	}
	return nil, nil
}

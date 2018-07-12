package oscli

import (
	"fmt"
	"testing"

	log "github.com/bobbae/logrus"
)

func TestInit(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestGetLimits(t *testing.T) {
	out, err := GetLimits()
	if err != nil {
		t.Errorf("cannot GetLimits, %v", err)
	}
	fmt.Println(out)
}

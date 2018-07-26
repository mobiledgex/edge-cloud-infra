package oscli

import (
	"fmt"
	"os"
	"testing"
)

var mexTestInfra = os.Getenv("MEX_TEST_INFRA")

func TestValidateNetwork(t *testing.T) {
	if mexTestInfra == "" {
		return
	}
	err := ValidateNetwork()
	if err != nil {
		t.Errorf("network not valididated, %v", err)
	}

	fmt.Println("net validated")
}

func TestPrepNetwork(t *testing.T) {
	if mexTestInfra == "" {
		return
	}
	err := PrepNetwork()
	if err != nil {
		t.Errorf("cannot prep network, %v", err)
	}
	fmt.Println("network prepped")
}

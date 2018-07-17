package oscli

import (
	"fmt"
	"testing"
)

func TestValidateNetwork(t *testing.T) {
	err := ValidateNetwork()
	if err != nil {
		t.Errorf("network not valididated, %v", err)
	}

	fmt.Println("net validated")
}

func TestPrepNetwork(t *testing.T) {
	err := PrepNetwork()
	if err != nil {
		t.Errorf("cannot prep network, %v", err)
	}
	fmt.Println("network prepped")
}

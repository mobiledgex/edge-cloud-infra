package oscli

import (
	"fmt"
	"strings"
	"testing"
)

var role1 = "test-1-generic"

func TestCreateKVM1(t *testing.T) {
	err := CreateKVM(role1, 1)
	if err != nil {
		t.Errorf("can't create kvm, %v", err)
		return
	}
	fmt.Println("created kvm")
}

func TestCreateKubernetesMaster(t *testing.T) {
	err := CreateKVM("k8s-master", 1)
	if err != nil {
		t.Errorf("can't create kubernetes master node, %v", err)
		return
	}
	fmt.Println("created k8s-master")
}

func TestCreateKubernetesNode1(t *testing.T) {
	err := CreateKVM("k8s-node", 2)
	if err != nil {
		t.Errorf("can't create kubernetes node 1, %v", err)
	}
	fmt.Println("created k8s-node 1")
}

func TestCreateKubernetesNode2(t *testing.T) {
	err := CreateKVM("k8s-node", 3)
	if err != nil {
		t.Errorf("can't create kubernetes node 2, %v", err)
	}
	fmt.Println("created k8s-node 2")
}

func TestDeleteKVM1(t *testing.T) {
	sl, err := ListServers()
	if err != nil {
		t.Errorf("can't get list of servers, %v", err)
		return
	}
	for _, s := range sl {
		if strings.HasPrefix(s.Name, "mex-") {
			sd, err := GetServerDetails(s.Name)
			if err != nil {
				t.Errorf("can't get server detail for %s, %v", s.Name, err)
				return
			}
			prop := sd.Properties
			props := strings.Split(prop, ",")
			for _, p := range props {
				kv := strings.Split(p, "=")
				if strings.Index(kv[0], "role") >= 0 { // extra space in front
					if strings.Index(kv[1], role1) >= 0 { // single quotes
						err := DeleteServer(s.Name)
						if err != nil {
							t.Errorf("can't delete %s, %v", s.Name, err)
							return
						}
						fmt.Println("delete", s.Name)
					}
				}
			}
		}
	}
}

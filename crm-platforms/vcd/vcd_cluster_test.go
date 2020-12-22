package vcd

import (
	//"context"
	"fmt"
	//"strings"
	"github.com/stretchr/testify/require"
	"testing"
	//"github.com/vmware/go-vcloud-director/v2/govcd"
	//"github.com/vmware/go-vcloud-director/v2/types/v56"
)

//  -cld -clst
func TestRMCluster(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("Removing cluster %s from cloudlet %s\n", *clstName, *cldName)
		err := tv.DeleteCluster(ctx, *clstName)
		if err != nil {
			fmt.Printf("Error deleting cluster %s : %s\n", *clstName, err.Error())
			return
		}
	}
	fmt.Printf("ClusterInst %s deleted successfully\n", *clstName)
}

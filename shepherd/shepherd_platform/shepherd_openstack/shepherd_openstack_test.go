package shepherd_openstack

import (
	"testing"

	"gotest.tools/assert"
)

func TestIpPoolRange(t *testing.T) {
	// single pool
	assert.Equal(t, uint64(20), getIpCountFromPools("10.10.10.1-10.10.10.20"))
	// several pools
	assert.Equal(t, uint64(31), getIpCountFromPools("10.10.10.1-10.10.10.20,10.10.10.30-10.10.10.40"))
	// ipv6 pool
	assert.Equal(t, uint64(18446744073709551614), getIpCountFromPools("2a01:598:4:4011::2-2a01:598:4:4011:ffff:ffff:ffff:ffff"))
	// empty pool
	assert.Equal(t, uint64(0), getIpCountFromPools(""))
	// invalid pool
	assert.Equal(t, uint64(0), getIpCountFromPools("invalid pool"))
}

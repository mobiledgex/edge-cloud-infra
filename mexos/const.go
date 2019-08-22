package mexos

const APIversion = "v1"

const MEXSubnetSeed = 100
const MEXSubnetLimit = 250

const (
	k8smasterRole = "k8s-master"
	k8snodeRole   = "k8s-node"
)

//For netspec components
//  netType,netName,netCIDR,netOptions
const (
	NetTypeVal       = 0
	NetNameVal       = 1
	NetCIDRVal       = 2
	NetFloatingIPVal = 3
	NetVnicType      = 4
	NetOptVal        = 5
)

package tdg

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/test-go/testify/require"
	"golang.org/x/net/context"
)

func TestBadIpv4(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelInfra)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// Intialize OperatorApiGw
	gw := &OperatorApiGw{}
	servers := operator.OperatorApiGwServers{}
	servers.QosSesAddr = "http://localhost:8081" // Not actually used because all tests fail before getting to the REST call.
	gw.Init("TDG", &servers)
	gw.SetQosSessionsApiKey("xxxx") // Not actually used because all tests fail before getting to the REST call.

	badIps := [9]string{"mobiledgex.net", "localhost", "192.168.0.BAD",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "500.0.0.1", "1.1.1.1.1",
		"0", "a.b.c.d", "-1.0.0.1"}
	log.SpanLog(ctx, log.DebugLevelDmereq, "TestBadIpv4", "badIps", badIps)

	// Test bad IPs for IpApplicationServer
	for _, badIp := range badIps {
		req := dme.QosPrioritySessionCreateRequest{}
		req.IpApplicationServer = badIp
		req.IpUserEquipment = "127.0.0.1"
		req.Profile = dme.QosSessionProfile_QOS_LOW_LATENCY

		_, err := gw.CreatePrioritySession(ctx, &req)
		require.Equal(t, err.Error(), "rpc error: code = InvalidArgument desc = Invalid Address for IpApplicationServer: "+badIp)
	}

	// Test bad IPs for IpUserEquipment
	for _, badIp := range badIps {
		req := dme.QosPrioritySessionCreateRequest{}
		req.IpApplicationServer = "192.168.0.1"
		req.IpUserEquipment = badIp
		req.Profile = dme.QosSessionProfile_QOS_LOW_LATENCY

		_, err := gw.CreatePrioritySession(ctx, &req)
		require.Equal(t, err.Error(), "rpc error: code = InvalidArgument desc = Invalid Address for IpUserEquipment: "+badIp)
	}
}

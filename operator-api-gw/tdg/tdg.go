package tdg

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"

	"fmt"
	"math/rand"
	"time"
)

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	LocationVerificationURL string
	QosPosEndpoint          string
}

// make a random number near the range of the base value (plus or minus 10)
func randomInRange(baseval int) float32 {
	rand.Seed(time.Now().UnixNano())
	min := baseval - 10
	if min < 0 {
		min = 0
	}
	max := baseval + 10
	result := rand.Intn(max-min) + min
	// add some fraction less than 1
	return float32(result) + rand.Float32()
}

// getQosResults currently just returns some fake results
func getQosResults(qosres *dme.QosPositionResult) {
	qosres.DluserthroughputMin = randomInRange(1)
	qosres.DluserthroughputMax = randomInRange(100)
	qosres.DluserthroughputAvg = randomInRange(50)
	qosres.UluserthroughputMin = randomInRange(1)
	qosres.UluserthroughputMax = randomInRange(50)
	qosres.UluserthroughputAvg = randomInRange(25)
	qosres.LatencyMin = randomInRange(20)
	qosres.LatencyMax = randomInRange(60)
	qosres.LatencyAvg = randomInRange(40)
}

func getQosPositionKpi(mreq *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getQosPositionKpi", "request", mreq)

	var mreply dme.QosPositionKpiReply

	mreply.Status = dme.ReplyStatus_RS_SUCCESS

	for _, p := range mreq.Positions {
		pid := p.Positionid
		var qosres dme.QosPositionResult

		qosres.Positionid = pid
		qosres.GpsLocation = p.GpsLocation
		getQosResults(&qosres)
		log.DebugLog(log.DebugLevelDmereq, "Position", "pid", pid, "qosres", qosres)

		mreply.PositionResults = append(mreply.PositionResults, &qosres)
	}

	getQosSvr.Send(&mreply)
	return nil

}

func (OperatorApiGw) GetOperatorName() string {
	return "TDG"
}

// Init is called once during startup.
func (*OperatorApiGw) Init(operatorName string) error {
	return nil // nothing to do
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey, locVerUrl string, tokSrvUrl string) (*dme.VerifyLocationReply, error) {
	return nil, fmt.Errorf("Verify Location not yet implemented")
}

func (*OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "TDG GetQOSPositionKPI", "request", mreq)

	var mreply dme.QosPositionKpiReply

	mreply.Status = dme.ReplyStatus_RS_SUCCESS

	for _, p := range mreq.Positions {
		pid := p.Positionid
		var qosres dme.QosPositionResult

		qosres.Positionid = pid
		qosres.GpsLocation = p.GpsLocation
		getQosResults(&qosres)
		log.DebugLog(log.DebugLevelDmereq, "Position", "pid", pid, "qosres", qosres)

		mreply.PositionResults = append(mreply.PositionResults, &qosres)
	}

	getQosSvr.Send(&mreply)
	return nil

}

package defaultoperator

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

	"fmt"
)

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	LocationVerificationURL string
	QosPosEndpoint          string
}

func (OperatorApiGw) GetOperatorName() string {
	return "default"
}

// Init is called once during startup.
func (*OperatorApiGw) Init(operatorName string) error {
	return nil // nothing to do
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey, locVerUrl string, tokSrvUrl string) (*dme.VerifyLocationReply, error) {
	return nil, fmt.Errorf("Verify Location not supported for this operator")
}

func (*OperatorApiGw) GetQOSPositionKPI(req *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	return fmt.Errorf("Get QOS Position KPI not supported for this operator")
}

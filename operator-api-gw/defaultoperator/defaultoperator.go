package defaultoperator

import (
	simulatedqos "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/defaultoperator/simulated-qos"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"

	"fmt"
)

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	VaultAddr string
	QosPosUrl string
	LocVerUrl string
	TokSrvUrl string
}

func (OperatorApiGw) GetOperatorName() string {
	return "default"
}

// Init is called once during startup.
func (o *OperatorApiGw) Init(operatorName, vaultAddr, qosPosUrl, locVerUrl, tokSrvUrl string) error {
	log.DebugLog(log.DebugLevelDmereq, "init for default operator")
	o.QosPosUrl = qosPosUrl
	o.LocVerUrl = locVerUrl
	o.TokSrvUrl = tokSrvUrl
	o.VaultAddr = vaultAddr
	return nil
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey) (*dme.VerifyLocationReply, error) {
	return nil, fmt.Errorf("Verify Location not supported for this operator")
}

func (*OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getting simulated results for operator with no QOS Pos implementation")
	return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
}

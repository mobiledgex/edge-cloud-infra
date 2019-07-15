package tdg

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"

	"fmt"

	simulatedqos "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/defaultoperator/simulated-qos"
	tdgclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-qos/qosclient"
)

var QosClientCert = "qosclient.crt"
var QosClientKey = "qosclient.key"
var QoServerCert = "qosserver.crt"

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	VaultAddr string
	QosPosUrl string
	LocVerUrl string
	TokSrvUrl string
}

func (OperatorApiGw) GetOperatorName() string {
	return "TDG"
}

// Init is called once during startup.
func (o *OperatorApiGw) Init(operatorName, vaultAddr, qosPosUrl, locVerUrl, tokSrvUrl string) error {
	log.DebugLog(log.DebugLevelDmereq, "init for tdg operator")
	o.QosPosUrl = qosPosUrl
	o.LocVerUrl = locVerUrl
	o.TokSrvUrl = tokSrvUrl
	o.VaultAddr = vaultAddr

	if o.QosPosUrl != "" {
		err := tdgclient.GetQosCertsFromVault(o.VaultAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey) (*dme.VerifyLocationReply, error) {
	return nil, fmt.Errorf("Verify Location not yet implemented")
}

func (o *OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "TDG GetQOSPositionKPI", "QosPosUrl", o.QosPosUrl, "request", mreq)

	if o.QosPosUrl == "" {
		log.DebugLog(log.DebugLevelDmereq, "No QosPosUrl, getting simulated results")
		return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
	}
	return tdgclient.GetQOSPositionKPIFromApiGW(o.QosPosUrl, mreq, getQosSvr)
}

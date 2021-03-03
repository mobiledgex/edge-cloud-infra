package tdg

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	locclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-loc/locclient"
	qosclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-qos/qosclient"
	"github.com/mobiledgex/edge-cloud-infra/version"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	simulatedloc "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-location"
	simulatedqos "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-qos"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

var QosClientCert = "qosclient.crt"
var QosClientKey = "qosclient.key"
var QoServerCert = "qosserver.crt"

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	Servers     *operator.OperatorApiGwServers
	vaultConfig *vault.Config
}

func (OperatorApiGw) GetOperatorName() string {
	return "TDG"
}

// Init is called once during startup.
func (o *OperatorApiGw) Init(operatorName string, servers *operator.OperatorApiGwServers) error {
	log.DebugLog(log.DebugLevelDmereq, "init for tdg operator", "servers", servers)
	o.Servers = servers
	vaultConfig, err := vault.BestConfig(o.Servers.VaultAddr)
	if err != nil {
		return err
	}
	o.vaultConfig = vaultConfig

	if o.Servers.QosPosUrl != "" {
		err := qosclient.GetQosCertsFromVault(vaultConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error {

	log.DebugLog(log.DebugLevelDmereq, "TDG VerifyLocation", "request", mreq)

	if o.Servers.LocVerUrl == "" {
		// because this is so often used for demos, it is better to fail in a clear way
		// than to give a defaultoperator/fake result
		return grpc.Errorf(codes.InvalidArgument, "DME has no location verification server")
	}
	if o.Servers.TokSrvUrl == "" {
		return grpc.Errorf(codes.InvalidArgument, "DME has no token server")
	}
	if mreq.VerifyLocToken == "" {
		return grpc.Errorf(codes.InvalidArgument, "no VerifyLocToken in request")
	}
	// this is checked in the DME, but since we dereference it we will check it in this code as well
	if mreq.GpsLocation == nil {
		return grpc.Errorf(codes.InvalidArgument, "no GpsLocation in request")
	}

	result := locclient.CallTDGLocationVerifyAPI(o.Servers.LocVerUrl, mreq.GpsLocation.Latitude, mreq.GpsLocation.Longitude, mreq.VerifyLocToken, o.Servers.TokSrvUrl)
	mreply.GpsLocationStatus = result.MatchEngineLocStatus
	mreply.GpsLocationAccuracyKm = result.DistanceRange
	log.DebugLog(log.DebugLevelDmereq, "TDG VerifyLocation result", "mreply", mreply)
	return nil
}

func (o *OperatorApiGw) GetLocation(mreq *dme.GetLocationRequest, mreply *dme.GetLocationReply) error {
	log.DebugLog(log.DebugLevelDmereq, "TDG GetLocation", "request", mreq)
	// We have no real implementation of this
	return simulatedloc.GetSimulatedClientLoc(mreq, mreply)
}

func (o *OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "TDG GetQOSPositionKPI", "QosPosUrl", o.Servers.QosPosUrl, "request", mreq)

	if o.Servers.QosPosUrl == "" {
		log.DebugLog(log.DebugLevelDmereq, "No QosPosUrl, getting simulated results")
		return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
	}
	return qosclient.GetQOSPositionFromApiGW(o.Servers.QosPosUrl, mreq, getQosSvr)
}

func (*OperatorApiGw) GetVersionProperties() map[string]string {
	return version.InfraBuildProps("TDGOperator")
}

package tdg

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	locclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-loc/locclient"
	qosclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-qos/qosclient"
	sessionsclient "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-sessions/sessionsclient"
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

var qosSessionsApiKey string

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
	o.Servers = servers
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

func (o *OperatorApiGw) CreatePrioritySession(ctx context.Context, req *dme.QosPrioritySessionCreateRequest) (*dme.QosPrioritySessionReply, error) {
	var reply *dme.QosPrioritySessionReply
	var err error
	log.SpanLog(ctx, log.DebugLevelDmereq, "TDG CreatePrioritySession", "req", req)
	// Only retrieve this from the vault if we don't already have it.
	if qosSessionsApiKey == "" {
		qosSessionsApiKey, err = sessionsclient.GetApiKeyFromVault(ctx, o.vaultConfig)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmereq, "GetApiKeyFromVault failed. QOS priority session creation not supported.", "err", err)
		}
		if qosSessionsApiKey == "" {
			return nil, status.Errorf(codes.Unauthenticated, "missing qosSessionsApiKey")
		}
	}
	// build a request and send it in CallTDGQosPriorityAPI()
	reqBody := sessionsclient.QosSessionRequest{}
	reqBody.UeAddr = req.IpUserEquipment
	reqBody.AsAddr = req.IpApplicationServer
	reqBody.UePorts = req.PortUserEquipment
	reqBody.AsPorts = req.PortApplicationServer
	reqBody.ProtocolIn = req.ProtocolIn.String()
	reqBody.ProtocolOut = req.ProtocolOut.String()
	reqBody.Qos = req.Profile.String()
	reqBody.Duration = int64(req.SessionDuration)
	if net.ParseIP(req.IpUserEquipment) == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Address for IpUserEquipment: %s", req.IpUserEquipment)
	}
	if net.ParseIP(req.IpApplicationServer) == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Address for IpApplicationServer: %s", req.IpApplicationServer)
	}
	reply, err = sessionsclient.CallTDGQosPriorityAPI(ctx, "", http.MethodPost, o.Servers.QosSesAddr, qosSessionsApiKey, reqBody)
	return reply, err
}

func (o *OperatorApiGw) DeletePrioritySession(ctx context.Context, req *dme.QosPrioritySessionDeleteRequest) (*dme.QosPrioritySessionDeleteReply, error) {
	sesInfo := sessionsclient.QosSessionRequest{}
	// Only the Qos (profile name) field is needed for delete.
	sesInfo.Qos = req.Profile.String()
	sessionId := req.SessionId
	log.SpanLog(ctx, log.DebugLevelDmereq, "TDG DeletePrioritySession", "sessionId", sessionId)
	// Get a generic QosPrioritySessionReply, then build a QosPrioritySessionDeleteReply based on the httpStatus.
	reply, err := sessionsclient.CallTDGQosPriorityAPI(ctx, sessionId, http.MethodDelete, o.Servers.QosSesAddr, qosSessionsApiKey, sesInfo)
	log.SpanLog(ctx, log.DebugLevelDmereq, "Response from TDG:", "reply", reply, "err", err)
	deleteReply := new(dme.QosPrioritySessionDeleteReply)
	if reply.HttpStatus == http.StatusNotFound {
		deleteReply.Status = dme.QosPrioritySessionDeleteReply_QDEL_NOT_FOUND
	} else if reply.HttpStatus == http.StatusNoContent {
		deleteReply.Status = dme.QosPrioritySessionDeleteReply_QDEL_DELETED
	} else {
		deleteReply.Status = dme.QosPrioritySessionDeleteReply_QDEL_UNKNOWN
	}
	return deleteReply, err
}

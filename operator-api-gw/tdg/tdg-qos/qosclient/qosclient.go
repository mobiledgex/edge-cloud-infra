package qosclient

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	tdgproto "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/proto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/vault"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

var clientCert = "qosclient.crt"
var clientKey = "qosclient.key"
var serverCert = "qosserver.crt"

var nextRequestId int64 = 1

func GetQosCertsFromVault(vaultConfig *vault.Config) error {
	log.DebugLog(log.DebugLevelDmereq, "GetQosCertsFromVault", "vaultAddr", vaultConfig.Addr)

	certs := []string{clientCert, clientKey, serverCert}
	for _, cert := range certs {

		certPath := fmt.Sprintf("/secret/data/accounts/tdg/qosapi/%s", cert)
		log.DebugLog(log.DebugLevelDmereq, "Fetching Cert", "certPath", certPath)
		fileName := "/tmp/" + cert
		err := infracommon.GetVaultDataToFile(vaultConfig, certPath, fileName)
		if err != nil {
			return grpc.Errorf(codes.Internal, "Unable to get cert from file: %s, %v", cert, err)
		}
	}
	return nil
}

func GetQOSPositionFromApiGW(serverUrl string, mreq *dme.QosPositionRequest, qosKpiServer dme.MatchEngineApi_GetQosPositionKpiServer) error {
	serverCertFile := "/tmp/" + serverCert
	clientCertFile := "/tmp/" + clientCert

	if mreq.Positions == nil {
		return grpc.Errorf(codes.InvalidArgument, "No positions requested")
	}

	// in case the responses come put of order, we need to be able to lookup the GPS coordinate of the request positionid
	var positionIdToGps = make(map[int64]*dme.Loc)

	log.DebugLog(log.DebugLevelDmereq, "Connecting to QOS API GW", "serverUrl", serverUrl)

	tlsConfig, err := edgetls.GetTLSClientConfig(serverUrl, clientCertFile, serverCertFile, false, nil)
	if err != nil {
		return grpc.Errorf(codes.Unavailable, "Unable get TLS Client config: %v", err)
	}

	transportCreds := credentials.NewTLS(tlsConfig)
	dialOption := grpc.WithTransportCredentials(transportCreds)
	conn, err := grpc.Dial(serverUrl, dialOption)
	if err != nil {
		return grpc.Errorf(codes.Unavailable, "Unable to connect to API GW: %s, %v", serverUrl, err)
	}
	defer conn.Close()
	ctx := context.TODO()
	defaultTimestamp := time.Now().Unix() + 1000
	var request tdgproto.QoSKPIRequest
	for _, p := range mreq.Positions {
		var posreq tdgproto.PositionKpiRequest
		posreq.Positionid = p.Positionid
		if p.GpsLocation == nil {
			return grpc.Errorf(codes.InvalidArgument, "Missing GPS Location in request")
		}
		positionIdToGps[posreq.Positionid] = p.GpsLocation
		posreq.Latitude = float32(p.GpsLocation.Latitude)
		posreq.Longitude = float32(p.GpsLocation.Longitude)
		posreq.Altitude = float32(p.GpsLocation.Altitude)
		if p.GpsLocation.Timestamp == nil {
			posreq.Timestamp = defaultTimestamp
		} else {
			posreq.Timestamp = p.GpsLocation.Timestamp.Seconds
		}
		request.Requests = append(request.Requests, &posreq)
	}
	if mreq.BandSelection != nil {
		request.Bandselection = new(tdgproto.BandSelection)
		request.Bandselection.RAT2G = mreq.BandSelection.Rat_2G
		request.Bandselection.RAT3G = mreq.BandSelection.Rat_3G
		request.Bandselection.RAT4G = mreq.BandSelection.Rat_4G
		request.Bandselection.RAT5G = mreq.BandSelection.Rat_5G
	}
	request.Ltecategory = mreq.LteCategory
	request.Requestid = nextRequestId
	nextRequestId++

	log.DebugLog(log.DebugLevelDmereq, "Sending request to API GW", "request", request)

	qosClient := tdgproto.NewQueryQoSClient(conn)
	stream, err := qosClient.QueryQoSKPI(ctx, &request)
	if err != nil {
		return fmt.Errorf("QueryQoSKPI error: %v", err)
	}
	stream.CloseSend()
	for {
		// convert the DT format to the MEX format and stream the replies
		var mreply dme.QosPositionKpiReply
		res, err := stream.Recv()
		if err == io.EOF {
			log.DebugLog(log.DebugLevelDmereq, "EOF received")
			err = nil
			break
		}
		if err != nil {
			break
		}
		log.DebugLog(log.DebugLevelDmereq, "Recv done", "resultLen", len(res.Results), "err", err)

		for _, r := range res.Results {
			var qosres dme.QosPositionKpiResult
			qosres.Positionid = r.Positionid
			gps, ok := positionIdToGps[qosres.Positionid]
			if !ok {
				return grpc.Errorf(codes.Internal, "PositionId %d found in response but not request", qosres.Positionid)
			}
			qosres.GpsLocation = gps
			qosres.UluserthroughputMin = r.UluserthroughputMin
			qosres.UluserthroughputMax = r.UluserthroughputMax
			qosres.UluserthroughputAvg = r.UluserthroughputAvg
			qosres.DluserthroughputMin = r.DluserthroughputMin
			qosres.DluserthroughputMax = r.DluserthroughputMax
			qosres.DluserthroughputAvg = r.DluserthroughputAvg
			qosres.LatencyMin = r.LatencyMin
			qosres.LatencyMin = r.LatencyMax
			qosres.LatencyMin = r.LatencyAvg

			mreply.PositionResults = append(mreply.PositionResults, &qosres)
		}
		mreply.Status = dme.ReplyStatus_RS_SUCCESS
		qosKpiServer.Send(&mreply)
	}

	log.DebugLog(log.DebugLevelDmereq, "Done receiving responses")
	return err

}

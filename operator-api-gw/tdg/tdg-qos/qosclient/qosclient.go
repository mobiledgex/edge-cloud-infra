package qosclient

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	tdgproto "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/proto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
)

var clientCert = "qosclient.crt"
var clientKey = "qosclient.key"
var serverCert = "qosserver.crt"
var nextRequestId int64 = 1

func GetQosCertsFromVault(vaultAddr string) error {
	log.DebugLog(log.DebugLevelDmereq, "GetQosCertsFromVault", "vaultAddr", vaultAddr)

	certs := []string{clientCert, clientKey, serverCert}
	for _, cert := range certs {

		certURL := fmt.Sprintf("%s/v1/secret/data/accounts/tdg/qosapi/%s", vaultAddr, cert)
		log.DebugLog(log.DebugLevelDmereq, "Fetching Cert", "certURL", certURL)
		fileName := "/tmp/" + cert
		err := mexos.GetVaultDataToFile(certURL, fileName)
		if err != nil {
			return fmt.Errorf("Unable to get cert from file: %s, %v", cert, err)
		}
	}
	return nil
}

func GetQOSPositionKPIFromApiGW(serverUrl string, mreq *dme.QosPositionKpiRequest, qosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {

	serverCertFile := "/tmp/" + serverCert
	clientCertFile := "/tmp/" + clientCert

	if mreq.Positions == nil {
		return fmt.Errorf("No positions requested")
	}

	// in case the responses come put of order, we need to be able to lookup the GPS coordinate of the request positionid
	var positionIdToGps = make(map[int64]*dme.Loc)

	log.DebugLog(log.DebugLevelDmereq, "Connecting to QOS API GW", "serverUrl", serverUrl)

	tlsConfig, err := edgetls.GetTLSClientConfig(serverUrl, clientCertFile, serverCertFile, false)
	if err != nil {
		return fmt.Errorf("Unable get TLS Client config: %v", err)
	}

	transportCreds := credentials.NewTLS(tlsConfig)
	dialOption := grpc.WithTransportCredentials(transportCreds)
	conn, err := grpc.Dial(serverUrl, dialOption)
	if err != nil {
		return fmt.Errorf("Unable to connect to API GW: %s, %v", serverUrl, err)
	}
	defer conn.Close()
	ctx := context.TODO()
	defaultTimestamp := time.Now().Unix() + 1000
	var request tdgproto.QoSKPIRequest
	for _, p := range mreq.Positions {
		var posreq tdgproto.PositionKpiRequest
		posreq.Positionid = p.Positionid
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

	request.Requestid = nextRequestId
	nextRequestId++

	qosClient := tdgproto.NewQueryQoSClient(conn)
	stream, err := qosClient.QueryQoSKPI(ctx, &request)
	stream.CloseSend()

	if err != nil {
		return fmt.Errorf("QueryQoSKPI error: %v", err)
	}

	for {
		log.DebugLog(log.DebugLevelDmereq, "Receiving responses", "serverUrl", serverUrl)
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
			var qosres dme.QosPositionResult
			qosres.Positionid = r.Positionid
			gps, ok := positionIdToGps[qosres.Positionid]
			if !ok {
				return fmt.Errorf("PositionId %d found in response but not request", qosres.Positionid)
			}
			qosres.GpsLocation = gps
			qosres.UluserthroughputMin = r.GetUluserthroughputMin()
			qosres.UluserthroughputMax = r.GetUluserthroughputMax()
			qosres.UluserthroughputAvg = r.GetUluserthroughputAvg()
			qosres.DluserthroughputMin = r.GetDluserthroughputMin()
			qosres.DluserthroughputMax = r.GetDluserthroughputMax()
			qosres.DluserthroughputAvg = r.GetDluserthroughputAvg()
			qosres.LatencyMin = r.GetLatencyMin()
			qosres.LatencyMin = r.GetLatencyMax()
			qosres.LatencyMin = r.GetLatencyAvg()

			mreply.PositionResults = append(mreply.PositionResults, &qosres)
		}
		mreply.Status = dme.ReplyStatus_RS_SUCCESS
		qosSvr.Send(&mreply)

	}
	log.DebugLog(log.DebugLevelDmereq, "Done receiving responses")
	return err

}

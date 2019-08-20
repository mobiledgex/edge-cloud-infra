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

var RequestTypeKPI = "RequestTypeKPI"
var RequestTypeClassifier = "RequestTypeClassifier"

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

func GetQOSPositionFromApiGW(serverUrl string, mreq *dme.QosPositionRequest, qosSvr interface{}, requestType string) error {
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

	switch requestType {
	case RequestTypeKPI:
		qosKpiServer, ok := qosSvr.(dme.MatchEngineApi_GetQosPositionKpiServer)
		if !ok {
			return fmt.Errorf("unable to cast client to GetQosPositionKpiServer")
		}
		qosClient := tdgproto.NewQueryQoSClient(conn)
		stream, err := qosClient.QueryQoSKPI(ctx, &request)
		if err != nil {
			return fmt.Errorf("Error getting stream: %v", err)
		}
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
				var qosres dme.QosPositionKpiResult
				qosres.Positionid = r.Positionid
				gps, ok := positionIdToGps[qosres.Positionid]
				if !ok {
					return fmt.Errorf("PositionId %d found in response but not request", qosres.Positionid)
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

	case RequestTypeClassifier:
		qosClassifierServer, ok := qosSvr.(dme.MatchEngineApi_GetQosPositionClassifierServer)
		if !ok {
			return fmt.Errorf("unable to cast client to GetQosPositionClassifierServer")
		}
		qosClient := tdgproto.NewQueryQoSClient(conn)
		stream, err := qosClient.QueryQoSKPIClassifier(ctx, &request)
		if err != nil {
			return fmt.Errorf("Error getting stream: %v", err)
		}
		stream.CloseSend()

		if err != nil {
			return fmt.Errorf("QueryQoSKPI error: %v", err)
		}
		for {
			log.DebugLog(log.DebugLevelDmereq, "Receiving responses", "serverUrl", serverUrl)
			// convert the DT format to the MEX format and stream the replies
			var mreply dme.QosPositionClassifierReply
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
				var qosres dme.QosPositionClassifierResult
				qosres.Positionid = r.Positionid
				gps, ok := positionIdToGps[qosres.Positionid]
				if !ok {
					return fmt.Errorf("PositionId %d found in response but not request", qosres.Positionid)
				}
				qosres.GpsLocation = gps
				qosres.UluserthroughputClass = r.UluserthroughputClass
				qosres.DluserthroughputClass = r.DluserthroughputClass
				qosres.LatencyClass = r.LatencyClass
				mreply.PositionResults = append(mreply.PositionResults, &qosres)
			}
			mreply.Status = dme.ReplyStatus_RS_SUCCESS
			qosClassifierServer.Send(&mreply)
		}
	default:
		// this is a bug
		return fmt.Errorf("invalid request type: %s", requestType)
	}
	log.DebugLog(log.DebugLevelDmereq, "Done receiving responses")
	return err

}

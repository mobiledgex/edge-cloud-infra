package locclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud-infra/operator-api-gw/tdg/tdg-loc/util"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type LocationResponseMessage struct {
	MatchingDegree string `json:"matchingDegree"`
	Message        string `json:"message"`
}

//format of the HTTP request body.  Token is used for validation of location, but
//IP address is still present to allow locations to be updated for the simulator
type LocationRequestMessage struct {
	Lat        float64       `json:"latitude" yaml:"lat"`
	Long       float64       `json:"longitude" yaml:"long"`
	Token      util.TDGToken `json:"token" yaml:"token"`
	Ipaddress  string        `json:"ipaddr,omitempty" yaml:"ipaddr"`
	ServiceURL string        `json:"serviceUrl,omitempty" yaml:"serviceUrl"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// CallTDGLocationVerifyAPI REST API client for the TDG implementation of Location verification API
func CallTDGLocationVerifyAPI(locVerUrl string, lat, long float64, token string, tokSrvUrl string) dmecommon.LocationResult {

	//for TDG, the serviceURL is the value of the query parameter "followURL" in the token service URL
	u, err := url.Parse(tokSrvUrl)
	if err != nil {
		// should never happen unless there is a provisioning error
		log.WarnLog("Error, cannot parse tokSrvUrl", "url", tokSrvUrl)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	qvals := u.Query()
	serviceURL := qvals.Get("followURL")
	if serviceURL == "" {
		log.WarnLog("Error, no followURL in tokSrvUrl", "url", tokSrvUrl, "qvals", qvals)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	// If the service URL needs to be urlencoded, uncomment this.  Currently it is not
	// serviceURL = url.PathEscape(serviceURL)
	var lrm LocationRequestMessage
	lrm.Lat = lat
	lrm.Long = long
	lrm.Token = util.TDGToken(token)
	lrm.ServiceURL = serviceURL

	b, err := json.Marshal(lrm)
	if err != nil {
		log.WarnLog("error in json mashal of request", "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}

	body := bytes.NewBufferString(string(b))
	req, err := http.NewRequest("POST", locVerUrl, body)

	if err != nil {
		log.WarnLog("error in http.NewRequest", "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	req.Header.Add("Content-Type", "application/json")
	username := os.Getenv("LOCAPI_USER")
	password := os.Getenv("LOCAPI_PASSWD")

	if username != "" {
		log.DebugLog(log.DebugLevelLocapi, "adding auth header", "username", username)
		req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	} else {
		log.DebugLog(log.DebugLevelLocapi, "no auth credentials")
	}
	client := &http.Client{}
	log.DebugLog(log.DebugLevelLocapi, "sending to api gw", "body:", body)

	resp, err := client.Do(req)

	if err != nil {
		log.WarnLog("Error in POST to TDG Loc service error", "error", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	defer resp.Body.Close()

	log.DebugLog(log.DebugLevelLocapi, "Received response", "statusCode:", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:
		log.DebugLog(log.DebugLevelLocapi, "200OK received")

	//treat 401 or 403 as a token issue.  Handling with TDG to be confirmed
	case http.StatusForbidden:
		fallthrough
	case http.StatusUnauthorized:
		log.WarnLog("returning VerifyLocationReply_LOC_ERROR_UNAUTHORIZED", "received code", resp.StatusCode)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_UNAUTHORIZED}
	default:
		log.WarnLog("returning VerifyLocationReply_LOC_ERROR_OTHER", "received code", resp.StatusCode)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}

	respBytes, resperr := ioutil.ReadAll(resp.Body)

	if resperr != nil {
		log.WarnLog("Error read response body", "resperr", resperr)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	var lrmResp LocationResponseMessage

	//resp = string(respBytes)
	err = json.Unmarshal(respBytes, &lrmResp)
	if err != nil {
		log.WarnLog("Error unmarshall response", "respBytes", respBytes, "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}

	log.DebugLog(log.DebugLevelLocapi, "unmarshalled location response", "lrmResp:", lrmResp)
	md, err := strconv.ParseInt(lrmResp.MatchingDegree, 10, 32)
	if err != nil {
		log.WarnLog("Error in LocationResult", "LocationResult", lrmResp.MatchingDegree, "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
	}
	if md < 0 {
		log.DebugLog(log.DebugLevelLocapi, "Invalid Matching degree received", "Message:", lrmResp.Message)
		if strings.Contains(lrmResp.Message, "invalidToken") {
			rc := dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_UNAUTHORIZED}
			log.DebugLog(log.DebugLevelLocapi, "Invalid token", "result", rc)
			return rc
		}
		rc := dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.VerifyLocationReply_LOC_ERROR_OTHER}
		log.DebugLog(log.DebugLevelLocapi, "other error", "result", rc)
		return rc
	}

	rc := dmecommon.GetDistanceAndStatusForLocationResult(uint32(md))
	log.DebugLog(log.DebugLevelLocapi, "Returning result", "Location Result", rc)

	return rc
}

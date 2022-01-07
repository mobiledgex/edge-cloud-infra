package sessionsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

func GetApiKeyFromVault(ctx context.Context, vaultConfig *vault.Config) (string, error) {
	apiKeyPath := "/secret/data/accounts/tdg/sessionsapi"
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetApiKeyFromVault", "vaultAddr", vaultConfig.Addr, "apiKeyPath", apiKeyPath)
	type ApiKeyData struct {
		Data string
	}
	var apiKeyData ApiKeyData
	err := vault.GetData(vaultConfig, apiKeyPath, 0, &apiKeyData)
	if err != nil {
		return "", fmt.Errorf("unable to fetch QOS API key from vault. err=%v", err)
	}
	apiKey := apiKeyData.Data

	return apiKey, err
}

// From https://staging-portal.hubraum.opsfactory.dev/de/products/617bd0928431ba00019948f4/summary
type QosSessionCommon struct {
	Duration              int64  `json:"duration" yaml:"duration"`
	UeAddr                string `json:"ueAddr" yaml:"ueAddr"`
	AsAddr                string `json:"asAddr" yaml:"asAddr"`
	UePorts               string `json:"uePorts" yaml:"uePorts"`
	AsPorts               string `json:"asPorts" yaml:"asPorts"`
	ProtocolIn            string `json:"protocolIn" yaml:"protocolIn"`
	ProtocolOut           string `json:"protocolOut" yaml:"protocolOut"`
	Qos                   string `json:"qos" yaml:"qos"`
	NotificationUri       string `json:"notificationUri" yaml:"notificationUri"`
	NotificationAuthToken string `json:"notificationAuthToken" yaml:"notificationAuthToken"`
}

type QosSessionRequest struct {
	QosSessionCommon
}

type QosSessionResponse struct {
	QosSessionCommon
	Id        string `json:"id" yaml:"id"`
	StartedAt int64  `json:"startedAt" yaml:"startedAt"`
	ExpiresAt int64  `json:"expiresAt" yaml:"expiresAt"`
}

// Build and send the request to the TDG API server
func sendRequest(ctx context.Context, method string, reqUrl string, apiKey string, body *bytes.Buffer) (int, string, error) {
	var req *http.Request
	var err error

	// I'm surprised this is necessary, and surprised that it works.
	// See https://github.com/golang/go/issues/32897
	if body == nil {
		req, err = http.NewRequest(method, reqUrl, nil)
	} else {
		req, err = http.NewRequest(method, reqUrl, body)
	}

	if err != nil {
		log.WarnLog("error in http.NewRequest", "err", err)
		return 0, "", err
	}

	log.SpanLog(ctx, log.DebugLevelDmereq, "Sending to TDG:", "method", method, "reqUrl:", reqUrl, "body:", body)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.WarnLog("Error in REST call to TDG QOS session priority service", "error", err)
		return 0, "", err
	}
	defer resp.Body.Close()

	log.SpanLog(ctx, log.DebugLevelDmereq, "Received response", "statusCode:", resp.StatusCode)

	respBytes, resperr := ioutil.ReadAll(resp.Body)
	if resperr != nil {
		log.WarnLog("Error reading response body", "resperr", resperr)
		return 0, "", resperr
	}
	respString := string(respBytes)
	log.SpanLog(ctx, log.DebugLevelDmereq, "Received response", "respString", respString)

	return resp.StatusCode, respString, nil
}

func buildQosUrl(ctx context.Context, profileName string, qosSesAddr string) (string, error) {
	var priorityType string
	if profileName == "LOW_LATENCY" { // LOW_LATENCY is the only valid latency profile.
		priorityType = "latency"
	} else if strings.HasPrefix(profileName, "THROUGHPUT") {
		priorityType = "throughput"
	} else {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Received invalid value", "profileName", profileName)
		return "", errors.New("Received invalid profileName" + profileName)
	}
	url := fmt.Sprintf("https://%s/5g-%s/sessions", qosSesAddr, priorityType) // Inserts either "latency" or "throughput".
	log.SpanLog(ctx, log.DebugLevelDmereq, "buildQosUrl", "url", url)
	return url, nil
}

// CallTDGQosPriorityAPI REST API client for the TDG implementation of QOS session priority API
func CallTDGQosPriorityAPI(ctx context.Context, sesId string, method string, qosSesAddr string, apiKey string, reqBody QosSessionRequest) (string, error) {
	qos := reqBody.Qos
	reqUrl, err := buildQosUrl(ctx, qos, qosSesAddr)
	if err != nil {
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "TDG CallTDGQosPriorityAPI", "qosSesAddr", qosSesAddr, "reqUrl", reqUrl, "reqBody", reqBody)
	out, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	body := bytes.NewBuffer(out)

	// For DELETE, we need to add the session ID to the URL
	if method == http.MethodDelete {
		reqUrl += "/" + sesId
	}
	status, respBody, err := sendRequest(ctx, method, reqUrl, apiKey, body)
	if err != nil {
		return "", err
	}

	var qsiResp QosSessionResponse
	var sessionId string

	if status == http.StatusConflict {
		log.SpanLog(ctx, log.DebugLevelDmereq, "409 Conflict received")
		// In this case, the session already exists. Look up the session info and see if it needs updated.
		// Example respBody:
		// "Found session 9aa00f58-38f6-4ed9-be8f-d375aad95721 already active until 2021-12-03T21:06:02Z"
		if strings.HasPrefix(respBody, "Found session") {
			words := strings.Split(respBody, " ")
			if len(words) < 3 {
				return "", fmt.Errorf(fmt.Sprintf("Could not parse response: %s", respBody))
			}
			sessionId = words[2]
			url := fmt.Sprintf("%s/%s", reqUrl, sessionId)
			status, respBody, err = sendRequest(ctx, http.MethodGet, url, apiKey, nil)
			if err != nil {
				return "", err
			}
			if status == http.StatusOK {
				respBytes := []byte(respBody)
				err = json.Unmarshal(respBytes, &qsiResp)
				if err != nil {
					log.WarnLog("Error unmarshalling response", "respBytes", respBytes, "err", err)
					return "", err
				}
				if qsiResp.Qos == reqBody.Qos {
					log.SpanLog(ctx, log.DebugLevelDmereq, "Requested QOS session already exists. Keeping it.", "qsiResp.Qos", qsiResp.Qos)
					sessionId = qsiResp.Id
				} else {
					log.SpanLog(ctx, log.DebugLevelDmereq, "Existing QOS profile doesn't match. Deleting session.")
					oldQos := qsiResp.Qos
					url, err := buildQosUrl(ctx, oldQos, qosSesAddr)
					url = fmt.Sprintf("%s/%s", url, sessionId)
					if err != nil {
						return "", err
					}
					status, _, err := sendRequest(ctx, http.MethodDelete, url, apiKey, nil)
					if err != nil {
						return "", err
					}
					if status == http.StatusNoContent {
						log.SpanLog(ctx, log.DebugLevelDmereq, "Successfully deleted QOS session")
					} else {
						return "", fmt.Errorf(fmt.Sprintf("Failed to delete existing QOS session: Error code: %d", status))
					}

					// Send new request to create session with desired QOS profile.
					body := bytes.NewBuffer(out)
					status, respBody, err = sendRequest(ctx, method, reqUrl, apiKey, body)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}

	// This value of 'status' can be from the initial call, or from the delete/retry attempt.
	if status == http.StatusCreated {
		log.SpanLog(ctx, log.DebugLevelDmereq, "201 Session Created received")
		respBytes := []byte(respBody)
		err = json.Unmarshal(respBytes, &qsiResp)
		if err != nil {
			log.WarnLog("Error unmarshalling response", "respBytes", respBytes, "err", err)
			return "", err
		}
		sessionId = qsiResp.Id
		log.SpanLog(ctx, log.DebugLevelDmereq, "unmarshalled response", "qsiResp:", qsiResp, "sessionId", sessionId)
	} else if status == http.StatusOK {
		log.SpanLog(ctx, log.DebugLevelDmereq, "200 OK received")
	} else if status == http.StatusNoContent {
		log.SpanLog(ctx, log.DebugLevelDmereq, "204 No Content received (session deleted)")
	} else {
		log.WarnLog("returning error", "received ", status)
		return "", fmt.Errorf(fmt.Sprintf("API call received unknown status: %d", status))
	}

	return sessionId, nil
}

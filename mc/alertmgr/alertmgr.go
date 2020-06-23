package alertmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/prometheus/alertmanager/api/v2/models"
)

var alertRefreshInterval = 30 * time.Second

const (
	AlertApi    string = "api/v2/alerts"
	ReceiverApi string = "api/v2/receivers"
	SilenceApi  string = "api/v2/silences"
)

// AlertMrgServer does two things - it periodically updates AlertManager about the
// current alerts on the system, and also handles configuration for the alert receivers
// i.e. backend handlers for the MC apis.
// NOTE: it does not perform any RBAC control here - this is done in ORM handlers
type AlertMrgServer struct {
	AlertMrgAddr            string
	McAlertmanagerAgentName string
	AlertCache              *edgeproto.AlertCache
	waitGrp                 sync.WaitGroup
	stop                    chan struct{}
}

// TODO - use version to track where this alert came from
func setAgentName() string {
	return "MasterControllerV1"
}

func NewAlertMgrServer(alertMgrAddr string, alertCache *edgeproto.AlertCache) *AlertMrgServer {
	server := AlertMrgServer{
		AlertMrgAddr:            alertMgrAddr,
		AlertCache:              alertCache,
		McAlertmanagerAgentName: setAgentName(),
	}
	return &server
}

// Update callback for a new alert - should send to alertmanager right away
func (s *AlertMrgServer) UpdateAlert(ctx context.Context, old *edgeproto.Alert, new *edgeproto.Alert) {
	s.AddAlerts(ctx, new)
}

func (s *AlertMrgServer) Start() {
	s.stop = make(chan struct{})
	s.waitGrp.Add(1)
	go s.runServer()
}

func (s *AlertMrgServer) runServer() {
	done := false
	for !done {
		// check if there are any new apps we need to start/stop scraping for
		select {
		case <-time.After(alertRefreshInterval):
			span := log.StartSpan(log.DebugLevelInfo, "alert-mgr")
			ctx := log.ContextWithSpan(context.Background(), span)
			log.SpanLog(ctx, log.DebugLevelInfo, "Sending Alerts to AlertMgr", "AlertMrgAddr",
				s.AlertMrgAddr)
			curAlerts := []*edgeproto.Alert{}
			s.AlertCache.Show(&edgeproto.Alert{}, func(obj *edgeproto.Alert) error {
				curAlerts = append(curAlerts, obj)
				return nil
			})
			// Send out alerts if any alerts need updating
			if len(curAlerts) > 0 {
				err := s.AddAlerts(ctx, curAlerts...)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "Error sending Alerts to AlertMgr", "AlertMrgAddr",
						s.AlertMrgAddr, "err", err)
				}
			}
			span.Finish()
		case <-s.stop:
			done = true
		}
	}
	s.waitGrp.Done()
}

func (s *AlertMrgServer) Stop() {
	close(s.stop)
	s.waitGrp.Wait()
}

func alertsToOpenAPIAlerts(alerts []*edgeproto.Alert) models.PostableAlerts {
	openAPIAlerts := models.PostableAlerts{}
	for _, a := range alerts {
		start := strfmt.DateTime(time.Unix(a.ActiveAt.Seconds, int64(a.ActiveAt.Nanos)))
		// Add region label to differentiate these at the global level
		labels := make(map[string]string)
		for k, v := range a.Labels {
			labels[k] = v
		}
		labels["region"] = a.Region
		openAPIAlerts = append(openAPIAlerts, &models.PostableAlert{
			Annotations: copyMap(a.Annotations),
			StartsAt:    start,
			Alert: models.Alert{
				Labels: copyMap(labels),
			},
		})
	}
	return openAPIAlerts
}

func copyMap(labels map[string]string) map[string]string {
	apiLabelSet := models.LabelSet{}
	for k, v := range labels {
		apiLabelSet[k] = v
	}
	return apiLabelSet
}

func alertManagerAlertsToEdgeprotoAlerts(openAPIAlerts models.GettableAlerts) []edgeproto.Alert {
	alerts := []edgeproto.Alert{}
	for _, openAPIAlert := range openAPIAlerts {
		alert := edgeproto.Alert{}
		if openAPIAlert.StartsAt != nil {
			alert.ActiveAt = dme.Timestamp{
				Seconds: time.Time(*openAPIAlert.StartsAt).Unix(),
				Nanos:   int32(time.Time(*openAPIAlert.StartsAt).UnixNano()),
			}
		}
		alert.Labels = copyMap(openAPIAlert.Labels)
		// Populate region with label value
		if region, found := alert.Labels["region"]; found {
			alert.Region = region
			delete(alert.Labels, "region")
		}
		alert.Annotations = copyMap(openAPIAlert.Annotations)
		alerts = append(alerts, alert)
	}
	return alerts
}

// Show all alerts in the alertmgr
// TODO Future: alerts api can take filters to make rbac simpler
func (s *AlertMrgServer) ShowAlerts(ctx context.Context, filter *edgeproto.Alert) ([]edgeproto.Alert, error) {
	data, err := s.alertMgrApi(ctx, "GET", AlertApi, "", nil)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to GET Alerts", "err", err, "filter", filter)
		return nil, err
	}
	openAPIAlerts := models.GettableAlerts{}
	err = json.Unmarshal(data, &openAPIAlerts)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to unmarshal Alerts", "err", err, "data", data)
		return nil, err
	}
	alerts := alertManagerAlertsToEdgeprotoAlerts(openAPIAlerts)
	return alerts, nil
}

// Marshal edgeproto.Alert into json payload suitabe for alertmanager api
func (s *AlertMrgServer) AddAlerts(ctx context.Context, alerts ...*edgeproto.Alert) error {

	openApiAlerts := alertsToOpenAPIAlerts(alerts)
	data, err := json.Marshal(openApiAlerts)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to marshal alerts", "err", err, "alerts", alerts)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "marshal alerts", "alerts", string(data))
	_, err = s.alertMgrApi(ctx, "POST", AlertApi, "", data)
	return err
}

// Common function to send an api call to alertmanager
func (s *AlertMrgServer) alertMgrApi(ctx context.Context, method, api, options string, payload []byte) ([]byte, error) {
	url := s.AlertMrgAddr + "/" + api
	if options != "" {
		url += "?" + options
	}
	client := http.DefaultClient
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create a new alerts request", "err", err, "url", url)
		return nil, err
	}
	req.Header.Set("User-Agent", s.McAlertmanagerAgentName)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to send request to the alertmanager", "err", err, "request", req)
		return nil, err
	}
	defer resp.Body.Close()
	// HTTP status 2xx is ok
	if resp.StatusCode/100 != 2 {
		log.SpanLog(ctx, log.DebugLevelInfo, "Alertmanager responded with an error", "request", req, "response", resp)
		return nil, fmt.Errorf("bad response status %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to read response body", "request", req, "response", resp)
		return nil, err
	}
	return body, nil
}

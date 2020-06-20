package alertmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
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
	// Add region label
	new.Labels["region"] = new.Region
	s.AlertMgrAddAlerts(ctx, new)
}

func (s *AlertMrgServer) Start() {
	s.stop = make(chan struct{})
	s.waitGrp.Add(1)
	go s.runServer()
}

func (s *AlertMrgServer) runServer() {
	// TODO - start thread to send alerts to alertMrg periordically
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
				obj.Labels["region"] = obj.Region
				curAlerts = append(curAlerts, obj)
				return nil
			})
			err := s.AlertMgrAddAlerts(ctx, curAlerts...)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Error sending Alerts to AlertMgr", "AlertMrgAddr",
					s.AlertMrgAddr, "err", err)
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
		openAPIAlerts = append(openAPIAlerts, &models.PostableAlert{
			Annotations: labelsToOpenAPILabelSet(a.Annotations),
			StartsAt:    start,
			Alert: models.Alert{
				Labels: labelsToOpenAPILabelSet(a.Labels),
			},
		})
	}

	return openAPIAlerts
}

func labelsToOpenAPILabelSet(labels map[string]string) models.LabelSet {
	apiLabelSet := models.LabelSet{}
	for k, v := range labels {
		apiLabelSet[k] = v
	}
	return apiLabelSet
}

// Marshal edgeproto.Alert into json payload suitabe for alertmanager api
func (s *AlertMrgServer) AlertMgrAddAlerts(ctx context.Context, alerts ...*edgeproto.Alert) error {

	openApiAlerts := alertsToOpenAPIAlerts(alerts)
	data, err := json.Marshal(openApiAlerts)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to marshal alerts", "err", err, "alerts", alerts)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "marshal alerts", "alerts", string(data))
	return s.alertMgrApi(ctx, "POST", AlertApi, data)
}

// Common function to send an api call to alertmanager
func (s *AlertMrgServer) alertMgrApi(ctx context.Context, method, api string, payload []byte) error {
	url := s.AlertMrgAddr + "/" + api
	client := http.DefaultClient
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to create a new alerts request", "err", err, "url", url)
		return err
	}
	req.Header.Set("User-Agent", s.McAlertmanagerAgentName)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to send request to the alertmanager", "err", err, "request", req)
		return err
	}
	// HTTP status 2xx is ok
	if resp.StatusCode/100 != 2 {
		log.SpanLog(ctx, log.DebugLevelInfo, "Alertmanager responded with an error", "request", req, "response", resp)
		return fmt.Errorf("bad response status %s", resp.Status)
	}
	return nil
}

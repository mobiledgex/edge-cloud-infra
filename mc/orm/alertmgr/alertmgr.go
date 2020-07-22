package alertmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"

	//"github.com/prometheus/alertmanager/api/v2/models"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	models "github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr/prometheus_structs/models"

	//	alertmanager_config "github.com/prometheus/alertmanager/config"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	alertmanager_config "github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr/prometheus_structs/config"
)

var alertRefreshInterval = 30 * time.Second

// Default alertmanager configuration
const DefaultAlertmanagerConfigFmt = `global:
  resolve_timeout: 5m
  smtp_from: %s
  smtp_smarthost: %s
  smtp_auth_username: %s
  smtp_auth_secret: %s
route:
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 2h
  receiver: default
receivers:
- name: default
`
const (
	AlertReceiverTypeEmail = "email"
	AlertReceiverTypeSlack = "slack"
	AlertSeverityError     = "error"
	AlertSeverityWarn      = "warning"
	AlertSeverityInfo      = "info"
)

// We will use this to read and write alertmanager config file
// Use AlertManagerGlobalConfig.String() to get the new file
// Use alertmanager_config.LoadFile(filename string) func to create AlertManagerConfig
// Use alertmanager_config.Load(s string) to test with example yamls
var AlertManagerConfig *alertmanager_config.Config
var configLock sync.RWMutex

const (
	AlertApi        string = "api/v2/alerts"
	ReceiverApi     string = "api/v2/receivers"
	SilenceApi      string = "api/v2/silences"
	ReloadConfigApi string = "-/reload"
)

// AlertMrgServer does two things - it periodically updates AlertManager about the
// current alerts on the system, and also handles configuration for the alert receivers
// i.e. backend handlers for the MC apis.
// NOTE: it does not perform any RBAC control here - this is done in ORM handlers
type AlertMrgServer struct {
	AlertMrgAddr            string
	McAlertmanagerAgentName string
	AlertMgrConfigPath      string
	AlertCache              *edgeproto.AlertCache
	vaultConfig             *vault.Config
	localVault              bool
	waitGrp                 sync.WaitGroup
	stop                    chan struct{}
}

type smtpInfo struct {
	Email string `json:"email"`
	User  string `json:"user"`
	Token string `json:"token"`
	Smtp  string `json:"smtp"`
}

// TODO - use version to track where this alert came from
func setAgentName() string {
	return "MasterControllerV1"
}

func NewAlertMgrServer(alertMgrAddr string, configPath string, vaultConfig *vault.Config, localVault bool, alertCache *edgeproto.AlertCache) (*AlertMrgServer, error) {
	server := AlertMrgServer{
		AlertMrgAddr:            alertMgrAddr,
		AlertCache:              alertCache,
		McAlertmanagerAgentName: setAgentName(),
		AlertMgrConfigPath:      configPath,
		vaultConfig:             vaultConfig,
		localVault:              localVault,
	}
	span := log.StartSpan(log.DebugLevelApi|log.DebugLevelInfo, "AlertMgrServer")
	ctx := log.ContextWithSpan(context.Background(), span)

	// We might need to wait for alertmanager to be up first
	for ii := 0; ii < 10; ii++ {
		_, err := server.alertMgrApi(ctx, "GET", "", "", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err := server.readConfigFile(ctx); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *AlertMrgServer) getAlertmanagertSmtpConfig(ctx context.Context) (*smtpInfo, error) {
	if s.localVault {
		log.SpanLog(ctx, log.DebugLevelApi, "Using dummy smtp credentials")
		return &testSmtpInfo, nil
	}
	alertMgrAcct := smtpInfo{}
	err := vault.GetData(s.vaultConfig,
		"/secret/data/accounts/alertmanagersmtp", 0, &alertMgrAcct)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to get data from vault", "err", err)
		return nil, err
	}
	return &alertMgrAcct, nil
}

// Load default configuration into Alertmanager
// Note configLock should be held prior to calling this
func (s *AlertMrgServer) loadDefaultConfigFileLocked(ctx context.Context) error {
	email, err := s.getAlertmanagertSmtpConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to get Smtp from vault", "err", err, "cfg", s.vaultConfig)
		return err
	}
	config := fmt.Sprintf(DefaultAlertmanagerConfigFmt, email.Email, email.Smtp+":587", email.User, email.Token)

	err = ioutil.WriteFile(s.AlertMgrConfigPath, []byte(config), 0644)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to write default alertmanager config", "err", err, "file", s.AlertMgrConfigPath)
		return err
	}
	// trigger reload of the config
	res, err := s.alertMgrApi(ctx, "POST", ReloadConfigApi, "", nil)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reeload alertmanager config", "err", err, "result", res)
		return err
	}
	return nil
}

// Read config from the alermgr config file.
// There are two passes here - one if a file exists and another if a file exists,
// but doesn' container required fields
func (s *AlertMrgServer) readConfigFile(ctx context.Context) error {
	// grab config lock
	configLock.Lock()
	defer configLock.Unlock()
	// Check that the config File exists
	file, err := os.Open(s.AlertMgrConfigPath)
	if err != nil {
		// Doesn't exist - need to load up a default config
		if os.IsNotExist(err) {
			log.SpanLog(ctx, log.DebugLevelInfo, "Loading default cofig - no file found")
			if err = s.loadDefaultConfigFileLocked(ctx); err != nil {
				return err
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to access alertmanager config", "err", err, "file", s.AlertMgrConfigPath)
			return err
		}
	}
	file.Close()
	// Read config
	AlertManagerConfig, err = alertmanager_config.LoadFile(s.AlertMgrConfigPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
			"file", s.AlertMgrConfigPath)
		return err
	}
	// Make sure that snmp defails are present
	if AlertManagerConfig.Global.SMTPSmarthost.Host == "" || AlertManagerConfig.Global.SMTPFrom == "" {
		log.SpanLog(ctx, log.DebugLevelInfo, "Writing correct default file")
		if err = s.loadDefaultConfigFileLocked(ctx); err != nil {
			return err
		}
		// Read config
		AlertManagerConfig, err = alertmanager_config.LoadFile(s.AlertMgrConfigPath)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
				"file", s.AlertMgrConfigPath)
			return err
		}
	}
	return nil
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
	res, err := s.alertMgrApi(ctx, "POST", AlertApi, "", data)
	log.SpanLog(ctx, log.DebugLevelInfo, "marshal alerts", "alerts", string(data), "err", err, "res", res)
	return err
}

// Note - this grabs configLock
func (s *AlertMrgServer) readAlertManagerConfigAndLock(ctx context.Context, config *alertmanager_config.Config) error {
	var err error
	// grab config lock
	configLock.Lock()

	// Read config
	config, err = alertmanager_config.LoadFile(s.AlertMgrConfigPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
			"file", s.AlertMgrConfigPath)
		return err
	}
	return nil
}

// Note - we should hold configLock prior to calling this function
func (s *AlertMrgServer) writeAlertManagerConfigLocked(ctx context.Context, config *alertmanager_config.Config) error {
	// write config out
	err := ioutil.WriteFile(s.AlertMgrConfigPath, []byte(config.String()), 0644)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to write alertmanager config file",
			"file", s.AlertMgrConfigPath, "config", config.String())
		return err
	}
	// restart alertmanager - call api
	res, err := s.alertMgrApi(ctx, "POST", ReloadConfigApi, "", nil)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload alertmanager config", "err", err, "res", res)
	}
	return err
}

func getAlertmgrReceiverName(receiver *ormapi.AlertReceiver) string {
	return receiver.Name + "-" + receiver.User + "-" + receiver.Severity + receiver.Type
}

// Receiver includes a route and a receiver which will receive the alert
// we create a route on the org tags for a given appInstance
func (s *AlertMrgServer) CreateReceiver(ctx context.Context, receiver *ormapi.AlertReceiver, routeMatchLabels map[string]string, cfg interface{}) error {
	// sanity - certain characters should not be part of the receiver name
	if strings.ContainsAny(receiver.Name, "-:") {
		return fmt.Errorf("Receiver name cannot contain dashes(\"-\"), or colons(\":\")")
	}

	// read file and greab a lock
	err := s.readAlertManagerConfigAndLock(ctx, AlertManagerConfig)
	defer configLock.Unlock()
	if err != nil {
		return err
	}

	// We create one entry per receiver, to make it simpler
	receiverName := getAlertmgrReceiverName(receiver)
	for _, rec := range AlertManagerConfig.Receivers {
		if rec.Name == receiverName {
			log.SpanLog(ctx, log.DebugLevelInfo, "Receiver Exists - delete it first")
			return fmt.Errorf("Receiver Exists - delete it first")
		}
	}
	// add a new reciever
	switch receiver.Type {
	case AlertReceiverTypeEmail:
		user, ok := cfg.(*ormapi.User)
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "Passed in struct is not a user struct")
			return fmt.Errorf("Passed in struct is not a user struct")
		}
		emailCfg := alertmanager_config.EmailConfig{
			To:   user.Email,
			Text: "TODO - write me",
		}
		rec := &alertmanager_config.Receiver{
			// to make the name unique - construct it with all the fields and username
			Name:         receiverName,
			EmailConfigs: []*alertmanager_config.EmailConfig{&emailCfg},
		}
		AlertManagerConfig.Receivers = append(AlertManagerConfig.Receivers, rec)
	case AlertReceiverTypeSlack:
		// TODO - need to figure out where to add slack details; as in which struct
		fallthrough
	default:
		log.SpanLog(ctx, log.DebugLevelInfo, "Unsupported receiver type", "type", receiver.Type,
			"receiver", receiver)
		return fmt.Errorf("Invalid receiver type - %s", receiver.Type)
	}
	// add route - match labels passed in
	route := alertmanager_config.Route{
		Receiver: receiverName,
		Match:    routeMatchLabels,
		Continue: false,
	}
	AlertManagerConfig.Route.Routes = append(AlertManagerConfig.Route.Routes, &route)
	// write config out
	return s.writeAlertManagerConfigLocked(ctx, AlertManagerConfig)
}

func (s *AlertMrgServer) DeleteReceiver(ctx context.Context, receiver *ormapi.AlertReceiver) error {
	// sanity - certain characters should not be part of the receiver name
	if strings.ContainsAny(receiver.Name, "-:") {
		return fmt.Errorf("Receiver name cannot contain dashes(\"-\"), or colons(\":\")")
	}

	// read file and greab a lock
	err := s.readAlertManagerConfigAndLock(ctx, AlertManagerConfig)
	defer configLock.Unlock()
	if err != nil {
		return err
	}

	// We create one entry per receiver, to make it simpler
	receiverName := getAlertmgrReceiverName(receiver)
	for ii, rec := range AlertManagerConfig.Receivers {
		if rec.Name == receiverName {
			log.SpanLog(ctx, log.DebugLevelInfo, "Found Receiver - now delete it")
			// remove from the receivers
			AlertManagerConfig.Receivers = append(AlertManagerConfig.Receivers[:ii],
				AlertManagerConfig.Receivers[ii+1:]...)
			// remove from routes
			for jj, route := range AlertManagerConfig.Route.Routes {
				if route.Receiver == receiverName {
					AlertManagerConfig.Route.Routes = append(AlertManagerConfig.Route.Routes[:jj],
						AlertManagerConfig.Route.Routes[jj+1:]...)
					break
				}
			}
			// write config out and return
			return s.writeAlertManagerConfigLocked(ctx, AlertManagerConfig)

		}
	}
	// nothing changed - just return nil
	return nil
}

// TODO - use local version of AlertManagerConfig
func (s *AlertMrgServer) ShowReceivers(ctx context.Context, filter ormapi.AlertReceiver) ([]ormapi.AlertReceiver, error) {
	// grab config lock
	// read config
	// show receivers
	return nil, nil
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

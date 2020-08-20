package alertmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"text/template"

	"github.com/mobiledgex/edge-cloud/log"

	//	alertmanager_config "github.com/prometheus/alertmanager/config"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	alertmanager_config "github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr/prometheus_structs/config"
)

// We will use this to read and write alertmanager config file
// Use AlertManagerGlobalConfig.String() to get the new file
// Use alertmanager_config.LoadFile(filename string) func to create AlertManagerConfig
// Use alertmanager_config.Load(s string) to test with example yamls
//var AlertManagerConfig *alertmanager_config.Config
var configLock sync.RWMutex

type SidecarReceiverConfig struct {
	Receiver alertmanager_config.Receiver `json:"receiver"`
	Route    alertmanager_config.Route    `json:"route,omitempty"`
}

type SidecarReceiverConfigs []SidecarReceiverConfig

type SidecarServer struct {
	alertMgrAddr       string
	alertMgrConfigPath string
	httpApiAddr        string
	server             *http.Server
}

func NewSidecarServer(target, path, apiAddr string) *SidecarServer {
	return &SidecarServer{
		alertMgrAddr:       target,
		alertMgrConfigPath: path,
		httpApiAddr:        apiAddr,
	}
}

// Get server address
func (s *SidecarServer) GetApiAddr() string {
	return s.httpApiAddr
}

// TODO - make this a TLS server
func (s *SidecarServer) Run() error {
	http.HandleFunc("/", s.proxyHandler)
	http.HandleFunc(AlertApi, s.proxyHandler)
	// http.HandleFunc(ReloadConfigApi, proxyHandler) - this should not be externally exposed
	http.HandleFunc(SilenceApi, s.proxyHandler)
	http.HandleFunc(ReceiverApi, s.proxyHandler)
	http.HandleFunc(mobiledgeXInitAlertmgr, s.initAlertmanager)
	http.HandleFunc(mobiledgeXReceiverApi, s.alertReceiver)

	listener, err := net.Listen("tcp4", s.httpApiAddr)
	if err != nil {
		return err
	}
	// For unit-tests we request the next available ports, so set it
	s.httpApiAddr = listener.Addr().String()
	s.server = &http.Server{
		Addr:    s.httpApiAddr,
		Handler: nil,
	}
	// detach and run the server
	go func() {
		var err error
		err = s.server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			log.FatalLog("Failed to run sidecar server", "err", err)
		}
	}()
	return nil
}

// Simple proxy to the Alertmanager we are connected to
func (s *SidecarServer) proxyHandler(w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse(s.alertMgrAddr)
	if err != nil {
		str := fmt.Sprintf("Proxy URL is not parsable - %s", s.alertMgrAddr)
		http.Error(w, str, http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Host = url.Host
	proxy.ServeHTTP(w, r)
}

func (s *SidecarServer) alertReceiver(w http.ResponseWriter, req *http.Request) {
	var writeConfig bool

	span := log.StartSpan(log.DebugLevelApi|log.DebugLevelInfo, "Alertmgr Sidecar Receiver")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	// Show Receivers
	if req.Method == http.MethodGet {
		config, err := s.readAlertManagerConfigAndLock(ctx)
		configLock.Unlock()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to read config request", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
		// return SidecarReceiverConfig[]
		receivers := SidecarReceiverConfigs{}
		for ii, rec := range config.Receivers {
			if rec.Name == "default" {
				continue
			}
			recConfig := SidecarReceiverConfig{
				Receiver: *config.Receivers[ii],
			}
			for jj, route := range config.Route.Routes {
				if route.Receiver == rec.Name {
					recConfig.Route = *config.Route.Routes[jj]
					receivers = append(receivers, recConfig)
					break
				}
			}
		}

		// marshal data and send it back
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(receivers)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to get marshal Receiver Config data", "err", err, "cfg", receivers)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	receiverConfig := SidecarReceiverConfig{}
	err := json.NewDecoder(req.Body).Decode(&receiverConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to decode request", "req", req)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// read file and grab a lock
	config, err := s.readAlertManagerConfigAndLock(ctx)
	defer configLock.Unlock()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to read config request", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeConfig = false
	if req.Method == http.MethodPost {
		// Create receiver
		writeConfig = true
		for _, rec := range config.Receivers {
			if rec.Name == receiverConfig.Receiver.Name {
				log.SpanLog(ctx, log.DebugLevelInfo, "Receiver Exists - delete it first")
				http.Error(w, "Receiver Exists - delete it first", http.StatusBadRequest)
				return
			}
		}
		config.Receivers = append(config.Receivers, &receiverConfig.Receiver)
		config.Route.Routes = append(config.Route.Routes, &receiverConfig.Route)
	} else if req.Method == http.MethodDelete {
		// Delete receiver
		for ii, rec := range config.Receivers {
			if rec.Name == receiverConfig.Receiver.Name {
				log.SpanLog(ctx, log.DebugLevelInfo, "Found Receiver - now delete it")
				// remove from the receivers
				config.Receivers = append(config.Receivers[:ii],
					config.Receivers[ii+1:]...)
				// remove from routes
				for jj, route := range config.Route.Routes {
					if route.Receiver == receiverConfig.Receiver.Name {
						config.Route.Routes = append(config.Route.Routes[:jj],
							config.Route.Routes[jj+1:]...)
						break
					}
				}
				// found something to delete
				writeConfig = true
				break
			}
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unsupported method", "req", req)
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	// No need to write config
	if !writeConfig {
		return
	}
	// write config out
	// NOTE: Alertmanager native unmarshal hides smtp password when marshalling.
	// See: https://github.com/prometheus/alertmanager/issues/1985
	// Instead our copy of the unmarshal code does not hide this. Hopefully 1985 will get addressed at some point
	err = s.writeAlertmanagerConfigLocked(ctx, bytes.NewBufferString(config.String()))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to write alertmanager configuration", "err", err, "config", config.String())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *SidecarServer) initAlertmanager(w http.ResponseWriter, req *http.Request) {
	span := log.StartSpan(log.DebugLevelApi|log.DebugLevelInfo, "Alertmgr Sidecar Init")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	// only support POST method
	if req.Method != http.MethodPost {
		log.SpanLog(ctx, log.DebugLevelInfo, "Only POST method is supported", "req", req)
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}
	smtpInfo := smtpInfo{}
	err := json.NewDecoder(req.Body).Decode(&smtpInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to decode request", "req", req)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.initConfigFile(ctx, &smtpInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Read config from the alertmgr config file.
// There are two passes here - one if a file exists and another if a file exists,
// but doesn't contain required fields
func (s *SidecarServer) initConfigFile(ctx context.Context, info *smtpInfo) error {
	// grab config lock
	configLock.Lock()
	defer configLock.Unlock()
	// Check that the config File exists
	file, err := os.Open(s.alertMgrConfigPath)
	if err != nil {
		// Doesn't exist - need to load up a default config
		if os.IsNotExist(err) {
			log.SpanLog(ctx, log.DebugLevelInfo, "Loading default config - no file found")
			if err = s.loadDefaultConfigFileLocked(ctx, info); err != nil {
				return err
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to access alertmanager config", "err", err, "file", s.alertMgrConfigPath)
			return err
		}
	}
	file.Close()
	// Read config
	config, err := alertmanager_config.LoadFile(s.alertMgrConfigPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
			"file", s.alertMgrConfigPath)
		return err
	}
	// Make sure that smtp defails are present
	if config.Global.SMTPSmarthost.Host == "" || config.Global.SMTPFrom == "" {
		log.SpanLog(ctx, log.DebugLevelInfo, "Writing correct default file")
		if err = s.loadDefaultConfigFileLocked(ctx, info); err != nil {
			return err
		}
		// Read config
		config, err = alertmanager_config.LoadFile(s.alertMgrConfigPath)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
				"file", s.alertMgrConfigPath)
			return err
		}
	}
	return nil
}

// Load default configuration into Alertmanager
// Note configLock should be held prior to calling this
func (s *SidecarServer) loadDefaultConfigFileLocked(ctx context.Context, info *smtpInfo) error {

	defaultConfigTemplate = template.Must(template.New("alertmanagerconfig").Parse(DefaultAlertmanagerConfigT))
	config := bytes.Buffer{}
	if err := defaultConfigTemplate.Execute(&config, info); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse the config template", "err", err)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Loading default config", "confog", config.String())
	return s.writeAlertmanagerConfigLocked(ctx, &config)
}

// Note - we should hold configLock prior to calling this function
func (s *SidecarServer) writeAlertmanagerConfigLocked(ctx context.Context, config *bytes.Buffer) error {
	err := ioutil.WriteFile(s.alertMgrConfigPath, config.Bytes(), 0644)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to write default alertmanager config", "err", err, "file", s.alertMgrConfigPath)
		return err
	}

	// trigger reload of the config
	res, err := alertMgrApi(ctx, s.alertMgrAddr, "POST", ReloadConfigApi, "", nil)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload alertmanager config", "err", err, "result", res)
		return err
	}
	return nil
}

// Note - this grabs configLock
func (s *SidecarServer) readAlertManagerConfigAndLock(ctx context.Context) (*alertmanager_config.Config, error) {
	// grab config lock
	configLock.Lock()

	// Read config
	config, err := alertmanager_config.LoadFile(s.alertMgrConfigPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to parse alertmanager config file", "err", err,
			"file", s.alertMgrConfigPath)
		return nil, err
	}
	return config, nil
}

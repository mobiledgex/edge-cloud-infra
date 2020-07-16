package alertmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/jarcoal/httpmock"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"

	//	open_api_models "github.com/prometheus/alertmanager/api/v2/models"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	open_api_models "github.com/mobiledgex/edge-cloud-infra/mc/alertmgr/prometheus_structs/models"

	// alertmanager_config "github.com/prometheus/alertmanager/config"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	alertmanager_config "github.com/mobiledgex/edge-cloud-infra/mc/alertmgr/prometheus_structs/config"

	//	"github.com/prometheus/common/model"
	// TODO - below is to replace the above for right now - once we update go and modules we can use prometheus directly
	model "github.com/mobiledgex/edge-cloud-infra/mc/alertmgr/prometheus_structs/model"

	"github.com/stretchr/testify/require"
)

type AlertmanagerMock struct {
	addr            string
	configFile      string
	alerts          map[string]model.Alert
	receivers       []*alertmanager_config.Receiver
	route           *alertmanager_config.Route
	AlertPosts      int
	AlertGets       int
	ReceiversGets   int
	SilencesGets    int
	SilencesPosts   int
	SilencesDeletes int
	ConfigReloads   int
}

func NewAlertmanagerMock(addr string, cfg string) *AlertmanagerMock {
	alertMgr := AlertmanagerMock{}
	alertMgr.addr = addr
	alertMgr.alerts = make(map[string]model.Alert)
	alertMgr.configFile = cfg
	if err := alertMgr.readConfig(); err != nil {
		fmt.Printf("Error reading config file, %v\n", err)
		return nil
	}
	alertMgr.registerMockResponders()
	return &alertMgr
}

func (s *AlertmanagerMock) readConfig() error {
	amCfg, err := alertmanager_config.LoadFile(s.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s.receivers = amCfg.Receivers
	s.route = amCfg.Route
	return nil
}

func (s *AlertmanagerMock) registerMockResponders() {
	// Create/Get Alerts
	s.registerCreateAlerts()
	s.registerGetAlerts()

	// Create/Delete/Get silences
	s.registerCreateSilences()
	s.registerGetSilences()
	s.registerDeleteSilences()

	// Get receivers
	s.registerGetReceivers()

	// Reload method
	s.registerConfigReload()
}

func (s *AlertmanagerMock) registerConfigReload() {
	httpmock.RegisterResponder("POST", s.addr+"/"+ReloadConfigApi,
		func(req *http.Request) (*http.Response, error) {
			err := s.readConfig()
			s.ConfigReloads++
			if err != nil {
				return httpmock.NewStringResponse(500, "Unable to read new configuration:"+err.Error()), nil
			}
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) registerCreateAlerts() {
	httpmock.RegisterResponder("POST", s.addr+"/"+AlertApi,
		func(req *http.Request) (*http.Response, error) {
			alerts := []model.Alert{}
			err := json.NewDecoder(req.Body).Decode(&alerts)
			if err != nil {
				return httpmock.NewStringResponse(500, "Unable to decode JSON body"), nil
			}
			for _, alert := range alerts {
				// set of labels is the key
				key := alert.Labels.String()
				_, found := s.alerts[key]
				if !found {
					s.alerts[key] = alert
				}
			}
			s.AlertPosts++
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) registerGetAlerts() {
	httpmock.RegisterResponder("GET", s.addr+"/"+AlertApi,
		func(req *http.Request) (*http.Response, error) {
			alerts := open_api_models.GettableAlerts{}
			for _, alert := range s.alerts {
				labels := open_api_models.LabelSet{}
				annotations := open_api_models.LabelSet{}
				for k, v := range alert.Labels {
					labels[string(k)] = string(v)
				}
				for k, v := range alert.Annotations {
					annotations[string(k)] = string(v)
				}

				start := strfmt.DateTime(alert.StartsAt)
				end := strfmt.DateTime(alert.EndsAt)

				alerts = append(alerts, &open_api_models.GettableAlert{
					Alert: open_api_models.Alert{
						Labels: labels,
					},
					Annotations: annotations,
					StartsAt:    &start,
					EndsAt:      &end,
				})
			}
			s.AlertGets++
			return httpmock.NewJsonResponse(200, alerts)
		},
	)
}

func (s *AlertmanagerMock) registerCreateSilences() {
	httpmock.RegisterResponder("POST", s.addr+"/"+SilenceApi,
		func(req *http.Request) (*http.Response, error) {
			// TODO
			s.SilencesPosts++
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) registerDeleteSilences() {
	httpmock.RegisterResponder("DELETE", s.addr+"/"+SilenceApi,
		func(req *http.Request) (*http.Response, error) {
			// TODO
			s.SilencesDeletes++
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) registerGetSilences() {
	httpmock.RegisterResponder("GET", s.addr+"/"+SilenceApi,
		func(req *http.Request) (*http.Response, error) {
			// TODO
			s.SilencesGets++
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) registerGetReceivers() {
	httpmock.RegisterResponder("GET", s.addr+"/"+ReceiverApi,
		func(req *http.Request) (*http.Response, error) {
			names := []string{}
			for _, receiver := range s.receivers {
				names = append(names, receiver.Name)
			}
			s.ReceiversGets++
			return httpmock.NewJsonResponse(200, names)
		},
	)
}

func (s *AlertmanagerMock) verifyAlertCnt(t *testing.T, cnt int) {
	require.Equal(t, cnt, len(s.alerts))
}

// verify the receiver is present and return this receiver
func (s *AlertmanagerMock) findReceiver(receiver *ormapi.AlertReceiver) *alertmanager_config.Receiver {
	name := getAlertmgrReceiverName(receiver)
	for ii, rec := range s.receivers {
		if rec.Name == name {
			return s.receivers[ii]
		}
	}
	return nil
}

func (s *AlertmanagerMock) findRouteByReceiver(receiver *ormapi.AlertReceiver) *alertmanager_config.Route {
	name := getAlertmgrReceiverName(receiver)
	for ii, route := range s.route.Routes {
		if route.Receiver == name {
			return s.route.Routes[ii]
		}
	}
	return nil
}

// Convert alert into alertmanager alert and check
func (s *AlertmanagerMock) verifyAlertPresent(t *testing.T, alert *edgeproto.Alert) {
	labelSet := model.LabelSet{}
	for k, v := range alert.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	labelSet[model.LabelName("region")] = model.LabelValue(alert.Region)
	key := labelSet.String()
	_, found := s.alerts[key]
	require.True(t, found)
}

func (s *AlertmanagerMock) verifyReceiversCnt(t *testing.T, cnt int) {
	require.Equal(t, cnt, len(s.receivers))
}

func (s *AlertmanagerMock) resetCounters() {
	s.AlertPosts = 0
	s.AlertGets = 0
	s.SilencesDeletes = 0
	s.SilencesGets = 0
	s.SilencesPosts = 0
	s.ReceiversGets = 0
	s.ConfigReloads = 0
}

func TestAlertMgrServer(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	// mock http to redirect requests
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// any requests that don't have a registered URL will be fetched normally
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)

	// // Run MC to test the actual MC apis
	// // master controller
	// addr := "127.0.0.1:9999"
	// var testAlertCache edgeproto.AlertCache
	// config := orm.ServerConfig{
	// 	ServAddr:        addr,
	// 	SqlAddr:         "127.0.0.1:5445",
	// 	RunLocal:        true,
	// 	InitLocal:       true,
	// 	IgnoreEnv:       true,
	// 	SkipVerifyEmail: true,
	// 	LocalVault:      true,
	// 	AlertCache:      &testAlertCache,
	// }

	// server, err := orm.RunServer(&config)
	// require.Nil(t, err, "run mc server")
	// defer server.Stop()

	testAlertMgrAddr := "http://dummyalertmgr.mobiledgex.net:9093"
	testAlertMgrConfig := "testAlertMgrConfig.yml"
	// start with clean configFile
	err := os.Remove(testAlertMgrConfig)
	if err != nil && !os.IsNotExist(err) {
		require.Fail(t, "cannot remove alertmanager config file", err)
	}
	fakeAlertmanager := NewAlertmanagerMock(testAlertMgrAddr, testAlertMgrConfig)
	require.NotNil(t, fakeAlertmanager)
	// Empty file - should have nothing
	fakeAlertmanager.verifyAlertCnt(t, 0)
	fakeAlertmanager.verifyReceiversCnt(t, 0)

	// Create a connection to fake alertmanager
	var testAlertCache edgeproto.AlertCache
	edgeproto.InitAlertCache(&testAlertCache)
	alertRefreshInterval = 100 * time.Millisecond
	testAlertMgrServer, err := NewAlertMgrServer(testAlertMgrAddr, testAlertMgrConfig,
		&vault.Config{}, true, &testAlertCache)
	require.Nil(t, err)
	require.NotNil(t, testAlertMgrServer)
	require.Equal(t, 1, fakeAlertmanager.ConfigReloads)
	// test readConfigFile - on an existing file
	err = testAlertMgrServer.readConfigFile(ctx)
	require.Nil(t, err)
	// We should still not have any configuration
	fakeAlertmanager.verifyAlertCnt(t, 0)
	// Default is one receiver
	fakeAlertmanager.verifyReceiversCnt(t, 1)
	testAlertCache.SetUpdatedCb(testAlertMgrServer.UpdateAlert)

	// Check that an alert notification triggers an api call to alertmgr
	testAlertCache.Update(ctx, &testAlerts[0], 0)
	// Test alertmgr create alert api
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlerts[0])
	require.Equal(t, 1, fakeAlertmanager.AlertPosts)

	// Start server after testing the watcher
	testAlertMgrServer.Start()
	// Wait refresh interval and check that the same alert is refreshed
	time.Sleep(alertRefreshInterval * 2)
	require.GreaterOrEqual(t, 2, fakeAlertmanager.AlertPosts)
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlerts[0])
	// Delete alert and check that alert doesn't get refreshed
	testAlertCache.Delete(ctx, &testAlerts[0], 0)
	// Make sure the last message was sent, since alertmgrserver is running in a separate thread
	time.Sleep(alertRefreshInterval)
	cnt := fakeAlertmanager.AlertPosts
	time.Sleep(alertRefreshInterval * 2)
	require.Equal(t, cnt, fakeAlertmanager.AlertPosts)
	// TODO - how to test alert timeout
	//fakeAlertmanager.verifyAlertCnt(t, 0)
	//    4.1. Can we delete alert from alertmgr right away?
	//    TODO
	// Create the alert again
	fakeAlertmanager.resetCounters()
	testAlertCache.Update(ctx, &testAlerts[0], 0)
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlerts[0])
	require.GreaterOrEqual(t, fakeAlertmanager.AlertPosts, 1)
	// Create the same alert, but in a different region
	testAlertCache.Update(ctx, &testAlerts[1], 0)
	fakeAlertmanager.verifyAlertCnt(t, 2)
	fakeAlertmanager.verifyAlertPresent(t, &testAlerts[1])
	// Test alertmgr show alert api
	alerts, err := testAlertMgrServer.ShowAlerts(ctx, nil)
	require.Nil(t, err)
	require.Equal(t, 1, fakeAlertmanager.AlertGets)
	require.Equal(t, 2, len(alerts))
	for _, alert := range alerts {
		val, found := alert.Labels["alertname"]
		require.True(t, found)
		require.Equal(t, cloudcommon.AlertAppInstDown, val)
		val, found = alert.Labels[edgeproto.AppKeyTagName]
		require.True(t, found)
		require.Equal(t, "testapp", val)
		val, found = alert.Labels[edgeproto.AppKeyTagOrganization]
		require.True(t, found)
		require.Equal(t, "testorg", val)
		val, found = alert.Labels[edgeproto.AppKeyTagVersion]
		require.True(t, found)
		require.Equal(t, "1.0", val)
		val, found = alert.Labels[edgeproto.CloudletKeyTagName]
		require.True(t, found)
		require.Equal(t, "testcloudlet", val)
		val, found = alert.Labels[cloudcommon.AlertHealthCheckStatus]
		require.True(t, found)
		require.Equal(t, strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)), val)
		if alert.Region != testRegions[0] {
			require.Equal(t, testRegions[1], alert.Region)
		}
	}
	// TODO - test for filter

	// 7. Test alertmgr create reciever api
	// Invalid receiver test
	err = testAlertMgrServer.CreateReceiver(ctx, &testAlertReceivers[0], testAlertReceiversMatchLabels[0], nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid receiver type")
	require.Contains(t, err.Error(), testAlertReceivers[0].Type)

	err = testAlertMgrServer.CreateReceiver(ctx, &testAlertReceivers[1], testAlertReceiversMatchLabels[0], &testAlertReceiverEmailCfg)
	require.Nil(t, err)
	require.Equal(t, 1, fakeAlertmanager.ConfigReloads)
	fakeAlertmanager.verifyReceiversCnt(t, 2)
	// Validate receivers
	receiver := fakeAlertmanager.findReceiver(&testAlertReceivers[1])
	require.NotNil(t, receiver)
	require.Len(t, receiver.EmailConfigs, 1)
	require.Equal(t, testAlertReceiverEmailCfg.Email, receiver.EmailConfigs[0].To)
	// check route and labels
	route := fakeAlertmanager.findRouteByReceiver(&testAlertReceivers[1])
	require.NotNil(t, route)
	for key, val := range testAlertReceiversMatchLabels[0] {
		routeLblVal, found := route.Match[key]
		require.True(t, found)
		require.Equal(t, val, routeLblVal)
	}
	// Delete non-existent receiver - nothing should change
	err = testAlertMgrServer.DeleteReceiver(ctx, &testAlertReceivers[0])
	require.Nil(t, err)
	require.Equal(t, 1, fakeAlertmanager.ConfigReloads)
	fakeAlertmanager.verifyReceiversCnt(t, 2)
	// Delete email receiver and verify it's deleted
	err = testAlertMgrServer.DeleteReceiver(ctx, &testAlertReceivers[1])
	require.Nil(t, err)
	require.Equal(t, 2, fakeAlertmanager.ConfigReloads)
	fakeAlertmanager.verifyReceiversCnt(t, 1)
	receiver = fakeAlertmanager.findReceiver(&testAlertReceivers[1])
	require.Nil(t, receiver)
	// Only receiver should be a default one
	require.Equal(t, "default", fakeAlertmanager.receivers[0].Name)
	route = fakeAlertmanager.findRouteByReceiver(&testAlertReceivers[1])
	require.Nil(t, route)
	// check routes
	route = fakeAlertmanager.findRouteByReceiver(&testAlertReceivers[1])
	require.Nil(t, route)

	// TODO - test silencers
	testAlertMgrServer.Stop()
}

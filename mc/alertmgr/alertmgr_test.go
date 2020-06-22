package alertmgr

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	open_api_models "github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

type AlertmanagerMock struct {
	addr            string
	alerts          map[string]model.Alert
	receivers       []open_api_models.Receiver
	AlertPosts      int
	AlertGets       int
	ReceiversGets   int
	SilencesGets    int
	SilencesPosts   int
	SilencesDeletes int
}

func NewAlertmanagerMock(addr string) *AlertmanagerMock {
	alertMgr := AlertmanagerMock{}
	alertMgr.addr = addr
	alertMgr.alerts = make(map[string]model.Alert)
	alertMgr.registerMockResponders()
	return &alertMgr
}

func (s *AlertmanagerMock) registerMockResponders() {
	// Create/Get Alerts
	s.registerCreateAlerts()
	s.registerGetAlerts()

	// Create/Delete/Get silences
	s.registerCreateSilences()
	s.registerGetSilences()
	s.rgisterDeleteSilences()

	// Get receivers
	s.registerGetReceivers()
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
			alerts := []model.Alert{}
			for _, alert := range s.alerts {
				alerts = append(alerts, alert)
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

func (s *AlertmanagerMock) rgisterDeleteSilences() {
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
			// TODO
			s.ReceiversGets++
			return httpmock.NewStringResponse(200, "Success"), nil
		},
	)
}

func (s *AlertmanagerMock) verifyEmpty(t *testing.T) {
	require.Equal(t, 0, len(s.alerts))
	require.Equal(t, 0, len(s.receivers))
}

func (s *AlertmanagerMock) verifyAlertCnt(t *testing.T, cnt int) {
	require.Equal(t, cnt, len(s.alerts))
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
}

var testRegion1 = "testRegion1"
var testRegion2 = "testRegion2"

var testAlertRootLbDown = edgeproto.Alert{
	Labels: map[string]string{
		"alertname":                        cloudcommon.AlertAppInstDown,
		cloudcommon.AlertLabelApp:          "testapp",
		cloudcommon.AlertLabelAppOrg:       "testorg",
		cloudcommon.AlertLabelAppVer:       "1.0",
		cloudcommon.AlertLabelCloudlet:     "testcloudlet",
		cloudcommon.AlertHealthCheckStatus: strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)),
	},
	Region: testRegion1,
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

	testAlertMgrAddr := "http://dummyalertmgr.mobiledgex.net:9093"

	fakeAlertmanager := NewAlertmanagerMock(testAlertMgrAddr)
	fakeAlertmanager.verifyEmpty(t)

	// Create a connection to fake alertmanager
	var testAlertCache edgeproto.AlertCache
	edgeproto.InitAlertCache(&testAlertCache)
	alertRefreshInterval = 100 * time.Millisecond
	testAlertMgrServer := NewAlertMgrServer(testAlertMgrAddr, &testAlertCache)
	require.NotNil(t, testAlertMgrServer)
	testAlertCache.SetUpdatedCb(testAlertMgrServer.UpdateAlert)

	// Check that an alert notification triggers an api call to alertmgr
	testAlertCache.Update(ctx, &testAlertRootLbDown, 0)
	// Test alertmgr create alert api
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlertRootLbDown)
	require.Equal(t, 1, fakeAlertmanager.AlertPosts)

	// Start server after testing the watcher
	testAlertMgrServer.Start()
	// Wait refresh interval and check that the same alert is refreshed
	time.Sleep(alertRefreshInterval * 2)
	require.GreaterOrEqual(t, 2, fakeAlertmanager.AlertPosts)
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlertRootLbDown)
	// Delete alert and check that alert doesn't get refreshed
	testAlertCache.Delete(ctx, &testAlertRootLbDown, 0)
	cnt := fakeAlertmanager.AlertPosts
	time.Sleep(alertRefreshInterval * 2)
	require.Equal(t, cnt, fakeAlertmanager.AlertPosts)
	// TODO - how to test alert timeout
	//fakeAlertmanager.verifyAlertCnt(t, 0)
	//    4.1. Can we delete alert from alertmgr right away?
	//    TODO
	// Create the alert again
	fakeAlertmanager.resetCounters()
	testAlertCache.Update(ctx, &testAlertRootLbDown, 0)
	fakeAlertmanager.verifyAlertCnt(t, 1)
	fakeAlertmanager.verifyAlertPresent(t, &testAlertRootLbDown)
	require.GreaterOrEqual(t, 1, fakeAlertmanager.AlertPosts)
	// Create the same alert, but in a different region
	testAlertRootLbDown.Region = testRegion2
	testAlertCache.Update(ctx, &testAlertRootLbDown, 0)
	fakeAlertmanager.verifyAlertCnt(t, 2)
	fakeAlertmanager.verifyAlertPresent(t, &testAlertRootLbDown)
	// Test alertmgr show alert api
	// TODO

	// 7. Test alertmgr create reciever api
	// ...

	testAlertMgrServer.Stop()
}

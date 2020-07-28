package alertmgr

import (
	"strconv"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var testSmtpInfo = smtpInfo{
	Email: "alerts@localhost",
	User:  "testuser",
	Token: "12345",
	Smtp:  "localhost",
	Port:  "25",
	Tls:   "false",
}

var testRegions = []string{"testRegion1", "testRegion2"}
var testUsers = []string{"testuser1", "testuser2"}

var testAlerts = []edgeproto.Alert{
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":                        cloudcommon.AlertAppInstDown,
			edgeproto.AppKeyTagName:            "testapp",
			edgeproto.AppKeyTagOrganization:    "testorg",
			edgeproto.AppKeyTagVersion:         "1.0",
			edgeproto.CloudletKeyTagName:       "testcloudlet",
			cloudcommon.AlertHealthCheckStatus: strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)),
		},
		Region: testRegions[0],
	},
	edgeproto.Alert{ //testAlerts[0], but in a different region
		Labels: map[string]string{
			"alertname":                        cloudcommon.AlertAppInstDown,
			edgeproto.AppKeyTagName:            "testapp",
			edgeproto.AppKeyTagOrganization:    "testorg",
			edgeproto.AppKeyTagVersion:         "1.0",
			edgeproto.CloudletKeyTagName:       "testcloudlet",
			cloudcommon.AlertHealthCheckStatus: strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)),
		},
		Region: testRegions[1],
	},
}

var testAlertReceivers = []ormapi.AlertReceiver{
	ormapi.AlertReceiver{
		Name:     "invalidReceiver",
		Type:     "invalidType",
		Severity: AlertSeverityError,
		User:     testUsers[0],
	},
	ormapi.AlertReceiver{
		Name:     "testorgemailreceiver",
		Type:     AlertReceiverTypeEmail,
		Severity: AlertSeverityError,
		User:     testUsers[0],
	},
}

var testAlertReceiversMatchLabels = []map[string]string{
	map[string]string{
		"test": "test",
	},
}

var testAlertReceiverEmailCfg = ormapi.User{
	Name:  testUsers[0],
	Email: "testuser1@testorg.net",
}

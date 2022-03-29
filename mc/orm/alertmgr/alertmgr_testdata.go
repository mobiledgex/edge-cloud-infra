package alertmgr

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var TestInitInfo = AlertmgrInitInfo{
	Email:          "alerts@localhost",
	User:           "testuser",
	Token:          "12345",
	Smtp:           "localhost",
	Port:           "25",
	Tls:            "false",
	ResolveTimeout: "2m",
}

var testRegions = []string{"testRegion1", "testRegion2"}
var testUsers = []string{"testuser1", "testuser2"}

var testAlerts = []edgeproto.Alert{
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":                        cloudcommon.AlertAppInstDown,
			"region":                           testRegions[0],
			edgeproto.AppKeyTagName:            "testapp",
			edgeproto.AppKeyTagOrganization:    "testorg",
			edgeproto.AppKeyTagVersion:         "1.0",
			edgeproto.CloudletKeyTagName:       "testcloudlet",
			cloudcommon.AlertSeverityLabel:     cloudcommon.AlertSeverityError,
			cloudcommon.AlertHealthCheckStatus: dme.HealthCheck_CamelName[int32(dme.HealthCheck_HEALTH_CHECK_ROOTLB_OFFLINE)],
		},
	},
	edgeproto.Alert{ //testAlerts[0], but in a different region
		Labels: map[string]string{
			"alertname":                        cloudcommon.AlertAppInstDown,
			"region":                           testRegions[1],
			edgeproto.AppKeyTagName:            "testapp",
			edgeproto.AppKeyTagOrganization:    "testorg",
			edgeproto.AppKeyTagVersion:         "1.0",
			edgeproto.CloudletKeyTagName:       "testcloudlet",
			cloudcommon.AlertSeverityLabel:     cloudcommon.AlertSeverityError,
			cloudcommon.AlertHealthCheckStatus: dme.HealthCheck_CamelName[int32(dme.HealthCheck_HEALTH_CHECK_ROOTLB_OFFLINE)],
		},
	},
	edgeproto.Alert{ // AlertAutoUndeploy alert
		Labels: map[string]string{
			"alertname":                     cloudcommon.AlertAutoUndeploy,
			"region":                        testRegions[1],
			edgeproto.AppKeyTagName:         "testapp",
			edgeproto.AppKeyTagOrganization: "testorg",
			edgeproto.AppKeyTagVersion:      "1.0",
			edgeproto.CloudletKeyTagName:    "testcloudlet",
		},
	},
}

var testAlertReceivers = []ormapi.AlertReceiver{
	ormapi.AlertReceiver{
		Name:     "invalidReceiver",
		Type:     "invalidType",
		Severity: cloudcommon.AlertSeverityError,
		User:     testUsers[0],
	},
	ormapi.AlertReceiver{
		Name:     "testorgemailreceiver",
		Type:     AlertReceiverTypeEmail,
		Severity: cloudcommon.AlertSeverityError,
		User:     testUsers[0],
		Email:    "testuser1@testorg.net",
		AppInst: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Name:         "testApp",
				Organization: "testAppOrg",
				Version:      "v1.0",
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "testCluster",
				},
				CloudletKey: edgeproto.CloudletKey{
					Name:         "testCloudlet",
					Organization: "testCloudletOrg",
				},
				Organization: "testClusterOrg",
			},
		},
	},
	ormapi.AlertReceiver{
		Name:         "testorgslackreceiverInvalidSlackData",
		Type:         AlertReceiverTypeSlack,
		Severity:     cloudcommon.AlertSeverityError,
		User:         testUsers[0],
		SlackChannel: "#alerts",
		SlackWebhook: "invalidURL",
		AppInst: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Name:         "testApp",
				Organization: "testAppOrg",
				Version:      "v1.0",
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "testCluster",
				},
				CloudletKey: edgeproto.CloudletKey{
					Name:         "testCloudlet",
					Organization: "testCloudletOrg",
				},
				Organization: "testClusterOrg",
			},
		},
	},
	ormapi.AlertReceiver{
		Name:         "testorgslackreceiver",
		Type:         AlertReceiverTypeSlack,
		Severity:     cloudcommon.AlertSeverityError,
		User:         testUsers[1],
		SlackChannel: "#alerts",
		SlackWebhook: "https://hooks.slack.com/foo",
		AppInst: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Name:         "testApp",
				Organization: "testAppOrg",
				Version:      "v1.0",
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "testCluster",
				},
				CloudletKey: edgeproto.CloudletKey{
					Name:         "testCloudlet",
					Organization: "testCloudletOrg",
				},
				Organization: "testClusterOrg",
			},
		},
	},
	ormapi.AlertReceiver{
		Name:     "testclusteremailreceiver",
		Type:     AlertReceiverTypeEmail,
		Severity: cloudcommon.AlertSeverityError,
		User:     testUsers[0],
		Email:    "testuser1@testorg.net",
		AppInst: edgeproto.AppInstKey{
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "testCluster",
				},
				CloudletKey: edgeproto.CloudletKey{
					Name:         "testCloudlet",
					Organization: "testCloudletOrg",
				},
				Organization: "testClusterOrg",
			},
		},
	},
}

package ormutil

import (
	"testing"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

func TestGetRegionObjSpecifiedFields(t *testing.T) {
	app := edgeproto.App{
		Key: edgeproto.AppKey{
			Organization: "Niantic",
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:   edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts: "tcp:443,tcp:10002,udp:10002",
		AccessType:  edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: edgeproto.FlavorKey{
			Name: "x1.tiny",
		},
		AllowServerless: true,
		ServerlessConfig: &edgeproto.ServerlessConfig{
			Vcpus: *edgeproto.NewUdec64(0, 500*edgeproto.DecMillis),
			Ram:   20,
		},
	}
	regionApp := ormapi.RegionApp{
		Region: "local",
		App:    app,
	}
	// The intent is to specify some fields with data, specify some fields
	// with empty data, and omit some fields with data, to make sure we
	// get everything we specified (regardless of its value), and omit
	// everything we didn't specify (regardless of its value).
	regionApp.App.Fields = []string{
		edgeproto.AppFieldKeyOrganization,
		edgeproto.AppFieldImageType,
		edgeproto.AppFieldAccessPorts,
		edgeproto.AppFieldInternalPorts,
		edgeproto.AppFieldAutoProvPolicies,
		edgeproto.AppFieldServerlessConfigVcpus,
		edgeproto.AppFieldServerlessConfigMinReplicas,
	}
	mm, err := GetRegionObjStructMapForUpdate(&regionApp)
	require.Nil(t, err)

	expected := map[string]interface{}{
		"Region": regionApp.Region,
		"App": map[string]interface{}{
			"Key": map[string]interface{}{
				"Organization": app.Key.Organization,
			},
			"ImageType":        edgeproto.ImageType_IMAGE_TYPE_DOCKER,
			"AccessPorts":      app.AccessPorts,
			"InternalPorts":    false,
			"AutoProvPolicies": []interface{}{},
			"ServerlessConfig": map[string]interface{}{
				"Vcpus":       *edgeproto.NewUdec64(0, 500*edgeproto.DecMillis),
				"MinReplicas": uint32(0),
			},
		},
	}
	require.Equal(t, expected, mm.Data)
}

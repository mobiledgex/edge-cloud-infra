// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "os"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import "github.com/mobiledgex/edge-cloud/cli"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestCreateApp(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.App) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionApp{}
	dat.Region = region
	dat.App = *in
	return mcClient.CreateApp(uri, token, dat)
}
func TestPermCreateApp(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.App{}
	in.Key.DeveloperKey.Name = org
	return TestCreateApp(mcClient, uri, token, region, in)
}

func TestDeleteApp(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.App) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionApp{}
	dat.Region = region
	dat.App = *in
	return mcClient.DeleteApp(uri, token, dat)
}
func TestPermDeleteApp(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.App{}
	in.Key.DeveloperKey.Name = org
	return TestDeleteApp(mcClient, uri, token, region, in)
}

func TestUpdateApp(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.App) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionApp{}
	dat.Region = region
	dat.App = *in
	return mcClient.UpdateApp(uri, token, dat)
}
func TestPermUpdateApp(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.App{}
	in.Key.DeveloperKey.Name = org
	return TestUpdateApp(mcClient, uri, token, region, in)
}

func TestShowApp(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.App) ([]edgeproto.App, int, error) {
	dat := &ormapi.RegionApp{}
	dat.Region = region
	dat.App = *in
	return mcClient.ShowApp(uri, token, dat)
}
func TestPermShowApp(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.App, int, error) {
	in := &edgeproto.App{}
	in.Key.DeveloperKey.Name = org
	return TestShowApp(mcClient, uri, token, region, in)
}

func RunMcAppApi(mcClient ormclient.Api, uri, token, region string, data *[]edgeproto.App, dataMap interface{}, rc *bool, mode string) {
	for ii, app := range *data {
		in := &ormapi.RegionApp{
			Region: region,
			App:    app,
		}
		switch mode {
		case "create":
			_, st, err := mcClient.CreateApp(uri, token, in)
			checkMcErr("CreateApp", st, err, rc)
		case "delete":
			_, st, err := mcClient.DeleteApp(uri, token, in)
			checkMcErr("DeleteApp", st, err, rc)
		case "update":
			objMap, err := cli.GetGenericObjFromList(dataMap, ii)
			if err != nil {
				fmt.Fprintf(os.Stderr, "bad dataMap for App: %v", err)
				os.Exit(1)
			}
			in.App.Fields = cli.GetSpecifiedFields(objMap, &in.App, cli.YamlNamespace)
			_, st, err := mcClient.UpdateApp(uri, token, in)
			checkMcErr("UpdateApp", st, err, rc)
		default:
			return
		}
	}
}

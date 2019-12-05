// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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

func TestCreateAutoScalePolicy(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AutoScalePolicy) (edgeproto.Result, int, error) {
	dat := &ormapi.RegionAutoScalePolicy{}
	dat.Region = region
	dat.AutoScalePolicy = *in
	return mcClient.CreateAutoScalePolicy(uri, token, dat)
}
func TestPermCreateAutoScalePolicy(mcClient *ormclient.Client, uri, token, region, org string) (edgeproto.Result, int, error) {
	in := &edgeproto.AutoScalePolicy{}
	in.Key.Developer = org
	return TestCreateAutoScalePolicy(mcClient, uri, token, region, in)
}

func TestDeleteAutoScalePolicy(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AutoScalePolicy) (edgeproto.Result, int, error) {
	dat := &ormapi.RegionAutoScalePolicy{}
	dat.Region = region
	dat.AutoScalePolicy = *in
	return mcClient.DeleteAutoScalePolicy(uri, token, dat)
}
func TestPermDeleteAutoScalePolicy(mcClient *ormclient.Client, uri, token, region, org string) (edgeproto.Result, int, error) {
	in := &edgeproto.AutoScalePolicy{}
	in.Key.Developer = org
	return TestDeleteAutoScalePolicy(mcClient, uri, token, region, in)
}

func TestUpdateAutoScalePolicy(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AutoScalePolicy) (edgeproto.Result, int, error) {
	dat := &ormapi.RegionAutoScalePolicy{}
	dat.Region = region
	dat.AutoScalePolicy = *in
	return mcClient.UpdateAutoScalePolicy(uri, token, dat)
}
func TestPermUpdateAutoScalePolicy(mcClient *ormclient.Client, uri, token, region, org string) (edgeproto.Result, int, error) {
	in := &edgeproto.AutoScalePolicy{}
	in.Key.Developer = org
	return TestUpdateAutoScalePolicy(mcClient, uri, token, region, in)
}

func TestShowAutoScalePolicy(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AutoScalePolicy) ([]edgeproto.AutoScalePolicy, int, error) {
	dat := &ormapi.RegionAutoScalePolicy{}
	dat.Region = region
	dat.AutoScalePolicy = *in
	return mcClient.ShowAutoScalePolicy(uri, token, dat)
}
func TestPermShowAutoScalePolicy(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.AutoScalePolicy, int, error) {
	in := &edgeproto.AutoScalePolicy{}
	in.Key.Developer = org
	return TestShowAutoScalePolicy(mcClient, uri, token, region, in)
}

func RunMcAutoScalePolicyApi(uri, token, region string, data *[]edgeproto.AutoScalePolicy, dataMap []map[string]interface{}, rc *bool, mode string) {
	var mcClient ormclient.Api
	for ii, autoScalePolicy := range *data {
		in := &ormapi.RegionAutoScalePolicy{
			Region:          region,
			AutoScalePolicy: autoScalePolicy,
		}
		switch mode {
		case "create":
			_, st, err := mcClient.CreateAutoScalePolicy(uri, token, in)
			checkMcErr("CreateAutoScalePolicy", st, err, rc)
		case "delete":
			_, st, err := mcClient.DeleteAutoScalePolicy(uri, token, in)
			checkMcErr("DeleteAutoScalePolicy", st, err, rc)
		case "update":
			in.AutoScalePolicy.Fields = cli.GetSpecifiedFields(dataMap[ii], &in.AutoScalePolicy, cli.YamlNamespace)
			_, st, err := mcClient.UpdateAutoScalePolicy(uri, token, in)
			checkMcErr("UpdateAutoScalePolicy", st, err, rc)
		default:
			return
		}
	}
}

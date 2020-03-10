// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestCreateCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.CreateCloudlet(uri, token, dat)
}
func TestPermCreateCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.OperatorKey.Name = org
	return TestCreateCloudlet(mcClient, uri, token, region, in)
}

func TestDeleteCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.DeleteCloudlet(uri, token, dat)
}
func TestPermDeleteCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.OperatorKey.Name = org
	return TestDeleteCloudlet(mcClient, uri, token, region, in)
}

func TestUpdateCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.UpdateCloudlet(uri, token, dat)
}
func TestPermUpdateCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.OperatorKey.Name = org
	return TestUpdateCloudlet(mcClient, uri, token, region, in)
}

func TestShowCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Cloudlet, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.ShowCloudlet(uri, token, dat)
}
func TestPermShowCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Cloudlet, int, error) {
	in := &edgeproto.Cloudlet{}
	return TestShowCloudlet(mcClient, uri, token, region, in)
}

func TestAddCloudletResMapping(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletResMap) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudletResMap{}
	dat.Region = region
	dat.CloudletResMap = *in
	return mcClient.AddCloudletResMapping(uri, token, dat)
}
func TestPermAddCloudletResMapping(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.CloudletResMap{}
	in.Key.OperatorKey.Name = org
	return TestAddCloudletResMapping(mcClient, uri, token, region, in)
}

func TestRemoveCloudletResMapping(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletResMap) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudletResMap{}
	dat.Region = region
	dat.CloudletResMap = *in
	return mcClient.RemoveCloudletResMapping(uri, token, dat)
}
func TestPermRemoveCloudletResMapping(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.CloudletResMap{}
	in.Key.OperatorKey.Name = org
	return TestRemoveCloudletResMapping(mcClient, uri, token, region, in)
}

func TestFindFlavorMatch(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, int, error) {
	dat := &ormapi.RegionFlavorMatch{}
	dat.Region = region
	dat.FlavorMatch = *in
	return mcClient.FindFlavorMatch(uri, token, dat)
}
func TestPermFindFlavorMatch(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.FlavorMatch, int, error) {
	in := &edgeproto.FlavorMatch{}
	in.Key.OperatorKey.Name = org
	return TestFindFlavorMatch(mcClient, uri, token, region, in)
}

func RunMcCloudletApi(mcClient ormclient.Api, uri, token, region string, data *[]edgeproto.Cloudlet, dataMap interface{}, rc *bool, mode string) {
	for ii, cloudlet := range *data {
		in := &ormapi.RegionCloudlet{
			Region:   region,
			Cloudlet: cloudlet,
		}
		switch mode {
		case "create":
			_, st, err := mcClient.CreateCloudlet(uri, token, in)
			checkMcErr("CreateCloudlet", st, err, rc)
		case "delete":
			_, st, err := mcClient.DeleteCloudlet(uri, token, in)
			checkMcErr("DeleteCloudlet", st, err, rc)
		case "update":
			objMap, err := cli.GetGenericObjFromList(dataMap, ii)
			if err != nil {
				fmt.Fprintf(os.Stderr, "bad dataMap for Cloudlet: %v", err)
				os.Exit(1)
			}
			in.Cloudlet.Fields = cli.GetSpecifiedFields(objMap, &in.Cloudlet, cli.YamlNamespace)
			_, st, err := mcClient.UpdateCloudlet(uri, token, in)
			checkMcErr("UpdateCloudlet", st, err, rc)
		case "show":
			_, st, err := mcClient.ShowCloudlet(uri, token, in)
			checkMcErr("ShowCloudlet", st, err, rc)
		default:
			return
		}
	}
}

func RunMcCloudletApi_CloudletResMap(mcClient ormclient.Api, uri, token, region string, data *[]edgeproto.CloudletResMap, dataMap interface{}, rc *bool, mode string) {
	for _, cloudletResMap := range *data {
		in := &ormapi.RegionCloudletResMap{
			Region:         region,
			CloudletResMap: cloudletResMap,
		}
		switch mode {
		case "addcloudletresmapping":
			_, st, err := mcClient.AddCloudletResMapping(uri, token, in)
			checkMcErr("AddCloudletResMapping", st, err, rc)
		case "removecloudletresmapping":
			_, st, err := mcClient.RemoveCloudletResMapping(uri, token, in)
			checkMcErr("RemoveCloudletResMapping", st, err, rc)
		default:
			return
		}
	}
}

func RunMcCloudletApi_FlavorMatch(mcClient ormclient.Api, uri, token, region string, data *[]edgeproto.FlavorMatch, dataMap interface{}, rc *bool, mode string) {
	for _, flavorMatch := range *data {
		in := &ormapi.RegionFlavorMatch{
			Region:      region,
			FlavorMatch: flavorMatch,
		}
		switch mode {
		case "find":
			_, st, err := mcClient.FindFlavorMatch(uri, token, in)
			checkMcErr("FindFlavorMatch", st, err, rc)
		default:
			return
		}
	}
}

func TestShowCloudletInfo(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletInfo) ([]edgeproto.CloudletInfo, int, error) {
	dat := &ormapi.RegionCloudletInfo{}
	dat.Region = region
	dat.CloudletInfo = *in
	return mcClient.ShowCloudletInfo(uri, token, dat)
}
func TestPermShowCloudletInfo(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.CloudletInfo, int, error) {
	in := &edgeproto.CloudletInfo{}
	in.Key.OperatorKey.Name = org
	return TestShowCloudletInfo(mcClient, uri, token, region, in)
}

func RunMcCloudletInfoApi(mcClient ormclient.Api, uri, token, region string, data *[]edgeproto.CloudletInfo, dataMap interface{}, rc *bool, mode string) {
	for _, cloudletInfo := range *data {
		in := &ormapi.RegionCloudletInfo{
			Region:       region,
			CloudletInfo: cloudletInfo,
		}
		switch mode {
		case "show":
			_, st, err := mcClient.ShowCloudletInfo(uri, token, in)
			checkMcErr("ShowCloudletInfo", st, err, rc)
		default:
			return
		}
	}
}

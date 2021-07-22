// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

package ormapi

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

// Request summary for CreateApp
// swagger:parameters CreateApp
type swaggerCreateApp struct {
	// in: body
	Body RegionApp
}

type RegionApp struct {
	// required: true
	// Region name
	Region string
	App    edgeproto.App
}

func (s *RegionApp) GetRegion() string {
	return s.Region
}

func (s *RegionApp) GetObj() interface{} {
	return &s.App
}

func (s *RegionApp) GetObjName() string {
	return "App"
}
func (s *RegionApp) GetObjFields() []string {
	return s.App.Fields
}

func (s *RegionApp) SetObjFields(fields []string) {
	s.App.Fields = fields
}

// Request summary for DeleteApp
// swagger:parameters DeleteApp
type swaggerDeleteApp struct {
	// in: body
	Body RegionApp
}

// Request summary for UpdateApp
// swagger:parameters UpdateApp
type swaggerUpdateApp struct {
	// in: body
	Body RegionApp
}

// Request summary for ShowApp
// swagger:parameters ShowApp
type swaggerShowApp struct {
	// in: body
	Body RegionApp
}

// Request summary for AddAppAutoProvPolicy
// swagger:parameters AddAppAutoProvPolicy
type swaggerAddAppAutoProvPolicy struct {
	// in: body
	Body RegionAppAutoProvPolicy
}

type RegionAppAutoProvPolicy struct {
	// required: true
	// Region name
	Region            string
	AppAutoProvPolicy edgeproto.AppAutoProvPolicy
}

func (s *RegionAppAutoProvPolicy) GetRegion() string {
	return s.Region
}

func (s *RegionAppAutoProvPolicy) GetObj() interface{} {
	return &s.AppAutoProvPolicy
}

func (s *RegionAppAutoProvPolicy) GetObjName() string {
	return "AppAutoProvPolicy"
}

// Request summary for RemoveAppAutoProvPolicy
// swagger:parameters RemoveAppAutoProvPolicy
type swaggerRemoveAppAutoProvPolicy struct {
	// in: body
	Body RegionAppAutoProvPolicy
}

// Request summary for AddAppAlertPolicy
// swagger:parameters AddAppAlertPolicy
type swaggerAddAppAlertPolicy struct {
	// in: body
	Body RegionAppAlertPolicy
}

type RegionAppAlertPolicy struct {
	// required: true
	// Region name
	Region         string
	AppAlertPolicy edgeproto.AppAlertPolicy
}

func (s *RegionAppAlertPolicy) GetRegion() string {
	return s.Region
}

func (s *RegionAppAlertPolicy) GetObj() interface{} {
	return &s.AppAlertPolicy
}

func (s *RegionAppAlertPolicy) GetObjName() string {
	return "AppAlertPolicy"
}

// Request summary for RemoveAppAlertPolicy
// swagger:parameters RemoveAppAlertPolicy
type swaggerRemoveAppAlertPolicy struct {
	// in: body
	Body RegionAppAlertPolicy
}

// Request summary for ShowCloudletsForAppDeployment
// swagger:parameters ShowCloudletsForAppDeployment
type swaggerShowCloudletsForAppDeployment struct {
	// in: body
	Body RegionDeploymentCloudletRequest
}

type RegionDeploymentCloudletRequest struct {
	// required: true
	// Region name
	Region                    string
	DeploymentCloudletRequest edgeproto.DeploymentCloudletRequest
}

func (s *RegionDeploymentCloudletRequest) GetRegion() string {
	return s.Region
}

func (s *RegionDeploymentCloudletRequest) GetObj() interface{} {
	return &s.DeploymentCloudletRequest
}

func (s *RegionDeploymentCloudletRequest) GetObjName() string {
	return "DeploymentCloudletRequest"
}

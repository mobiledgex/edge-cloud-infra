// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: node.proto

package ormclient

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/mobiledgex/edge-cloud/protogen"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) ShowNode(uri, token string, in *ormapi.RegionNode) ([]edgeproto.Node, int, error) {
	out := edgeproto.Node{}
	outlist := []edgeproto.Node{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowNode", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type NodeApiClient interface {
	ShowNode(uri, token string, in *ormapi.RegionNode) ([]edgeproto.Node, int, error)
}

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: node.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "context"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
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

func TestShowNode(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Node) ([]edgeproto.Node, int, error) {
	dat := &ormapi.RegionNode{}
	dat.Region = region
	dat.Node = *in
	return mcClient.ShowNode(uri, token, dat)
}
func TestPermShowNode(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Node, int, error) {
	in := &edgeproto.Node{}
	return TestShowNode(mcClient, uri, token, region, in)
}

func (s *TestClient) ShowNode(ctx context.Context, in *edgeproto.Node) ([]edgeproto.Node, error) {
	inR := &ormapi.RegionNode{
		Region: s.Region,
		Node:   *in,
	}
	out, status, err := s.McClient.ShowNode(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

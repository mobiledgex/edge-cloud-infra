# Build proto files from edge-cloud

PROTODIR	= ../edge-cloud/edgeproto
# NOTE:
# Because we CD to protodir to run these commands, all relative paths
# passed to protoc command (like GOPATH) are relative to PROTODIR
GOPATH		= ../../../..
GW		= $(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)
APIS		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/googleapis)
GOGO		= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)
INFRA		= $(shell go list -f '{{ .Dir }}' -m github.com/mobiledgex/edge-cloud-infra)
INCLUDE		= -I. -I${GW} -I${APIS} -I${GOGO} -I${GOPATH}
BUILTIN		= Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto
OUTDIR		= ${INFRA}/mc/orm
OUTAPI		= ${INFRA}/mc/ormapi
OUTCLIENT	= ${INFRA}/mc/ormclient

build:
	(cd ${PROTODIR}; protoc ${INCLUDE} --mc2_out=${BUILTIN},genapi,pkg=ormapi:${OUTAPI} *.proto)
	(cd ${PROTODIR}; protoc ${INCLUDE} --mc2_out=${BUILTIN}:${OUTDIR} *.proto)
	(cd ${PROTODIR}; protoc ${INCLUDE} --mc2_out=${BUILTIN},gentest,suffix=_mc2_test.go:${OUTDIR} *.proto)
	(cd ${PROTODIR}; protoc ${INCLUDE} --mc2_out=${BUILTIN},genclient,pkg=ormclient,suffix=_client.go:${OUTCLIENT} *.proto)

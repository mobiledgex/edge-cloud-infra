package main

import (
	"context"
	"fmt"
	"io"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var testDialOpt grpc.DialOption

func goAppInstApi(ctx context.Context, inst *edgeproto.AppInst, action cloudcommon.Action, reason, policyName string) error {
	span := log.StartSpan(log.DebugLevelApi, "auto-prov deploy "+action.String(), opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
	span.SetTag("AppInst", inst.Key)
	span.SetTag("reason", reason)
	defer span.Finish()
	ctx = log.ContextWithSpan(context.Background(), span)

	log.SpanLog(ctx, log.DebugLevelApi, "auto-prov deploy "+action.String(), "AppInst", inst.Key, "reason", reason, "policyName", policyName)
	if action != cloudcommon.Create && action != cloudcommon.Delete {
		log.SpanLog(ctx, log.DebugLevelApi, "invalid action", "action", action.String())
		return fmt.Errorf("invalid action")
	}
	err := runAppInstApi(ctx, inst, action, reason, policyName)
	log.SpanLog(ctx, log.DebugLevelApi, "auto-prov deploy result", "err", err)
	return err
}

func runAppInstApi(ctx context.Context, inst *edgeproto.AppInst, action cloudcommon.Action, reason, policyName string) error {
	opts := []grpc.DialOption{}
	if dialOpts != nil {
		opts = append(opts, dialOpts)
	}
	if testDialOpt != nil {
		opts = append(opts, testDialOpt)
	}
	opts = append(opts, grpc.WithBlock(),
		grpc.WithWaitForHandshake(),
		grpc.WithUnaryInterceptor(log.UnaryClientTraceGrpc),
		grpc.WithStreamInterceptor(log.StreamClientTraceGrpc))
	conn, err := grpc.Dial(*ctrlAddr, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	kvPairs := []string{
		cloudcommon.CallerAutoProv, "",
		cloudcommon.AutoProvReason, reason,
		cloudcommon.AutoProvPolicyName, policyName}
	ctx = metadata.AppendToOutgoingContext(ctx, kvPairs...)
	client := edgeproto.NewAppInstApiClient(conn)
	var stream edgeproto.AppInstApi_CreateAppInstClient
	switch action {
	case cloudcommon.Create:
		stream, err = client.CreateAppInst(ctx, inst)
	case cloudcommon.Delete:
		stream, err = client.DeleteAppInst(ctx, inst)
	}
	if err != nil {
		return err
	}
	for {
		_, err = stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}
	}
	return err
}

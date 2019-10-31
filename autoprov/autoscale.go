package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
)

func autoScale(ctx context.Context, name string, alert *edgeproto.Alert) error {
	if alert.State != "firing" {
		return nil
	}
	inst := edgeproto.ClusterInst{}
	inst.Key.Developer = alert.Labels[cloudcommon.AlertLabelDev]
	inst.Key.ClusterKey.Name = alert.Labels[cloudcommon.AlertLabelCluster]
	inst.Key.CloudletKey.Name = alert.Labels[cloudcommon.AlertLabelCloudlet]
	inst.Key.CloudletKey.OperatorKey.Name = alert.Labels[cloudcommon.AlertLabelOperator]

	nodecountStr := alert.Annotations[cloudcommon.AlertKeyNodeCount]
	nodecount, err := strconv.Atoi(nodecountStr)
	if err != nil {
		return fmt.Errorf("failed to parse nodecount %s, %v", nodecountStr, err)
	}

	if name == cloudcommon.AlertAutoScaleUp {
		inst.NumNodes = uint32(nodecount + 1)
	} else {
		lowStr := alert.Annotations[cloudcommon.AlertKeyLowCpuNodeCount]
		lowcount, err := strconv.Atoi(lowStr)
		if err != nil {
			return fmt.Errorf("failed to parse lowcpu count %s, %v", lowStr, err)
		}
		minNodesStr := alert.Annotations[cloudcommon.AlertKeyMinNodes]
		minNodes, err := strconv.Atoi(minNodesStr)
		if err != nil {
			return fmt.Errorf("failed to parse min nodes %s, %v", minNodesStr, err)
		}
		newCount := nodecount - lowcount
		if newCount < minNodes {
			newCount = minNodes
		}
		inst.NumNodes = uint32(newCount)
	}
	inst.Fields = []string{edgeproto.ClusterInstFieldNumNodes}

	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to controller %s failed, %v", *ctrlAddr, err)
	}
	defer conn.Close()

	log.SpanLog(ctx, log.DebugLevelApi, "auto scaling clusterinst", "alert", alert, "ClusterInst", inst)
	client := edgeproto.NewClusterInstApiClient(conn)
	stream, err := client.UpdateClusterInst(ctx, &inst)
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

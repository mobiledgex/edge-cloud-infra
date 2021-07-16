package vcd

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (v *VcdPlatform) initDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc("govcdcmd", v.runVcdCliCommand)
}

func (v *VcdPlatform) runVcdCliCommand(ctx context.Context, req *edgeproto.DebugRequest) string {

	if req.Args == "" {
		return "please specify vcd command in args field"
	}
	rd := csv.NewReader(strings.NewReader(req.Args))
	rd.Comma = ' '
	args, err := rd.Read()
	if err != nil {
		return fmt.Sprintf("failed to split args string, %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "runVcdCliCommand ", "args[0]", args[0], "args[1:]...", args[1:])
	out, err := v.TimedVcdCliCommand(ctx, args[0], args[1:]...)
	if err != nil {
		return fmt.Sprintf("given vcd cmd command failed: %v, %s", err, string(out))
	}
	return string(out)
}

func (v *VcdPlatform) TimedVcdCliCommand(ctx context.Context, name string, a ...string) ([]byte, error) {
	parmstr := strings.Join(a, " ")
	start := time.Now()
	log.SpanLog(ctx, log.DebugLevelInfra, "govcdcmd Command Start", "name", name, "parms", parmstr)
	newSh := infracommon.Sh(v.vcdVars)

	out, err := newSh.Command(name, a).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "govcdcmd command returned error", "parms", parmstr, "out", string(out), "err", err, "elapsed time", time.Since(start))
		return out, err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "govcdcmd Command Done", "parmstr", parmstr, "elapsed time", time.Since(start))
	return out, nil
}

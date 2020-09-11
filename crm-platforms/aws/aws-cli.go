package aws

import (
	"context"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"
)

func (a *AWSPlatform) TimedAwsCommand(ctx context.Context, name string, p ...string) ([]byte, error) {
	parmstr := strings.Join(p, " ")
	start := time.Now()

	log.SpanLog(ctx, log.DebugLevelInfra, "AWS Command Start", "name", name, "parms", parmstr)
	newSh := sh.NewSession()
	//envvar stuff here

	out, err := newSh.Command(name, p).CombinedOutput()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "AWS command returned error", "parms", parmstr, "out", string(out), "err", err, "elapsed time", time.Since(start))
		return out, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "AWS Command Done", "parmstr", parmstr, "elapsed time", time.Since(start))
	return out, nil
}

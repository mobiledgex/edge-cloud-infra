package orm

import (
	"context"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

func RunCommandValidateInput(ctx context.Context, rc *RegionContext, obj *edgeproto.ExecRequest) error {
	if obj.Cmd != nil {
		sanitizedCmd, err := util.RunCommandSanitize(obj.Cmd.Command)
		if err != nil {
			return err
		}
		obj.Cmd.Command = sanitizedCmd
	}
	return nil
}

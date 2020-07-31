package orm

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

type RegionContext struct {
	region    string
	username  string
	conn      *grpc.ClientConn
	skipAuthz bool
}

func connectController(ctx context.Context, region string) (*grpc.ClientConn, error) {
	addr, err := getControllerAddrForRegion(ctx, region)
	if err != nil {
		return nil, err
	}
	return connectGrpcAddr(addr)
}

func connectNotifyRoot(ctx context.Context) (*grpc.ClientConn, error) {
	if serverConfig.NotifyAddrs == "" {
		return nil, fmt.Errorf("No parent notify address specified, cannot connect to notify root")
	}
	addrs := strings.Split(serverConfig.NotifyAddrs, ",")
	return connectGrpcAddr(addrs[0])
}

func connectGrpcAddr(addr string) (*grpc.ClientConn, error) {
	dialOption, err := tls.GetTLSClientDialOption(addr, serverConfig.ClientCert, false)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(addr, dialOption,
		grpc.WithUnaryInterceptor(log.UnaryClientTraceGrpc),
		grpc.WithStreamInterceptor(log.StreamClientTraceGrpc),
	)
}

func getControllerAddrForRegion(ctx context.Context, region string) (string, error) {
	ctrl, err := getControllerObj(ctx, region)
	if err != nil {
		return "", err
	}
	return ctrl.Address, nil
}

func getControllerObj(ctx context.Context, region string) (*ormapi.Controller, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	ctrl := ormapi.Controller{
		Region: region,
	}
	db := loggedDB(ctx)
	res := db.Where(&ctrl).First(&ctrl)
	if res.Error != nil {
		if res.RecordNotFound() {
			return nil, fmt.Errorf("region \"%s\" not found", region)
		}
		return nil, res.Error
	}
	return &ctrl, nil
}

func CreateController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return bindErr(c, err)
	}
	err = CreateControllerObj(ctx, claims, &ctrl)
	return setReply(c, err, Msg("Controller registered"))
}

func CreateControllerObj(ctx context.Context, claims *UserClaims, ctrl *ormapi.Controller) error {
	if ctrl.Region == "" {
		return fmt.Errorf("Controller Region not specified")
	}
	if ctrl.Address == "" {
		return fmt.Errorf("Controller Address not specified")
	}
	if err := authorized(ctx, claims.Username, "", ResourceControllers, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	err := db.Create(ctrl).Error
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func DeleteController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return bindErr(c, err)
	}
	err = DeleteControllerObj(ctx, claims, &ctrl)
	return setReply(c, err, Msg("Controller deregistered"))
}

func DeleteControllerObj(ctx context.Context, claims *UserClaims, ctrl *ormapi.Controller) error {
	if err := authorized(ctx, claims.Username, "", ResourceControllers, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	err := db.Delete(ctrl).Error
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func ShowController(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "Show controllers", "context", c, "clainms", claims)
	ctrls, err := ShowControllerObj(ctx, claims)
	return setReply(c, err, ctrls)
}

func ShowControllerObj(ctx context.Context, claims *UserClaims) ([]ormapi.Controller, error) {
	ctrls := []ormapi.Controller{}
	db := loggedDB(ctx)
	err := db.Find(&ctrls).Error
	if err != nil {
		return nil, dbErr(err)
	}
	return ctrls, nil
}

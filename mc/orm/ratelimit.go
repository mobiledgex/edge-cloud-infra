package orm

import (
	"context"
	fmt "fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ApiType int

const (
	Admin ApiType = iota
	Developer
	Operator
)

type ApiAction int

const (
	Create ApiAction = iota
	Delete
	Show
)

func rateLimit(ctx context.Context, api, usr, org, ip string) (bool, error) {
	rateLimitCtx := ratelimit.Context{Context: ctx}
	rateLimitCtx.Api = api
	rateLimitCtx.User = usr
	rateLimitCtx.Org = org
	rateLimitCtx.Ip = ip
	limit, err := rateLimitMgr.Limit(rateLimitCtx)
	if limit {
		errMsg := fmt.Sprintf("%s is rejected, please retry later.", api)
		if err != nil {
			errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
		}
		return true, status.Errorf(codes.ResourceExhausted, errMsg)

	}
	return false, nil
}

func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := getClaims(c)
		if err != nil {
			return err
		}

		rateLimitCtx := ratelimit.Context{Context: c.Request().Context()}
		rateLimitCtx.Api = c.Path()
		rateLimitCtx.User = claims.Username
		rateLimitCtx.Ip = c.RealIP()
		// TODO: by org???
		limit, err := rateLimitMgr.Limit(rateLimitCtx)
		if limit {
			log.DebugLog(log.DebugLevelInfo, "BLAH: error ratelimiting", "err", err)
			errMsg := fmt.Sprintf("%s is rejected, please retry later.", c.Path())
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return echo.NewHTTPError(http.StatusTooManyRequests, errMsg)

		}
		return next(c)
	}
}

func getOrgFromRequest(c echo.Context) (string, error) {
	var inter interface{}
	if err := c.Bind(&inter); err != nil {
		// log
		log.DebugLog(log.DebugLevelInfo, "BLAH: error binding", "err", err)
		return "", fmt.Errorf("BLAH error binding: %s", err)
	}
	log.DebugLog(log.DebugLevelInfo, "BLAH: binded")
	switch typ := inter.(type) {
	case *ormapi.RegionAppInstMetrics:
		// switch based on selector
		return typ.AppInst.AppKey.Organization, nil
	case *ormapi.RegionAppInst:
		return typ.AppInst.Key.AppKey.Organization, nil
	default:
		log.DebugLog(log.DebugLevelInfo, "BLAH: unknown req", "req type", typ)
		return "", nil
	}
}

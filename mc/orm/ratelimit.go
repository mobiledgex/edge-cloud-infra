package orm

import (
	fmt "fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
)

type ApiType int

const (
	Mc ApiType = iota
	Dme
	Controller
)

var GlobalMcApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	Key: edgeproto.RateLimitSettingsKey{
		ApiName:         edgeproto.GlobalApiName,
		RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
	},
	FlowSettings: []*edgeproto.FlowSettings{
		&edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 10000,
			BurstSize:     500,
		},
	},
}

var GlobalMcApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	Key: edgeproto.RateLimitSettingsKey{
		ApiName:         edgeproto.GlobalApiName,
		RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
	},
	FlowSettings: []*edgeproto.FlowSettings{
		&edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1000,
			BurstSize:     100,
		},
	},
}

var GlobalMcApiPerUserRateLimitSettings = &edgeproto.RateLimitSettings{
	Key: edgeproto.RateLimitSettingsKey{
		ApiName:         edgeproto.GlobalApiName,
		RateLimitTarget: edgeproto.RateLimitTarget_PER_USER,
	},
	FlowSettings: []*edgeproto.FlowSettings{
		&edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1000,
			BurstSize:     100,
		},
	},
}

var UserCreateFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	Key: edgeproto.RateLimitSettingsKey{
		ApiName:         "/usercreate",
		RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
	},
	FlowSettings: []*edgeproto.FlowSettings{
		&edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 100,
			BurstSize:     5,
		},
	},
}

var UserCreatePerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	Key: edgeproto.RateLimitSettingsKey{
		ApiName:         "/usercreate",
		RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
	},
	FlowSettings: []*edgeproto.FlowSettings{
		&edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create ctx with rateLimitInfo
		ctx := GetContext(c)
		getClaims(ctx)
		rateLimitInfo := &ratelimit.LimiterInfo{
			Api: c.Path(),
			Ip:  c.RealIP(),
		}
		ctx = ratelimit.NewLimiterInfoContext(ctx, rateLimitInfo)
		// Rate limit
		limit, err := rateLimitMgr.Limit(ctx)
		if limit {
			errMsg := fmt.Sprintf("%s is rejected, please retry later.", c.Path())
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return echo.NewHTTPError(http.StatusTooManyRequests, errMsg)

		}
		return next(c)
	}
}

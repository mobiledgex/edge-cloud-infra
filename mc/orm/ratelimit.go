package orm

import (
	fmt "fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

/*type ApiType int

const (
	Admin ApiType = iota
	Developer
	Operator
)*/

type ApiType int

const (
	Mc ApiType = iota
	Controller
)

type ApiAuthType int

const (
	NoAuth ApiAuthType = iota
	Auth
)

type ApiActionType int

const (
	Default ApiActionType = iota
	Create
	Delete
	Update
	Show
	ShowMetrics
	ShowUsage
)

var DefaultNoAuthMcApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 2,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcCreateApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcDeleteApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcUpdateApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcDefaultApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcShowApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 10,
			BurstSize:     3,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcShowMetricsApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultMcShowUsageApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var TestMcApiEndpointRateLimitSettings = &edgeproto.ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 100,
			BurstSize:     100,
		},
	},
	EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
		FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 100,
			BurstSize:     100,
		},
	},
}

/*func rateLimit(ctx context.Context, api, usr, org, ip string) (bool, error) {
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
}*/

func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := GetContext(c)
		rateLimitInfo := &ratelimit.LimiterInfo{
			Api:         c.Path(),
			Ip:          c.RealIP(),
			RateLimited: true,
		}
		ctx = ratelimit.NewLimiterInfoContext(ctx, rateLimitInfo)
		//rateLimitCtx.Ip = c.RealIP()
		// TODO: by org???
		limit, err := rateLimitMgr.Limit(ctx)
		if limit {
			log.DebugLog(log.DebugLevelInfo, "BLAH: error ratelimiting", "err", err)
			errMsg := fmt.Sprintf("%s is rejected, please retry later.", c.Path())
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return echo.NewHTTPError(http.StatusTooManyRequests, errMsg)

		}
		c = NewEchoContext(c, ctx)
		log.DebugLog(log.DebugLevelInfo, "BLAH: new echo context", "c", c, "c.ctx", GetContext(c))
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

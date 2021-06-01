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

var McControllerApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
}

var McControllerApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM,
	ReqsPerSecond: 10,
}

var NoAuthMcApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     10,
}

var NoAuthMcApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 10,
	BurstSize:     2,
}

var McCreateApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McCreateApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McDeleteApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McDeleteApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McUpdateApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McUpdateApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McDefaultApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McDefaultApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McShowApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 300,
	BurstSize:     25,
}

var McShowApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McShowMetricsApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McShowMetricsApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

var McShowUsageApiFullEndpointRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 200,
	BurstSize:     25,
}

var McShowUsageApiPerIpRateLimitSettings = &edgeproto.RateLimitSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 50,
	BurstSize:     5,
}

func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create ctx with rateLimitInfo
		ctx := GetContext(c)
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

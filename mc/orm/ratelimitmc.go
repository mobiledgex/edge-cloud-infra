package orm

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var GlobalMcApiAllRequestsRateLimitSettings = &ormapi.McRateLimitSettings{
	ApiName:         edgeproto.GlobalApiName,
	RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
	FlowSettings: []edgeproto.FlowSettings{
		edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 10000,
			BurstSize:     500,
		},
	},
}

var GlobalMcApiPerIpRateLimitSettings = &ormapi.McRateLimitSettings{
	ApiName:         edgeproto.GlobalApiName,
	RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
	FlowSettings: []edgeproto.FlowSettings{
		edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1000,
			BurstSize:     100,
		},
	},
}

var GlobalMcApiPerUserRateLimitSettings = &ormapi.McRateLimitSettings{
	ApiName:         edgeproto.GlobalApiName,
	RateLimitTarget: edgeproto.RateLimitTarget_PER_USER,
	FlowSettings: []edgeproto.FlowSettings{
		edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1000,
			BurstSize:     100,
		},
	},
}

var UserCreateAllRequestsRateLimitSettings = &ormapi.McRateLimitSettings{
	ApiName:         "/api/v1/usercreate",
	RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
	FlowSettings: []edgeproto.FlowSettings{
		edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 100,
			BurstSize:     5,
		},
	},
}

var UserCreatePerIpRateLimitSettings = &ormapi.McRateLimitSettings{
	ApiName:         "/api/v1/usercreate",
	RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
	FlowSettings: []edgeproto.FlowSettings{
		edgeproto.FlowSettings{
			FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 2,
			BurstSize:     2,
		},
	},
}

func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := GetContext(c)
		if getDisableRateLimit(ctx) {
			return nil
		}
		// Create ctx with rateLimitInfo
		callerInfo := &ratelimit.CallerInfo{
			Api: c.Path(),
		}
		claims, err := getClaims(c)
		if err != nil {
			// use IP if cannot get claims
			callerInfo.Ip = c.RealIP()
		} else {
			// use Username if can get claims
			callerInfo.User = claims.Username
		}
		// Rate limit
		err = rateLimitMgr.Limit(ctx, callerInfo)
		if err != nil {
			errMsg := fmt.Sprintf("%s is rejected, please retry later.", c.Path())
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return echo.NewHTTPError(http.StatusTooManyRequests, errMsg)

		}
		return next(c)
	}
}

func getDisableRateLimit(ctx context.Context) bool {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for disableRateLimit", "err", err)
		return false
	}
	return config.DisableRateLimit
}

func getMaxNumRateLimiters(ctx context.Context) int {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for maxNumRateLimiters", "err", err)
		return 0
	}
	return config.MaxNumRateLimiters
}

func InitRateLimitMc(ctx context.Context) error {
	if getDisableRateLimit(ctx) {
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelApi, "init ratelimit")
	db := loggedDB(ctx)

	// Create Global RateLimitSettings and UserCreate RateLimitSettings
	err := createRateLimitDbEntry(db, GlobalMcApiAllRequestsRateLimitSettings)
	if err != nil {
		return fmt.Errorf("Unable to create Global AllRequests RateLimitSettings - error: %s", err.Error())
	}
	err = createRateLimitDbEntry(db, GlobalMcApiPerIpRateLimitSettings)
	if err != nil {
		return fmt.Errorf("Unable to create Global PerIP RateLimitSettings - error: %s", err.Error())
	}
	err = createRateLimitDbEntry(db, GlobalMcApiPerUserRateLimitSettings)
	if err != nil {
		return fmt.Errorf("Unable to create Global PerUser RateLimitSettings - error: %s", err.Error())
	}
	err = createRateLimitDbEntry(db, UserCreateAllRequestsRateLimitSettings)
	if err != nil {
		return fmt.Errorf("Unable to create UserCreate AllRequests RateLimitSettings - error: %s", err.Error())
	}
	err = createRateLimitDbEntry(db, UserCreatePerIpRateLimitSettings)
	if err != nil {
		return fmt.Errorf("Unable to create UserCreate PerIP RateLimitSettings - error: %s", err.Error())
	}

	// Init RateLimitMgr and add Global RateLimitSettings and UserCreate RateLimitSettings
	rateLimitMgr = ratelimit.NewRateLimitManager(serverConfig.DisableRateLimit, defaultConfig.MaxNumRateLimiters)
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(GlobalMcApiAllRequestsRateLimitSettings), convertToRateLimitSettings(GlobalMcApiPerIpRateLimitSettings), convertToRateLimitSettings(GlobalMcApiPerUserRateLimitSettings))
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(UserCreateAllRequestsRateLimitSettings), convertToRateLimitSettings(UserCreatePerIpRateLimitSettings), nil)
	return nil
}

// Create MC RateLimit settings for an API endpoint type
func CreateRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)

	// Validate (make sure apiendpointype is not set)

	// Create RateLimitSettings
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	wrapper, err := convertToRateLimitSettingsGormWrapper(&in)
	if err != nil {
		return err
	}

	err = db.Create(wrapper).Error
	if err != nil {
		return fmt.Errorf("Unable to create RateLimitSettings %v - error: %s", in, err.Error())
	}

	rateLimitMgr.UpdateRateLimitSettings(convertToRateLimitSettings(&in))

	return nil
}

// Update MC RateLimit settings for an API endpoint type
func UpdateRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)

	// validate

	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	search := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	tx := db.Begin()

	var found ormapi.RateLimitSettingsGormWrapper
	if err = tx.Where(search).First(&found).Error; err != nil {
		return err
	}

	new, err := convertToRateLimitSettingsGormWrapper(&in)
	if err != nil {
		return err
	}

	err = tx.Model(&found).Updates(new).Error
	if err != nil {
		return err
	}

	updatedmc, err := convertToMcRateLimitSettings(&found)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	rateLimitMgr.UpdateRateLimitSettings(convertToRateLimitSettings(updatedmc))

	return nil
}

// Delete MC RateLimit settings for an API endpoint type (ie. no rate limiting)
func DeleteRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)

	// validate

	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	wrapper := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	r := db.Delete(wrapper)
	if r.Error != nil {
		return dbErr(r.Error)
	}
	if r.RowsAffected == 0 {
		return fmt.Errorf("RateLimitSettings %v not found", in)
	}

	key := edgeproto.RateLimitSettingsKey{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}
	rateLimitMgr.RemoveRateLimitSettings(key)

	return nil
}

func ShowRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)

	// validate

	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	wrapper := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	r := db.Where(wrapper)
	if r.RecordNotFound() {
		return fmt.Errorf("Specified Key not found")
	}
	if r.Error != nil {
		return dbErr(r.Error)
	}

	var records []*ormapi.RateLimitSettingsGormWrapper
	if err := r.Find(&records).Error; err != nil {
		return fmt.Errorf("Unable to find records, error %s", err.Error())
	}
	show := make([]ormapi.McRateLimitSettings, 0)
	for _, record := range records {
		settings, err := convertToMcRateLimitSettings(record)
		if err != nil {
			return fmt.Errorf("Unable to convert to RateLimitSettingsMc - error %s", err.Error())
		}
		show = append(show, *settings)
	}
	return setReply(c, &show)
}

// Helper function that converts to edgeproto.RateLimitSettings for RateLimitMgr calls
func convertToRateLimitSettings(mcsettings *ormapi.McRateLimitSettings) *edgeproto.RateLimitSettings {
	flowsettings := make([]*edgeproto.FlowSettings, 0)
	if mcsettings.FlowSettings == nil || len(mcsettings.FlowSettings) == 0 {
		flowsettings = nil
	} else {
		for _, settings := range mcsettings.FlowSettings {
			flowsettings = append(flowsettings, &settings)
		}
	}

	maxreqssettings := make([]*edgeproto.MaxReqsSettings, 0)
	if mcsettings.MaxReqsSettings == nil || len(mcsettings.MaxReqsSettings) == 0 {
		maxreqssettings = nil
	} else {
		for _, settings := range mcsettings.MaxReqsSettings {
			maxreqssettings = append(maxreqssettings, &settings)
		}
	}

	return &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         mcsettings.ApiName,
			RateLimitTarget: mcsettings.RateLimitTarget,
		},
		FlowSettings:    flowsettings,
		MaxReqsSettings: maxreqssettings,
	}
}

func createRateLimitDbEntry(db *gorm.DB, settings *ormapi.McRateLimitSettings) error {
	wrapper, err := convertToRateLimitSettingsGormWrapper(settings)
	if err != nil {
		return err
	}
	return db.Create(wrapper).Error
}

func convertToMcRateLimitSettings(r *ormapi.RateLimitSettingsGormWrapper) (*ormapi.McRateLimitSettings, error) {
	settings := &ormapi.McRateLimitSettings{
		ApiName:         r.ApiName,
		RateLimitTarget: r.RateLimitTarget,
	}

	if r.FlowSettings != nil && string(r.FlowSettings) != "" {
		var fsettings []edgeproto.FlowSettings
		err := json.Unmarshal(r.FlowSettings, &fsettings)
		if err != nil {
			return nil, err
		}
		settings.FlowSettings = fsettings
	}

	if r.MaxReqsSettings != nil && string(r.MaxReqsSettings) != "" {
		var msettings []edgeproto.MaxReqsSettings
		err := json.Unmarshal(r.MaxReqsSettings, &msettings)
		if err != nil {
			return nil, err
		}
		settings.MaxReqsSettings = msettings
	}
	return settings, nil
}

func convertToRateLimitSettingsGormWrapper(r *ormapi.McRateLimitSettings) (*ormapi.RateLimitSettingsGormWrapper, error) {
	wrapper := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         r.ApiName,
		RateLimitTarget: r.RateLimitTarget,
	}
	if r.FlowSettings != nil {
		b, err := json.Marshal(r.FlowSettings)
		if err != nil {
			return nil, err
		}
		wrapper.FlowSettings = b
	}
	if r.MaxReqsSettings != nil {
		b, err := json.Marshal(r.MaxReqsSettings)
		if err != nil {
			return nil, err
		}
		wrapper.MaxReqsSettings = b
	}
	return wrapper, nil
}

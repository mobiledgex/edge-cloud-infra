package orm

import (
	"context"
	fmt "fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Default McRateLimitSettings structs

var GlobalAllRequestsMcRateLimitSettings = &ormapi.McRateLimitSettings{
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

var GlobalPerIpMcRateLimitSettings = &ormapi.McRateLimitSettings{
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

var GlobalPerUserMcRateLimitSettings = &ormapi.McRateLimitSettings{
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

var UserCreateAllRequestsMcRateLimitSettings = &ormapi.McRateLimitSettings{
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

var UserCreatePerIpMcRateLimitSettings = &ormapi.McRateLimitSettings{
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

/*
 * Intialize McRateLimitSettings for MC APIs
 * Store default McRateLimitSettings in postgres
 * Add RateLimitSettings to RateLimitMgr
 */
func InitRateLimitMc(ctx context.Context) error {
	if getDisableRateLimit(ctx) {
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelApi, "init ratelimit")
	db := loggedDB(ctx)

	createFunc := func(settings interface{}) error {
		return db.Create(settings).Error
	}

	// Create Global RateLimitSettings and UserCreate RateLimitSettings entries in postgres
	var err error
	if err = executeDbOperationOnMcRateLimitSettings(GlobalAllRequestsMcRateLimitSettings, createFunc); err != nil {
		return fmt.Errorf("Unable to create Global AllRequests RateLimitSettings - error: %s", err.Error())
	}
	if err = executeDbOperationOnMcRateLimitSettings(GlobalPerIpMcRateLimitSettings, createFunc); err != nil {
		return fmt.Errorf("Unable to create Global PerIP RateLimitSettings - error: %s", err.Error())
	}
	if err = executeDbOperationOnMcRateLimitSettings(GlobalPerUserMcRateLimitSettings, createFunc); err != nil {
		return fmt.Errorf("Unable to create Global PerUser RateLimitSettings - error: %s", err.Error())
	}
	if err = executeDbOperationOnMcRateLimitSettings(UserCreateAllRequestsMcRateLimitSettings, createFunc); err != nil {
		return fmt.Errorf("Unable to create UserCreate AllRequests RateLimitSettings - error: %s", err.Error())
	}
	if err = executeDbOperationOnMcRateLimitSettings(UserCreatePerIpMcRateLimitSettings, createFunc); err != nil {
		return fmt.Errorf("Unable to create UserCreate PerIP RateLimitSettings - error: %s", err.Error())
	}

	// Init RateLimitMgr and add Global RateLimitSettings and UserCreate RateLimitSettings
	rateLimitMgr = ratelimit.NewRateLimitManager(getDisableRateLimit(ctx), getMaxNumPerIpRateLimiters(ctx), getMaxNumPerUserRateLimiters(ctx))
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(GlobalAllRequestsMcRateLimitSettings), convertToRateLimitSettings(GlobalPerIpMcRateLimitSettings), convertToRateLimitSettings(GlobalPerUserMcRateLimitSettings))
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(UserCreateAllRequestsMcRateLimitSettings), convertToRateLimitSettings(UserCreatePerIpMcRateLimitSettings), nil)
	return nil
}

// Generates unique id for McRateLimitFlowSettings or McRateLimitMaxReqsSettings based on the index the setting occurs in the slice
func generateId(apiName string, rateLimitTarget edgeproto.RateLimitTarget, idx int) string {
	return fmt.Sprintf("%s-%s-%d", apiName, rateLimitTarget, idx)
}

/*
 * Pulls out all FlowSettings and MaxReqsSettings from McRateLimitSettings and converts each to McRateLimitFlowSettings and McRateLimitMaxReqsSettings respectively
 * Executes provided db operation on each
 */
func executeDbOperationOnMcRateLimitSettings(settings *ormapi.McRateLimitSettings, operation func(settings interface{}) error) error {
	apiName := settings.ApiName
	rateLimitTarget := settings.RateLimitTarget

	// Add FlowSettings to postgres
	for idx, flowsetting := range settings.FlowSettings {
		mcflowsettings := &ormapi.McRateLimitFlowSettings{
			Id:              generateId(apiName, rateLimitTarget, idx),
			ApiName:         apiName,
			RateLimitTarget: rateLimitTarget,
			FlowAlgorithm:   flowsetting.FlowAlgorithm,
			ReqsPerSecond:   flowsetting.ReqsPerSecond,
			BurstSize:       flowsetting.BurstSize,
		}
		err := operation(mcflowsettings)
		if err != nil {
			return err
		}
	}

	// Add MaxReqsSettings to postgres
	for idx, maxreqssetting := range settings.MaxReqsSettings {
		mcmaxreqssettings := &ormapi.McRateLimitMaxReqsSettings{
			Id:               generateId(apiName, rateLimitTarget, idx),
			ApiName:          apiName,
			RateLimitTarget:  rateLimitTarget,
			MaxReqsAlgorithm: maxreqssetting.MaxReqsAlgorithm,
			MaxRequests:      maxreqssetting.MaxRequests,
			Interval:         int64(maxreqssetting.Interval),
		}
		err := operation(mcmaxreqssettings)
		if err != nil {
			return err
		}
	}
	return nil
}

// Echo middleware function that handles rate limiting for MC APIs
func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := GetContext(c)

		// Check if rate limiting is disabled, if disabled continue
		if getDisableRateLimit(ctx) {
			return next(c)
		}

		// Create callerInfo
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
		if err = rateLimitMgr.Limit(ctx, callerInfo); err != nil {
			errMsg := fmt.Sprintf("%s is rejected, please retry later.", c.Path())
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return echo.NewHTTPError(http.StatusTooManyRequests, errMsg)

		}
		return next(c)
	}
}

// Helper function that grabs the DisableRateLimit bool from the Config struct
func getDisableRateLimit(ctx context.Context) bool {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for disableRateLimit", "err", err)
		return defaultConfig.DisableRateLimit
	}
	return config.DisableRateLimit
}

// Helper function that grabs the MaxNumPerIpRateLimiters int from the Config struct
func getMaxNumPerIpRateLimiters(ctx context.Context) int {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for maxNumPerIpRateLimiters", "err", err)
		return defaultConfig.MaxNumPerIpRateLimiters
	}
	return config.MaxNumPerIpRateLimiters
}

// Helper function that grabs the MaxNumPerUserRateLimiters int from the Config struct
func getMaxNumPerUserRateLimiters(ctx context.Context) int {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for maxNumPerUserRateLimiters", "err", err)
		return defaultConfig.MaxNumPerUserRateLimiters
	}
	return config.MaxNumPerUserRateLimiters
}

// Create MC RateLimit settings
func CreateRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)

	// Check if rate limiting is disabled
	if getDisableRateLimit(ctx) {
		return fmt.Errorf("DisableRateLimit must be false to create ratelimitsettingsmc")
	}

	// Validate rbac
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}

	// Get McRateLimitSettings from request
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Insert new value into db
	createFunc := func(settings interface{}) error {
		db := loggedDB(ctx)
		if err = db.Create(settings).Error; err != nil {
			return fmt.Errorf("Unable to create RateLimitSettings %v - error: %s", in, err.Error())
		}
		return nil
	}

	err = executeDbOperationOnMcRateLimitSettings(&in, createFunc)
	if err != nil {
		return err
	}

	// Update RateLimitMgr with new RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(convertToRateLimitSettings(&in))
	return nil
}

// Update MC RateLimit settings
func UpdateRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)

	// Check if rate limiting is disabled
	if getDisableRateLimit(ctx) {
		return fmt.Errorf("DisableRateLimit must be false to update ratelimitsettingsmc")
	}

	// Validate rbac
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}

	// Get McRateLimitSettings from request
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	updateFunc := func(settings interface{}) error {
		var search interface{}
		var old interface{}
		switch typ := settings.(type) {
		case *ormapi.McRateLimitFlowSettings:
			search = &ormapi.McRateLimitFlowSettings{
				Id: typ.Id,
			}
			old = &ormapi.McRateLimitFlowSettings{}
		case *ormapi.McRateLimitMaxReqsSettings:
			search = &ormapi.McRateLimitMaxReqsSettings{
				Id: typ.Id,
			}
			old = &ormapi.McRateLimitMaxReqsSettings{}
		default:
			return fmt.Errorf("Unknown settings type %v", typ)
		}

		db := loggedDB(ctx)
		// Search for entry with corresponding primary key
		if err = db.Where(search).First(old).Error; err != nil {
			return err
		}
		// Update found entry with new values
		if err = db.Model(old).Updates(settings).Error; err != nil {
			return err
		}
		return nil
	}

	err = executeDbOperationOnMcRateLimitSettings(&in, updateFunc)
	if err != nil {
		return err
	}

	// Update RateLimitMgr with updated RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(convertToRateLimitSettings(&in))
	return nil
}

// Delete MC RateLimit settings (ie. no rate limiting for specified api and ratelimittarget)
func DeleteRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)

	// Check if rate limiting is disabled
	if getDisableRateLimit(ctx) {
		return fmt.Errorf("DisableRateLimit must be false to delete ratelimitsettingsmc")
	}

	// Validate rbac
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}

	// Get McRateLimitSettings from request
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	search := &ormapi.McRateLimitFlowSettings{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	// Search for all entries with specified primary keys (if fields are not specified, fields are left out of search)
	db := loggedDB(ctx)
	r := db.Where(search)
	if r.RecordNotFound() {
		return fmt.Errorf("Specified Key not found")
	}
	if r.Error != nil {
		return dbErr(r.Error)
	}

	mcflowrecords := make([]*ormapi.McRateLimitFlowSettings, 0)
	mcmaxreqsrecords := make([]*ormapi.McRateLimitMaxReqsSettings, 0)
	if err = r.Find(&mcflowrecords).Error; err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Unable to find records for flow", "error", err.Error())
	}
	if err = r.Find(&mcmaxreqsrecords).Error; err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Unable to find records for maxreqs", "error", err.Error())
	}

	for _, record := range mcflowrecords {
		// Remove entry from db
		d := db.Delete(record)
		if d.Error != nil {
			return dbErr(d.Error)
		}
		if d.RowsAffected == 0 {
			return fmt.Errorf("Unable to delete flow record: %v", record)
		}
	}

	for _, record := range mcmaxreqsrecords {
		// Remove entry from db
		d := db.Delete(record)
		if d.Error != nil {
			return dbErr(d.Error)
		}
		if d.RowsAffected == 0 {
			return fmt.Errorf("Unable to delete maxreqs record: %v", record)
		}
	}

	// Remove RateLimitSettings from RateLimitMgr
	key := edgeproto.RateLimitSettingsKey{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}
	rateLimitMgr.RemoveRateLimitSettings(key)
	return nil
}

// Show MC RateLimit settings
func ShowRateLimitSettingsMc(c echo.Context) error {
	ctx := GetContext(c)

	// Check if rate limiting is disabled
	if getDisableRateLimit(ctx) {
		return fmt.Errorf("DisableRateLimit must be false to show ratelimitsettingsmc")
	}

	// Validate rbac
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionManage); err != nil {
		return err
	}

	// Get McRateLimitSettings from request
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	search := &ormapi.McRateLimitFlowSettings{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	// Search for all entries with specified primary keys (if fields are not specified, fields are left out of search)
	db := loggedDB(ctx)
	r := db.Where(search)
	if r.RecordNotFound() {
		return fmt.Errorf("Specified Key not found")
	}
	if r.Error != nil {
		return dbErr(r.Error)
	}

	mcflowrecords := make([]*ormapi.McRateLimitFlowSettings, 0)
	mcmaxreqsrecords := make([]*ormapi.McRateLimitMaxReqsSettings, 0)
	if err = r.Find(&mcflowrecords).Error; err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Unable to find records for flow", "error", err.Error())
	}
	if err = r.Find(&mcmaxreqsrecords).Error; err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Unable to find records for maxreqs", "error", err.Error())
	}

	show := convertToMcRateLimitSettings(mcflowrecords, mcmaxreqsrecords)
	return setReply(c, &show)
}

func convertToEdgeProtoFlowSettings(flowsettings *ormapi.McRateLimitFlowSettings) edgeproto.FlowSettings {
	return edgeproto.FlowSettings{
		FlowAlgorithm: flowsettings.FlowAlgorithm,
		ReqsPerSecond: flowsettings.ReqsPerSecond,
		BurstSize:     flowsettings.BurstSize,
	}
}

func convertToEdgeProtoMaxReqsSettings(maxreqssettings *ormapi.McRateLimitMaxReqsSettings) edgeproto.MaxReqsSettings {
	return edgeproto.MaxReqsSettings{
		MaxReqsAlgorithm: maxreqssettings.MaxReqsAlgorithm,
		MaxRequests:      maxreqssettings.MaxRequests,
		Interval:         edgeproto.Duration(maxreqssettings.Interval),
	}
}

// Helper function to convert lists of McRateLimitFlowSettings and McRateLimitMaxReqsSettings to McRateLimitSettings to return to api caller
func convertToMcRateLimitSettings(flowsettings []*ormapi.McRateLimitFlowSettings, maxreqssettings []*ormapi.McRateLimitMaxReqsSettings) []*ormapi.McRateLimitSettings {
	settingsmap := make(map[edgeproto.RateLimitSettingsKey]*ormapi.McRateLimitSettings)

	for _, flowsetting := range flowsettings {
		key := edgeproto.RateLimitSettingsKey{
			ApiName:         flowsetting.ApiName,
			RateLimitTarget: flowsetting.RateLimitTarget,
		}
		mcratelimitsetting, ok := settingsmap[key]
		if !ok || mcratelimitsetting == nil {
			mcratelimitsetting = &ormapi.McRateLimitSettings{
				ApiName:         flowsetting.ApiName,
				RateLimitTarget: flowsetting.RateLimitTarget,
				FlowSettings:    make([]edgeproto.FlowSettings, 0),
				MaxReqsSettings: make([]edgeproto.MaxReqsSettings, 0),
			}
		}
		mcratelimitsetting.FlowSettings = append(mcratelimitsetting.FlowSettings, convertToEdgeProtoFlowSettings(flowsetting))
		settingsmap[key] = mcratelimitsetting
	}

	for _, maxreqssetting := range maxreqssettings {
		key := edgeproto.RateLimitSettingsKey{
			ApiName:         maxreqssetting.ApiName,
			RateLimitTarget: maxreqssetting.RateLimitTarget,
		}
		mcratelimitsetting, ok := settingsmap[key]
		if !ok || mcratelimitsetting == nil {
			mcratelimitsetting = &ormapi.McRateLimitSettings{
				ApiName:         maxreqssetting.ApiName,
				RateLimitTarget: maxreqssetting.RateLimitTarget,
				FlowSettings:    make([]edgeproto.FlowSettings, 0),
				MaxReqsSettings: make([]edgeproto.MaxReqsSettings, 0),
			}
		}
		mcratelimitsetting.MaxReqsSettings = append(mcratelimitsetting.MaxReqsSettings, convertToEdgeProtoMaxReqsSettings(maxreqssetting))
		settingsmap[key] = mcratelimitsetting
	}

	mcratelimitsettings := make([]*ormapi.McRateLimitSettings, 0)
	for _, settings := range settingsmap {
		mcratelimitsettings = append(mcratelimitsettings, settings)
	}
	return mcratelimitsettings
}

// Helper function that converts ormapi.McRateLimitSettings to edgeproto.RateLimitSettings for RateLimitMgr calls
func convertToRateLimitSettings(mcsettings *ormapi.McRateLimitSettings) *edgeproto.RateLimitSettings {
	// Init RateLimitSettings with key
	settings := &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         mcsettings.ApiName,
			RateLimitTarget: mcsettings.RateLimitTarget,
		},
	}

	// Add FlowSettings
	if mcsettings.FlowSettings != nil && len(mcsettings.FlowSettings) > 0 {
		flowsettings := make([]*edgeproto.FlowSettings, 0)
		for _, settings := range mcsettings.FlowSettings {
			flowsettings = append(flowsettings, &settings)
		}
		settings.FlowSettings = flowsettings
	}

	// Add MaxReqsSettings
	if mcsettings.MaxReqsSettings != nil && len(mcsettings.MaxReqsSettings) > 0 {
		maxreqssettings := make([]*edgeproto.MaxReqsSettings, 0)
		for _, settings := range mcsettings.MaxReqsSettings {
			maxreqssettings = append(maxreqssettings, &settings)
		}
		settings.MaxReqsSettings = maxreqssettings
	}
	return settings
}

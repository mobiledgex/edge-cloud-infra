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

	// Create Global RateLimitSettings and UserCreate RateLimitSettings entries in postgres
	var err error
	if err = createRateLimitDbEntry(db, GlobalAllRequestsMcRateLimitSettings); err != nil {
		return fmt.Errorf("Unable to create Global AllRequests RateLimitSettings - error: %s", err.Error())
	}
	if err = createRateLimitDbEntry(db, GlobalPerIpMcRateLimitSettings); err != nil {
		return fmt.Errorf("Unable to create Global PerIP RateLimitSettings - error: %s", err.Error())
	}
	if err = createRateLimitDbEntry(db, GlobalPerUserMcRateLimitSettings); err != nil {
		return fmt.Errorf("Unable to create Global PerUser RateLimitSettings - error: %s", err.Error())
	}
	if err = createRateLimitDbEntry(db, UserCreateAllRequestsMcRateLimitSettings); err != nil {
		return fmt.Errorf("Unable to create UserCreate AllRequests RateLimitSettings - error: %s", err.Error())
	}
	if err = createRateLimitDbEntry(db, UserCreatePerIpMcRateLimitSettings); err != nil {
		return fmt.Errorf("Unable to create UserCreate PerIP RateLimitSettings - error: %s", err.Error())
	}

	// Init RateLimitMgr and add Global RateLimitSettings and UserCreate RateLimitSettings
	rateLimitMgr = ratelimit.NewRateLimitManager(getDisableRateLimit(ctx), getMaxNumRateLimiters(ctx))
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(GlobalAllRequestsMcRateLimitSettings), convertToRateLimitSettings(GlobalPerIpMcRateLimitSettings), convertToRateLimitSettings(GlobalPerUserMcRateLimitSettings))
	rateLimitMgr.CreateApiEndpointLimiter(convertToRateLimitSettings(UserCreateAllRequestsMcRateLimitSettings), convertToRateLimitSettings(UserCreatePerIpMcRateLimitSettings), nil)
	return nil
}

// Helper function that converts McRateLimitSettings into RateLimitSettingsGormWrapper and then creates entry in db
func createRateLimitDbEntry(db *gorm.DB, settings *ormapi.McRateLimitSettings) error {
	wrapper, err := convertToRateLimitSettingsGormWrapper(settings)
	if err != nil {
		return err
	}
	return db.Create(wrapper).Error
}

// Echo middleware function that handles rate limiting for MC APIs
func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := GetContext(c)

		// Check if rate limiting is disabled
		if getDisableRateLimit(ctx) {
			return nil
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

// Helper function that grabs the MaxNumRateLimiters int from the Config struct
func getMaxNumRateLimiters(ctx context.Context) int {
	config, err := getConfig(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "unable to check config for maxNumRateLimiters", "err", err)
		return defaultConfig.MaxNumRateLimiters
	}
	return config.MaxNumRateLimiters
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

	// Convert McRateLimitSettings to RateLimitSettingsGormWrapper to easily store in postgres
	wrapper, err := convertToRateLimitSettingsGormWrapper(&in)
	if err != nil {
		return err
	}

	// Insert new value into db
	db := loggedDB(ctx)
	if err = db.Create(wrapper).Error; err != nil {
		return fmt.Errorf("Unable to create RateLimitSettings %v - error: %s", in, err.Error())
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

	// Get McRateLimitSettings from request and convert to gorm wrapper struct
	in := ormapi.McRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	new, err := convertToRateLimitSettingsGormWrapper(&in)
	if err != nil {
		return err
	}

	// Create RateLimitSettingsGormWrapper with primary keys for lookup
	search := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	// Begin transaction
	db := loggedDB(ctx)
	tx := db.Begin()

	// Search for entry with corresponding primary kesy
	var old ormapi.RateLimitSettingsGormWrapper
	if err = tx.Where(search).First(&old).Error; err != nil {
		return err
	}

	// Update found entry with new values
	if err = tx.Model(&old).Updates(new).Error; err != nil {
		return err
	}

	// Convert updated entry to McRateLimitSettings
	updatedmc, err := convertToMcRateLimitSettings(&old)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	// Update RateLimitMgr with updated RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(convertToRateLimitSettings(updatedmc))
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

	// Create RateLimitSettingsGormWrapper with primary keys for lookup
	search := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         in.ApiName,
		RateLimitTarget: in.RateLimitTarget,
	}

	// Remove entry from db
	db := loggedDB(ctx)
	r := db.Delete(search)
	if r.Error != nil {
		return dbErr(r.Error)
	}
	if r.RowsAffected == 0 {
		return fmt.Errorf("RateLimitSettings %v not found", in)
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

	// Create RateLimitSettingsGormWrapper with primary keys for lookup
	search := &ormapi.RateLimitSettingsGormWrapper{
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
	var records []*ormapi.RateLimitSettingsGormWrapper
	if err := r.Find(&records).Error; err != nil {
		return fmt.Errorf("Unable to find records, error %s", err.Error())
	}

	// Create list of McRateLimitSettings from db search results
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

// Helper function to convert RateLimitSettingsGormWrapper to McRateLimitSettings to return to api caller
func convertToMcRateLimitSettings(r *ormapi.RateLimitSettingsGormWrapper) (*ormapi.McRateLimitSettings, error) {
	// Init McRateLimitSettings
	settings := &ormapi.McRateLimitSettings{
		ApiName:         r.ApiName,
		RateLimitTarget: r.RateLimitTarget,
	}

	// Unmarshal []byte into []edgeproto.FlowSettings
	if r.FlowSettings != nil && string(r.FlowSettings) != "" {
		var fsettings []edgeproto.FlowSettings
		err := json.Unmarshal(r.FlowSettings, &fsettings)
		if err != nil {
			return nil, err
		}
		settings.FlowSettings = fsettings
	}

	// Unmarshal []byte into []edgeproto.MaxReqsSettings
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

// Helper function to convert McRateLimitSettings to RateLimitSettingsGormWrapper to store in postgres
func convertToRateLimitSettingsGormWrapper(r *ormapi.McRateLimitSettings) (*ormapi.RateLimitSettingsGormWrapper, error) {
	// Init RateLimitSettingsGormWrapper with primary keys
	wrapper := &ormapi.RateLimitSettingsGormWrapper{
		ApiName:         r.ApiName,
		RateLimitTarget: r.RateLimitTarget,
	}

	// Marshal slice of FlowSettings into []byte
	if r.FlowSettings != nil {
		b, err := json.Marshal(r.FlowSettings)
		if err != nil {
			return nil, err
		}
		wrapper.FlowSettings = b
	}

	// Marshal slice of MaxReqsSettings into []byte
	if r.MaxReqsSettings != nil {
		b, err := json.Marshal(r.MaxReqsSettings)
		if err != nil {
			return nil, err
		}
		wrapper.MaxReqsSettings = b
	}
	return wrapper, nil
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

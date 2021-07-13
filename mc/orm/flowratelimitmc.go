package orm

import (
	fmt "fmt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

// Create MC RateLimit Flow settings
func CreateFlowRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitFlowSettings from request
	in := ormapi.McRateLimitFlowSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Create McRateLimitFlowSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	if err := tx.Create(&in).Error; err != nil {
		return fmt.Errorf("Unable to create FlowRateLimitSettings %v - error: %s", in, err.Error())
	}

	// Get all records with same apiname and ratelimittarget to build edgeproto.RateLimitSettings
	mcflowrecords, mcmaxreqsrecords, err := getAllEntriesForApiAndTarget(ctx, db, in.ApiName, in.RateLimitTarget)
	if err != nil {
		tx.Rollback()
		return err
	}
	mcflowrecords = append(mcflowrecords, &in)

	// Get edgeproto.RateLimitSettings to update RateLimiMgr
	ratelimitsettings, err := buildRateLimitSettings(in.ApiName, in.RateLimitTarget, mcflowrecords, mcmaxreqsrecords)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit create changes
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// Update RateLimitMgr with new RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(ratelimitsettings)
	return nil
}

// Update MC RateLimit flow settings
func UpdateFlowRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitFlowSettings from request
	in := ormapi.McRateLimitFlowSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Update McRateLimitFlowSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	searchFlow := &ormapi.McRateLimitFlowSettings{
		FlowSettingsName: in.FlowSettingsName,
	}
	err = tx.Model(searchFlow).Updates(&in).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("Unable to find McRateLimitFlowSettings for specified name: %s", in.FlowSettingsName)
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get all records with same apiname and ratelimittarget to build edgeproto.RateLimitSettings
	mcflowrecords, mcmaxreqsrecords, err := getAllEntriesForApiAndTarget(ctx, db, in.ApiName, in.RateLimitTarget)
	if err != nil {
		tx.Rollback()
		return err
	}
	mcflowrecords = append(mcflowrecords, &in)

	// Build edgeproto.RateLimitSettings to update RateLimiMgr
	ratelimitsettings, err := buildRateLimitSettings(in.ApiName, in.RateLimitTarget, mcflowrecords, mcmaxreqsrecords)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit updates to entry
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// Update RateLimitMgr with new RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(ratelimitsettings)
	return nil
}

// Delete MC RateLimit flow settings
func DeleteFlowRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitFlowSettings from request
	in := ormapi.McRateLimitFlowSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Remove McRateLimitFlowSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	err = tx.Delete(&in).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("Unable to find McRateLimitFlowSettings for specified name: %s", in.FlowSettingsName)
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get all records with same apiname and ratelimittarget to build edgeproto.RateLimitSettings
	mcflowrecords, mcmaxreqsrecords, err := getAllEntriesForApiAndTarget(ctx, db, in.ApiName, in.RateLimitTarget)
	if err != nil {
		tx.Rollback()
		return err
	}
	mcflowrecords = append(mcflowrecords, &in)

	// Build edgeproto.RateLimitSettings to update RateLimiMgr
	ratelimitsettings, err := buildRateLimitSettings(in.ApiName, in.RateLimitTarget, mcflowrecords, mcmaxreqsrecords)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit deleted entry
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// Update RateLimitMgr with new RateLimitSettings
	rateLimitMgr.UpdateRateLimitSettings(ratelimitsettings)
	return nil
}

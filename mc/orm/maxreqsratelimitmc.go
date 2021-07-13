package orm

import (
	fmt "fmt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

// Create MC RateLimit MaxReqs settings
func CreateMaxReqsRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitMaxReqsSettings from request
	in := ormapi.McRateLimitMaxReqsSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Create McRateLimitMaxReqsSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	if err := tx.Create(&in).Error; err != nil {
		return fmt.Errorf("Unable to create MaxReqsRateLimitSettings %v - error: %s", in, err.Error())
	}

	// Get all records with same apiname and ratelimittarget to build edgeproto.RateLimitSettings
	mcflowrecords, mcmaxreqsrecords, err := getAllEntriesForApiAndTarget(ctx, db, in.ApiName, in.RateLimitTarget)
	if err != nil {
		tx.Rollback()
		return err
	}
	mcmaxreqsrecords = append(mcmaxreqsrecords, &in)

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

// Update MC RateLimit maxreqs settings
func UpdateMaxReqsRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitMaxReqsSettings from request
	in := ormapi.McRateLimitMaxReqsSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Update McRateLimitMaxReqsSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	searchMaxReqs := &ormapi.McRateLimitMaxReqsSettings{
		MaxReqsSettingsName: in.MaxReqsSettingsName,
	}
	err = tx.Model(searchMaxReqs).Updates(&in).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("Unable to find McRateLimitMaxReqsSettings for specified name: %s", in.MaxReqsSettingsName)
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
	mcmaxreqsrecords = append(mcmaxreqsrecords, &in)

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

// Delete MC RateLimit maxreqs settings
func DeleteMaxReqsRateLimitSettingsMc(c echo.Context) error {
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

	// Get McRateLimitMaxReqsSettings from request
	in := ormapi.McRateLimitMaxReqsSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}

	// Remove McRateLimitMaxReqsSettings entry
	db := loggedDB(ctx)
	tx := db.Begin()
	err = tx.Delete(&in).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("Unable to find McRateLimitMaxReqsSettings for specified name: %s", in.MaxReqsSettingsName)
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
	mcmaxreqsrecords = append(mcmaxreqsrecords, &in)

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

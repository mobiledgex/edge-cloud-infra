package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/billing/zuora"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/rbac"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var BillingOrgTypeSelf = "self"
var BillingOrgTypeParent = "parent"

func CreateBillingOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	org := ormapi.BillingOrganization
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("billing org", org.Name)

	err = CreateBillingOrgObj(ctx, claims, &org)
	return setReply(c, err, Msg("Billing Organization created"))
}

func createBillingOrgObj(ctx context.Context, claims *UserClaims, org *ormapi.BillingOrganization) error {
	if org.Name == "" {
		return fmt.Errorf("Name not specified")
	}
	err := ValidName(org.Name)
	if err != nil {
		return err
	}
	if org.Type == BillingOrgTypeSelf {
		orgCheck, _ := orgExists(ctx, org.Name)
		if orgCheck == nil {
			return fmt.Errorf("Organization %s not found, cannot create a self BillingOrganization for it", org.Name)
		}
		if orgCheck.Type != OrgTypeDeveloper {
			return fmt.Errorf("Cannot create BillingOrganizations for operator orgs")
		}
		if err := authorized(ctx, claims.Username, org.Name, ResourceUsers, ActionManage); err != nil {
			return fmt.Errorf("Not authorized to create a Billing Organization")
		}
	} else if org.Type == BillingOrgTypeParent {
		orgCheck, _ := orgExists(ctx, org.Name)
		if orgCheck != nil {
			return fmt.Errorf("Cannot create a parent BillingOrganization. Name %s is already in use by an Organization", org.Name)
		}
	} else {
		return fmt.Errorf("Invalid type: %s. Type must be either \"%s\" or \"%s\"", org.Type, BillingOrgTypeSelf, BillingOrgTypeParent)
	}

	err := createZuoraAccount(ctx, org)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	err = db.Create(&org).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"organizations_pkey") {
			return fmt.Errorf("Billing Organization with name %s (case-insensitive) already exists", org.Name)
		}
		return dbErr(err)
	}

	// If its a self billing org, just use the same group that was created for the regular org.
	if org.Type == BillingOrgTypeParent {
		// TODO: set user to admin role of billing organization
	}
	return nil
}

func billingOrgExists(ctx context.Context, orgName string) (*ormapi.BillingOrganization, error) {
	lookup := ormapi.BillingOrganization{
		Name: orgName,
	}
	db := loggedDB(ctx)
	org := ormapi.BillingOrganization{}
	res := db.Where(&lookup).First(&org)
	if res.RecordNotFound() {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &org, nil
}

func createZuoraAccount(ctx context.Context, info *ormapi.BillingOrganization) error {
	accountInfo := zuora.AccountInfo{OrgName: info.Name}
	billTo := zuora.CustomerBillToContact{
		FirstName: info.Name,
		LastName:  info.Name,
		WorkEmail: info.Email,
		Address1:  info.Address,
		City:      info.City,
		Country:   info.Country,
		State:     info.State,
	}
	currency := "USD" // for now, later on have a function that selects it based on address?
	err := zuora.CreateCustomer(info.Name, currency, &billTo, &accountInfo)
	if err != nil {
		return err
	}

	//put the account info in the db
	db := loggedDB(ctx)
	err = db.Create(&accountInfo).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"accountinfo_pkey") {
			return fmt.Errorf("AccountInfo with name %s (case-insensitive) already exists", info.Name)
		}
		return dbErr(err)
	}
	return nil
}

func DeleteBillingOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	org := ormapi.BillingOrganization{}
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("billing org", org.Name)

	err = deleteBillingOrgObj(ctx, claims, &org)
	return setReply(c, err, Msg("Billing Organization deleted"))
}

func deleteBillingOrgObj(ctx context.Context, claims *UserClaims, org *ormapi.BillingOrganization) error {
	if org.Name == "" {
		return fmt.Errorf("Organization name not specified")
	}
	if err := authorized(ctx, claims.Username, org.Name, ResourceUsers, ActionManage); err != nil {
		return err
	}
	// mark org for delete in progress
	db := loggedDB(ctx)
	doMark := true
	err := markOrgForDelete(db, org.Name, doMark)
	if err != nil {
		return err
	}

	// check to see if orgs related to this BillingOrg are still up
	err = billingOrgInUse(ctx, org.Name)
	if err != nil {
		undoerr := markOrgForDelete(db, org.Name, !doMark)
		if undoerr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "undo mark org for delete", "undoerr", undoerr)
		}
		return err
	}

	// delete org
	err = db.Delete(&org).Error
	if err != nil {
		undoerr := markOrgForDelete(db, org.Name, !doMark)
		if undoerr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "undo mark org for delete", "undoerr", undoerr)
		}
		if strings.Contains(err.Error(), "violates foreign key constraint \"org_cloudlet_pools_org_fkey\"") {
			return fmt.Errorf("Cannot delete organization because it is referenced by an OrgCloudletPool")
		}
		return dbErr(err)
	}
	// delete all casbin groups associated with org
	groups, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return dbErr(err)
	}
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if len(strs) == 2 && strs[0] == org.Name {
			err = enforcer.RemoveGroupingPolicy(ctx, grp[0], grp[1])
			if err != nil {
				return dbErr(err)
			}
		}
	}
	gitlabDeleteGroup(ctx, org)
	artifactoryDeleteGroupObjects(ctx, org.Name, "")
	return nil
}

func cancelZuoraSubscription(ctx context.Context, orgName string) error {
	//cancel the zuora subscription and remove it from the db
	// get the full accountInfo
	info, err := GetAccountObj(ctx, orgName)
	if err != nil {
		return err
	}
	// remove account from db
	db := loggedDB(ctx)
	err = db.Delete(info).Error
	if err != nil {
		return dbErr(err)
	}
	err = zuora.CancelSubscription(info)
	if err != nil {
		return err
	}
	return nil
}

func GetAccountObj(ctx context.Context, orgName string) (*zuora.AccountInfo, error) {
	if orgName == "" {
		return nil, fmt.Errorf("no orgName specified")
	}
	acc := zuora.AccountInfo{
		OrgName: orgName,
	}
	db := loggedDB(ctx)
	res := db.Where(&acc).First(&acc)
	if res.Error != nil {
		if res.RecordNotFound() {
			return nil, fmt.Errorf("account \"%s\" not found", orgName)
		}
		return nil, res.Error
	}
	return &acc, nil
}

//TODO: write this
func billingOrgInUse(ctx context.Context, orgName string) error {
	ctrls, err := ShowControllerObj(ctx, nil)
	if err != nil {
		return err
	}
	errs := make([]string, 0)
	var mux sync.Mutex
	var wg sync.WaitGroup

	for _, ctrl := range ctrls {
		wg.Add(1)
		go func(c ormapi.Controller) {
			err := orgInUseRegion(ctx, c, orgName)
			if err != nil {
				mux.Lock()
				errs = append(errs, fmt.Sprintf("region %s: %v", c.Region, err))
				mux.Unlock()
			}
			wg.Done()
		}(ctrl)
	}
	wg.Wait()
	if len(errs) == 0 {
		return nil
	}
	sort.Strings(errs)
	return fmt.Errorf("Organization %s in use or check failed: %s", orgName, strings.Join(errs, "; "))
}

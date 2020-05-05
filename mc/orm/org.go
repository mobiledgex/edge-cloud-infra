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

// Organization Type names for ORM database
var OrgTypeAdmin = "admin"
var OrgTypeDeveloper = "developer"
var OrgTypeOperator = "operator"

func CreateOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
<<<<<<< HEAD
	org := ormapi.Organization{}
	if err := c.Bind(&org); err != nil {
		return bindErr(c, err)
=======
	orgInfo := ormapi.OrgInfo{}
	if err := c.Bind(&orgInfo); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
>>>>>>> started billing org
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("org", orgInfo.Name)

	err = CreateOrgObj(ctx, claims, &orgInfo)
	return setReply(c, err, Msg("Organization created"))
}

func createZuoraAccount(ctx context.Context, info *ormapi.BillingOrganization) error {
	//create the account in zuora
	if info.Type != OrgTypeDeveloper || info.Name == "mobiledgex" {
		return nil
	}
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

func CreateOrgObj(ctx context.Context, claims *UserClaims, orgInfo *ormapi.OrgInfo) error {
	org := ormapi.Organization{
		Name: orgInfo.Name,
		Type: orgInfo.Type,
	}
	if org.Name == "" {
		return fmt.Errorf("Name not specified")
	}
	err := ValidName(orgInfo.Name)
	if err != nil {
		return err
	}
	// any user can create their own organization

	role := ""
	if orgInfo.Type == OrgTypeDeveloper {
		role = RoleDeveloperManager
	} else if org.Type == OrgTypeOperator {
		role = RoleOperatorManager
	} else {
		return fmt.Errorf("Organization type must be %s, or %s", OrgTypeDeveloper, OrgTypeOperator)
	}
	if strings.ToLower(claims.Username) == strings.ToLower(org.Name) {
		return fmt.Errorf("org name cannot be same as existing user name")
	}
	if strings.ToLower(org.Name) == strings.ToLower(cloudcommon.OrganizationMobiledgeX) || strings.ToLower(org.Name) == strings.ToLower(cloudcommon.OrganizationEdgeBox) {
		if err := authorized(ctx, claims.Username, "", ResourceUsers, ActionManage); err != nil {
			return fmt.Errorf("Not authorized to create reserved org %s", org.Name)
		}
	}
	createBilling, err := checkBillingInfo(orgInfo)
	if err != nil {
		return err
	}
	// make sure that the name isnt already in use in a billingOrg
	exists, _ := billingOrgExists(ctx, orgInfo.Name)
	if exists != nil {
		return fmt.Errorf("Org name not available")
	}
	db := loggedDB(ctx)
	err = db.Create(&org).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"organizations_pkey") {
			return fmt.Errorf("Organization with name %s (case-insensitive) already exists", org.Name)
		}
		return dbErr(err)
	}
	if serverConfig.Billing && createBilling {
		// TODO: create a billingOrganization
		// add this user to the admin role of the billingOrganization
		billOrg := ormapi.BillingOrganization{
			Name:       orgInfo.Name,
			Type:       BillingOrgTypeSelf,
			FirstName:  orgInfo.FirstName,
			LastName:   orgInfo.LastName,
			Email:      orgInfo.Email,
			Address:    orgInfo.Address,
			City:       orgInfo.City,
			Country:    orgInfo.Country,
			State:      orgInfo.State,
			PostalCode: orgInfo.PostalCode,
			Phone:      orgInfo.Phone,
		}
		err = createZuoraAccount(ctx, &billOrg)
		if err != nil {
			// delete the org
			db.Delete(&org)
			return err
		}
	}
	// set user to admin role of organization
	psub := rbac.GetCasbinGroup(org.Name, claims.Username)
	err = enforcer.AddGroupingPolicy(ctx, psub, role)
	if err != nil {
		return dbErr(err)
	}

	gitlabCreateGroup(ctx, orgInfo)
	r := ormapi.Role{
		Org:      org.Name,
		Username: claims.Username,
		Role:     role,
	}
	gitlabAddGroupMember(ctx, &r, org.Type)

	artifactoryCreateGroupObjects(ctx, org.Name, org.Type)
	artifactoryAddUserToGroup(ctx, &r, org.Type)

	return nil
}

func DeleteOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	org := ormapi.Organization{}
	if err := c.Bind(&org); err != nil {
		return bindErr(c, err)
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("org", org.Name)

	err = DeleteOrgObj(ctx, claims, &org)
	return setReply(c, err, Msg("Organization deleted"))
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

func DeleteOrgObj(ctx context.Context, claims *UserClaims, org *ormapi.Organization) error {
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

	// check for Controller objects belonging to org
	err = orgInUse(ctx, org.Name)
	if err != nil {
		undoerr := markOrgForDelete(db, org.Name, !doMark)
		if undoerr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "undo mark org for delete", "undoerr", undoerr)
		}
		return err
	}

	// delete org
	if serverConfig.Billing {
		err = cancelZuoraSubscription(ctx, org.Name)
		// if the accountInfo is not in the db, go ahead and delete the org
		if !strings.Contains(err.Error(), fmt.Sprintf("account \"%s\" not found", org.Name)) {
			return err
		}
	}
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

func UpdateOrg(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	// Pull json directly so we can unmarshal twice.
	// First time is to do lookup, second time is to apply
	// modified fields.
	body, err := ioutil.ReadAll(c.Request().Body)
	in := ormapi.Organization{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		return bindErr(c, err)
	}
	if in.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Organization name not specified"))
	}

	lookup := ormapi.Organization{
		Name: in.Name,
	}
	org := ormapi.Organization{}
	db := loggedDB(ctx)
	res := db.Where(&lookup).First(&org)
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("Organization not found"))
	}
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, MsgErr(dbErr(res.Error)))
	}
	oldType := org.Type

	if err := authorized(ctx, claims.Username, in.Name, ResourceUsers, ActionManage); err != nil {
		return err
	}

	// apply specified fields
	err = json.Unmarshal(body, &org)
	if err != nil {
		return bindErr(c, err)
	}
	if org.Type != oldType {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change Organization type"))
	}

	err = db.Save(&org).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MsgErr(dbErr(err)))
	}
	return nil
}

// Show Organizations that current user belongs to.
func ShowOrg(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	orgs, err := ShowOrgObj(ctx, claims)
	return setReply(c, err, orgs)
}

func ShowOrgObj(ctx context.Context, claims *UserClaims) ([]ormapi.Organization, error) {
	orgs := []ormapi.Organization{}
	db := loggedDB(ctx)
	err := authorized(ctx, claims.Username, "", ResourceUsers, ActionView)
	if err == nil {
		// super user, show all orgs
		err := db.Find(&orgs).Error
		if err != nil {
			return nil, dbErr(err)
		}
	} else {
		// show orgs for current user
		groupings, err := enforcer.GetGroupingPolicy()
		if err != nil {
			return nil, dbErr(err)
		}
		for _, grp := range groupings {
			if len(grp) < 2 {
				continue
			}
			orguser := strings.Split(grp[0], "::")
			if len(orguser) > 1 && orguser[1] == claims.Username {
				org := ormapi.Organization{}
				org.Name = orguser[0]
				err := db.Where(&org).First(&org).Error
				if err != nil {
					return nil, dbErr(err)
				}
				orgs = append(orgs, org)
			}
		}
	}
	return orgs, nil
}

func GetAllOrgs(ctx context.Context) (map[string]*ormapi.Organization, error) {
	orgsT := make(map[string]*ormapi.Organization)
	orgs := []ormapi.Organization{}

	db := loggedDB(ctx)
	err := db.Find(&orgs).Error
	if err != nil {
		return orgsT, err
	}
	for ii, _ := range orgs {
		orgsT[orgs[ii].Name] = &orgs[ii]
	}
	return orgsT, err
}

func getOrgType(orgName string, allOrgs map[string]*ormapi.Organization) string {
	if allOrgs != nil {
		if org, ok := allOrgs[orgName]; ok {
			return org.Type
		}
	}
	return ""
}

func orgExists(ctx context.Context, orgName string) (*ormapi.Organization, error) {
	lookup := ormapi.Organization{
		Name: orgName,
	}
	db := loggedDB(ctx)
	org := ormapi.Organization{}
	res := db.Where(&lookup).First(&org)
	if res.RecordNotFound() {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// SQL lookup by org name is case-insensitive.
	// Make sure org name matches (case-sensitive).
	if org.Name != orgName {
		return nil, fmt.Errorf("lookup %s but found %s", orgName, org.Name)
	}
	return &org, nil
}

// Marking an org for delete must be done transactionally so other threads
// cannot accidentally run the delete in parallel.
func markOrgForDelete(db *gorm.DB, name string, mark bool) (reterr error) {
	tx := db.Begin()
	defer func() {
		if reterr != nil {
			tx.Rollback()
		}
	}()
	// lookup org
	lookup := ormapi.Organization{
		Name: name,
	}
	findOrg := ormapi.Organization{}
	res := tx.Where(&lookup).First(&findOrg)
	if res.RecordNotFound() {
		return echo.NewHTTPError(http.StatusBadRequest, "org not found")
	}
	if res.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, res.Error.Error())
	}
	if mark {
		if findOrg.DeleteInProgress {
			return echo.NewHTTPError(http.StatusBadRequest, "org already being deleted")
		}
		findOrg.DeleteInProgress = true
	} else {
		findOrg.DeleteInProgress = false
	}
	err := tx.Save(&findOrg).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return tx.Commit().Error
}

func orgInUse(ctx context.Context, orgName string) error {
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

func orgInUseRegion(ctx context.Context, c ormapi.Controller, orgName string) error {
	conn, err := connectGrpcAddr(c.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	api := edgeproto.NewOrganizationApiClient(conn)
	org := edgeproto.Organization{
		Name: orgName,
	}
	res, err := api.OrganizationInUse(ctx, &org)
	if err != nil {
		return err
	}
	if res.Code == 0 {
		return nil
	}
	return fmt.Errorf(res.Message)
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

// TODO: move everything to its own file
var BillingOrgTypeSelf = "self"
var BillingOrgTypeParent = "parent"

// checks to see if the user entered in billing information and if they did, that it is all there
func checkBillingInfo(orgInfo *ormapi.OrgInfo) (bool, error) {
	if orgInfo.Type == OrgTypeOperator {
		return false, nil
	}
	fn := orgInfo.FirstName
	ln := orgInfo.LastName
	e := orgInfo.Email
	a := orgInfo.Address
	cy := orgInfo.City
	c := orgInfo.Country
	s := orgInfo.State
	p := orgInfo.PostalCode
	if fn == "" && ln == "" && e == "" && a == "" && cy == "" && c == "" && s == "" && p == "" {
		return false, nil
		// at least one of them was not blank, make sure none of them are
	} else if fn == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include FirstName")
	} else if ln == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include LastName")
	} else if e == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include email")
	} else if a == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include address")
	} else if cy == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include city")
	} else if c == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include country")
	} else if s == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include state")
	} else if p == "" {
		return false, fmt.Errorf("If sepcifying billing information, must include postalCode")
	} else {
		return true, nil
	}
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

func createBillingOrgObj(ctx context.Context, claims *UserClaims, org *ormapi.BillingOrganization) error {
	err := ValidName(org.Name)
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

	// TODO: set user to admin role of organization
	return nil
}

func CreateBillingOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	bOrg := ormapi.BillingOrganization{}
	if err := c.Bind(&bOrg); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("org", orgInfo.Name)

	err = CreateOrgObj(ctx, claims, &orgInfo)
	return setReply(c, err, Msg("Organization created"))
}

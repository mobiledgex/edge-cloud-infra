package orm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/federation"
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func sendFederationRequest(method, fedAddr, endpoint string, reqData, replyData interface{}) error {
	restClient := &ormclient.Client{}
	if unitTest {
		restClient.ForceDefaultTransport = true
	}
	requestUrl := fmt.Sprintf("http://%s%s", fedAddr, endpoint)
	status, err := restClient.HttpJsonSend(method, requestUrl, "", reqData, replyData)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("Failed to get response for %s request to URL %s, status=%s", method, requestUrl, http.StatusText(status))
	}
	return nil
}

func GetFederator(ctx context.Context, fedType, operatorId, countryCode string) (*ormapi.Federator, error) {
	if fedType == "" {
		return nil, fmt.Errorf("Missing federation type")
	}
	if err := fedcommon.IsValidFederationType(fedType); err != nil {
		return nil, err
	}
	label := "self"
	if fedType == fedcommon.TypePartner {
		label = "partner"
	}
	if operatorId == "" {
		return nil, fmt.Errorf("Missing %s operator ID", label)
	}
	if countryCode == "" {
		return nil, fmt.Errorf("Missing %s country code", label)
	}
	// get self federation information
	db := loggedDB(ctx)
	fedObj := ormapi.Federator{}
	lookup := ormapi.Federator{
		OperatorId:  operatorId,
		CountryCode: countryCode,
		Type:        fedType,
	}
	res := db.Where(&lookup).First(&fedObj)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Federation (%s) with operator ID %q and country code %q doesn't exist", label, operatorId, countryCode)
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	return &fedObj, nil
}

// Create self federator for an operator belonging to a set of regions labelled by a country code
func CreateSelfFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing Operator ID")
	}
	if opFed.CountryCode == "" {
		return fmt.Errorf("Missing country code")
	}
	if opFed.MCC == "" {
		return fmt.Errorf("Missing MCC")
	}
	if len(opFed.MNCs) == 0 {
		return fmt.Errorf("Missing MNCs. Please specify one or more MNCs")
	}
	if len(opFed.Regions) == 0 {
		return fmt.Errorf("Missing regions. Please specify one or more regions")
	}
	if err := fedcommon.ValidateCountryCode(opFed.CountryCode); err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// ensure that valid regions are passed
	for _, region := range opFed.Regions {
		_, err = getControllerObj(ctx, region)
		if err != nil {
			return fmt.Errorf("Invalid region specified: %s, %v", region, err)
		}
	}
	// ensure that operator ID is a valid operator org
	org, err := orgExists(ctx, opFed.OperatorId)
	if err != nil {
		return fmt.Errorf("Invalid operator ID specified")
	}
	if org.Type != OrgTypeOperator {
		return fmt.Errorf("Invalid operator ID, must be a valid operator org")
	}

	db := loggedDB(ctx)
	fedKey := uuid.New().String()
	fedStore := ormapi.Federator{}
	fedStore.FederationId = fedKey
	fedStore.FederationAddr = serverConfig.FederationAddr
	fedStore.Type = fedcommon.TypeSelf
	fedStore.OperatorId = opFed.OperatorId
	fedStore.CountryCode = opFed.CountryCode
	fedStore.Regions = strings.Join(opFed.Regions, fedcommon.Delimiter)
	fedStore.MCC = opFed.MCC
	fedStore.MNCs = strings.Join(opFed.MNCs, fedcommon.Delimiter)
	fedStore.LocatorEndPoint = opFed.LocatorEndPoint
	if err := db.Create(&fedStore).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			// UUID collision
			return fmt.Errorf("Federation ID collision for operator ID %s, country code %s. Please retry again", opFed.OperatorId, opFed.CountryCode)
		}
		return ormutil.DbErr(err)
	}

	opFedOut := ormapi.Federator{
		FederationId: fedKey,
	}
	return c.JSON(http.StatusOK, &opFedOut)
}

// Update self federator attributes and notify associated
// partner federators who have access to self federator zones
func UpdateSelfFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}
	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)

	update := false
	if opFed.MCC != selfFed.MCC {
		update = true
		selfFed.MCC = opFed.MCC
	}
	curMNCs := strings.Split(selfFed.MNCs, ",")
	if len(curMNCs) != len(opFed.MNCs) {
		update = true
		selfFed.MNCs = strings.Join(opFed.MNCs, fedcommon.Delimiter)
	} else {
		newMNCsMap := make(map[string]struct{})
		for _, nm := range opFed.MNCs {
			newMNCsMap[nm] = struct{}{}
		}
		for _, cm := range curMNCs {
			if _, ok := newMNCsMap[cm]; !ok {
				update = true
				selfFed.MNCs = strings.Join(opFed.MNCs, fedcommon.Delimiter)
				break
			}
		}
	}
	if opFed.LocatorEndPoint != selfFed.LocatorEndPoint {
		update = true
		selfFed.LocatorEndPoint = opFed.LocatorEndPoint
	}
	curRegs := strings.Split(selfFed.Regions, ",")
	newRegsMap := make(map[string]struct{})
	for _, nr := range opFed.Regions {
		newRegsMap[nr] = struct{}{}
	}
	for _, cr := range curRegs {
		if _, ok := newRegsMap[cr]; !ok {
			return fmt.Errorf("Cannot delete region %q. Only new regions can be added", cr)
		}
	}
	if len(opFed.Regions) != len(curRegs) {
		selfFed.Regions = strings.Join(opFed.Regions, fedcommon.Delimiter)
		update = true
	}
	if !update {
		return fmt.Errorf("nothing to update")
	}
	err = db.Save(selfFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	lookup := ormapi.FederatorRole{
		SelfFederationId: selfFed.FederationId,
	}
	partnerFederatorRoles := []ormapi.FederatorRole{}
	res := db.Where(&lookup).Find(&partnerFederatorRoles)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	for _, partnerRole := range partnerFederatorRoles {
		// Notify all the partner federators who have access to self zones
		if !fedcommon.ValueExistsInDelimitedList(partnerRole.Role, fedcommon.RoleShareZonesWithPartner) {
			continue
		}
		// get partner federator information
		db := loggedDB(ctx)
		partnerFed := ormapi.Federator{
			FederationId: partnerRole.PartnerFederationId,
		}
		res := db.Where(&partnerFed).First(&partnerFed)
		if res.RecordNotFound() {
			// this should not happen
			continue
		}
		if res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		// Notify partner federator about the update
		opConf := federation.UpdateMECNetConf{
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerFed.FederationId,
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			MCC:              selfFed.MCC,
			MNC:              strings.Split(selfFed.MNCs, ","),
			LocatorEndPoint:  selfFed.LocatorEndPoint,
		}
		err = sendFederationRequest("PUT", partnerFed.FederationAddr, federation.OperatorPartnerAPI, &opConf, nil)
		if err != nil {
			return err
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Updated self federator attributes successfully"))
}

// Delete self federator, given that there are no more
// partner federators associated with it
func DeleteSelfFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}
	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)

	lookup := ormapi.FederatorRole{
		SelfFederationId: selfFed.FederationId,
	}
	partnerFederatorRoles := []ormapi.FederatorRole{}
	res := db.Where(&lookup).Find(&partnerFederatorRoles)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if len(partnerFederatorRoles) > 0 {
		return fmt.Errorf("Self federator is associated with multiple partner federators. Please delete all those associations before deleting the federator")
	}
	// Ensure that no zone exists for this federator
	zoneLookup := ormapi.FederatorZone{
		FederationId: selfFed.FederationId,
	}
	selfZones := []ormapi.FederatorZone{}
	res = db.Where(&zoneLookup).Find(&selfZones)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if len(selfZones) > 0 {
		// This will ensure that no zones are used by any developer or partner federators
		return fmt.Errorf("Please delete all the associated zones before deleting the federator")
	}
	if err := db.Delete(&selfFed).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted self federator successfully"))
}

func ShowSelfFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}

	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	outFeds := []ormapi.FederatorRequest{}

	db := loggedDB(ctx)
	feds := []ormapi.Federator{}
	err = db.Where(&opFed).Find(&feds).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, fed := range feds {
		// Do not display federation ID
		outFed := ormapi.FederatorRequest{}
		outFed.FederationAddr = fed.FederationAddr
		outFed.Type = fed.Type
		outFed.OperatorId = fed.OperatorId
		outFed.CountryCode = fed.CountryCode
		outFed.Regions = strings.Split(fed.Regions, fedcommon.Delimiter)
		outFed.MCC = fed.MCC
		outFed.MNCs = strings.Split(fed.MNCs, fedcommon.Delimiter)
		outFed.LocatorEndPoint = fed.LocatorEndPoint
		outFeds = append(outFeds, outFed)
	}
	return c.JSON(http.StatusOK, outFeds)
}

// A self federator will add a partner federator. This gives self
// federator access to all the zones of the partner federator
func AddPartnerFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorPartnerRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opFed.SelfOperatorId == "" {
		return fmt.Errorf("Missing self operator ID")
	}
	if opFed.SelfCountryCode == "" {
		return fmt.Errorf("Missing self country code")
	}
	if opFed.PartnerOperatorId == "" {
		return fmt.Errorf("Missing partner operator ID")
	}
	if opFed.PartnerCountryCode == "" {
		return fmt.Errorf("Missing partner country code")
	}
	if opFed.PartnerFederationId == "" {
		return fmt.Errorf("Missing partner federation ID")
	}
	if opFed.PartnerFederationAddr == "" {
		return fmt.Errorf("Missing partner federation access address")
	}

	if err := authorized(ctx, claims.Username, opFed.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	// get federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	// call REST API /operator/partner
	opRegReq := federation.OperatorRegistrationRequest{
		OrigFederationId:   selfFed.FederationId,
		DestFederationId:   opFed.PartnerFederationId,
		OperatorId:         selfFed.OperatorId,
		CountryCode:        selfFed.CountryCode,
		OrigFederationAddr: selfFed.FederationAddr,
	}
	opRegRes := federation.OperatorRegistrationResponse{}
	err = sendFederationRequest("POST", opFed.PartnerFederationAddr, federation.OperatorPartnerAPI, &opRegReq, &opRegRes)
	if err != nil {
		return err
	}
	partnerFed := ormapi.Federator{
		FederationId:    opFed.PartnerFederationId,
		FederationAddr:  opFed.PartnerFederationAddr,
		OperatorId:      opFed.PartnerOperatorId,
		CountryCode:     opFed.PartnerCountryCode,
		Type:            fedcommon.TypePartner,
		MCC:             opRegRes.MCC,
		MNCs:            strings.Join(opRegRes.MNC, ","),
		LocatorEndPoint: opRegRes.LocatorEndPoint,
	}
	db := loggedDB(ctx)
	if err := db.Create(&partnerFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s",
				opFed.PartnerOperatorId, opFed.PartnerCountryCode)
		}
		return ormutil.DbErr(err)
	}

	// Store partner zones in DB
	for _, partnerZone := range opRegRes.PartnerZone {
		zoneObj := ormapi.FederatorZone{}
		zoneObj.FederationId = opRegRes.OrigFederationId
		zoneObj.ZoneId = partnerZone.ZoneId
		zoneObj.GeoLocation = partnerZone.GeoLocation
		zoneObj.City = partnerZone.City
		zoneObj.Locality = partnerZone.Locality
		if err := db.Create(&zoneObj).Error; err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
				return fmt.Errorf("Zone Id %q already exists", zoneObj.ZoneId)
			}
			return ormutil.DbErr(err)
		}
	}

	// Store partner federator role
	err = fedcommon.AddOrUpdatePartnerFederatorRole(db, selfFed, &partnerFed, fedcommon.RoleAccessPartnerZones)
	if err != nil {
		return err
	}

	return ormutil.SetReply(c, ormutil.Msg("Added partner federator successfully"))
}

func RemovePartnerFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorPartnerRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opFed.SelfOperatorId == "" {
		return fmt.Errorf("Missing self operator ID")
	}
	if opFed.SelfCountryCode == "" {
		return fmt.Errorf("Missing self country code")
	}
	if opFed.PartnerOperatorId == "" {
		return fmt.Errorf("Missing partner operator ID")
	}
	if opFed.PartnerCountryCode == "" {
		return fmt.Errorf("Missing partner country code")
	}

	if err := authorized(ctx, claims.Username, opFed.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	// get partner federator information
	partnerFed, err := GetFederator(ctx, fedcommon.TypePartner, opFed.PartnerOperatorId, opFed.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Check if all the partner zones are unused before deleting the partner federator
	lookup := ormapi.FederatorZone{
		FederationId: partnerFed.FederationId,
	}
	partnerZones := []ormapi.FederatorZone{}
	db := loggedDB(ctx)
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, pZone := range partnerZones {
		regLookup := ormapi.FederatorRegisteredZone{
			ZoneId:       pZone.ZoneId,
			FederationId: selfFed.FederationId,
		}
		regZone := ormapi.FederatorRegisteredZone{}
		res := db.Where(&regLookup).First(&regZone)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		if regZone.ZoneId != "" {
			return fmt.Errorf("Cannot remove partner federator as partner zone %q is registered locally. Please deregister it before removing the federation partner", regZone.ZoneId)
		}
	}

	// call REST API /operator/partner
	opFedReq := federation.FederationRequest{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: partnerFed.FederationId,
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
	}
	err = sendFederationRequest("DELETE", partnerFed.FederationAddr, federation.OperatorPartnerAPI, &opFedReq, nil)
	if err != nil {
		return err
	}

	// Delete all the local copy of partner OP zones
	for _, pZone := range partnerZones {
		if err := db.Delete(pZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
		regZone := ormapi.FederatorRegisteredZone{
			ZoneId: pZone.ZoneId,
		}
		if err := db.Delete(regZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
		shZone := ormapi.FederatorSharedZone{
			ZoneId: pZone.ZoneId,
		}
		if err := db.Delete(shZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
	}

	// Delete partner OP
	if err := db.Delete(&partnerFed).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Removed partner federator successfully"))
}

func ShowPartnerFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}

	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	// get list of all partner federators for the self federator
	db := loggedDB(ctx)
	lookup := ormapi.FederatorRole{
		SelfFederationId: selfFed.FederationId,
	}
	fedRoles := []ormapi.FederatorRole{}
	err = db.Where(&lookup).Find(&fedRoles).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	outFeds := []ormapi.FederatorRequest{}
	for _, fedRole := range fedRoles {
		fedLookup := ormapi.Federator{
			FederationId: fedRole.PartnerFederationId,
			Type:         fedcommon.TypePartner,
		}
		partnerFed := ormapi.Federator{}
		err = db.Where(&fedLookup).First(&partnerFed).Error
		if err != nil {
			return ormutil.DbErr(err)
		}
		// Do not display federation ID
		outFed := ormapi.FederatorRequest{}
		outFed.FederationAddr = partnerFed.FederationAddr
		outFed.Type = partnerFed.Type
		outFed.OperatorId = partnerFed.OperatorId
		outFed.CountryCode = partnerFed.CountryCode
		outFed.MCC = partnerFed.MCC
		outFed.MNCs = strings.Split(partnerFed.MNCs, fedcommon.Delimiter)
		outFed.LocatorEndPoint = partnerFed.LocatorEndPoint
		outFeds = append(outFeds, outFed)
	}
	return c.JSON(http.StatusOK, outFeds)
}

func ShowPartnerFederatorRole(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.Federator{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}

	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	// validate self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.FederatorRole{
		SelfFederationId: selfFed.FederationId,
	}
	fedRoles := []ormapi.FederatorRole{}
	err = db.Where(&lookup).Find(&fedRoles).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	out := []ormapi.FederatorRoleResponse{}
	for _, fedRole := range fedRoles {
		partnerFed := ormapi.Federator{
			FederationId: fedRole.PartnerFederationId,
		}
		res := db.Where(&partnerFed).First(&partnerFed)
		if res.RecordNotFound() {
			// this should not happen
			continue
		}
		if res.Error != nil {
			return ormutil.DbErr(err)
		}
		// Do not display federation ID
		resp := ormapi.FederatorRoleResponse{
			PartnerOperatorId:  partnerFed.OperatorId,
			PartnerCountryCode: partnerFed.CountryCode,
			PartnerRole:        fedRole.Role,
		}
		out = append(out, resp)
	}
	return c.JSON(http.StatusOK, out)
}

func CreateSelfFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZone := ormapi.FederatorZoneDetails{}
	if err := c.Bind(&opZone); err != nil {
		return ormutil.BindErr(err)
	}
	//sanity check
	if opZone.ZoneId == "" {
		return fmt.Errorf("Missing zone ID")
	}
	if opZone.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}
	if opZone.Region == "" {
		return fmt.Errorf("Missing region")
	}
	if len(opZone.Cloudlets) == 0 {
		return fmt.Errorf("Missing cloudlets")
	}
	if len(opZone.Cloudlets) > 1 {
		return fmt.Errorf("Only one cloudlet supported for now")
	}
	if err := fedcommon.ValidateGeoLocation(opZone.GeoLocation); err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, opZone.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federation information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opZone.OperatorId, opZone.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		FederationId: selfFed.FederationId,
		ZoneId:       opZone.ZoneId,
	}
	existingFed := ormapi.FederatorZone{}
	res := db.Where(&lookup).First(&existingFed)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if existingFed.ZoneId != "" {
		return fmt.Errorf("Zone %q already exists", opZone.ZoneId)
	}

	rc := ormutil.RegionContext{
		Region:    opZone.CountryCode,
		Username:  claims.Username,
		SkipAuthz: true,
		Database:  database,
	}
	cloudletMap := make(map[string]edgeproto.Cloudlet)
	cloudletLookup := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: selfFed.OperatorId,
		},
	}
	err = ctrlclient.ShowCloudletStream(ctx, &rc, &cloudletLookup, connCache, nil, func(cloudlet *edgeproto.Cloudlet) error {
		cloudletMap[cloudlet.Key.Name] = *cloudlet
		return nil
	})
	if err != nil {
		return err
	}
	for _, cl := range opZone.Cloudlets {
		if _, ok := cloudletMap[cl]; !ok {
			return fmt.Errorf("Cloudlet %q doesn't exist", cl)
		}
	}

	az := ormapi.FederatorZone{}
	az.FederationId = selfFed.FederationId
	az.ZoneId = opZone.ZoneId
	az.GeoLocation = opZone.GeoLocation
	az.State = opZone.State
	az.Locality = opZone.Locality
	az.Region = opZone.Region
	az.Cloudlets = strings.Join(opZone.Cloudlets, fedcommon.Delimiter)
	if err := db.Create(&az).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone with same zone ID %q already exists for operator ID %s, country code %s", az.ZoneId, selfFed.OperatorId, selfFed.CountryCode)
		}
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Created zone successfully"))
}

func DeleteSelfFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZone := ormapi.FederatorZoneDetails{}
	if err := c.Bind(&opZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opZone.ZoneId == "" {
		return fmt.Errorf("Missing zone ID")
	}
	if opZone.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}
	if err := authorized(ctx, claims.Username, opZone.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opZone.OperatorId, opZone.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:       opZone.ZoneId,
		FederationId: selfFed.FederationId,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s does not exist", opZone.ZoneId)
	}

	shLookup := ormapi.FederatorSharedZone{
		ZoneId: opZone.ZoneId,
	}
	shZone := ormapi.FederatorSharedZone{}
	res := db.Where(&shLookup).First(&shZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if shZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %q as it shared with partner federator with operator ID %q and country code %q. Please unshare it before deleting it", shZone.ZoneId, shZone.OperatorId, shZone.CountryCode)
	}

	regLookup := ormapi.FederatorRegisteredZone{
		ZoneId: opZone.ZoneId,
	}
	regZone := ormapi.FederatorRegisteredZone{}
	res = db.Where(&regLookup).First(&regZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %q as it registered by partner federator with operator ID %q and country code %q. Please deregister it before deleting it", regZone.ZoneId, regZone.OperatorId, regZone.CountryCode)
	}

	if err := db.Delete(&existingZone).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted federator zone successfully"))
}

func ShowFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZoneReq := ormapi.FederatorZoneDetails{}
	if err := c.Bind(&opZoneReq); err != nil {
		return ormutil.BindErr(err)
	}
	if opZoneReq.OperatorId == "" {
		return fmt.Errorf("Missing operator ID")
	}
	if err := authorized(ctx, claims.Username, opZoneReq.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, opZoneReq.OperatorId, opZoneReq.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	opZones := []ormapi.FederatorZone{}
	lookup := ormapi.FederatorZone{
		FederationId: selfFed.FederationId,
		ZoneId:       opZoneReq.ZoneId,
	}
	err = db.Where(&lookup).Find(&opZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	fedZones := []ormapi.FederatorZoneDetails{}
	for _, opZone := range opZones {
		opRegZones := []ormapi.FederatorRegisteredZone{}
		regLookup := ormapi.FederatorRegisteredZone{
			ZoneId: opZoneReq.ZoneId,
		}
		res := db.Where(&regLookup).Find(&opRegZones)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}

		opShZones := []ormapi.FederatorSharedZone{}
		shLookup := ormapi.FederatorSharedZone{
			ZoneId: opZoneReq.ZoneId,
		}
		res = db.Where(&shLookup).Find(&opShZones)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		zoneOut := ormapi.FederatorZoneDetails{}
		zoneOut.ZoneId = opZone.ZoneId
		zoneOut.GeoLocation = opZone.GeoLocation
		zoneOut.City = opZone.City
		zoneOut.State = opZone.State
		zoneOut.Locality = opZone.Locality
		zoneOut.Cloudlets = strings.Split(opZone.Cloudlets, fedcommon.Delimiter)
		for _, opRegZone := range opRegZones {
			regZone := fmt.Sprintf("%s/%s", opRegZone.OperatorId, opRegZone.CountryCode)
			zoneOut.RegisteredByOPs = append(zoneOut.RegisteredByOPs, regZone)
		}
		for _, opShZone := range opShZones {
			shZone := fmt.Sprintf("%s/%s", opShZone.OperatorId, opShZone.CountryCode)
			zoneOut.SharedWithOPs = append(zoneOut.SharedWithOPs, shZone)
		}

		fedZones = append(fedZones, zoneOut)
	}
	return c.JSON(http.StatusOK, fedZones)
}

func ShareSelfFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	shZone := ormapi.FederatorZoneShare{}
	if err := c.Bind(&shZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if shZone.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be shared")
	}
	if shZone.SelfOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator whose zone is to be shared")
	}
	if shZone.SelfCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator whose zone is to be shared")
	}
	if shZone.PartnerOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator with whom the zone is to be shared")
	}
	if shZone.PartnerCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator with whom the zone is to be shared")
	}
	if err := authorized(ctx, claims.Username, shZone.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, shZone.SelfOperatorId, shZone.SelfCountryCode)
	if err != nil {
		return err
	}
	// get partner federator information
	partnerFed, err := GetFederator(ctx, fedcommon.TypePartner, shZone.PartnerOperatorId, shZone.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Only share with those partner federators who are permitted to access our zones
	db := loggedDB(ctx)
	roleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerFederatorRole := ormapi.FederatorRole{}
	res := db.Where(&roleLookup).Find(&partnerFederatorRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if !fedcommon.ValueExistsInDelimitedList(partnerFederatorRole.Role, fedcommon.RoleShareZonesWithPartner) {
		return fmt.Errorf("Federator with operator ID %q and country code %q is not allowed to access our zones",
			partnerFed.OperatorId, partnerFed.CountryCode)
	}

	// Check if zone exists
	lookup := ormapi.FederatorZone{
		ZoneId:       shZone.ZoneId,
		FederationId: selfFed.FederationId,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", shZone.ZoneId)
	}

	// Notify federated partner about new zone
	opZoneShare := federation.NotifyPartnerOperatorZone{
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
		OrigFederationId: selfFed.FederationId,
		DestFederationId: partnerFed.FederationId,
		PartnerZone: federation.ZoneInfo{
			ZoneId:      existingZone.ZoneId,
			GeoLocation: existingZone.GeoLocation,
			City:        existingZone.City,
			State:       existingZone.State,
			Locality:    existingZone.Locality,
			EdgeCount:   len(existingZone.Cloudlets),
		},
	}
	err = sendFederationRequest("POST", partnerFed.FederationAddr, federation.OperatorNotifyZoneAPI, &opZoneShare, nil)
	if err != nil {
		return err
	}

	// Mark zone as shared in DB
	shareZone := ormapi.FederatorSharedZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: partnerFed.FederationId,
		OperatorId:   partnerFed.OperatorId,
		CountryCode:  partnerFed.CountryCode,
	}
	if err := db.Create(&shareZone).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Zone %q shared with partner federator(s) successfully", shareZone.ZoneId)))
}

func UnshareSelfFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	unshZone := ormapi.FederatorZoneShare{}
	if err := c.Bind(&unshZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if unshZone.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be unshared")
	}
	if unshZone.SelfOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator whose zone is to be unshared")
	}
	if unshZone.SelfCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator whose zone is to be unshared")
	}
	if unshZone.PartnerOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator with whom the zone is to be unshared")
	}
	if unshZone.PartnerCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator with whom the zone is to be unshared")
	}
	if err := authorized(ctx, claims.Username, unshZone.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, unshZone.SelfOperatorId, unshZone.SelfCountryCode)
	if err != nil {
		return err
	}
	// get partner federator information
	partnerFed, err := GetFederator(ctx, fedcommon.TypePartner, unshZone.PartnerOperatorId, unshZone.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Only unshare with those partner federators who are permitted to access our zones
	db := loggedDB(ctx)
	roleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerFederatorRole := ormapi.FederatorRole{}
	res := db.Where(&roleLookup).Find(&partnerFederatorRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if !fedcommon.ValueExistsInDelimitedList(partnerFederatorRole.Role, fedcommon.RoleShareZonesWithPartner) {
		return fmt.Errorf("Federator with operator ID %q and country code %q is not allowed to access our zones",
			partnerFed.OperatorId, partnerFed.CountryCode)
	}

	// Check if zone exists
	lookup := ormapi.FederatorZone{
		ZoneId:       unshZone.ZoneId,
		FederationId: selfFed.FederationId,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", unshZone.ZoneId)
	}

	// Notify federated partner about deleted zone
	opZoneUnShare := federation.ZoneRequest{
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
		OrigFederationId: selfFed.FederationId,
		DestFederationId: partnerFed.FederationId,
		Zone:             existingZone.ZoneId,
	}
	err = sendFederationRequest("DELETE", partnerFed.FederationAddr, federation.OperatorNotifyZoneAPI, &opZoneUnShare, nil)
	if err != nil {
		return err
	}

	// Delete zone from shared list in DB
	unshareZone := ormapi.FederatorSharedZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: partnerFed.FederationId,
		OperatorId:   partnerFed.OperatorId,
		CountryCode:  partnerFed.CountryCode,
	}
	if err := db.Delete(&unshareZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {

			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Zone %s unshared from partner federator successfully", unshareZone.ZoneId)))
}

func RegisterPartnerFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.FederatorZoneRegister{}
	if err := c.Bind(&reg); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if reg.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be registered")
	}
	if reg.SelfOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator who wants to register a partner zone")
	}
	if reg.SelfCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator who wants to register a partner zone")
	}
	if reg.PartnerOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator whose zone is to be registered")
	}
	if reg.PartnerCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator whose zone is to be registered")
	}
	if err := authorized(ctx, claims.Username, reg.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, reg.SelfOperatorId, reg.SelfCountryCode)
	if err != nil {
		return err
	}
	// get partner federator information
	partnerFed, err := GetFederator(ctx, fedcommon.TypePartner, reg.PartnerOperatorId, reg.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Only register with those partner federators whose zones can be accessed by self federator
	db := loggedDB(ctx)
	roleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerFederatorRole := ormapi.FederatorRole{}
	res := db.Where(&roleLookup).Find(&partnerFederatorRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if !fedcommon.ValueExistsInDelimitedList(partnerFederatorRole.Role, fedcommon.RoleAccessPartnerZones) {
		return fmt.Errorf("Cannot access zones of partner federator with operator ID %q and country code %q",
			partnerFed.OperatorId, partnerFed.CountryCode)
	}

	// Check if zone exists
	lookup := ormapi.FederatorZone{
		ZoneId:       reg.ZoneId,
		FederationId: partnerFed.FederationId,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", reg.ZoneId)
	}

	// Notify federated partner about zone registration
	opZoneReg := federation.OperatorZoneRegister{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: partnerFed.FederationId,
		Operator:         partnerFed.OperatorId,
		Country:          partnerFed.CountryCode,
		Zones:            []string{existingZone.ZoneId},
	}
	opZoneRes := federation.OperatorZoneRegisterResponse{}
	err = sendFederationRequest("POST", partnerFed.FederationAddr, federation.OperatorZoneAPI, &opZoneReg, &opZoneRes)
	if err != nil {
		return err
	}

	// Mark zone as registered in DB
	regZone := ormapi.FederatorRegisteredZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: partnerFed.FederationId,
		OperatorId:   partnerFed.OperatorId,
		CountryCode:  partnerFed.CountryCode,
	}
	if err := db.Create(&regZone).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Partner federator zone %q registered successfully", regZone.ZoneId)))
}

func DeregisterPartnerFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.FederatorZoneRegister{}
	if err := c.Bind(&reg); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if reg.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be deregistered")
	}
	if reg.SelfOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator who wants to deregister a partner zone")
	}
	if reg.SelfCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator who wants to deregister a partner zone")
	}
	if reg.PartnerOperatorId == "" {
		return fmt.Errorf("Must specify the operator ID of the federator whose zone is to be deregistered")
	}
	if reg.PartnerCountryCode == "" {
		return fmt.Errorf("Must specify the country code of the federator whose zone is to be deregistered")
	}
	if err := authorized(ctx, claims.Username, reg.SelfOperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetFederator(ctx, fedcommon.TypeSelf, reg.SelfOperatorId, reg.SelfCountryCode)
	if err != nil {
		return err
	}
	// get partner federator information
	partnerFed, err := GetFederator(ctx, fedcommon.TypePartner, reg.PartnerOperatorId, reg.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Only deregister with those partner federators whose zones can be accessed by self federator
	db := loggedDB(ctx)
	roleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerFederatorRole := ormapi.FederatorRole{}
	res := db.Where(&roleLookup).Find(&partnerFederatorRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(err)
	}
	if !fedcommon.ValueExistsInDelimitedList(partnerFederatorRole.Role, fedcommon.RoleAccessPartnerZones) {
		return fmt.Errorf("Cannot access zones of partner federator with operator ID %q and country code %q",
			partnerFed.OperatorId, partnerFed.CountryCode)
	}

	// Check if zone exists
	lookup := ormapi.FederatorZone{
		ZoneId:       reg.ZoneId,
		FederationId: partnerFed.FederationId,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", reg.ZoneId)
	}

	// Notify federated partner about zone deregistration
	opZoneReg := federation.ZoneRequest{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: partnerFed.FederationId,
		Operator:         partnerFed.OperatorId,
		Country:          partnerFed.CountryCode,
		Zone:             existingZone.ZoneId,
	}
	err = sendFederationRequest("DELETE", partnerFed.FederationAddr, federation.OperatorZoneAPI, &opZoneReg, nil)
	if err != nil {
		return err
	}

	// Mark zone as deregistered in DB
	deregZone := ormapi.FederatorRegisteredZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: partnerFed.FederationId,
		OperatorId:   partnerFed.OperatorId,
		CountryCode:  partnerFed.CountryCode,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {

			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Partner federator zone %q deregistered successfully", existingZone.ZoneId)))
}

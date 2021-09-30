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
	if fedAddr == "" {
		return fmt.Errorf("Missing partner federation address")
	}
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

func fedAuthorized(ctx context.Context, username, operatorId string) error {
	if operatorId == "" {
		return fmt.Errorf("Missing self operator ID %q", operatorId)
	}
	return authorized(ctx, username, operatorId, ResourceCloudlets, ActionManage)
}

func GetSelfFederator(ctx context.Context, operatorId, countryCode string) (*ormapi.Federator, error) {
	if operatorId == "" {
		return nil, fmt.Errorf("Missing self operator ID")
	}
	if countryCode == "" {
		return nil, fmt.Errorf("Missing self country code")
	}
	db := loggedDB(ctx)
	fedObj := ormapi.Federator{
		OperatorId:  operatorId,
		CountryCode: countryCode,
		Type:        fedcommon.TypeSelf,
	}
	res := db.Where(&fedObj).First(&fedObj)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federator %s does not exist", fedcommon.FederatorStr(operatorId, countryCode))
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	return &fedObj, nil
}

func GetPartnerFederator(ctx context.Context, partnerOperatorId, partnerCountryCode string) (*ormapi.Federator, error) {
	if partnerOperatorId == "" {
		return nil, fmt.Errorf("Missing partner operator ID %q", partnerOperatorId)
	}
	if partnerCountryCode == "" {
		return nil, fmt.Errorf("Missing partner country code %q", partnerCountryCode)
	}

	db := loggedDB(ctx)
	partnerLookup := ormapi.Federator{
		OperatorId:  partnerOperatorId,
		CountryCode: partnerCountryCode,
		Type:        fedcommon.TypePartner,
	}
	partnerFed := ormapi.Federator{}
	res := db.Where(&partnerLookup).First(&partnerFed)
	if !res.RecordNotFound() && res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Partner federator (%s) does not exist",
			fedcommon.FederatorStr(partnerOperatorId, partnerCountryCode))
	}

	return &partnerFed, nil
}

func GetPartnerFederatorWithRole(ctx context.Context, selfOperatorId, selfCountryCode, partnerOperatorId, partnerCountryCode string, role fedcommon.FederatorRole) (*ormapi.Federator, error) {
	partnerFed, err := GetPartnerFederator(ctx, partnerOperatorId, partnerCountryCode)
	if err != nil {
		return nil, err
	}

	if selfOperatorId == "" {
		return nil, fmt.Errorf("Missing self operator ID %q", selfOperatorId)
	}
	if selfCountryCode == "" {
		return nil, fmt.Errorf("Missing self country code %q", selfCountryCode)
	}

	db := loggedDB(ctx)
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     selfOperatorId,
		SelfCountryCode:    selfCountryCode,
		PartnerOperatorId:  partnerOperatorId,
		PartnerCountryCode: partnerCountryCode,
		PartnerRole:        role,
	}
	res := db.Where(&partnerFederation).First(&partnerFederation)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Partner federator (%s) with role %q does not exist",
			fedcommon.FederatorStr(partnerOperatorId, partnerCountryCode), role)
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	return partnerFed, nil
}

func GetPartnerFederatorsWithRole(ctx context.Context, selfOperatorId, selfCountryCode string, role fedcommon.FederatorRole) ([]ormapi.Federator, error) {
	db := loggedDB(ctx)
	lookup := ormapi.Federation{
		SelfOperatorId:  selfOperatorId,
		SelfCountryCode: selfCountryCode,
	}
	partnerFederations := []ormapi.Federation{}
	res := db.Where(&lookup).Find(&partnerFederations)
	if !res.RecordNotFound() && res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	outFeds := []ormapi.Federator{}
	for _, partnerFederation := range partnerFederations {
		if role != fedcommon.RoleAny && partnerFederation.PartnerRole != role {
			continue
		}
		partnerFed := ormapi.Federator{
			OperatorId:  partnerFederation.PartnerOperatorId,
			CountryCode: partnerFederation.PartnerCountryCode,
		}
		res := db.Where(&partnerFed).Find(&partnerFed)
		if res.RecordNotFound() {
			// this should not happen
			continue
		}
		if res.Error != nil {
			return nil, ormutil.DbErr(res.Error)
		}
		outFeds = append(outFeds, partnerFed)
	}
	return outFeds, nil
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
	if err := fedAuthorized(ctx, claims.Username, opFed.OperatorId); err != nil {
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
	fedStore.FederationKey = fedKey
	fedStore.FederationAddr = serverConfig.FederationAddr
	fedStore.OperatorId = opFed.OperatorId
	fedStore.CountryCode = opFed.CountryCode
	fedStore.Type = fedcommon.TypeSelf
	fedStore.Regions = strings.Join(opFed.Regions, fedcommon.Delimiter)
	fedStore.MCC = opFed.MCC
	fedStore.MNCs = strings.Join(opFed.MNCs, fedcommon.Delimiter)
	fedStore.LocatorEndPoint = opFed.LocatorEndPoint
	if err := db.Create(&fedStore).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Self federator %s already exists",
				fedcommon.FederatorStr(opFed.OperatorId, opFed.CountryCode))
		}
		return ormutil.DbErr(err)
	}

	opFedOut := ormapi.Federator{
		FederationKey: fedKey,
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
	if err := fedAuthorized(ctx, claims.Username, opFed.OperatorId); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetSelfFederator(ctx, opFed.OperatorId, opFed.CountryCode)
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

	// Notify all the partner federators who have access to self zones
	partnerFeds, err := GetPartnerFederatorsWithRole(ctx, selfFed.OperatorId, selfFed.CountryCode, fedcommon.RoleAccessToSelfZones)
	if err != nil {
		return err
	}
	for _, partnerFed := range partnerFeds {
		opConf := federation.UpdateMECNetConf{
			OrigFederationId: selfFed.FederationKey,
			DestFederationId: partnerFed.FederationKey,
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
	if err := fedAuthorized(ctx, claims.Username, opFed.OperatorId); err != nil {
		return err
	}
	// get federator information
	selfFed, err := GetSelfFederator(ctx, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	partnerFeds, err := GetPartnerFederatorsWithRole(ctx, selfFed.OperatorId, selfFed.CountryCode, fedcommon.RoleAny)
	if err != nil {
		return err
	}
	if len(partnerFeds) > 0 {
		return fmt.Errorf("Self federator is associated with multiple partner federators. Please delete all those associations before deleting the federator")
	}
	// Ensure that no zone exists for this federator
	zoneLookup := ormapi.FederatorZone{
		OperatorId:  selfFed.OperatorId,
		CountryCode: selfFed.CountryCode,
	}
	selfZones := []ormapi.FederatorZone{}
	res := db.Where(&zoneLookup).Find(&selfZones)
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

// A self federator will add a partner federator. This is done as
// part of federation planning. This does not form federation with
// partner federator
func AddPartnerFederator(c echo.Context) error {
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
	if opFed.SelfOperatorId == "" {
		return fmt.Errorf("Missing self operator ID")
	}
	if opFed.SelfCountryCode == "" {
		return fmt.Errorf("Missing self country code")
	}
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing partner operator ID")
	}
	if opFed.CountryCode == "" {
		return fmt.Errorf("Missing partner country code")
	}
	if opFed.FederationKey == "" {
		return fmt.Errorf("Missing partner federation key")
	}
	if opFed.FederationAddr == "" {
		return fmt.Errorf("Missing partner federation access address")
	}

	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
		return err
	}

	// validate self federator
	_, err = GetSelfFederator(ctx, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	partnerFed := ormapi.Federator{
		OperatorId:      opFed.OperatorId,
		CountryCode:     opFed.CountryCode,
		Type:            fedcommon.TypePartner,
		FederationKey:   opFed.FederationKey,
		FederationAddr:  opFed.FederationAddr,
		MCC:             opFed.MCC,
		MNCs:            strings.Join(opFed.MNCs, fedcommon.Delimiter),
		LocatorEndPoint: opFed.LocatorEndPoint,
	}
	if err := db.Create(&partnerFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Partner federator (%s) already exists",
				fedcommon.FederatorStr(opFed.OperatorId, opFed.CountryCode))
		}
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Added partner federator successfully"))
}

func RemovePartnerFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
		return err
	}

	// validate self federator
	selfFed, err := GetSelfFederator(ctx, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	partnerFed, err := GetPartnerFederator(ctx, opFed.OperatorId, opFed.CountryCode)
	if err != nil {
		return err
	}

	// check if federation with partner federator exists
	db := loggedDB(ctx)
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     selfFed.OperatorId,
		SelfCountryCode:    selfFed.CountryCode,
		PartnerOperatorId:  partnerFed.OperatorId,
		PartnerCountryCode: partnerFed.CountryCode,
	}
	res := db.Where(&partnerFederation).First(nil)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}

	if !res.RecordNotFound() {
		return fmt.Errorf("Cannot delete partner federator (%s) "+
			"as it is part of federation",
			fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
	}

	// Delete partner federator
	if err := db.Delete(partnerFed).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Removed partner federator successfully"))
}

func ShowFederator(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}

	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
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
		outFed.Type = fed.Type
		outFed.FederationAddr = fed.FederationAddr
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
	// sanity check
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
	if err := fedAuthorized(ctx, claims.Username, opZone.OperatorId); err != nil {
		return err
	}
	// get self federation information
	selfFed, err := GetSelfFederator(ctx, opZone.OperatorId, opZone.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		OperatorId:  opZone.OperatorId,
		CountryCode: opZone.CountryCode,
		ZoneId:      opZone.ZoneId,
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
	az.OperatorId = selfFed.OperatorId
	az.CountryCode = selfFed.CountryCode
	az.ZoneId = opZone.ZoneId
	az.GeoLocation = opZone.GeoLocation
	az.State = opZone.State
	az.Locality = opZone.Locality
	az.Region = opZone.Region
	az.Cloudlets = strings.Join(opZone.Cloudlets, fedcommon.Delimiter)
	if err := db.Create(&az).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone with same zone ID %q already exists for federator (%s)",
				az.ZoneId, fedcommon.FederatorStr(selfFed.OperatorId, selfFed.CountryCode))
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
	if err := fedAuthorized(ctx, claims.Username, opZone.OperatorId); err != nil {
		return err
	}
	// get federator information
	_, err = GetSelfFederator(ctx, opZone.OperatorId, opZone.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:      opZone.ZoneId,
		OperatorId:  opZone.OperatorId,
		CountryCode: opZone.CountryCode,
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
		ZoneId:           opZone.ZoneId,
		OwnerOperatorId:  opZone.OperatorId,
		OwnerCountryCode: opZone.CountryCode,
	}
	shZone := ormapi.FederatorSharedZone{}
	res := db.Where(&shLookup).First(&shZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if shZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %q as it is shared with partner federator "+
			"(%s). Please unshare it before deleting it", shZone.ZoneId,
			fedcommon.FederatorStr(shZone.SharedWithOperatorId, shZone.SharedWithCountryCode))
	}

	regLookup := ormapi.FederatorRegisteredZone{
		ZoneId:           opZone.ZoneId,
		OwnerOperatorId:  opZone.OperatorId,
		OwnerCountryCode: opZone.CountryCode,
	}
	regZone := ormapi.FederatorRegisteredZone{}
	res = db.Where(&regLookup).First(&regZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %q as it registered by partner federator "+
			"(%s). Please deregister it before deleting it", regZone.ZoneId,
			fedcommon.FederatorStr(regZone.RegisteredByOperatorId, regZone.RegisteredByCountryCode))
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
	if err := fedAuthorized(ctx, claims.Username, opZoneReq.OperatorId); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetSelfFederator(ctx, opZoneReq.OperatorId, opZoneReq.CountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	opZones := []ormapi.FederatorZone{}
	lookup := ormapi.FederatorZone{
		OperatorId:  selfFed.OperatorId,
		CountryCode: selfFed.CountryCode,
		ZoneId:      opZoneReq.ZoneId,
	}
	err = db.Where(&lookup).Find(&opZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	fedZones := []ormapi.FederatorZoneDetails{}
	for _, opZone := range opZones {
		opRegZones := []ormapi.FederatorRegisteredZone{}
		regLookup := ormapi.FederatorRegisteredZone{
			ZoneId:           opZoneReq.ZoneId,
			OwnerOperatorId:  opZoneReq.OperatorId,
			OwnerCountryCode: opZoneReq.CountryCode,
		}
		res := db.Where(&regLookup).Find(&opRegZones)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}

		opShZones := []ormapi.FederatorSharedZone{}
		shLookup := ormapi.FederatorSharedZone{
			ZoneId:           opZoneReq.ZoneId,
			OwnerOperatorId:  opZoneReq.OperatorId,
			OwnerCountryCode: opZoneReq.CountryCode,
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
			regZone := fmt.Sprintf("%s/%s", opRegZone.RegisteredByOperatorId, opRegZone.RegisteredByCountryCode)
			zoneOut.RegisteredByFederators = append(zoneOut.RegisteredByFederators, regZone)
		}
		for _, opShZone := range opShZones {
			shZone := fmt.Sprintf("%s/%s", opShZone.SharedWithOperatorId, opShZone.SharedWithCountryCode)
			zoneOut.SharedWithFederators = append(zoneOut.SharedWithFederators, shZone)
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
	shZone := ormapi.FederatorZoneRequest{}
	if err := c.Bind(&shZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if shZone.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be shared")
	}
	if err := fedAuthorized(ctx, claims.Username, shZone.SelfOperatorId); err != nil {
		return err
	}

	// get self federator information
	selfFed, err := GetSelfFederator(ctx, shZone.SelfOperatorId, shZone.SelfCountryCode)
	if err != nil {
		return err
	}

	// get partner federator information
	partnerFed, err := GetPartnerFederator(ctx, shZone.PartnerOperatorId, shZone.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:      shZone.ZoneId,
		OperatorId:  selfFed.OperatorId,
		CountryCode: selfFed.CountryCode,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", shZone.ZoneId)
	}

	// Only share with those partner federators who are permitted to access our zones
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     shZone.SelfOperatorId,
		SelfCountryCode:    shZone.SelfCountryCode,
		PartnerOperatorId:  shZone.PartnerOperatorId,
		PartnerCountryCode: shZone.PartnerCountryCode,
		PartnerRole:        fedcommon.RoleAccessToSelfZones,
	}
	res := db.Where(&partnerFederation).First(nil)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	// Only share if federation exists with partner federator, else
	// just add it to the DB (federation planning)
	if res.Error == nil {
		opZoneShare := federation.NotifyPartnerOperatorZone{
			OrigFederationId: selfFed.FederationKey,
			DestFederationId: partnerFed.FederationKey,
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
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
	}

	// Mark zone as shared in DB
	shareZone := ormapi.FederatorSharedZone{
		ZoneId:                existingZone.ZoneId,
		OwnerOperatorId:       existingZone.OperatorId,
		OwnerCountryCode:      existingZone.CountryCode,
		SharedWithOperatorId:  partnerFed.OperatorId,
		SharedWithCountryCode: partnerFed.CountryCode,
	}
	if err := db.Create(&shareZone).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(
		fmt.Sprintf("Zone %q shared with partner federator (%s) successfully",
			shareZone.ZoneId, fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))))
}

func UnshareSelfFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	unshZone := ormapi.FederatorZoneRequest{}
	if err := c.Bind(&unshZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if unshZone.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be unshared")
	}
	if err := fedAuthorized(ctx, claims.Username, unshZone.SelfOperatorId); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetSelfFederator(ctx, unshZone.SelfOperatorId, unshZone.SelfCountryCode)
	if err != nil {
		return err
	}
	// get partner federator information
	partnerFed, err := GetPartnerFederator(ctx, unshZone.PartnerOperatorId, unshZone.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:      unshZone.ZoneId,
		OperatorId:  selfFed.OperatorId,
		CountryCode: selfFed.CountryCode,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", unshZone.ZoneId)
	}

	// Only unshare with those partner federators who are permitted to access our zones
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     unshZone.SelfOperatorId,
		SelfCountryCode:    unshZone.SelfCountryCode,
		PartnerOperatorId:  unshZone.PartnerOperatorId,
		PartnerCountryCode: unshZone.PartnerCountryCode,
		PartnerRole:        fedcommon.RoleAccessToSelfZones,
	}
	res := db.Where(&partnerFederation).First(nil)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	// Only unshare if federation exists with partner federator, else
	// just add it to the DB (federation planning)
	if res.Error == nil {
		// Notify federated partner about deleted zone
		opZoneUnShare := federation.ZoneRequest{
			OrigFederationId: selfFed.FederationKey,
			DestFederationId: partnerFed.FederationKey,
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			Zone:             existingZone.ZoneId,
		}
		err = sendFederationRequest("DELETE", partnerFed.FederationAddr, federation.OperatorNotifyZoneAPI, &opZoneUnShare, nil)
		if err != nil {
			return err
		}
	}

	// Delete zone from shared list in DB
	unshareZone := ormapi.FederatorSharedZone{
		ZoneId:                existingZone.ZoneId,
		OwnerOperatorId:       existingZone.OperatorId,
		OwnerCountryCode:      existingZone.CountryCode,
		SharedWithOperatorId:  partnerFed.OperatorId,
		SharedWithCountryCode: partnerFed.CountryCode,
	}
	if err := db.Delete(&unshareZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {

			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Zone %s unshared from partner federator (%s) successfully",
		unshareZone.ZoneId, fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))))
}

func RegisterPartnerFederatorZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.FederatorZoneRequest{}
	if err := c.Bind(&reg); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if reg.ZoneId == "" {
		return fmt.Errorf("Must specify the zone which is to be registered")
	}
	if err := fedAuthorized(ctx, claims.Username, reg.SelfOperatorId); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetSelfFederator(ctx, reg.SelfOperatorId, reg.SelfCountryCode)
	if err != nil {
		return err
	}
	// Only register with those partner federator whose zones can be accessed by self federator
	partnerFed, err := GetPartnerFederatorWithRole(ctx,
		reg.SelfOperatorId, reg.SelfCountryCode,
		reg.PartnerOperatorId, reg.PartnerCountryCode,
		fedcommon.RoleShareZonesWithSelf,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:      reg.ZoneId,
		OperatorId:  partnerFed.OperatorId,
		CountryCode: partnerFed.CountryCode,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", reg.ZoneId)
	}

	// Notify partner federator about zone registration
	opZoneReg := federation.OperatorZoneRegister{
		OrigFederationId: selfFed.FederationKey,
		DestFederationId: partnerFed.FederationKey,
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
		ZoneId:                  existingZone.ZoneId,
		OwnerOperatorId:         existingZone.OperatorId,
		OwnerCountryCode:        existingZone.CountryCode,
		RegisteredByOperatorId:  selfFed.OperatorId,
		RegisteredByCountryCode: selfFed.CountryCode,
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
	reg := ormapi.FederatorZoneRequest{}
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
	if err := fedAuthorized(ctx, claims.Username, reg.SelfOperatorId); err != nil {
		return err
	}
	// get self federator information
	selfFed, err := GetSelfFederator(ctx, reg.SelfOperatorId, reg.SelfCountryCode)
	if err != nil {
		return err
	}
	// Only deregister with those partner federator whose zones can be accessed by self federator
	partnerFed, err := GetPartnerFederatorWithRole(ctx,
		reg.SelfOperatorId, reg.SelfCountryCode,
		reg.PartnerOperatorId, reg.PartnerCountryCode,
		fedcommon.RoleShareZonesWithSelf,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		ZoneId:      reg.ZoneId,
		OperatorId:  partnerFed.OperatorId,
		CountryCode: partnerFed.CountryCode,
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
		OrigFederationId: selfFed.FederationKey,
		DestFederationId: partnerFed.FederationKey,
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
		ZoneId:                  existingZone.ZoneId,
		OwnerOperatorId:         existingZone.OperatorId,
		OwnerCountryCode:        existingZone.CountryCode,
		RegisteredByOperatorId:  selfFed.OperatorId,
		RegisteredByCountryCode: selfFed.CountryCode,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {

			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg(fmt.Sprintf("Partner federator zone %q deregistered successfully", existingZone.ZoneId)))
}

// Creates a directed federation between self federator and partner federator.
// This gives self federator access to all the zones of the partner federator
// which it is willing to share
func CreateFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederationRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
		return err
	}

	// get federator information
	selfFed, err := GetSelfFederator(ctx, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	// get partner federator information
	partnerFed, err := GetPartnerFederator(ctx, opFed.PartnerOperatorId, opFed.PartnerCountryCode)
	if err != nil {
		return err
	}

	// Check that there is no existing federation using the same partner
	// federator with role to share zones with self. This constraint is required
	// as there cannot be two partner federators with same operatorID/countryCode
	// sharing zones with MC. This will cause confusion for developers as to which
	// partner federator to choose
	db := loggedDB(ctx)
	partnerFederation := ormapi.Federation{
		PartnerOperatorId:  opFed.PartnerOperatorId,
		PartnerCountryCode: opFed.PartnerCountryCode,
		PartnerRole:        fedcommon.RoleShareZonesWithSelf,
	}
	res := db.Where(&partnerFederation).First(nil)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if res.Error == nil {
		return fmt.Errorf("Federation with partner federator (%s) providing access to partner federator zones already exists",
			fedcommon.FederatorStr(opFed.PartnerOperatorId, opFed.PartnerCountryCode))
	}

	// check that there is no existing federation with partner federator
	partnerFederation = ormapi.Federation{
		SelfOperatorId:     opFed.SelfOperatorId,
		SelfCountryCode:    opFed.SelfCountryCode,
		PartnerOperatorId:  opFed.PartnerOperatorId,
		PartnerCountryCode: opFed.PartnerCountryCode,
		PartnerRole:        fedcommon.RoleShareZonesWithSelf,
	}
	res = db.Where(&partnerFederation).First(nil)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if res.Error == nil {
		return fmt.Errorf("Federation already exists with partner federator (%s)",
			fedcommon.FederatorStr(opFed.PartnerOperatorId, opFed.PartnerCountryCode))
	}

	// Call federation API
	opRegReq := federation.OperatorRegistrationRequest{
		OrigFederationId: selfFed.FederationKey,
		DestFederationId: partnerFed.FederationKey,
		OperatorId:       selfFed.OperatorId,
		CountryCode:      selfFed.CountryCode,
	}
	opRegRes := federation.OperatorRegistrationResponse{}
	err = sendFederationRequest("POST", partnerFed.FederationAddr, federation.OperatorPartnerAPI, &opRegReq, &opRegRes)
	if err != nil {
		return err
	}
	// Store partner zones in DB
	for _, partnerZone := range opRegRes.PartnerZone {
		zoneObj := ormapi.FederatorZone{}
		zoneObj.OperatorId = opFed.PartnerOperatorId
		zoneObj.CountryCode = opFed.PartnerCountryCode
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

	// Add federation with partner federator
	if err := db.Create(&partnerFederation).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Federation with partner federator %s already exists",
				fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
		}
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg(
		fmt.Sprintf("Created directed federation with partner federator (%s) successfully",
			fedcommon.FederatorStr(opFed.PartnerOperatorId, opFed.PartnerCountryCode))))
}

// Delete directed federation between self federator and partner federator.
// Partner federator will no longer have access to any of self federator zones
func DeleteFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederationRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
		return err
	}

	// get self federator information
	selfFed, err := GetSelfFederator(ctx, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}

	// get partner federator information with appropriate federation role
	partnerFed, err := GetPartnerFederatorWithRole(ctx,
		opFed.SelfOperatorId, opFed.SelfCountryCode,
		opFed.PartnerOperatorId, opFed.PartnerCountryCode,
		fedcommon.RoleShareZonesWithSelf,
	)
	if err != nil {
		return err
	}

	// Check if all the partner zones are unused before deleting the partner federator
	lookup := ormapi.FederatorZone{
		OperatorId:  partnerFed.OperatorId,
		CountryCode: partnerFed.CountryCode,
	}
	partnerZones := []ormapi.FederatorZone{}
	db := loggedDB(ctx)
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, pZone := range partnerZones {
		regLookup := ormapi.FederatorRegisteredZone{
			ZoneId:                  pZone.ZoneId,
			OwnerOperatorId:         pZone.OperatorId,
			OwnerCountryCode:        pZone.CountryCode,
			RegisteredByOperatorId:  selfFed.OperatorId,
			RegisteredByCountryCode: selfFed.CountryCode,
		}
		regZone := ormapi.FederatorRegisteredZone{}
		res := db.Where(&regLookup).First(&regZone)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		if regZone.ZoneId != "" {
			return fmt.Errorf("Cannot delete federation with partner federator (%s) as partner "+
				"zone %q is registered locally. Please deregister it before removing the federation partner",
				fedcommon.FederatorStr(pZone.OperatorId, pZone.CountryCode), regZone.ZoneId)
		}
	}

	// call federation API
	opFedReq := federation.FederationRequest{
		OrigFederationId: selfFed.FederationKey,
		DestFederationId: partnerFed.FederationKey,
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
	}
	err = sendFederationRequest("DELETE", partnerFed.FederationAddr, federation.OperatorPartnerAPI, &opFedReq, nil)
	if err != nil {
		return err
	}

	// Delete all the local copy of partner federator zones
	for _, pZone := range partnerZones {
		if err := db.Delete(pZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
		regZone := ormapi.FederatorRegisteredZone{
			ZoneId:           pZone.ZoneId,
			OwnerOperatorId:  pZone.OperatorId,
			OwnerCountryCode: pZone.CountryCode,
		}
		if err := db.Delete(regZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
		shZone := ormapi.FederatorSharedZone{
			ZoneId:           pZone.ZoneId,
			OwnerOperatorId:  pZone.OperatorId,
			OwnerCountryCode: pZone.CountryCode,
		}
		if err := db.Delete(shZone).Error; err != nil {
			if err != gorm.ErrRecordNotFound {

				return ormutil.DbErr(err)
			}
		}
	}

	// delete partner federation
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     opFed.SelfOperatorId,
		SelfCountryCode:    opFed.SelfCountryCode,
		PartnerOperatorId:  opFed.PartnerOperatorId,
		PartnerCountryCode: opFed.PartnerCountryCode,
		PartnerRole:        fedcommon.RoleShareZonesWithSelf,
	}
	if err := db.Delete(&partnerFederation).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg(
		fmt.Sprintf("Deleted directed federation with partner federator (%s) successfully",
			fedcommon.FederatorStr(opFed.PartnerOperatorId, opFed.PartnerCountryCode))))
}

func ShowFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.FederatorRequest{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	if err := fedAuthorized(ctx, claims.Username, opFed.SelfOperatorId); err != nil {
		return err
	}

	// validate self federator information
	_, err = GetSelfFederator(ctx, opFed.SelfOperatorId, opFed.SelfCountryCode)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.Federation{
		SelfOperatorId:  opFed.SelfOperatorId,
		SelfCountryCode: opFed.SelfCountryCode,
	}
	outFeds := []ormapi.Federation{}
	res := db.Where(&lookup).Find(&outFeds)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	return c.JSON(http.StatusOK, outFeds)
}

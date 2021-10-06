package federation

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlclient"
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
	"github.com/mobiledgex/edge-cloud-infra/mc/gormlog"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
)

const (
	// Federation APIs
	OperatorPartnerAPI    = "/operator/partner"
	OperatorZoneAPI       = "/operator/zone"
	OperatorNotifyZoneAPI = "/operator/notify/zone"
)

type PartnerApi struct {
	Database  *gorm.DB
	ConnCache ctrlclient.ClientConnMgr
}

func (p *PartnerApi) loggedDB(ctx context.Context) *gorm.DB {
	return gormlog.LoggedDB(ctx, p.Database)
}

// E/W-BoundInterface APIs for Federation between multiple Operator Platforms (federators)
// These are the standard interfaces which are called by other federators for unified edge platform experience
func (p *PartnerApi) InitAPIs(e *echo.Echo) {
	// Create directed federation with partner federator
	e.POST(OperatorPartnerAPI, p.FederationOperatorPartnerCreate)
	// Update attributes of an existing federation with a partner federator
	e.PUT(OperatorPartnerAPI, p.FederationOperatorPartnerUpdate)
	// Remove existing federation with a partner federator
	e.DELETE(OperatorPartnerAPI, p.FederationOperatorPartnerDelete)
	// Register a partner federator zone
	e.POST(OperatorZoneAPI, p.FederationOperatorZoneRegister)
	// Deregister a partner federator zone
	e.DELETE(OperatorZoneAPI, p.FederationOperatorZoneDeRegister)
	// Notify partner federator about a new zone being added
	e.POST(OperatorNotifyZoneAPI, p.FederationOperatorZoneShare)
	// Notify partner federator about a zone being unshared
	e.DELETE(OperatorNotifyZoneAPI, p.FederationOperatorZoneUnShare)
}

func (p *PartnerApi) ValidateAndGetFederatorInfo(ctx context.Context, origKey, destKey, origOperatorId, origCountryCode string) (*ormapi.Federator, *ormapi.Federator, error) {
	// sanity check
	if origKey == "" {
		return nil, nil, fmt.Errorf("Missing origin federation key")
	}
	if destKey == "" {
		return nil, nil, fmt.Errorf("Missing destination federation key")
	}
	if origOperatorId == "" {
		return nil, nil, fmt.Errorf("Missing origin Operator ID")
	}
	if origCountryCode == "" {
		return nil, nil, fmt.Errorf("Missing origin country code")
	}
	db := p.loggedDB(ctx)

	// We are using destination federation key to get self federator info.
	// A federator is identified by operatorID/countryCode, but the partner
	// federator does not send these details. One reason for this could be
	// Hence, MC uses the destination federation key to multiplex federation
	// requests to appropriate self federator
	selfFed := ormapi.Federator{
		FederationKey: destKey,
		Type:          fedcommon.TypeSelf,
	}
	res := db.Where(&selfFed).First(&selfFed)
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Destination federator does not exist")
	}
	if res.Error != nil {
		return nil, nil, ormutil.DbErr(res.Error)
	}
	// validate destination (self) federation key
	if selfFed.FederationKey != destKey {
		return nil, nil, fmt.Errorf("Invalid destination federation key")
	}

	partnerFed := ormapi.Federator{
		OperatorId:  origOperatorId,
		CountryCode: origCountryCode,
		Type:        fedcommon.TypePartner,
	}
	res = db.Where(&partnerFed).First(&partnerFed)
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Origin federator %s does not exist",
			fedcommon.FederatorStr(origOperatorId, origCountryCode))
	}
	if res.Error != nil {
		return nil, nil, ormutil.DbErr(res.Error)
	}
	// validate origin (partner) federationkey/operator/country
	if partnerFed.FederationKey != origKey {
		return nil, nil, fmt.Errorf("Invalid origin federation key")
	}
	return &selfFed, &partnerFed, nil
}

func (p *PartnerApi) ValidateWithRoleAndGetFederatorInfo(ctx context.Context, origKey, destKey, origOperatorId, origCountryCode string, role fedcommon.FederatorRole) (*ormapi.Federator, *ormapi.Federator, error) {
	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(ctx, origKey, destKey, origOperatorId, origCountryCode)
	if err != nil {
		return nil, nil, err
	}

	// check that there is no existing federation with partner federator
	db := p.loggedDB(ctx)
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     selfFed.OperatorId,
		SelfCountryCode:    selfFed.CountryCode,
		PartnerOperatorId:  origOperatorId,
		PartnerCountryCode: origCountryCode,
		PartnerRole:        role,
	}
	res := db.Where(&partnerFederation).First(&partnerFederation)
	if !res.RecordNotFound() && res.Error != nil {
		return nil, nil, ormutil.DbErr(res.Error)
	}
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Partner federator (%s) with role %q does not exist",
			fedcommon.FederatorStr(origOperatorId, origCountryCode), role)
	}
	return selfFed, partnerFed, nil
}

// Remote partner federator requests to create the federation, which
// allows its developers and subscribers to run their applications
// on our cloudlets
func (p *PartnerApi) FederationOperatorPartnerCreate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.OperatorId,
		opRegReq.CountryCode,
	)
	if err != nil {
		return err
	}

	// check that there is no existing federation with partner federator
	db := p.loggedDB(ctx)
	partnerFederation := ormapi.Federation{
		SelfOperatorId:     selfFed.OperatorId,
		SelfCountryCode:    selfFed.CountryCode,
		PartnerOperatorId:  opRegReq.OperatorId,
		PartnerCountryCode: opRegReq.CountryCode,
		PartnerRole:        fedcommon.RoleAccessToSelfZones,
	}
	res := db.Where(&partnerFederation).First(&partnerFederation)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if res.Error == nil {
		return fmt.Errorf("Federation already exists with partner federator (%s)",
			fedcommon.FederatorStr(opRegReq.OperatorId, opRegReq.CountryCode))
	}

	// Get list of zones to be shared with partner federator
	opShZones := []ormapi.FederatorSharedZone{}
	lookup := ormapi.FederatorSharedZone{
		OwnerOperatorId:       selfFed.OperatorId,
		OwnerCountryCode:      selfFed.CountryCode,
		SharedWithOperatorId:  partnerFed.OperatorId,
		SharedWithCountryCode: partnerFed.CountryCode,
	}
	err = db.Where(&lookup).Find(&opShZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	out := OperatorRegistrationResponse{}
	out.OrigFederationId = selfFed.FederationKey
	out.DestFederationId = opRegReq.OrigFederationId
	out.OrigOperatorId = selfFed.OperatorId
	out.PartnerOperatorId = opRegReq.OperatorId
	out.MCC = selfFed.MCC
	out.MNC = strings.Split(selfFed.MNCs, fedcommon.Delimiter)
	out.LocatorEndPoint = selfFed.LocatorEndPoint
	for _, opShZone := range opShZones {
		zoneLookup := ormapi.FederatorZone{
			ZoneId:          opShZone.ZoneId,
			SelfOperatorId:  opShZone.OwnerOperatorId,
			SelfCountryCode: opShZone.OwnerCountryCode,
			OperatorId:      opShZone.OwnerOperatorId,
			CountryCode:     opShZone.OwnerCountryCode,
		}
		opZone := ormapi.FederatorZone{}
		err = db.Where(&zoneLookup).First(&opZone).Error
		if err != nil {
			return ormutil.DbErr(err)
		}

		partnerZone := ZoneInfo{
			ZoneId:      opZone.ZoneId,
			GeoLocation: opZone.GeoLocation,
			City:        opZone.City,
			State:       opZone.State,
			Locality:    opZone.Locality,
			EdgeCount:   len(strings.Split(opZone.Cloudlets, fedcommon.Delimiter)),
		}
		out.PartnerZone = append(out.PartnerZone, partnerZone)
	}

	// Add federation with partner federator
	if err := db.Create(&partnerFederation).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Federation with partner federator %s already exists",
				fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
		}
		return ormutil.DbErr(err)
	}

	// Return with list of zones to be shared
	return c.JSON(http.StatusOK, out)
}

// Remote partner federator sends this request to us to notify about
// the change in its MNC, MCC or locator URL
func (p *PartnerApi) FederationOperatorPartnerUpdate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opConf := UpdateMECNetConf{}
	if err := c.Bind(&opConf); err != nil {
		return err
	}

	_, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opConf.OrigFederationId,
		opConf.DestFederationId,
		opConf.Operator,
		opConf.Country,
		fedcommon.RoleAccessToSelfZones,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	update := false
	if opConf.MCC != "" {
		partnerFed.MCC = opConf.MCC
		update = true
	}
	if len(opConf.MNC) > 0 {
		partnerFed.MNCs = strings.Join(opConf.MNC, ",")
		update = true
	}
	if opConf.LocatorEndPoint != "" {
		partnerFed.LocatorEndPoint = opConf.LocatorEndPoint
		update = true
	}

	if !update {
		return fmt.Errorf("Nothing to update")
	}

	err = db.Save(partnerFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	return c.JSON(http.StatusOK, "Federation attributes of partner federator updated successfully")
}

// Remote partner federator requests to delete the federation, which
// disallows its developers and subscribers to run their applications
// on our cloudlets
func (p *PartnerApi) FederationOperatorPartnerDelete(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opFedReq := FederationRequest{}
	if err := c.Bind(&opFedReq); err != nil {
		return err
	}

	selfFed, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opFedReq.OrigFederationId,
		opFedReq.DestFederationId,
		opFedReq.Operator,
		opFedReq.Country,
		fedcommon.RoleAccessToSelfZones,
	)
	if err != nil {
		return err
	}

	// Check if all the self zones are deregistered by partner federator
	db := p.loggedDB(ctx)
	lookup := ormapi.FederatorZone{
		SelfOperatorId:  selfFed.OperatorId,
		SelfCountryCode: selfFed.CountryCode,
		OperatorId:      selfFed.OperatorId,
		CountryCode:     selfFed.CountryCode,
	}
	partnerZones := []ormapi.FederatorZone{}
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, pZone := range partnerZones {
		regLookup := ormapi.FederatorRegisteredZone{
			ZoneId:                  pZone.ZoneId,
			OwnerOperatorId:         pZone.OperatorId,
			OwnerCountryCode:        pZone.CountryCode,
			RegisteredByOperatorId:  partnerFed.OperatorId,
			RegisteredByCountryCode: partnerFed.CountryCode,
		}
		regZone := ormapi.FederatorRegisteredZone{}
		res := db.Where(&regLookup).First(&regZone)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		if regZone.ZoneId != "" {
			return fmt.Errorf("Cannot delete partner federation as zone %q of self federator (%s) "+
				"is registered by partner federator. Please deregister it before deleting the "+
				"partner federation", regZone.ZoneId,
				fedcommon.FederatorStr(pZone.OperatorId, pZone.CountryCode))
		}
	}

	// Delete partner federator role
	partnerRole := ormapi.Federation{
		SelfOperatorId:     selfFed.OperatorId,
		SelfCountryCode:    selfFed.CountryCode,
		PartnerOperatorId:  partnerFed.OperatorId,
		PartnerCountryCode: partnerFed.CountryCode,
		PartnerRole:        fedcommon.RoleAccessToSelfZones,
	}
	if err = db.Delete(&partnerRole).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted partner federation successfully"))
}

// Remote partner federator sends this request to us to register
// our zone i.e cloudlet. Once our cloudlet is registered,
// remote partner federator can then make it accessible to its
// developers or subscribers
func (p *PartnerApi) FederationOperatorZoneRegister(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorZoneRegister{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}
	if len(opRegReq.Zones) == 0 {
		return fmt.Errorf("Must specify one zone ID")
	}
	selfFed, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
		fedcommon.RoleAccessToSelfZones,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := p.loggedDB(ctx)
	zoneId := opRegReq.Zones[0]
	lookup := ormapi.FederatorZone{
		ZoneId:          zoneId,
		SelfOperatorId:  selfFed.OperatorId,
		SelfCountryCode: selfFed.CountryCode,
		OperatorId:      selfFed.OperatorId,
		CountryCode:     selfFed.CountryCode,
	}
	existingFed := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found for self federator %s", zoneId,
			fedcommon.FederatorStr(selfFed.OperatorId, selfFed.CountryCode))
	}

	// Store registration details in DB
	regZone := ormapi.FederatorRegisteredZone{
		ZoneId:                  zoneId,
		OwnerOperatorId:         selfFed.OperatorId,
		OwnerCountryCode:        selfFed.CountryCode,
		RegisteredByOperatorId:  partnerFed.OperatorId,
		RegisteredByCountryCode: partnerFed.CountryCode,
	}
	if err := db.Create(&regZone).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone ID %q is already registered by %s", zoneId,
				fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
		}
		return ormutil.DbErr(err)
	}

	// Share zone details
	resp := OperatorZoneRegisterResponse{}
	resp.LeadOperatorId = selfFed.OperatorId
	resp.PartnerOperatorId = opRegReq.Operator
	resp.FederationId = selfFed.FederationKey
	resp.Zone = ZoneRegisterDetails{
		ZoneId:            zoneId,
		RegistrationToken: selfFed.FederationKey,
	}
	return c.JSON(http.StatusOK, resp)
}

// Remote partner federator deregisters our zone i.e. cloudlet.
// This will ensure that our cloudlet is no longer accessible
// to remote partner federator's developers or subscribers
func (p *PartnerApi) FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := ZoneRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}
	if opRegReq.Zone == "" {
		return fmt.Errorf("Must specify zone ID")
	}
	selfFed, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
		fedcommon.RoleAccessToSelfZones,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := p.loggedDB(ctx)
	zoneId := opRegReq.Zone
	lookup := ormapi.FederatorZone{
		ZoneId:          zoneId,
		SelfOperatorId:  selfFed.OperatorId,
		SelfCountryCode: selfFed.CountryCode,
		OperatorId:      selfFed.OperatorId,
		CountryCode:     selfFed.CountryCode,
	}
	existingFed := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found for self federator %s", zoneId,
			fedcommon.FederatorStr(selfFed.OperatorId, selfFed.CountryCode))
	}

	// Delete registration details from DB
	deregZone := ormapi.FederatorRegisteredZone{
		ZoneId:                  zoneId,
		OwnerOperatorId:         selfFed.OperatorId,
		OwnerCountryCode:        selfFed.CountryCode,
		RegisteredByOperatorId:  partnerFed.OperatorId,
		RegisteredByCountryCode: partnerFed.CountryCode,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return ormutil.DbErr(err)
		}
	}
	return c.JSON(http.StatusOK, "Zone deregistered successfully")
}

// Remote partner federator sends this request to us to share its zone.
// It is triggered by partner federator when it has a new zone available
// and it wishes to share the zone with us
func (p *PartnerApi) FederationOperatorZoneShare(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opZoneShare := NotifyPartnerOperatorZone{}
	if err := c.Bind(&opZoneShare); err != nil {
		return err
	}

	if opZoneShare.PartnerZone.ZoneId == "" {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
		fedcommon.RoleShareZonesWithSelf,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := p.loggedDB(ctx)
	zoneId := opZoneShare.PartnerZone.ZoneId
	lookup := ormapi.FederatorZone{
		ZoneId:          zoneId,
		SelfOperatorId:  selfFed.OperatorId,
		SelfCountryCode: selfFed.CountryCode,
		OperatorId:      partnerFed.OperatorId,
		CountryCode:     partnerFed.CountryCode,
	}
	existingZone := ormapi.FederatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId != "" {
		return fmt.Errorf("Zone ID %q already exists for partner federator %s", zoneId,
			fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
	}

	zoneObj := ormapi.FederatorZone{}
	zoneObj.ZoneId = zoneId
	zoneObj.SelfOperatorId = selfFed.OperatorId
	zoneObj.SelfCountryCode = selfFed.CountryCode
	zoneObj.OperatorId = partnerFed.OperatorId
	zoneObj.CountryCode = partnerFed.CountryCode
	zoneObj.GeoLocation = opZoneShare.PartnerZone.GeoLocation
	zoneObj.City = opZoneShare.PartnerZone.City
	zoneObj.Locality = opZoneShare.PartnerZone.Locality
	if err := db.Create(&zoneObj).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone ID %q already exists", zoneObj.ZoneId)
		}
		return ormutil.DbErr(err)
	}

	// Mark zone as shared in DB
	shareZone := ormapi.FederatorSharedZone{
		ZoneId:                zoneObj.ZoneId,
		OwnerOperatorId:       zoneObj.OperatorId,
		OwnerCountryCode:      zoneObj.CountryCode,
		SharedWithOperatorId:  zoneObj.SelfOperatorId,
		SharedWithCountryCode: zoneObj.SelfCountryCode,
	}
	if err := db.Create(&shareZone).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return ormutil.DbErr(err)
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Added zone successfully"))
}

// Remote partner federator sends this request to us to unshare its zone.
// This api is triggered when the remote partner federator decides to unshare
// one of its zone with us
func (p *PartnerApi) FederationOperatorZoneUnShare(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opZone := ZoneRequest{}
	if err := c.Bind(&opZone); err != nil {
		return err
	}

	if opZone.Zone == "" {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := p.ValidateWithRoleAndGetFederatorInfo(
		ctx,
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
		fedcommon.RoleShareZonesWithSelf,
	)
	if err != nil {
		return err
	}

	// Check if zone exists
	db := p.loggedDB(ctx)
	zoneId := opZone.Zone
	lookup := ormapi.FederatorZone{
		ZoneId:          zoneId,
		SelfOperatorId:  selfFed.OperatorId,
		SelfCountryCode: selfFed.CountryCode,
		OperatorId:      partnerFed.OperatorId,
		CountryCode:     partnerFed.CountryCode,
	}
	existingZone := ormapi.FederatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q does not exist for partner federator (%s)", zoneId,
			fedcommon.FederatorStr(partnerFed.OperatorId, partnerFed.CountryCode))
	}

	// Ensure that this zone is not registered by us
	regLookup := ormapi.FederatorRegisteredZone{
		ZoneId:                  existingZone.ZoneId,
		OwnerOperatorId:         partnerFed.OperatorId,
		OwnerCountryCode:        partnerFed.CountryCode,
		RegisteredByOperatorId:  selfFed.OperatorId,
		RegisteredByCountryCode: selfFed.CountryCode,
	}
	regZone := ormapi.FederatorRegisteredZone{}
	res = db.Where(&regLookup).First(&regZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot remove partner federator zone %q as it is registered locally. "+
			"Please deregister it before unsharing the zone", regZone.ZoneId)
	}

	// Delete zone from shared list in DB
	unshareZone := ormapi.FederatorSharedZone{
		ZoneId:                existingZone.ZoneId,
		OwnerOperatorId:       partnerFed.OperatorId,
		OwnerCountryCode:      partnerFed.CountryCode,
		SharedWithOperatorId:  selfFed.OperatorId,
		SharedWithCountryCode: selfFed.CountryCode,
	}
	if err := db.Delete(&unshareZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return ormutil.DbErr(err)
		}
	}

	if err := db.Delete(&lookup).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Removed zone successfully"))
}

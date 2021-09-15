package federation

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/gormlog"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
)

const (
	// Federation Types
	TypeSelf    = "self"
	TypePartner = "partner"

	// Federation Partner Roles
	RoleAccessZones = "access" // Can access partner OP zones, but cannot share zones with partner OP
	RoleShareZones  = "share"  // Can only share zones with partner OP, but cannot access partner OP's zones

	// Federation APIs
	OperatorPartnerAPI    = "/operator/partner"
	OperatorZoneAPI       = "/operator/zone"
	OperatorNotifyZoneAPI = "/operator/notify/zone"
)

type FederationObj struct {
	Database *gorm.DB
	Echo     *echo.Echo
}

func (f *FederationObj) loggedDB(ctx context.Context) *gorm.DB {
	return gormlog.LoggedDB(ctx, f.Database)
}

// E/W-BoundInterface APIs for Federation between multiple Operator Platforms (OPs)
// These are the standard interfaces which are called by other OPs for unified edge platform experience
func (f *FederationObj) InitFederationAPIs() {
	// Create directed federation with partner OP
	f.Echo.POST(OperatorPartnerAPI, f.FederationOperatorPartnerCreate)
	// Update attributes of an existing federation with a partner OP
	f.Echo.PUT(OperatorPartnerAPI, f.FederationOperatorPartnerUpdate)
	// Remove existing federation with a partner OP
	f.Echo.DELETE(OperatorPartnerAPI, f.FederationOperatorPartnerDelete)
	// Register a partner OP zone
	f.Echo.POST(OperatorZoneAPI, f.FederationOperatorZoneRegister)
	// Deregister a partner OP zone
	f.Echo.DELETE(OperatorZoneAPI, f.FederationOperatorZoneDeRegister)
	// Notify partner OP about a new zone being added
	f.Echo.POST(OperatorNotifyZoneAPI, f.FederationOperatorZoneShare)
	// Notify partner OP about a zone being unshared
	f.Echo.DELETE(OperatorNotifyZoneAPI, f.FederationOperatorZoneUnShare)
}

func (f *FederationObj) ValidateSelfAndGetFederationInfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, error) {
	// sanity check
	if origKey == "" {
		return nil, fmt.Errorf("Missing origin federation key")
	}
	if destKey == "" {
		return nil, fmt.Errorf("Missing destination federation key")
	}
	if operatorId == "" {
		return nil, fmt.Errorf("Missing origin Operator ID")
	}
	if countryCode == "" {
		return nil, fmt.Errorf("Missing origin country code")
	}
	db := f.loggedDB(ctx)
	// validate destination federation key
	selfFed := &ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: TypeSelf,
	}
	res := db.Where(&lookup).First(selfFed)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federation doesn't exist")
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	if selfFed.FederationId != destKey {
		return nil, fmt.Errorf("Invalid destination federation key")
	}
	return selfFed, nil
}

func (f *FederationObj) ValidateAndGetFederationinfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, *ormapi.OperatorFederation, error) {
	selfFed, err := f.ValidateSelfAndGetFederationInfo(ctx, origKey, destKey, operatorId, countryCode)
	if err != nil {
		return nil, nil, err
	}
	db := f.loggedDB(ctx)
	// validate origin federationkey/operator/country
	partnerFed := &ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: origKey,
		Type:         TypePartner,
	}
	res := db.Where(&fedLookup).First(partnerFed)
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Origin federation doesn't exist")
	}
	if res.Error != nil {
		return nil, nil, ormutil.DbErr(res.Error)
	}
	if partnerFed.OperatorId != operatorId {
		return nil, nil, fmt.Errorf("Invalid origin operator ID %s", operatorId)
	}
	if partnerFed.CountryCode != countryCode {
		return nil, nil, fmt.Errorf("Invalid origin country code %s", countryCode)
	}
	return selfFed, partnerFed, nil
}

func (f *FederationObj) ShowOperatorZone(ctx context.Context, opZoneReq *ormapi.OperatorZoneCloudletMap) ([]ormapi.OperatorZoneCloudletMap, error) {
	db := f.loggedDB(ctx)
	opZones := []ormapi.OperatorZone{}
	lookup := ormapi.OperatorZone{
		FederationId: opZoneReq.FederationId,
		ZoneId:       opZoneReq.ZoneId,
	}
	err := db.Where(&lookup).Find(&opZones).Error
	if err != nil {
		return nil, ormutil.DbErr(err)
	}

	fedZones := []ormapi.OperatorZoneCloudletMap{}
	for _, opZone := range opZones {
		clLookup := ormapi.OperatorZoneCloudlet{
			ZoneId: opZone.ZoneId,
		}
		opCloudlets := []ormapi.OperatorZoneCloudlet{}
		err = db.Where(&clLookup).Find(&opCloudlets).Error
		if err != nil {
			return nil, ormutil.DbErr(err)
		}
		opRegZones := []ormapi.OperatorRegisteredZone{}
		regLookup := ormapi.OperatorRegisteredZone{
			FederationId: opZoneReq.FederationId,
			ZoneId:       opZoneReq.ZoneId,
		}
		res := db.Where(&regLookup).Find(&opRegZones)
		if !res.RecordNotFound() && res.Error != nil {
			return nil, ormutil.DbErr(res.Error)
		}

		zoneOut := ormapi.OperatorZoneCloudletMap{}
		zoneOut.ZoneId = opZone.ZoneId
		zoneOut.GeoLocation = opZone.GeoLocation
		zoneOut.City = opZone.City
		zoneOut.State = opZone.State
		zoneOut.Locality = opZone.Locality
		zoneOut.Cloudlets = []string{}
		for _, opCl := range opCloudlets {
			zoneOut.Cloudlets = append(zoneOut.Cloudlets, opCl.CloudletName)
		}
		for _, opRegZone := range opRegZones {
			regZone := fmt.Sprintf("%s/%s", opRegZone.OperatorId, opRegZone.CountryCode)
			zoneOut.RegisteredOPs = append(zoneOut.RegisteredOPs, regZone)
		}

		fedZones = append(fedZones, zoneOut)
	}

	return fedZones, nil
}

// Create directed federation with partner OP. By Federation create request,
// the API initiator OP say ‘A’ request OP 'B' to allow its developers/subscribers
// to run their application on edge sites of OP 'B'
func (f *FederationObj) FederationOperatorPartnerCreate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	selfFed, err := f.ValidateSelfAndGetFederationInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.OperatorId,
		opRegReq.CountryCode,
	)
	if err != nil {
		return err
	}

	// Get list of zones to be shared with partner OP
	opZoneReq := &ormapi.OperatorZoneCloudletMap{
		FederationId: selfFed.FederationId,
	}
	opZones, err := f.ShowOperatorZone(ctx, opZoneReq)
	if err != nil {
		return err
	}
	out := OperatorRegistrationResponse{}
	out.OrigOperatorId = selfFed.OperatorId
	out.PartnerOperatorId = opRegReq.OperatorId
	out.OrigFederationId = selfFed.FederationId
	out.DestFederationId = opRegReq.OrigFederationId
	out.MCC = selfFed.MCC
	out.MNC = strings.Split(selfFed.MNCs, ",")
	out.LocatorEndPoint = selfFed.LocatorEndPoint
	for _, opZone := range opZones {
		partnerZone := ZoneInfo{
			ZoneId:      opZone.ZoneId,
			GeoLocation: opZone.GeoLocation,
			City:        opZone.City,
			State:       opZone.State,
			Locality:    opZone.Locality,
			EdgeCount:   len(opZone.Cloudlets),
		}
		out.PartnerZone = append(out.PartnerZone, partnerZone)
	}

	// Store partner OP, this mean partner OP is now federated
	db := f.loggedDB(ctx)
	partnerOP := ormapi.OperatorFederation{
		FederationId:   opRegReq.OrigFederationId,
		OperatorId:     opRegReq.OperatorId,
		CountryCode:    opRegReq.CountryCode,
		Type:           TypePartner,
		FederationAddr: opRegReq.OrigFederationAddr,
	}
	fedLookup := ormapi.OperatorFederation{
		FederationId: opRegReq.OrigFederationId,
	}
	existingPartnerOP := ormapi.OperatorFederation{}
	err = db.Where(&fedLookup).First(&existingPartnerOP).Error
	if err == nil {
		if FederationRoleExists(&existingPartnerOP, RoleShareZones) {
			return fmt.Errorf("Partner OP is already federated")
		} else {
			err = AddFederationRole(&existingPartnerOP, RoleShareZones)
			if err != nil {
				return err
			}
			err = db.Save(existingPartnerOP).Error
			if err != nil {
				return ormutil.DbErr(err)
			}
		}
	} else {
		err = AddFederationRole(&partnerOP, RoleShareZones)
		if err != nil {
			return err
		}
		if err := db.Create(&partnerOP).Error; err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
				return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s", partnerOP.OperatorId, partnerOP.CountryCode)
			}
			return ormutil.DbErr(err)
		}
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, out)
}

// Federation Agent sends this request to partner OP federation
// Agent to update its MNC, MCC or locator URL
func (f *FederationObj) FederationOperatorPartnerUpdate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opConf := UpdateMECNetConf{}
	if err := c.Bind(&opConf); err != nil {
		return err
	}

	_, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opConf.OrigFederationId,
		opConf.DestFederationId,
		opConf.Operator,
		opConf.Country,
	)
	if err != nil {
		return err
	}

	db := f.loggedDB(ctx)
	save := false
	if opConf.MCC != "" {
		partnerFed.MCC = opConf.MCC
		save = true
	}
	if len(opConf.MNC) > 0 {
		partnerFed.MNCs = strings.Join(opConf.MNC, ",")
		save = true
	}
	if opConf.LocatorEndPoint != "" {
		partnerFed.LocatorEndPoint = opConf.LocatorEndPoint
		save = true
	}

	if !save {
		return fmt.Errorf("Nothing to update")
	}

	err = db.Save(partnerFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, "Federation attributes of partner OP updated successfully")
}

// Remove existing federation with a partner OP. By Federation delete
// request, the API initiator OP say A is requesting to the partner
// OP B to disallow A applications access to OP B edges
func (f *FederationObj) FederationOperatorPartnerDelete(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opFedReq := FederationRequest{}
	if err := c.Bind(&opFedReq); err != nil {
		return err
	}

	_, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opFedReq.OrigFederationId,
		opFedReq.DestFederationId,
		opFedReq.Operator,
		opFedReq.Country,
	)
	if err != nil {
		return err
	}

	db := f.loggedDB(ctx)
	if !FederationRoleExists(partnerFed, RoleShareZones) {
		return fmt.Errorf("No zones are shared with this OP")
	}
	err = RemoveFederationRole(partnerFed, RoleShareZones)
	if err != nil {
		return err
	}
	if partnerFed.Role == "" {
		if err := db.Delete(partnerFed).Error; err != nil {
			return ormutil.DbErr(err)
		}
	} else {
		err = db.Save(partnerFed).Error
		if err != nil {
			return ormutil.DbErr(err)
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted partner OP successfully"))
}

// Operator platform sends this request to partner OP, to register a
// partner OP zone. It is only after successful registration that
// partner OP allow access to its zones. This api shall be triggered
// for each partner zone
func (f *FederationObj) FederationOperatorZoneRegister(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorZoneRegister{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	if len(opRegReq.Zones) == 0 {
		return fmt.Errorf("Must specify one zone ID")
	}

	selfFed, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, RoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
	}

	resp := OperatorZoneRegisterResponse{}
	zoneId := opRegReq.Zones[0]
	db := f.loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: zoneId,
	}
	existingFed := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", zoneId)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot register partner zones, only self zones are allowed to be registered")
	}
	regZone := ormapi.OperatorRegisteredZone{
		ZoneId:       zoneId,
		FederationId: opRegReq.OrigFederationId,
		OperatorId:   opRegReq.Operator,
		CountryCode:  opRegReq.Country,
	}
	if err := db.Create(&regZone).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("ZoneId %s is already registered by %s", zoneId, opRegReq.Operator)
		}
		return ormutil.DbErr(err)
	}
	resp.LeadOperatorId = selfFed.FederationId
	resp.PartnerOperatorId = opRegReq.Operator
	resp.FederationId = opRegReq.OrigFederationId
	resp.Zone = ZoneRegisterDetails{
		ZoneId:            zoneId,
		RegistrationToken: selfFed.FederationId,
	}

	// Share zone registration details
	return c.JSON(http.StatusOK, resp)
}

// Deregister a partner OP zone. By zone deregistration request,
// the API initiator OP say A is indicating to the partner OP
// say B, that it will no longer access partner OP zone
func (f *FederationObj) FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := ZoneRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	if len(opRegReq.Zone) == 0 {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, RoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
	}

	db := f.loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opRegReq.Zone,
	}
	existingFed := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", opRegReq.Zone)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot deregister partner zones, only self zones are allowed to be deregistered")
	}
	deregZone := ormapi.OperatorRegisteredZone{
		ZoneId:       opRegReq.Zone,
		FederationId: opRegReq.OrigFederationId,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Zone %s is already deregistered for operator %s", opRegReq.Zone, opRegReq.Operator)
		}
		return ormutil.DbErr(err)
	}
	return c.JSON(http.StatusOK, "Zone deregistered successfully")
}

// OP notifies its partner MEcs whenever it has a new zone.
// This api is triggered when OP has a new zone available and
// it wishes to share the zone with its partner OP. This request
// is triggered only when federation already exists
func (f *FederationObj) FederationOperatorZoneShare(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opZoneShare := NotifyPartnerOperatorZone{}
	if err := c.Bind(&opZoneShare); err != nil {
		return err
	}

	if opZoneShare.PartnerZone.ZoneId == "" {
		return fmt.Errorf("Must specify zone ID")
	}

	_, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, RoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
	}

	db := f.loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opZoneShare.PartnerZone.ZoneId,
	}
	existingZone := ormapi.OperatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if existingZone.ZoneId != "" {
		return fmt.Errorf("Zone %s already exists", opZoneShare.PartnerZone.ZoneId)
	}
	zoneObj := ormapi.OperatorZone{}
	zoneObj.FederationId = opZoneShare.OrigFederationId
	zoneObj.ZoneId = opZoneShare.PartnerZone.ZoneId
	zoneObj.GeoLocation = opZoneShare.PartnerZone.GeoLocation
	zoneObj.City = opZoneShare.PartnerZone.City
	zoneObj.Locality = opZoneShare.PartnerZone.Locality
	if err := db.Create(&zoneObj).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("ZoneId %s already exists", zoneObj.ZoneId)
		}
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Added zone successfully"))
}

// OP notifies its partner MECs whenever it unshares a zone.
// This api is triggered when a OP decides to unshare one of its
// zone with one of the federated partners. This is used when
// federation already exists
func (f *FederationObj) FederationOperatorZoneUnShare(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opZone := ZoneRequest{}
	if err := c.Bind(&opZone); err != nil {
		return err
	}

	if opZone.Zone == "" {
		return fmt.Errorf("Must specify zone ID")
	}

	_, partnerFed, err := f.ValidateAndGetFederationinfo(
		ctx,
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, RoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
	}

	db := f.loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opZone.Zone,
	}
	existingZone := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s does not exist", opZone.Zone)
	}
	if err := db.Delete(&lookup).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted zone successfully"))
}

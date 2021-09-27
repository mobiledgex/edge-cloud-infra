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

func (p *PartnerApi) ValidateSelfAndGetFederatorInfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.Federator, error) {
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
	db := p.loggedDB(ctx)
	// validate destination federation key
	selfFed := &ormapi.Federator{}
	lookup := ormapi.Federator{
		Type: fedcommon.TypeSelf,
	}
	res := db.Where(&lookup).First(selfFed)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federator doesn't exist")
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
	}
	if selfFed.FederationId != destKey {
		return nil, fmt.Errorf("Invalid destination federation key")
	}
	return selfFed, nil
}

func (p *PartnerApi) ValidateAndGetFederatorInfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.Federator, *ormapi.Federator, error) {
	selfFed, err := p.ValidateSelfAndGetFederatorInfo(ctx, origKey, destKey, operatorId, countryCode)
	if err != nil {
		return nil, nil, err
	}
	db := p.loggedDB(ctx)
	// validate origin federationkey/operator/country
	partnerFed := &ormapi.Federator{}
	fedLookup := ormapi.Federator{
		FederationId: origKey,
		Type:         fedcommon.TypePartner,
	}
	res := db.Where(&fedLookup).First(partnerFed)
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Origin federator doesn't exist")
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

// Remote partner federator requests to create the federation, which
// allows its developers and subscribers to run their applications
// on our cloudlets
func (p *PartnerApi) FederationOperatorPartnerCreate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	selfFed, err := p.ValidateSelfAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.OperatorId,
		opRegReq.CountryCode,
	)
	if err != nil {
		return err
	}

	// Get list of zones to be shared with partner federator
	db := p.loggedDB(ctx)
	opZones := []ormapi.FederatorZone{}
	lookup := ormapi.FederatorZone{
		FederationId: selfFed.FederationId,
	}
	err = db.Where(&lookup).Find(&opZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}

	out := OperatorRegistrationResponse{}
	out.OrigOperatorId = selfFed.OperatorId
	out.PartnerOperatorId = opRegReq.OperatorId
	out.OrigFederationId = selfFed.FederationId
	out.DestFederationId = opRegReq.OrigFederationId
	out.MCC = selfFed.MCC
	out.MNC = strings.Split(selfFed.MNCs, fedcommon.Delimiter)
	out.LocatorEndPoint = selfFed.LocatorEndPoint
	for _, opZone := range opZones {
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

	// Store partner federator, this mean partner federator is now federated
	partnerFed := ormapi.Federator{
		FederationId:   opRegReq.OrigFederationId,
		OperatorId:     opRegReq.OperatorId,
		CountryCode:    opRegReq.CountryCode,
		Type:           fedcommon.TypePartner,
		FederationAddr: opRegReq.OrigFederationAddr,
	}
	if err := db.Create(&partnerFed).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return ormutil.DbErr(err)
		}
	}

	// Add/Update partner federator role
	err = fedcommon.AddOrUpdatePartnerFederatorRole(db, selfFed, &partnerFed, fedcommon.RoleShareZonesWithPartner)
	if err != nil {
		return err
	}

	// Share list of zones to be shared
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

	_, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opConf.OrigFederationId,
		opConf.DestFederationId,
		opConf.Operator,
		opConf.Country,
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

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opFedReq.OrigFederationId,
		opFedReq.DestFederationId,
		opFedReq.Operator,
		opFedReq.Country,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	partnerRoleLookup := ormapi.FederatorRole{
		SelfFederationId:    selfFed.FederationId,
		PartnerFederationId: partnerFed.FederationId,
	}
	partnerRole := ormapi.FederatorRole{}
	res := db.Where(&partnerRoleLookup).First(&partnerRole)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if res.RecordNotFound() || !fedcommon.ValueExistsInDelimitedList(partnerRole.Role, fedcommon.RoleShareZonesWithPartner) {
		return fmt.Errorf("No zones are shared with partner federator")
	}

	err = fedcommon.RemoveFromDelimitedList(&partnerRole.Role, fedcommon.RoleShareZonesWithPartner)
	if err != nil {
		return err
	}
	if partnerRole.Role == "" {
		if err := db.Delete(&partnerRole).Error; err != nil {
			return ormutil.DbErr(err)
		}
		if err := db.Delete(&partnerFed).Error; err != nil {
			return ormutil.DbErr(err)
		}
	} else {
		err = db.Save(&partnerRole).Error
		if err != nil {
			return ormutil.DbErr(err)
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted partner federator successfully"))
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

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	ok, err := fedcommon.FederationRoleExists(db, selfFed, partnerFed, fedcommon.RoleShareZonesWithPartner)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("No zones are shared with partner federator")
	}

	resp := OperatorZoneRegisterResponse{}
	zoneId := opRegReq.Zones[0]
	lookup := ormapi.FederatorZone{
		ZoneId: zoneId,
	}
	existingFed := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", zoneId)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot register partner federator zones, only self zones are allowed to be registered")
	}
	regZone := ormapi.FederatorRegisteredZone{
		ZoneId:       zoneId,
		FederationId: opRegReq.OrigFederationId,
		OperatorId:   opRegReq.Operator,
		CountryCode:  opRegReq.Country,
	}
	if err := db.Create(&regZone).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone ID %q is already registered by %q", zoneId, opRegReq.Operator)
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

	// Share zone details
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

	if len(opRegReq.Zone) == 0 {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	ok, err := fedcommon.FederationRoleExists(db, selfFed, partnerFed, fedcommon.RoleShareZonesWithPartner)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("No zones are shared with partner federator")
	}

	lookup := ormapi.FederatorZone{
		ZoneId: opRegReq.Zone,
	}
	existingFed := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone ID %q not found", opRegReq.Zone)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot deregister partner federator zones, only self zones are allowed to be deregistered")
	}
	deregZone := ormapi.FederatorRegisteredZone{
		ZoneId:       opRegReq.Zone,
		FederationId: opRegReq.OrigFederationId,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Zone ID %q is already deregistered for operator %q", opRegReq.Zone, opRegReq.Operator)
		}
		return ormutil.DbErr(err)
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

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	ok, err := fedcommon.FederationRoleExists(db, selfFed, partnerFed, fedcommon.RoleAccessPartnerZones)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Self federator does not have access to partner federator zones")
	}

	lookup := ormapi.FederatorZone{
		ZoneId: opZoneShare.PartnerZone.ZoneId,
	}
	existingZone := ormapi.FederatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if existingZone.ZoneId != "" {
		return fmt.Errorf("Zone ID %q already exists", opZoneShare.PartnerZone.ZoneId)
	}
	zoneObj := ormapi.FederatorZone{}
	zoneObj.FederationId = opZoneShare.OrigFederationId
	zoneObj.ZoneId = opZoneShare.PartnerZone.ZoneId
	zoneObj.GeoLocation = opZoneShare.PartnerZone.GeoLocation
	zoneObj.City = opZoneShare.PartnerZone.City
	zoneObj.Locality = opZoneShare.PartnerZone.Locality
	if err := db.Create(&zoneObj).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone ID %q already exists", zoneObj.ZoneId)
		}
		return ormutil.DbErr(err)
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

	selfFed, partnerFed, err := p.ValidateAndGetFederatorInfo(
		ctx,
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
	)
	if err != nil {
		return err
	}

	db := p.loggedDB(ctx)
	ok, err := fedcommon.FederationRoleExists(db, selfFed, partnerFed, fedcommon.RoleAccessPartnerZones)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Self federator does not have access to partner federator zones")
	}

	lookup := ormapi.FederatorZone{
		ZoneId: opZone.Zone,
	}
	existingZone := ormapi.FederatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone ID %q does not exist", opZone.Zone)
	}

	// Ensure that this zone is not registered by us
	regLookup := ormapi.FederatorRegisteredZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: selfFed.FederationId,
	}
	regZone := ormapi.FederatorRegisteredZone{}
	res := db.Where(&regLookup).First(&regZone)
	if !res.RecordNotFound() && res.Error != nil {
		return ormutil.DbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot remove partner federator zone %q as it is registered locally. Please deregister it before unsharing the zone", regZone.ZoneId)
	}

	if err := db.Delete(&lookup).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted zone successfully"))
}

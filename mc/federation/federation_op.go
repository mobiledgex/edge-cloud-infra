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

// E/W-BoundInterface APIs for Federation between multiple Operator Platforms (OPs)
// These are the standard interfaces which are called by other OPs for unified edge platform experience
func (p *PartnerApi) InitAPIs(e *echo.Echo) {
	// Create directed federation with partner OP
	e.POST(OperatorPartnerAPI, p.FederationOperatorPartnerCreate)
	// Update attributes of an existing federation with a partner OP
	e.PUT(OperatorPartnerAPI, p.FederationOperatorPartnerUpdate)
	// Remove existing federation with a partner OP
	e.DELETE(OperatorPartnerAPI, p.FederationOperatorPartnerDelete)
	// Register a partner OP zone
	e.POST(OperatorZoneAPI, p.FederationOperatorZoneRegister)
	// Deregister a partner OP zone
	e.DELETE(OperatorZoneAPI, p.FederationOperatorZoneDeRegister)
	// Notify partner OP about a new zone being added
	e.POST(OperatorNotifyZoneAPI, p.FederationOperatorZoneShare)
	// Notify partner OP about a zone being unshared
	e.DELETE(OperatorNotifyZoneAPI, p.FederationOperatorZoneUnShare)
}

func (p *PartnerApi) ValidateSelfAndGetFederationInfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, error) {
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
	selfFed := &ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: fedcommon.TypeSelf,
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

func (p *PartnerApi) ValidateAndGetFederationinfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, *ormapi.OperatorFederation, error) {
	selfFed, err := p.ValidateSelfAndGetFederationInfo(ctx, origKey, destKey, operatorId, countryCode)
	if err != nil {
		return nil, nil, err
	}
	db := p.loggedDB(ctx)
	// validate origin federationkey/operator/country
	partnerFed := &ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: origKey,
		Type:         fedcommon.TypePartner,
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

func (p *PartnerApi) ShowOperatorZone(ctx context.Context, opZoneReq *ormapi.OperatorZoneCloudletMap) ([]ormapi.OperatorZoneCloudletMap, error) {
	db := p.loggedDB(ctx)
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

// Remote partner OP requests to create the federation, which
// allows its developers and subscribers to run their applications
// on our cloudlets
func (p *PartnerApi) FederationOperatorPartnerCreate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := OperatorRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	selfFed, err := p.ValidateSelfAndGetFederationInfo(
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
	opZones, err := p.ShowOperatorZone(ctx, opZoneReq)
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
	db := p.loggedDB(ctx)
	partnerOP := ormapi.OperatorFederation{
		FederationId:   opRegReq.OrigFederationId,
		OperatorId:     opRegReq.OperatorId,
		CountryCode:    opRegReq.CountryCode,
		Type:           fedcommon.TypePartner,
		FederationAddr: opRegReq.OrigFederationAddr,
	}
	fedLookup := ormapi.OperatorFederation{
		FederationId: opRegReq.OrigFederationId,
	}
	existingPartnerOP := ormapi.OperatorFederation{}
	err = db.Where(&fedLookup).First(&existingPartnerOP).Error
	if err == nil {
		if fedcommon.FederationRoleExists(&existingPartnerOP, fedcommon.RoleShareZones) {
			return fmt.Errorf("Partner OP is already federated")
		} else {
			err = fedcommon.AddFederationRole(&existingPartnerOP, fedcommon.RoleShareZones)
			if err != nil {
				return err
			}
			err = db.Save(existingPartnerOP).Error
			if err != nil {
				return ormutil.DbErr(err)
			}
		}
	} else {
		err = fedcommon.AddFederationRole(&partnerOP, fedcommon.RoleShareZones)
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

// Remote partner OP sends this request to us to notify about
// the change in its MNC, MCC or locator URL
func (p *PartnerApi) FederationOperatorPartnerUpdate(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opConf := UpdateMECNetConf{}
	if err := c.Bind(&opConf); err != nil {
		return err
	}

	_, partnerFed, err := p.ValidateAndGetFederationinfo(
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

	return c.JSON(http.StatusOK, "Federation attributes of partner OP updated successfully")
}

// Remote partner OP requests to delete the federation, which
// disallows its developers and subscribers to run their applications
// on our cloudlets
func (p *PartnerApi) FederationOperatorPartnerDelete(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opFedReq := FederationRequest{}
	if err := c.Bind(&opFedReq); err != nil {
		return err
	}

	_, partnerFed, err := p.ValidateAndGetFederationinfo(
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
	if !fedcommon.FederationRoleExists(partnerFed, fedcommon.RoleShareZones) {
		return fmt.Errorf("No zones are shared with this OP")
	}
	err = fedcommon.RemoveFederationRole(partnerFed, fedcommon.RoleShareZones)
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

// Remote partner OP sends this request to us to register
// our zone i.e cloudlet. Once our cloudlet is registered,
// remote partner OP can then make it accessible to its
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

	selfFed, partnerFed, err := p.ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !fedcommon.FederationRoleExists(partnerFed, fedcommon.RoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
	}

	resp := OperatorZoneRegisterResponse{}
	zoneId := opRegReq.Zones[0]
	db := p.loggedDB(ctx)
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

	// Share zone details
	return c.JSON(http.StatusOK, resp)
}

// Remote partner OP deregisters our zone i.e. cloudlet.
// This will ensure that our cloudlet is no longer accessible
// to remote partner OP's developers or subscribers
func (p *PartnerApi) FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	opRegReq := ZoneRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return err
	}

	if len(opRegReq.Zone) == 0 {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := p.ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !fedcommon.FederationRoleExists(partnerFed, fedcommon.RoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
	}

	db := p.loggedDB(ctx)
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

// Remote partner OP sends this request to us to share its zone.
// It is triggered by partner OP when it has a new zone available
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

	_, partnerFed, err := p.ValidateAndGetFederationinfo(
		ctx,
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
	)
	if err != nil {
		return err
	}

	if !fedcommon.FederationRoleExists(partnerFed, fedcommon.RoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
	}

	db := p.loggedDB(ctx)
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

// Remote partner OP sends this request to us to unshare its zone.
// This api is triggered when the remote partner OP decides to unshare
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

	_, partnerFed, err := p.ValidateAndGetFederationinfo(
		ctx,
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
	)
	if err != nil {
		return err
	}

	if !fedcommon.FederationRoleExists(partnerFed, fedcommon.RoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
	}

	db := p.loggedDB(ctx)
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

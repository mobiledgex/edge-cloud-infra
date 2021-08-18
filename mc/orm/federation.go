package orm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var FederationTypeSelf = "self"
var FederationTypePartner = "partner"

type ZoneStatus int

var (
	ZoneStatusNone     ZoneStatus = 0
	ZoneStatusRegister ZoneStatus = 1
)

const (
	// Federation APIs
	F_API_OPERATOR_PARTNER = "/operator/partner"
	F_API_OPERATOR_ZONE    = "/operator/zone"
)

func CreateSelfFederation(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return bindErr(err)
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
	mncs := strings.Split(opFed.MNCs, ",")
	if len(mncs) == 0 {
		return fmt.Errorf("Missing MNCs")
	}
	// ensure that country code is a valid region
	_, err = getControllerObj(ctx, opFed.CountryCode)
	if err != nil {
		return fmt.Errorf("Invalid country code specified: %s, %v", opFed.CountryCode, err)
	}
	// ensure that operator ID is a valid operator org
	org, err := orgExists(ctx, opFed.OperatorId)
	if err != nil {
		return fmt.Errorf("Invalid operator ID specified")
	}
	if org.Type != OrgTypeOperator {
		return fmt.Errorf("Invalid operator ID, must be a valid operator org")
	}
	if err := authorized(ctx, claims.Username, opFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	opFed.Type = FederationTypeSelf

	db := loggedDB(ctx)
	// ensure that federation is not created already
	existingObj := ormapi.OperatorFederation{}
	db.Where(&opFed).First(existingObj)
	if existingObj.FederationId != "" {
		return fmt.Errorf("Federation already exists for operator ID %s, country code %s", opFed.OperatorId, opFed.CountryCode)
	}

	fedKey := uuid.New().String()
	opFed.FederationId = fedKey
	if err := db.Create(&opFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			// UUID collision
			return fmt.Errorf("Federation key collision for operator ID %s, country code %s. Please retry again", opFed.OperatorId, opFed.CountryCode)
		}
	}

	opFedOut := ormapi.OperatorFederation{
		FederationId: fedKey,
	}
	return c.JSON(http.StatusOK, &opFedOut)
}

func getSelfFederationInfo(ctx context.Context) (*ormapi.OperatorFederation, error) {
	// get self federation information
	db := loggedDB(ctx)
	selfFed := ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: FederationTypeSelf,
	}
	err := db.Where(&lookup).First(&selfFed).Error
	if err != nil {
		return nil, fmt.Errorf("Self federation doesn't exist")
	}
	return &selfFed, nil
}

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

func CreatePartnerFederation(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return bindErr(err)
	}
	// sanity check
	if opFed.OperatorId == "" {
		return fmt.Errorf("Missing Operator ID")
	}
	if opFed.CountryCode == "" {
		return fmt.Errorf("Missing country code")
	}
	if opFed.FederationId == "" {
		return fmt.Errorf("Missing partner federation key")
	}
	if opFed.FederationAddr == "" {
		return fmt.Errorf("Missing partner federation access address")
	}
	opFed.Type = FederationTypePartner

	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}

	if err := authorized(ctx, claims.Username, selfFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}

	// call REST API /operator/partner
	opRegReq := ormapi.OPRegistrationRequest{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: opFed.FederationId,
		OperatorId:       selfFed.OperatorId,
		CountryCode:      selfFed.CountryCode,
	}
	opRegRes := ormapi.OPRegistrationResponse{}
	err = sendFederationRequest("POST", opFed.FederationAddr, F_API_OPERATOR_PARTNER, &opRegReq, &opRegRes)
	if err != nil {
		return err
	}
	opFed.MCC = opRegRes.MCC
	opFed.MNCs = strings.Join(opRegRes.MNC, ",")
	opFed.LocatorEndPoint = opRegRes.LocatorEndPoint
	db := loggedDB(ctx)
	if err := db.Create(&opFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s", opFed.OperatorId, opFed.CountryCode)
		}
	}

	// Store partner zones in DB
	for _, partnerZone := range opRegRes.PartnerZone {
		zoneObj := ormapi.OperatorZone{}
		zoneObj.FederationId = opRegRes.OrigFederationId
		zoneObj.ZoneId = partnerZone.ZoneId
		zoneObj.GeoLocation = partnerZone.GeoLocation
		zoneObj.City = partnerZone.City
		zoneObj.Locality = partnerZone.Locality
		if err := db.Create(&zoneObj).Error; err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
				return fmt.Errorf("ZoneId %s already exists", zoneObj.ZoneId)
			}
		}
	}

	return setReply(c, Msg("Added partner OP successfully"))
}

func CreateFederationZone(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZone := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&opZone); err != nil {
		return bindErr(err)
	}
	// sanity check
	if opZone.ZoneId == "" {
		return fmt.Errorf("Missing Zone ID")
	}
	if len(opZone.Cloudlets) == 0 {
		return fmt.Errorf("Missing cloudlets")
	}
	if opZone.GeoLocation == "" {
		return fmt.Errorf("Missing geo location")
	}
	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, selfFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		FederationId: selfFed.FederationId,
		ZoneId:       opZone.ZoneId,
	}
	existingFed := ormapi.OperatorZone{}
	db.Where(&lookup).First(&existingFed)
	if existingFed.ZoneId != "" {
		return fmt.Errorf("Zone %s already exists", opZone.ZoneId)
	}

	rc := RegionContext{
		region:    selfFed.CountryCode,
		username:  claims.Username,
		skipAuthz: true,
	}
	cloudletMap := make(map[string]edgeproto.Cloudlet)
	cloudletLookup := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: selfFed.OperatorId,
		},
	}
	authzCloudlet := AuthzCloudlet{}
	err = authzCloudlet.populate(ctx, rc.region, rc.username, selfFed.OperatorId, ResourceCloudlets, ActionView)
	if err != nil {
		return err
	}
	err = ShowCloudletStream(ctx, &rc, &cloudletLookup, func(cloudlet *edgeproto.Cloudlet) error {
		authzOk, filterOutput := authzCloudlet.Ok(cloudlet)
		if authzOk {
			if filterOutput {
				authzCloudlet.Filter(cloudlet)
			}
			cloudletMap[cloudlet.Key.Name] = *cloudlet
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, cl := range opZone.Cloudlets {
		if _, ok := cloudletMap[cl]; !ok {
			return fmt.Errorf("Cloudlet %s doesn't exist", cl)
		}
	}

	az := ormapi.OperatorZone{}
	az.FederationId = selfFed.FederationId
	az.ZoneId = opZone.ZoneId
	az.GeoLocation = opZone.GeoLocation
	az.State = opZone.State
	az.Locality = opZone.Locality
	if err := db.Create(&az).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone already exists for operator ID %s, country code %s", selfFed.OperatorId, selfFed.CountryCode)
		}
	}

	zCloudlet := ormapi.OperatorZoneCloudlet{}
	zCloudlet.ZoneId = opZone.ZoneId
	for _, cloudlet := range opZone.Cloudlets {
		zCloudlet.CloudletName = cloudlet
	}
	if err := db.Create(&zCloudlet).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Zone cloudlet already exists for operator ID %s, country code %s", selfFed.OperatorId, selfFed.CountryCode)
		}
	}

	return setReply(c, Msg("Created OP zone successfully"))
}

func ShowOperatorZone(ctx context.Context, opZoneReq *ormapi.OperatorZoneCloudletMap) ([]ormapi.OperatorZoneCloudletMap, error) {
	db := loggedDB(ctx)
	opZones := []ormapi.OperatorZone{}
	lookup := ormapi.OperatorZone{
		FederationId: opZoneReq.FederationId,
		ZoneId:       opZoneReq.ZoneId,
	}
	err := db.Where(&lookup).Find(&opZones).Error
	if err != nil {
		return nil, dbErr(err)
	}

	fedZones := []ormapi.OperatorZoneCloudletMap{}
	for _, opZone := range opZones {
		clLookup := ormapi.OperatorZoneCloudlet{
			ZoneId: opZone.ZoneId,
		}
		opCloudlets := []ormapi.OperatorZoneCloudlet{}
		err = db.Where(&clLookup).Find(&opCloudlets).Error
		if err != nil {
			return nil, dbErr(err)
		}

		zoneOut := ormapi.OperatorZoneCloudletMap{}
		zoneOut.FederationId = opZone.FederationId
		zoneOut.ZoneId = opZone.ZoneId
		zoneOut.GeoLocation = opZone.GeoLocation
		zoneOut.City = opZone.City
		zoneOut.State = opZone.State
		zoneOut.Locality = opZone.Locality
		zoneOut.Cloudlets = []string{}
		for _, opCl := range opCloudlets {
			zoneOut.Cloudlets = append(zoneOut.Cloudlets, opCl.CloudletName)
		}
		fedZones = append(fedZones, zoneOut)
	}

	return fedZones, nil
}

func ShowFederationZone(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZoneReq := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&opZoneReq); err != nil {
		return bindErr(err)
	}
	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, selfFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	fedZones, err := ShowOperatorZone(ctx, &opZoneReq)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fedZones)
}

func RegisterFederationZone(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&reg); err != nil {
		return bindErr(err)
	}
	// sanity check
	if reg.ZoneId == "" {
		return fmt.Errorf("Must specify zone ID")
	}
	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, selfFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: reg.ZoneId,
	}
	existingFed := ormapi.OperatorZone{}
	db.Where(&lookup).First(&existingFed)
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", reg.ZoneId)
	}
	if existingFed.FederationId == selfFed.FederationId {
		return fmt.Errorf("Cannot register self zones, only federated zones are allowed to be registered")
	}
	if existingFed.Status == int(ZoneStatusRegister) {
		return fmt.Errorf("Zone %s is already registered", reg.ZoneId)
	}

	// get federation information
	fedInfo := ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: existingFed.FederationId,
		Type:         FederationTypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingFed.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := ormapi.OPZoneRegister{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zones:            []string{existingFed.ZoneId},
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingFed.FederationId,
	}
	opZoneRes := ormapi.OPZoneRegisterResponse{}
	err = sendFederationRequest("POST", fedInfo.FederationAddr, F_API_OPERATOR_ZONE, &opZoneReg, &opZoneRes)
	if err != nil {
		return err
	}

	// Mark zone as registered in DB
	existingFed.Status = int(ZoneStatusRegister)
	if err := db.Model(&existingFed).Updates(&existingFed).Error; err != nil {
		return dbErr(err)
	}
	return setReply(c, Msg("Partner OP zone registered successfully"))
}

func DeRegisterFederationZone(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&reg); err != nil {
		return bindErr(err)
	}
	// sanity check
	if reg.ZoneId == "" {
		return fmt.Errorf("Must specify zone ID")
	}
	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}
	if err := authorized(ctx, claims.Username, selfFed.OperatorId,
		ResourceCloudlets, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: reg.ZoneId,
	}
	existingFed := ormapi.OperatorZone{}
	db.Where(&lookup).First(&existingFed)
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", reg.ZoneId)
	}
	if existingFed.FederationId == selfFed.FederationId {
		return fmt.Errorf("Cannot register self zones, only federated zones are allowed to be registered")
	}
	if existingFed.Status == int(ZoneStatusNone) {
		return fmt.Errorf("Zone %s is already deregistered", reg.ZoneId)
	}

	// get federation information
	fedInfo := ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: existingFed.FederationId,
		Type:         FederationTypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingFed.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := ormapi.OPZoneDeRegister{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zone:             existingFed.ZoneId,
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingFed.FederationId,
	}
	err = sendFederationRequest("DELETE", fedInfo.FederationAddr, F_API_OPERATOR_ZONE, &opZoneReg, nil)
	if err != nil {
		return err
	}

	// Mark zone as deregistered in DB
	existingFed.Status = int(ZoneStatusNone)
	if err := db.Model(&existingFed).Updates(&existingFed).Error; err != nil {
		return dbErr(err)
	}
	return setReply(c, Msg("Partner OP zone deregistered successfully"))
}

func VerifyFederationAccess(ctx context.Context, selfFed *ormapi.OperatorFederation, origKey, destKey, operatorId, countryCode string) error {
	// sanity check
	if origKey == "" {
		return fmt.Errorf("Missing origin federation key")
	}
	if destKey == "" {
		return fmt.Errorf("Missing destination federation key")
	}
	if operatorId == "" {
		return fmt.Errorf("Missing self Operator ID")
	}
	if countryCode == "" {
		return fmt.Errorf("Missing self country code")
	}
	// validate destination federation key
	if selfFed.FederationId != destKey {
		return fmt.Errorf("Invalid destination federation key")
	}
	return nil
}

func FederationOperatorPartner(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}

	err = VerifyFederationAccess(
		ctx,
		selfFed,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.OperatorId,
		opRegReq.CountryCode,
	)
	if err != nil {
		return err
	}

	// Get list of zones to be shared to partner OP
	opZoneReq := &ormapi.OperatorZoneCloudletMap{
		FederationId: selfFed.FederationId,
	}
	opZones, err := ShowOperatorZone(ctx, opZoneReq)
	if err != nil {
		return err
	}
	out := ormapi.OPRegistrationResponse{}
	out.OrigOperatorId = selfFed.OperatorId
	out.PartnerOperatorId = opRegReq.OperatorId
	out.OrigFederationId = selfFed.FederationId
	out.DestFederationId = opRegReq.OrigFederationId
	out.MCC = selfFed.MCC
	out.MNC = strings.Split(selfFed.MNCs, ",")
	out.LocatorEndPoint = selfFed.LocatorEndPoint
	for _, opZone := range opZones {
		partnerZone := ormapi.OPZoneInfo{
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
	db := loggedDB(ctx)
	partnerOP := ormapi.OperatorFederation{
		FederationId: opRegReq.OrigFederationId,
		OperatorId:   opRegReq.OperatorId,
		CountryCode:  opRegReq.CountryCode,
		Type:         FederationTypePartner,
	}
	if err := db.Create(&partnerOP).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s", partnerOP.OperatorId, partnerOP.CountryCode)
		}
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, out)
}

func FederationOperatorZoneRegister(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPZoneRegister{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	if len(opRegReq.Zones) != 0 {
		return fmt.Errorf("Must specify one zone ID")
	}

	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}

	err = VerifyFederationAccess(
		ctx,
		selfFed,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	resp := ormapi.OPZoneRegisterResponse{}
	zoneId := opRegReq.Zones[0]
	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: zoneId,
	}
	existingFed := ormapi.OperatorZone{}
	db.Where(&lookup).First(&existingFed)
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", zoneId)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot register partner zones, only self zones are allowed to be registered")
	}
	if existingFed.Status != int(ZoneStatusRegister) {
		// Mark zone as registered in DB
		existingFed.Status = int(ZoneStatusRegister)
		if err := db.Model(&existingFed).Updates(&existingFed).Error; err != nil {
			return dbErr(err)
		}
	}
	resp.LeadOperatorId = selfFed.FederationId
	resp.PartnerOperatorId = opRegReq.Operator
	resp.FederationId = opRegReq.OrigFederationId
	resp.Zone = ormapi.OPZoneRegisterDetails{
		ZoneId:            zoneId,
		RegistrationToken: selfFed.FederationId,
	}

	// Share zone registration details
	return c.JSON(http.StatusOK, resp)
}

func FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPZoneDeRegister{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	if len(opRegReq.Zone) != 0 {
		return fmt.Errorf("Must specify zone ID")
	}

	// get self federation information
	selfFed, err := getSelfFederationInfo(ctx)
	if err != nil {
		return err
	}

	err = VerifyFederationAccess(
		ctx,
		selfFed,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opRegReq.Zone,
	}
	existingFed := ormapi.OperatorZone{}
	db.Where(&lookup).First(&existingFed)
	if existingFed.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", opRegReq.Zone)
	}
	if existingFed.FederationId != selfFed.FederationId {
		return fmt.Errorf("Cannot deregister partner zones, only self zones are allowed to be deregistered")
	}
	if existingFed.Status != int(ZoneStatusNone) {
		// Mark zone as deregistered in DB
		existingFed.Status = int(ZoneStatusNone)
		if err := db.Model(&existingFed).Updates(&existingFed).Error; err != nil {
			return dbErr(err)
		}
	}
	return c.JSON(http.StatusOK, "Zone deregistered successfully")
}

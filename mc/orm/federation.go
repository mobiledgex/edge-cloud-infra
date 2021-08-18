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
	F_API_OPERATOR_PARTNER     = "/operator/partner"
	F_API_OPERATOR_ZONE        = "/operator/zone"
	F_API_OPERATOR_NOTIFY_ZONE = "/operator/notify/zone"
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

func UpdatePartnerFederation(c echo.Context) error {
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
	if opFed.FederationId == "" {
		return fmt.Errorf("Missing partner federation key")
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

	db := loggedDB(ctx)
	old := ormapi.OperatorFederation{}
	err = db.Where(&opFed).First(old).Error
	if err != nil {
		return dbErr(err)
	}

	update := false
	if opFed.MCC != old.MCC {
		update = true
		old.MCC = opFed.MCC
	}
	if opFed.MNCs != old.MNCs {
		update = true
		old.MNCs = opFed.MNCs
	}
	if opFed.LocatorEndPoint != old.LocatorEndPoint {
		update = true
		old.LocatorEndPoint = opFed.LocatorEndPoint
	}

	// call REST API /operator/partner
	opConf := ormapi.OPUpdateMECNetConf{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: opFed.FederationId,
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
		MCC:              old.MCC,
		MNC:              strings.Split(old.MNCs, ","),
		LocatorEndPoint:  old.LocatorEndPoint,
	}
	err = sendFederationRequest("PUT", opFed.FederationAddr, F_API_OPERATOR_PARTNER, &opConf, nil)
	if err != nil {
		return err
	}

	if update {
		err = db.Save(old).Error
		if err != nil {
			return dbErr(err)
		}
	}

	return setReply(c, Msg("Updated OP successfully"))
}

func DeletePartnerFederation(c echo.Context) error {
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
	if opFed.FederationId == "" {
		return fmt.Errorf("Missing partner federation key")
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

	db := loggedDB(ctx)
	err = db.Where(&opFed).First(opFed).Error
	if err != nil {
		return dbErr(err)
	}

	// call REST API /operator/partner
	opFedReq := ormapi.OPFederationRequest{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: opFed.FederationId,
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
	}
	err = sendFederationRequest("DELETE", opFed.FederationAddr, F_API_OPERATOR_PARTNER, &opFedReq, nil)
	if err != nil {
		return err
	}

	// Delete all the partner gMEC zones
	lookup := ormapi.OperatorZone{
		FederationId: opFed.FederationId,
	}
	partnerZones := []ormapi.OperatorZone{}
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return dbErr(err)
	}
	for _, pZone := range partnerZones {
		if err := db.Delete(pZone).Error; err != nil {
			// ignore err
			continue
		}
	}

	// Delete partner gMEC
	if err := db.Delete(&opFed).Error; err != nil {
		return dbErr(err)
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
	res := db.Where(&lookup).First(&existingFed)
	if !res.RecordNotFound() && res.Error != nil {
		return dbErr(res.Error)
	}
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

func DeleteFederationZone(c echo.Context) error {
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
		ZoneId: opZone.ZoneId,
	}
	existingZone := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return dbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s does not exist", opZone.ZoneId)
	}
	if existingZone.FederationId != selfFed.FederationId {
		return fmt.Errorf("Can only delete self zones")
	}

	if err := db.Delete(&existingZone).Error; err != nil {
		return dbErr(err)
	}

	zCloudlet := ormapi.OperatorZoneCloudlet{}
	zCloudlet.ZoneId = opZone.ZoneId
	if err := db.Delete(&zCloudlet).Error; err != nil {
		return dbErr(err)
	}

	return setReply(c, Msg("Deleted zone successfully"))
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
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return dbErr(err)
	}
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
	opZoneReg := ormapi.OPZoneRequest{
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

// Federation APIs
// ===============

// Create directed federation with partner gMEC. By Federation create request,
// the API initiator gMEC say ‘A’ request gMEC 'B' to allow its developers/subscribers
// to run their application on edge sites of gMEC 'B'
func FederationOperatorPartnerCreate(c echo.Context) error {
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
		MCC:          selfFed.MCC,
		MNCs:         selfFed.MNCs,
	}
	if err := db.Create(&partnerOP).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s", partnerOP.OperatorId, partnerOP.CountryCode)
		}
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, out)
}

// Federation Agent sends this request to partner gMEC federation
// Agent to update its MNC, MCC or locator URL
func FederationOperatorPartnerUpdate(c echo.Context) error {
	ctx := GetContext(c)
	opConf := ormapi.OPUpdateMECNetConf{}
	if err := c.Bind(&opConf); err != nil {
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
		opConf.OrigFederationId,
		opConf.DestFederationId,
		opConf.Operator,
		opConf.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorFederation{
		FederationId: opConf.OrigFederationId,
	}
	opFed := &ormapi.OperatorFederation{}
	err = db.Where(&lookup).First(opFed).Error
	if err != nil {
		return dbErr(err)
	}

	save := false
	if opConf.MCC != "" {
		opFed.MCC = opConf.MCC
		save = true
	}
	if len(opConf.MNC) > 0 {
		opFed.MNCs = strings.Join(opConf.MNC, ",")
		save = true
	}
	if opConf.LocatorEndPoint != "" {
		opFed.LocatorEndPoint = opConf.LocatorEndPoint
		save = true
	}

	if !save {
		return fmt.Errorf("Nothing to update")
	}

	err = db.Save(opFed).Error
	if err != nil {
		return dbErr(err)
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, "Federation attributes updated successfully")
}

// Remove existing federation with a partner gMEC. By Federation delete
// request, the API initiator gMEC say A is requesting to the partner
// gMEC B to disallow A applications access to gMEC B edges
func FederationOperatorPartnerDelete(c echo.Context) error {
	ctx := GetContext(c)
	opFedReq := ormapi.OPFederationRequest{}
	if err := c.Bind(&opFedReq); err != nil {
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
		opFedReq.OrigFederationId,
		opFedReq.DestFederationId,
		opFedReq.Operator,
		opFedReq.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorFederation{
		FederationId: opFedReq.OrigFederationId,
	}
	opFed := &ormapi.OperatorFederation{}
	err = db.Where(&lookup).First(opFed).Error
	if err != nil {
		return dbErr(err)
	}

	if err := db.Delete(&opFed).Error; err != nil {
		return dbErr(err)
	}
	return setReply(c, Msg("Deleted partner OP successfully"))
}

// gMEC platform sends this request to partner gMEC, to register a
// partner gMEC zone. It is only after successful registration that
// partner gMEC allow access to its zones. This api shall be triggered
// for each partner zone
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
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return dbErr(err)
	}
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

// Deregister a partner gMEC zone. By zone deregistration request,
// the API initiator gMEC say A is indicating to the partner gMEC
// say B, that it will no longer access partner gMEC zone
func FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPZoneRequest{}
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
	err = db.Where(&lookup).First(&existingFed).Error
	if err != nil {
		return dbErr(err)
	}
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

// gMEC notifies its partner MEcs whenever it has a new zone.
// This api is triggered when gMEC has a new zone available and
// it wishes to share the zone with its partner gMEC. This request
// is triggered only when federation already exists
func FederationOperatorZoneShare(c echo.Context) error {
	ctx := GetContext(c)
	opZoneShare := ormapi.OPZoneNotify{}
	if err := c.Bind(&opZoneShare); err != nil {
		return bindErr(err)
	}

	if opZoneShare.PartnerZone.ZoneId == "" {
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
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opZoneShare.PartnerZone.ZoneId,
	}
	existingZone := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return dbErr(err)
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
	}

	return setReply(c, Msg("Added zone successfully"))
}

// gMEC notifies its partner MECs whenever it unshares a zone.
// This api is triggered when a gMEC decides to unshare one of its
// zone with one of the federated partners. This is used when
// federation already exists
func FederationOperatorZoneUnShare(c echo.Context) error {
	ctx := GetContext(c)
	opZone := ormapi.OPZoneRequest{}
	if err := c.Bind(&opZone); err != nil {
		return bindErr(err)
	}

	if opZone.Zone == "" {
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
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opZone.Zone,
	}
	existingZone := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return dbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s does not exist", opZone.Zone)
	}
	if err := db.Delete(&lookup).Error; err != nil {
		return dbErr(err)
	}

	return setReply(c, Msg("Deleted zone successfully"))
}

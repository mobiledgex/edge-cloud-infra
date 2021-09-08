package orm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const (
	// Federation Types
	FederationTypeSelf    = "self"
	FederationTypePartner = "partner"

	// Federation Partner Roles
	FederationRoleAccessZones = "access" // Can access partner OP zones, but cannot share zones with partner OP
	FederationRoleShareZones  = "share"  // Can only share zones with partner OP, but cannot access partner OP's zones

	// Federation APIs
	F_API_OPERATOR_PARTNER     = "/operator/partner"
	F_API_OPERATOR_ZONE        = "/operator/zone"
	F_API_OPERATOR_NOTIFY_ZONE = "/operator/notify/zone"
)

// Manage federation roles
func AddFederationRole(fedObj *ormapi.OperatorFederation, inRole string) error {
	if fedObj.Role == "" {
		fedObj.Role = inRole
		return nil
	}
	roles := strings.Split(fedObj.Role, ",")
	for _, role := range roles {
		if role == inRole {
			// role already present
			return nil
		}
	}
	roles = append(roles, inRole)
	fedObj.Role = strings.Join(roles, ",")
	return nil
}

func RemoveFederationRole(fedObj *ormapi.OperatorFederation, inRole string) error {
	roles := strings.Split(fedObj.Role, ",")
	for ii, role := range roles {
		if role == inRole {
			roles = append(roles[:ii], roles[ii+1:]...)
			break
		}
	}
	fedObj.Role = strings.Join(roles, ",")
	return nil
}

func FederationRoleExists(fedObj *ormapi.OperatorFederation, inRole string) bool {
	roles := strings.Split(fedObj.Role, ",")
	roleMap := make(map[string]struct{})
	for _, role := range roles {
		roleMap[role] = struct{}{}
	}
	_, matchRes := roleMap[inRole]
	return matchRes
}

func CreateFederation(c echo.Context) error {
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
	opFed.FederationAddr = serverConfig.FederationAddr
	if err := db.Create(&opFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			// UUID collision
			return fmt.Errorf("Federation key collision for operator ID %s, country code %s. Please retry again", opFed.OperatorId, opFed.CountryCode)
		}
		return dbErr(err)
	}

	opFedOut := ormapi.OperatorFederation{
		FederationId: fedKey,
	}
	return c.JSON(http.StatusOK, &opFedOut)
}

func UpdateFederation(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
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

	db := loggedDB(ctx)

	update := false
	if opFed.MCC != selfFed.MCC {
		update = true
		selfFed.MCC = opFed.MCC
	}
	if opFed.MNCs != selfFed.MNCs {
		update = true
		selfFed.MNCs = opFed.MNCs
	}
	if opFed.LocatorEndPoint != selfFed.LocatorEndPoint {
		update = true
		selfFed.LocatorEndPoint = opFed.LocatorEndPoint
	}
	if update {
		err = db.Save(selfFed).Error
		if err != nil {
			return dbErr(err)
		}
	}

	partnerLookup := ormapi.OperatorFederation{
		Type: FederationTypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return dbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones are shared
		if !FederationRoleExists(&partnerOP, FederationRoleShareZones) {
			continue
		}
		// Notify partner OP about the update
		opConf := ormapi.OPUpdateMECNetConf{
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			MCC:              selfFed.MCC,
			MNC:              strings.Split(selfFed.MNCs, ","),
			LocatorEndPoint:  selfFed.LocatorEndPoint,
		}
		err = sendFederationRequest("PUT", partnerOP.FederationAddr, F_API_OPERATOR_PARTNER, &opConf, nil)
		if err != nil {
			return err
		}
	}

	return setReply(c, Msg("Updated OP successfully"))
}

func DeleteFederation(c echo.Context) error {
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
		return fmt.Errorf("Missing Federation ID")
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

	// ensure that no partner OP exists
	partnerLookup := ormapi.OperatorFederation{
		Type: FederationTypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return dbErr(err)
	}
	if len(partnerOPs) > 0 {
		return fmt.Errorf("Cannot delete federation as there are multiple partner OPs associated with it. Please remove all partner federations before deleting the federation")
	}

	opFed.Type = FederationTypeSelf
	if err := db.Delete(&opFed).Error; err != nil {
		return dbErr(err)
	}

	return setReply(c, Msg("Deleted federation successfully"))
}

func getSelfFederationInfo(ctx context.Context) (*ormapi.OperatorFederation, error) {
	// get self federation information
	db := loggedDB(ctx)
	selfFed := ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: FederationTypeSelf,
	}
	res := db.Where(&lookup).First(&selfFed)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federation doesn't exist")
	}
	if res.Error != nil {
		return nil, dbErr(res.Error)
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

func AddFederationPartner(c echo.Context) error {
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
		OrigFederationId:   selfFed.FederationId,
		DestFederationId:   opFed.FederationId,
		OperatorId:         selfFed.OperatorId,
		CountryCode:        selfFed.CountryCode,
		OrigFederationAddr: selfFed.FederationAddr,
	}
	opRegRes := ormapi.OPRegistrationResponse{}
	err = sendFederationRequest("POST", opFed.FederationAddr, F_API_OPERATOR_PARTNER, &opRegReq, &opRegRes)
	if err != nil {
		return err
	}
	opFed.Type = FederationTypePartner
	err = AddFederationRole(&opFed, FederationRoleAccessZones)
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
		return dbErr(err)
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
			return dbErr(err)
		}
	}

	return setReply(c, Msg("Added partner OP successfully"))
}

func RemoveFederationPartner(c echo.Context) error {
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
	err = db.Where(&opFed).First(&opFed).Error
	if err != nil {
		return dbErr(err)
	}

	// Check if all the zones are deregistered
	lookup := ormapi.OperatorZone{
		FederationId: opFed.FederationId,
	}
	partnerZones := []ormapi.OperatorZone{}
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return dbErr(err)
	}
	for _, pZone := range partnerZones {
		regLookup := ormapi.OperatorRegisteredZone{
			ZoneId:       pZone.ZoneId,
			FederationId: opFed.FederationId,
		}
		regZone := ormapi.OperatorRegisteredZone{}
		res := db.Where(&regLookup).First(&regZone)
		if !res.RecordNotFound() && res.Error != nil {
			return dbErr(res.Error)
		}
		if regZone.ZoneId != "" {
			return fmt.Errorf("Cannot remove federation partner as partner zone %s is registered locally. Please deregister it before removing the federation partner", regZone.ZoneId)
		}
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

	// Delete all the local copy of partner OP zones
	for _, pZone := range partnerZones {
		if err := db.Delete(pZone).Error; err != nil {
			// ignore err
			continue
		}
	}

	// Delete partner OP
	if err := db.Delete(&opFed).Error; err != nil {
		return dbErr(err)
	}

	return setReply(c, Msg("Added partner OP successfully"))
}

func ShowFederation(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
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

	db := loggedDB(ctx)
	feds := []ormapi.OperatorFederation{}
	err = db.Where(&opFed).Find(&feds).Error
	if err != nil {
		return dbErr(err)
	}
	for ii, _ := range feds {
		// Do not display federation ID
		feds[ii].FederationId = ""
	}
	return c.JSON(http.StatusOK, feds)
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
		return dbErr(err)
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
		return dbErr(err)
	}

	// Share zone details with partner OPs
	zoneInfo := ormapi.OPZoneInfo{
		ZoneId:      opZone.ZoneId,
		GeoLocation: opZone.GeoLocation,
		City:        opZone.City,
		State:       opZone.State,
		Locality:    opZone.Locality,
		EdgeCount:   len(opZone.Cloudlets),
	}
	partnerLookup := ormapi.OperatorFederation{
		Type: FederationTypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return dbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones are shared
		if !FederationRoleExists(&partnerOP, FederationRoleShareZones) {
			continue
		}
		// Notify federated partner about new zone
		opZoneShare := ormapi.OPZoneNotify{
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			PartnerZone:      zoneInfo,
		}
		err = sendFederationRequest("POST", partnerOP.FederationAddr, F_API_OPERATOR_NOTIFY_ZONE, &opZoneShare, nil)
		if err != nil {
			return err
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

	regLookup := ormapi.OperatorRegisteredZone{
		ZoneId:       opZone.ZoneId,
		FederationId: existingZone.FederationId,
	}
	regZone := ormapi.OperatorRegisteredZone{}
	res := db.Where(&regLookup).First(&regZone)
	if !res.RecordNotFound() && res.Error != nil {
		return dbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %s as it registered by partner OP. Please deregister it before deleting it", regZone.ZoneId)
	}

	if err := db.Delete(&existingZone).Error; err != nil {
		return dbErr(err)
	}

	zCloudlet := ormapi.OperatorZoneCloudlet{}
	zCloudlet.ZoneId = opZone.ZoneId
	if err := db.Delete(&zCloudlet).Error; err != nil {
		return dbErr(err)
	}

	// Notify deleted zone details with partner OPs
	partnerLookup := ormapi.OperatorFederation{
		Type: FederationTypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return dbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones were shared
		if !FederationRoleExists(&partnerOP, FederationRoleShareZones) {
			continue
		}
		// Notify federated partner about deleted zone
		opZoneUnShare := ormapi.OPZoneRequest{
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			Zone:             opZone.ZoneId,
		}
		err = sendFederationRequest("DELETE", partnerOP.FederationAddr, F_API_OPERATOR_NOTIFY_ZONE, &opZoneUnShare, nil)
		if err != nil {
			return err
		}
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
		opRegZones := []ormapi.OperatorRegisteredZone{}
		regLookup := ormapi.OperatorRegisteredZone{
			FederationId: opZoneReq.FederationId,
			ZoneId:       opZoneReq.ZoneId,
		}
		res := db.Where(&regLookup).Find(&opRegZones)
		if !res.RecordNotFound() && res.Error != nil {
			return nil, dbErr(res.Error)
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
	existingZone := ormapi.OperatorZone{}
	err = db.Where(&lookup).First(&existingZone).Error
	if err != nil {
		return dbErr(err)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", reg.ZoneId)
	}
	if existingZone.FederationId == selfFed.FederationId {
		return fmt.Errorf("Cannot register self zones, only federated zones are allowed to be registered")
	}

	// get federation information
	fedInfo := ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: existingZone.FederationId,
		Type:         FederationTypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingZone.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := ormapi.OPZoneRegister{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zones:            []string{existingZone.ZoneId},
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingZone.FederationId,
	}
	opZoneRes := ormapi.OPZoneRegisterResponse{}
	err = sendFederationRequest("POST", fedInfo.FederationAddr, F_API_OPERATOR_ZONE, &opZoneReg, &opZoneRes)
	if err != nil {
		return err
	}

	// Mark zone as registered in DB
	regZone := ormapi.OperatorRegisteredZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: fedInfo.FederationId,
		OperatorId:   fedInfo.OperatorId,
		CountryCode:  fedInfo.CountryCode,
	}
	if err := db.Create(&regZone).Error; err != nil {
		if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return dbErr(err)
		}
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
	existingZone := ormapi.OperatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && res.Error != nil {
		return dbErr(res.Error)
	}
	if existingZone.ZoneId == "" {
		return fmt.Errorf("Zone %s not found", reg.ZoneId)
	}
	if existingZone.FederationId == selfFed.FederationId {
		return fmt.Errorf("Cannot deregister self zones, only federated zones are allowed to be deregistered")
	}

	// get federation information
	fedInfo := ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: existingZone.FederationId,
		Type:         FederationTypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingZone.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := ormapi.OPZoneRequest{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zone:             existingZone.ZoneId,
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingZone.FederationId,
	}
	err = sendFederationRequest("DELETE", fedInfo.FederationAddr, F_API_OPERATOR_ZONE, &opZoneReg, nil)
	if err != nil {
		return err
	}

	// Mark zone as deregistered in DB
	deregZone := ormapi.OperatorRegisteredZone{
		ZoneId:       existingZone.ZoneId,
		FederationId: existingZone.FederationId,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err != gorm.ErrRecordNotFound {

			return dbErr(err)
		}
	}
	return setReply(c, Msg("Partner OP zone deregistered successfully"))
}

func ValidateSelfAndGetFederationInfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, error) {
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
	db := loggedDB(ctx)
	// validate destination federation key
	selfFed := ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: FederationTypeSelf,
	}
	res := db.Where(&lookup).First(&selfFed)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federation doesn't exist")
	}
	if res.Error != nil {
		return nil, dbErr(res.Error)
	}
	if selfFed.FederationId != destKey {
		return nil, fmt.Errorf("Invalid destination federation key")
	}
	return &selfFed, nil
}

func ValidateAndGetFederationinfo(ctx context.Context, origKey, destKey, operatorId, countryCode string) (*ormapi.OperatorFederation, *ormapi.OperatorFederation, error) {
	selfFed, err := ValidateSelfAndGetFederationInfo(ctx, origKey, destKey, operatorId, countryCode)
	if err != nil {
		return nil, nil, err
	}
	db := loggedDB(ctx)
	// validate origin federationkey/operator/country
	partnerFed := ormapi.OperatorFederation{}
	fedLookup := ormapi.OperatorFederation{
		FederationId: origKey,
		Type:         FederationTypePartner,
	}
	res := db.Where(&fedLookup).First(&partnerFed)
	if res.RecordNotFound() {
		return nil, nil, fmt.Errorf("Origin federation doesn't exist")
	}
	if res.Error != nil {
		return nil, nil, dbErr(res.Error)
	}
	if partnerFed.OperatorId != operatorId {
		return nil, nil, fmt.Errorf("Invalid origin operator ID %s", operatorId)
	}
	if partnerFed.CountryCode != countryCode {
		return nil, nil, fmt.Errorf("Invalid origin country code %s", countryCode)
	}
	return selfFed, &partnerFed, nil
}

// Federation APIs
// ===============

// Create directed federation with partner OP. By Federation create request,
// the API initiator OP say ‘A’ request OP 'B' to allow its developers/subscribers
// to run their application on edge sites of OP 'B'
func FederationOperatorPartnerCreate(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPRegistrationRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	selfFed, err := ValidateSelfAndGetFederationInfo(
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
		FederationId:   opRegReq.OrigFederationId,
		OperatorId:     opRegReq.OperatorId,
		CountryCode:    opRegReq.CountryCode,
		Type:           FederationTypePartner,
		FederationAddr: opRegReq.OrigFederationAddr,
	}
	fedLookup := ormapi.OperatorFederation{
		FederationId: opRegReq.OrigFederationId,
	}
	existingPartnerOP := ormapi.OperatorFederation{}
	err = db.Where(&fedLookup).First(&existingPartnerOP).Error
	if err == nil {
		if FederationRoleExists(&existingPartnerOP, FederationRoleShareZones) {
			return fmt.Errorf("Partner OP is already federated")
		} else {
			err = AddFederationRole(&existingPartnerOP, FederationRoleShareZones)
			if err != nil {
				return err
			}
			err = db.Save(existingPartnerOP).Error
			if err != nil {
				return dbErr(err)
			}
		}
	} else {
		err = AddFederationRole(&partnerOP, FederationRoleShareZones)
		if err != nil {
			return err
		}
		if err := db.Create(&partnerOP).Error; err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
				return fmt.Errorf("Partner federation already exists for operator ID %s, country code %s", partnerOP.OperatorId, partnerOP.CountryCode)
			}
			return dbErr(err)
		}
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, out)
}

// Federation Agent sends this request to partner OP federation
// Agent to update its MNC, MCC or locator URL
func FederationOperatorPartnerUpdate(c echo.Context) error {
	ctx := GetContext(c)
	opConf := ormapi.OPUpdateMECNetConf{}
	if err := c.Bind(&opConf); err != nil {
		return bindErr(err)
	}

	_, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opConf.OrigFederationId,
		opConf.DestFederationId,
		opConf.Operator,
		opConf.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
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
		return dbErr(err)
	}

	// Share list of zones to be shared
	return c.JSON(http.StatusOK, "Federation attributes of partner OP updated successfully")
}

// Remove existing federation with a partner OP. By Federation delete
// request, the API initiator OP say A is requesting to the partner
// OP B to disallow A applications access to OP B edges
func FederationOperatorPartnerDelete(c echo.Context) error {
	ctx := GetContext(c)
	opFedReq := ormapi.OPFederationRequest{}
	if err := c.Bind(&opFedReq); err != nil {
		return bindErr(err)
	}

	_, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opFedReq.OrigFederationId,
		opFedReq.DestFederationId,
		opFedReq.Operator,
		opFedReq.Country,
	)
	if err != nil {
		return err
	}

	db := loggedDB(ctx)
	if !FederationRoleExists(partnerFed, FederationRoleShareZones) {
		return fmt.Errorf("No zones are shared with this OP")
	}
	err = RemoveFederationRole(partnerFed, FederationRoleShareZones)
	if err != nil {
		return err
	}
	if partnerFed.Role == "" {
		if err := db.Delete(partnerFed).Error; err != nil {
			return dbErr(err)
		}
	} else {
		err = db.Save(partnerFed).Error
		if err != nil {
			return dbErr(err)
		}
	}

	return setReply(c, Msg("Deleted partner OP successfully"))
}

// Operator platform sends this request to partner OP, to register a
// partner OP zone. It is only after successful registration that
// partner OP allow access to its zones. This api shall be triggered
// for each partner zone
func FederationOperatorZoneRegister(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPZoneRegister{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	if len(opRegReq.Zones) == 0 {
		return fmt.Errorf("Must specify one zone ID")
	}

	selfFed, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, FederationRoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
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
		return dbErr(err)
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

// Deregister a partner OP zone. By zone deregistration request,
// the API initiator OP say A is indicating to the partner OP
// say B, that it will no longer access partner OP zone
func FederationOperatorZoneDeRegister(c echo.Context) error {
	ctx := GetContext(c)
	opRegReq := ormapi.OPZoneRequest{}
	if err := c.Bind(&opRegReq); err != nil {
		return bindErr(err)
	}

	if len(opRegReq.Zone) == 0 {
		return fmt.Errorf("Must specify zone ID")
	}

	selfFed, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opRegReq.OrigFederationId,
		opRegReq.DestFederationId,
		opRegReq.Operator,
		opRegReq.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, FederationRoleShareZones) {
		return fmt.Errorf("No zones are shared with partner OP")
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
	deregZone := ormapi.OperatorRegisteredZone{
		ZoneId:       opRegReq.Zone,
		FederationId: opRegReq.OrigFederationId,
	}
	if err := db.Delete(&deregZone).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Zone %s is already deregistered for operator %s", opRegReq.Zone, opRegReq.Operator)
		}
		return dbErr(err)
	}
	return c.JSON(http.StatusOK, "Zone deregistered successfully")
}

// OP notifies its partner MEcs whenever it has a new zone.
// This api is triggered when OP has a new zone available and
// it wishes to share the zone with its partner OP. This request
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

	_, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opZoneShare.OrigFederationId,
		opZoneShare.DestFederationId,
		opZoneShare.Operator,
		opZoneShare.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, FederationRoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
	}

	db := loggedDB(ctx)
	lookup := ormapi.OperatorZone{
		ZoneId: opZoneShare.PartnerZone.ZoneId,
	}
	existingZone := ormapi.OperatorZone{}
	res := db.Where(&lookup).First(&existingZone)
	if !res.RecordNotFound() && res.Error != nil {
		return dbErr(res.Error)
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
		return dbErr(err)
	}

	return setReply(c, Msg("Added zone successfully"))
}

// OP notifies its partner MECs whenever it unshares a zone.
// This api is triggered when a OP decides to unshare one of its
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

	_, partnerFed, err := ValidateAndGetFederationinfo(
		ctx,
		opZone.OrigFederationId,
		opZone.DestFederationId,
		opZone.Operator,
		opZone.Country,
	)
	if err != nil {
		return err
	}

	if !FederationRoleExists(partnerFed, FederationRoleAccessZones) {
		return fmt.Errorf("OP does not have access to partner OP zones")
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

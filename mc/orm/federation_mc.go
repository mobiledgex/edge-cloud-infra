package orm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/federation"
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func CreateFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
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
	opFed.Type = fedcommon.TypeSelf

	db := loggedDB(ctx)
	fedKey := uuid.New().String()
	opFed.FederationId = fedKey
	opFed.FederationAddr = serverConfig.FederationAddr
	if err := db.Create(&opFed).Error; err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			// UUID collision
			return fmt.Errorf("Federation key collision for operator ID %s, country code %s. Please retry again", opFed.OperatorId, opFed.CountryCode)
		}
		return ormutil.DbErr(err)
	}

	opFedOut := ormapi.OperatorFederation{
		FederationId: fedKey,
	}
	return c.JSON(http.StatusOK, &opFedOut)
}

func UpdateFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
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
			return ormutil.DbErr(err)
		}
	}

	partnerLookup := ormapi.OperatorFederation{
		Type: fedcommon.TypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones are shared
		if !fedcommon.FederationRoleExists(&partnerOP, fedcommon.RoleShareZones) {
			continue
		}
		// Notify partner OP about the update
		opConf := federation.UpdateMECNetConf{
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			MCC:              selfFed.MCC,
			MNC:              strings.Split(selfFed.MNCs, ","),
			LocatorEndPoint:  selfFed.LocatorEndPoint,
		}
		err = sendFederationRequest("PUT", partnerOP.FederationAddr, federation.OperatorPartnerAPI, &opConf, nil)
		if err != nil {
			return err
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Updated OP successfully"))
}

func DeleteFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
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
		Type: fedcommon.TypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	if len(partnerOPs) > 0 {
		return fmt.Errorf("Cannot delete federation as there are multiple partner OPs associated with it. Please remove all partner federations before deleting the federation")
	}

	opFed.Type = fedcommon.TypeSelf
	if err := db.Delete(&opFed).Error; err != nil {
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted federation successfully"))
}

func getSelfFederationInfo(ctx context.Context) (*ormapi.OperatorFederation, error) {
	// get self federation information
	db := loggedDB(ctx)
	selfFed := ormapi.OperatorFederation{}
	lookup := ormapi.OperatorFederation{
		Type: fedcommon.TypeSelf,
	}
	res := db.Where(&lookup).First(&selfFed)
	if res.RecordNotFound() {
		return nil, fmt.Errorf("Self federation doesn't exist")
	}
	if res.Error != nil {
		return nil, ormutil.DbErr(res.Error)
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
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
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
	opRegReq := federation.OperatorRegistrationRequest{
		OrigFederationId:   selfFed.FederationId,
		DestFederationId:   opFed.FederationId,
		OperatorId:         selfFed.OperatorId,
		CountryCode:        selfFed.CountryCode,
		OrigFederationAddr: selfFed.FederationAddr,
	}
	opRegRes := federation.OperatorRegistrationResponse{}
	err = sendFederationRequest("POST", opFed.FederationAddr, federation.OperatorPartnerAPI, &opRegReq, &opRegRes)
	if err != nil {
		return err
	}
	opFed.Type = fedcommon.TypePartner
	err = fedcommon.AddFederationRole(&opFed, fedcommon.RoleAccessZones)
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
		return ormutil.DbErr(err)
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
			return ormutil.DbErr(err)
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Added partner OP successfully"))
}

func RemoveFederationPartner(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opFed.FederationId == "" {
		return fmt.Errorf("Missing partner federation key")
	}
	opFed.Type = fedcommon.TypePartner

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
		return ormutil.DbErr(err)
	}

	// Check if all the zones are deregistered
	lookup := ormapi.OperatorZone{
		FederationId: opFed.FederationId,
	}
	partnerZones := []ormapi.OperatorZone{}
	err = db.Where(&lookup).Find(&partnerZones).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, pZone := range partnerZones {
		regLookup := ormapi.OperatorRegisteredZone{
			ZoneId:       pZone.ZoneId,
			FederationId: opFed.FederationId,
		}
		regZone := ormapi.OperatorRegisteredZone{}
		res := db.Where(&regLookup).First(&regZone)
		if !res.RecordNotFound() && res.Error != nil {
			return ormutil.DbErr(res.Error)
		}
		if regZone.ZoneId != "" {
			return fmt.Errorf("Cannot remove federation partner as partner zone %s is registered locally. Please deregister it before removing the federation partner", regZone.ZoneId)
		}
	}

	// call REST API /operator/partner
	opFedReq := federation.FederationRequest{
		OrigFederationId: selfFed.FederationId,
		DestFederationId: opFed.FederationId,
		Operator:         selfFed.OperatorId,
		Country:          selfFed.CountryCode,
	}
	err = sendFederationRequest("DELETE", opFed.FederationAddr, federation.OperatorPartnerAPI, &opFedReq, nil)
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
		return ormutil.DbErr(err)
	}

	return ormutil.SetReply(c, ormutil.Msg("Added partner OP successfully"))
}

func ShowFederation(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opFed := ormapi.OperatorFederation{}
	if err := c.Bind(&opFed); err != nil {
		return ormutil.BindErr(err)
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
		return ormutil.DbErr(err)
	}
	for ii, _ := range feds {
		// Do not display federation ID
		feds[ii].FederationId = ""
	}
	return c.JSON(http.StatusOK, feds)
}

func CreateFederationZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZone := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&opZone); err != nil {
		return ormutil.BindErr(err)
	}
	// sanity check
	if opZone.ZoneId == "" {
		return fmt.Errorf("Missing Zone ID")
	}
	if len(opZone.Cloudlets) == 0 {
		return fmt.Errorf("Missing cloudlets")
	}
	if len(opZone.Cloudlets) > 1 {
		return fmt.Errorf("Only one cloudlet supported for now")
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
		return ormutil.DbErr(res.Error)
	}
	if existingFed.ZoneId != "" {
		return fmt.Errorf("Zone %s already exists", opZone.ZoneId)
	}

	rc := ormutil.RegionContext{
		Region:    selfFed.CountryCode,
		Username:  claims.Username,
		SkipAuthz: true,
	}
	cloudletMap := make(map[string]edgeproto.Cloudlet)
	cloudletLookup := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: selfFed.OperatorId,
		},
	}
	err = ctrlapi.ShowCloudletStream(ctx, &rc, &cloudletLookup, connCache, nil, func(cloudlet *edgeproto.Cloudlet) error {
		cloudletMap[cloudlet.Key.Name] = *cloudlet
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
			return fmt.Errorf("Zone with same zone ID %q already exists for operator ID %s, country code %s", az.ZoneId, selfFed.OperatorId, selfFed.CountryCode)
		}
		return ormutil.DbErr(err)
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
		return ormutil.DbErr(err)
	}

	// Share zone details with partner OPs
	zoneInfo := federation.ZoneInfo{
		ZoneId:      opZone.ZoneId,
		GeoLocation: opZone.GeoLocation,
		City:        opZone.City,
		State:       opZone.State,
		Locality:    opZone.Locality,
		EdgeCount:   len(opZone.Cloudlets),
	}
	partnerLookup := ormapi.OperatorFederation{
		Type: fedcommon.TypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones are shared
		if !fedcommon.FederationRoleExists(&partnerOP, fedcommon.RoleShareZones) {
			continue
		}
		// Notify federated partner about new zone
		opZoneShare := federation.NotifyPartnerOperatorZone{
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			PartnerZone:      zoneInfo,
		}
		err = sendFederationRequest("POST", partnerOP.FederationAddr, federation.OperatorNotifyZoneAPI, &opZoneShare, nil)
		if err != nil {
			return err
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Created OP zone successfully"))
}

func DeleteFederationZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZone := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&opZone); err != nil {
		return ormutil.BindErr(err)
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
		return ormutil.DbErr(err)
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
		return ormutil.DbErr(res.Error)
	}
	if regZone.ZoneId != "" {
		return fmt.Errorf("Cannot delete zone %s as it registered by partner OP. Please deregister it before deleting it", regZone.ZoneId)
	}

	if err := db.Delete(&existingZone).Error; err != nil {
		return ormutil.DbErr(err)
	}

	zCloudlet := ormapi.OperatorZoneCloudlet{}
	zCloudlet.ZoneId = opZone.ZoneId
	if err := db.Delete(&zCloudlet).Error; err != nil {
		return ormutil.DbErr(err)
	}

	// Notify deleted zone details with partner OPs
	partnerLookup := ormapi.OperatorFederation{
		Type: fedcommon.TypePartner,
	}
	partnerOPs := []ormapi.OperatorFederation{}
	err = db.Where(&partnerLookup).Find(&partnerOPs).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	for _, partnerOP := range partnerOPs {
		// Only notify those partners with whom the zones were shared
		if !fedcommon.FederationRoleExists(&partnerOP, fedcommon.RoleShareZones) {
			continue
		}
		// Notify federated partner about deleted zone
		opZoneUnShare := federation.ZoneRequest{
			Operator:         selfFed.OperatorId,
			Country:          selfFed.CountryCode,
			OrigFederationId: selfFed.FederationId,
			DestFederationId: partnerOP.FederationId,
			Zone:             opZone.ZoneId,
		}
		err = sendFederationRequest("DELETE", partnerOP.FederationAddr, federation.OperatorNotifyZoneAPI, &opZoneUnShare, nil)
		if err != nil {
			return err
		}
	}

	return ormutil.SetReply(c, ormutil.Msg("Deleted zone successfully"))
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

func ShowFederationZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	opZoneReq := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&opZoneReq); err != nil {
		return ormutil.BindErr(err)
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
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&reg); err != nil {
		return ormutil.BindErr(err)
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
		return ormutil.DbErr(err)
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
		Type:         fedcommon.TypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingZone.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := federation.OperatorZoneRegister{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zones:            []string{existingZone.ZoneId},
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingZone.FederationId,
	}
	opZoneRes := federation.OperatorZoneRegisterResponse{}
	err = sendFederationRequest("POST", fedInfo.FederationAddr, federation.OperatorZoneAPI, &opZoneReg, &opZoneRes)
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
			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg("Partner OP zone registered successfully"))
}

func DeRegisterFederationZone(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reg := ormapi.OperatorZoneCloudletMap{}
	if err := c.Bind(&reg); err != nil {
		return ormutil.BindErr(err)
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
		return ormutil.DbErr(res.Error)
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
		Type:         fedcommon.TypePartner,
	}
	err = db.Where(&fedLookup).First(&fedInfo).Error
	if err != nil {
		return fmt.Errorf("Partner federation %s doesn't exist", existingZone.FederationId)
	}

	// Notify federated partner about zone registration
	opZoneReg := federation.ZoneRequest{
		Operator:         fedInfo.OperatorId,
		Country:          fedInfo.CountryCode,
		Zone:             existingZone.ZoneId,
		OrigFederationId: selfFed.FederationId,
		DestFederationId: existingZone.FederationId,
	}
	err = sendFederationRequest("DELETE", fedInfo.FederationAddr, federation.OperatorZoneAPI, &opZoneReg, nil)
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

			return ormutil.DbErr(err)
		}
	}
	return ormutil.SetReply(c, ormutil.Msg("Partner OP zone deregistered successfully"))
}

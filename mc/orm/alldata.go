package orm

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"google.golang.org/grpc"
)

type RegionContext struct {
	region    string
	username  string
	conn      *grpc.ClientConn
	skipAuthz bool
}

func newResCb(c echo.Context, desc string) func(*edgeproto.Result) {
	return func(res *edgeproto.Result) {
		streamReplyMsg(c, desc, res.Message)
	}
}

func CreateData(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	data := ormapi.AllData{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	// stream back responses
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	hadErr := false
	for _, ctrl := range data.Controllers {
		desc := fmt.Sprintf("Create Controller region %s", ctrl.Region)
		err := CreateControllerObj(ctx, claims, &ctrl)
		streamReply(c, desc, err, &hadErr)
	}
	for _, org := range data.Orgs {
		desc := fmt.Sprintf("Create Organization %s", org.Name)
		err := CreateOrgObj(ctx, claims, &org)
		streamReply(c, desc, err, &hadErr)
	}
	for _, role := range data.Roles {
		desc := fmt.Sprintf("Add User Role %v", role)
		err := AddUserRoleObj(ctx, claims, &role)
		streamReply(c, desc, err, &hadErr)
	}
	for _, regionData := range data.RegionData {
		conn, err := connectController(ctx, regionData.Region)
		if err != nil {
			desc := fmt.Sprintf("Connect %s Controller", regionData.Region)
			streamReply(c, desc, err, &hadErr)
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.username = claims.Username
		rc.region = regionData.Region
		rc.conn = conn

		appdata := &regionData.AppData
		for _, flavor := range appdata.Flavors {
			desc := fmt.Sprintf("Create Flavor %s", flavor.Key.Name)
			_, err = CreateFlavorObj(ctx, rc, &flavor)
			streamReply(c, desc, err, &hadErr)
		}
		for _, restagtbl := range appdata.ResTagTables {
			desc := fmt.Sprintf("Create ResTagTable %v", restagtbl.Key)
			_, err = CreateResTagTableObj(ctx, rc, &restagtbl)
			streamReply(c, desc, err, &hadErr)
		}
		for _, cloudlet := range appdata.Cloudlets {
			desc := fmt.Sprintf("Create Cloudlet %v", cloudlet.Key)
			cb := newResCb(c, desc)
			err = CreateCloudletStream(ctx, rc, &cloudlet, cb)
			streamReply(c, desc, err, &hadErr)
		}
		for _, pool := range appdata.CloudletPools {
			desc := fmt.Sprintf("Create CloudletPool %v", pool.Key)
			_, err := CreateCloudletPoolObj(ctx, rc, &pool)
			streamReply(c, desc, err, &hadErr)
		}
		for _, member := range appdata.CloudletPoolMembers {
			desc := fmt.Sprintf("Create CloudletPoolMember %v", member)
			_, err := CreateCloudletPoolMemberObj(ctx, rc, &member)
			streamReply(c, desc, err, &hadErr)
		}
		for _, policy := range appdata.AutoScalePolicies {
			desc := fmt.Sprintf("Create AutoScalePolicy %v", policy.Key)
			_, err := CreateAutoScalePolicyObj(ctx, rc, &policy)
			streamReply(c, desc, err, &hadErr)
		}
		for _, cinst := range appdata.ClusterInsts {
			desc := fmt.Sprintf("Create ClusterInst %v", cinst.Key)
			cb := newResCb(c, desc)
			err = CreateClusterInstStream(ctx, rc, &cinst, cb)
			streamReply(c, desc, err, &hadErr)
		}
		for _, app := range appdata.Applications {
			desc := fmt.Sprintf("Create App %v", app.Key)
			_, err = CreateAppObj(ctx, rc, &app)
			streamReply(c, desc, err, &hadErr)
		}
		for _, appinst := range appdata.AppInstances {
			desc := fmt.Sprintf("Create AppInst %v", appinst.Key)
			cb := newResCb(c, desc)
			err = CreateAppInstStream(ctx, rc, &appinst, cb)
			streamReply(c, desc, err, &hadErr)
		}
	}
	for _, oc := range data.OrgCloudletPools {
		desc := fmt.Sprintf("Create OrgCloudletPool %v", oc)
		err := CreateOrgCloudletPoolObj(ctx, claims, &oc)
		streamReply(c, desc, err, &hadErr)
	}
	if hadErr {
		streamErr(c, "Some error encountered")
	}
	return nil
}

func DeleteData(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	data := ormapi.AllData{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	// stream back responses
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	hadErr := false
	for _, oc := range data.OrgCloudletPools {
		desc := fmt.Sprintf("Delete OrgCloudletPool %v", oc)
		err := DeleteOrgCloudletPoolObj(ctx, claims, &oc)
		streamReply(c, desc, err, &hadErr)
	}
	for _, regionData := range data.RegionData {
		conn, err := connectController(ctx, regionData.Region)
		if err != nil {
			desc := fmt.Sprintf("Connect %s Controller", regionData.Region)
			streamReply(c, desc, err, &hadErr)
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.username = claims.Username
		rc.region = regionData.Region
		rc.conn = conn

		appdata := &regionData.AppData

		for _, appinst := range appdata.AppInstances {
			desc := fmt.Sprintf("Delete AppInst %v", appinst.Key)
			cb := newResCb(c, desc)
			err = DeleteAppInstStream(ctx, rc, &appinst, cb)
			streamReply(c, desc, err, &hadErr)
		}
		for _, app := range appdata.Applications {
			desc := fmt.Sprintf("Delete App %v", app.Key)
			_, err = DeleteAppObj(ctx, rc, &app)
			streamReply(c, desc, err, &hadErr)
		}
		for _, cinst := range appdata.ClusterInsts {
			desc := fmt.Sprintf("Delete ClusterInst %v", cinst.Key)
			cb := newResCb(c, desc)
			err = DeleteClusterInstStream(ctx, rc, &cinst, cb)
			streamReply(c, desc, err, &hadErr)
		}
		for _, policy := range appdata.AutoScalePolicies {
			desc := fmt.Sprintf("Delete AutoScalePolicy %v", policy.Key)
			_, err := DeleteAutoScalePolicyObj(ctx, rc, &policy)
			streamReply(c, desc, err, &hadErr)
		}
		for _, member := range appdata.CloudletPoolMembers {
			desc := fmt.Sprintf("Delete CloudletPoolMember %v", member)
			_, err := DeleteCloudletPoolMemberObj(ctx, rc, &member)
			streamReply(c, desc, err, &hadErr)
		}
		for _, pool := range appdata.CloudletPools {
			desc := fmt.Sprintf("Delete CloudletPool %v", pool.Key)
			_, err := DeleteCloudletPoolObj(ctx, rc, &pool)
			streamReply(c, desc, err, &hadErr)
		}
		for _, cloudlet := range appdata.Cloudlets {
			desc := fmt.Sprintf("Delete Cloudlet %v", cloudlet.Key)
			cb := newResCb(c, desc)
			err = DeleteCloudletStream(ctx, rc, &cloudlet, cb)
			streamReply(c, desc, err, &hadErr)
		}
		for _, flavor := range appdata.Flavors {
			desc := fmt.Sprintf("Delete Flavor %s", flavor.Key.Name)
			_, err = DeleteFlavorObj(ctx, rc, &flavor)
			streamReply(c, desc, err, &hadErr)
		}
		for _, restagtbl := range appdata.ResTagTables {
			desc := fmt.Sprintf("Delete Restable %s", restagtbl.Key.Name)
			_, err = DeleteResTagTableObj(ctx, rc, &restagtbl)
			streamReply(c, desc, err, &hadErr)
		}
	}
	// roles must be deleted after orgs, otherwise we may delete the
	// role that's needed to be able to delete the org.
	for _, org := range data.Orgs {
		desc := fmt.Sprintf("Delete Organization %s", org.Name)
		err := DeleteOrgObj(ctx, claims, &org)
		streamReply(c, desc, err, &hadErr)
	}
	for _, role := range data.Roles {
		desc := fmt.Sprintf("Remove User Role %v", role)
		err := RemoveUserRoleObj(ctx, claims, &role)
		streamReply(c, desc, err, &hadErr)
	}
	for _, ctrl := range data.Controllers {
		desc := fmt.Sprintf("Delete Controller region %s", ctrl.Region)
		err := DeleteControllerObj(ctx, claims, &ctrl)
		streamReply(c, desc, err, &hadErr)
	}
	if hadErr {
		streamErr(c, "Some error encountered")
	}
	return nil
}

func ShowData(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	data := ormapi.AllData{}

	ctrls, err := ShowControllerObj(ctx, claims)
	if err == nil {
		data.Controllers = ctrls
	}
	orgs, err := ShowOrgObj(ctx, claims)
	if err == nil {
		data.Orgs = orgs
	}
	roles, err := ShowUserRoleObj(ctx, claims.Username)
	if err == nil {
		data.Roles = roles
	}
	ocs, err := ShowOrgCloudletPoolObj(ctx, claims.Username)
	if err == nil {
		data.OrgCloudletPools = ocs
	}

	// Iterate over all controllers. We need to look up
	// controllers this time without enforcement check.
	ctrls = []ormapi.Controller{}
	db := loggedDB(ctx)
	err = db.Find(&ctrls).Error
	if err != nil {
		return c.JSON(http.StatusOK, data)
	}
	for _, ctrl := range ctrls {
		conn, err := connectControllerAddr(ctrl.Address)
		if err != nil {
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.username = claims.Username
		rc.region = ctrl.Region
		rc.conn = conn

		regionData := &ormapi.RegionData{}
		regionData.Region = ctrl.Region
		appdata := &regionData.AppData

		cloudlets, err := ShowCloudletObj(ctx, rc, &edgeproto.Cloudlet{})
		if err == nil {
			appdata.Cloudlets = cloudlets
		}
		pools, err := ShowCloudletPoolObj(ctx, rc, &edgeproto.CloudletPool{})
		if err == nil {
			appdata.CloudletPools = pools
		}
		members, err := ShowCloudletPoolMemberObj(ctx, rc, &edgeproto.CloudletPoolMember{})
		if err == nil {
			appdata.CloudletPoolMembers = members
		}
		flavors, err := ShowFlavorObj(ctx, rc, &edgeproto.Flavor{})
		if err == nil {
			appdata.Flavors = flavors
		}
		aspolicies, err := ShowAutoScalePolicyObj(ctx, rc, &edgeproto.AutoScalePolicy{})
		if err == nil {
			appdata.AutoScalePolicies = aspolicies
		}
		cinsts, err := ShowClusterInstObj(ctx, rc, &edgeproto.ClusterInst{})
		if err == nil {
			appdata.ClusterInsts = cinsts
		}
		apps, err := ShowAppObj(ctx, rc, &edgeproto.App{})
		if err == nil {
			appdata.Applications = apps
		}
		appinsts, err := ShowAppInstObj(ctx, rc, &edgeproto.AppInst{})
		if err == nil {
			appdata.AppInstances = appinsts
		}
		restbls, err := ShowResTagTableObj(ctx, rc, &edgeproto.ResTagTable{})
		if err == nil {
			appdata.ResTagTables = restbls
		}
		if len(flavors) > 0 ||
			len(cloudlets) > 0 || len(cinsts) > 0 ||
			len(apps) > 0 || len(appinsts) > 0 {
			data.RegionData = append(data.RegionData, *regionData)
		}
	}
	return c.JSON(http.StatusOK, data)
}

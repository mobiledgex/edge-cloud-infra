package orm

import (
	"context"
	fmt "fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// AuthzCloudlet provides an efficient way to check if the user
// can see the Cloudlet and Create/Delete ClusterInsts/AppInsts on the Cloudlet,
// based on an Organization's cloudlet pool associations.
type AuthzCloudlet struct {
	orgs             map[string]struct{}
	cloudletPoolSide map[edgeproto.CloudletKey]int
	allowAll         bool
	admin            bool
}

const myPool int = 1
const notMyPool int = 2

func (s *AuthzCloudlet) populate(ctx context.Context, region, username, orgfilter, resource, action string, authops ...authOp) error {
	opts := authOptions{}
	for _, op := range authops {
		op(&opts)
	}

	// Get all orgs user has specified resource+action permissions for
	orgs, err := enforcer.GetAuthorizedOrgs(ctx, username, resource, action)
	if err != nil {
		return err
	}

	// special cases
	if _, found := orgs[""]; found {
		// User is an admin. If no filter is specified,
		// then access for all cloudlets is provided.
		s.admin = true
		if orgfilter == "" {
			s.allowAll = true
			return nil
		} else {
			// ensure access (admin may not have explicit perms
			// for specified org).
			orgs[orgfilter] = struct{}{}
		}
	}
	if orgfilter != "" {
		// Filter only cloudlets for specified org.
		if _, found := orgs[orgfilter]; !found {
			return echo.ErrForbidden
		}
		orgs = make(map[string]struct{})
		orgs[orgfilter] = struct{}{}
	}

	if len(orgs) == 0 {
		// no access to any orgs for given resource/action
		return echo.ErrForbidden
	}

	if opts.requiresOrg != "" {
		if err := checkRequiresOrg(ctx, opts.requiresOrg, s.admin); err != nil {
			return err
		}
	}

	s.orgs = orgs

	// get pools associated with orgs
	db := loggedDB(ctx)
	op := ormapi.OrgCloudletPool{}
	op.Region = region
	if orgfilter != "" {
		op.Org = orgfilter
	}
	ops := []ormapi.OrgCloudletPool{}
	err = db.Where(&op).Find(&ops).Error
	if err != nil {
		return err
	}
	ops = getAccessGranted(ops)

	mypools := make(map[edgeproto.CloudletPoolKey]struct{})
	for _, op := range ops {
		if _, found := orgs[op.Org]; !found {
			// no perms for org
			continue
		}
		poolKey := edgeproto.CloudletPoolKey{
			Name:         op.CloudletPool,
			Organization: op.CloudletPoolOrg,
		}
		mypools[poolKey] = struct{}{}
	}

	// get pools membership
	rc := RegionContext{
		region:    region,
		username:  username,
		skipAuthz: true,
	}
	// build map of cloudlets associated with all cloudlet pools
	s.cloudletPoolSide = make(map[edgeproto.CloudletKey]int)
	err = ShowCloudletPoolStream(ctx, &rc, &edgeproto.CloudletPool{}, func(pool *edgeproto.CloudletPool) {
		for _, name := range pool.Cloudlets {
			cloudletKey := edgeproto.CloudletKey{
				Name:         name,
				Organization: pool.Key.Organization,
			}
			// cloudlet may belong to multiple pools, if any pool
			// is ours, allow access.
			side, found := s.cloudletPoolSide[cloudletKey]
			if !found {
				side = notMyPool
			}
			if _, found := mypools[pool.Key]; found {
				side = myPool
			}
			s.cloudletPoolSide[cloudletKey] = side
		}
	})
	return err
}

// Ok checks if user is authorized to perform resource+action on cloudlet.
// Ok may be called many times, once for each cloudlet in a show command,
// so operates on the cached database data, rather than having to call into
// the database/regional controller each time.
func (s *AuthzCloudlet) Ok(obj *edgeproto.Cloudlet) (bool, bool) {
	filterOutput := false
	if s.allowAll {
		return true, filterOutput
	}
	if _, found := s.orgs[obj.Key.Organization]; found {
		// operator has access to cloudlets created by their org,
		// regardless of whether that cloudlet belongs to
		// developer pools or not.
		return true, filterOutput
	}

	// if user doesn't belong to operator role for this cloudlet and is not admin,
	// then set filterOutput to true, so that operator data which is meant to be hidden
	// is filtered for that user
	filterOutput = true

	// First determine if cloudlet is "public" or "private".
	// "Public" cloudlets do not belong to any cloudlet pool.
	// "Private" cloudlets belong to one or more cloudlet pools.
	poolSide, found := s.cloudletPoolSide[obj.Key]
	if found {
		// "Private" cloudlet, accessible if it belongs to one
		// of our pools
		return poolSide == myPool, filterOutput
	} else {
		// "Public" cloudlet, accessible by all
		return true, filterOutput
	}
}

func (s *AuthzCloudlet) Filter(obj *edgeproto.Cloudlet) {
	// filter cloudlet details not required for developer role
	output := *obj
	*obj = edgeproto.Cloudlet{}
	obj.Key = output.Key
	obj.Location = output.Location
	obj.State = output.State
	obj.IpSupport = output.IpSupport
	obj.NumDynamicIps = output.NumDynamicIps
	obj.MaintenanceState = output.MaintenanceState
	obj.PlatformType = output.PlatformType
	obj.ResTagMap = output.ResTagMap
	obj.TrustPolicy = output.TrustPolicy
	obj.TrustPolicyState = output.TrustPolicyState
}

func authzCreateClusterInst(ctx context.Context, region, username string, obj *edgeproto.ClusterInst, resource, action string) error {
	if !isBillable(ctx, obj.Key.Organization) {
		return echo.ErrForbidden
	}
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, obj.Key.Organization, resource, action, withRequiresOrg(obj.Key.Organization))
	if err != nil {
		return err
	}
	cloudlet := edgeproto.Cloudlet{
		Key: obj.Key.CloudletKey,
	}
	if authzOk, _ := authzCloudlet.Ok(&cloudlet); !authzOk {
		return echo.ErrForbidden
	}
	return nil
}

func authzCreateAppInst(ctx context.Context, region, username string, obj *edgeproto.AppInst, resource, action string) error {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, obj.Key.AppKey.Organization, resource, action, withRequiresOrg(obj.Key.AppKey.Organization))
	if err != nil {
		return err
	}
	cloudlet := edgeproto.Cloudlet{
		Key: obj.Key.ClusterInstKey.CloudletKey,
	}
	if authzOk, _ := authzCloudlet.Ok(&cloudlet); !authzOk {
		return echo.ErrForbidden
	}
	// Developers can't create AppInsts on other developer's ClusterInsts,
	// except for autoclusters where ClusterInst org is MobiledgeX.
	autocluster := false
	if strings.HasPrefix(obj.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
		if obj.Key.ClusterInstKey.Organization != cloudcommon.OrganizationMobiledgeX {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Autocluster AppInst's ClusterInst organization must be %s", cloudcommon.OrganizationMobiledgeX))
		}
		autocluster = true
	}
	if !authzCloudlet.admin && !autocluster && obj.Key.ClusterInstKey.Organization != "" && obj.Key.ClusterInstKey.Organization != obj.Key.AppKey.Organization {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("AppInst organization must match ClusterInst organization"))
	}
	return nil
}

func authzCreateAutoProvPolicy(ctx context.Context, region, username string, obj *edgeproto.AutoProvPolicy, resource, action string) error {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, obj.Key.Organization, resource, action, withRequiresOrg(obj.Key.Organization))
	if err != nil {
		return err
	}
	for _, apCloudlet := range obj.Cloudlets {
		cloudlet := edgeproto.Cloudlet{
			Key: apCloudlet.Key,
		}
		if authzOk, _ := authzCloudlet.Ok(&cloudlet); !authzOk {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("No permissions for Cloudlet %s", cloudlet.Key.GetKeyString()))
		}
	}
	return nil
}

func authzUpdateAutoProvPolicy(ctx context.Context, region, username string, obj *edgeproto.AutoProvPolicy, resource, action string) error {
	// handled the same as create
	return authzCreateAutoProvPolicy(ctx, region, username, obj, resource, action)
}

func authzAddAutoProvPolicyCloudlet(ctx context.Context, region, username string, obj *edgeproto.AutoProvPolicyCloudlet, resource, action string) error {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, obj.Key.Organization, resource, action, withRequiresOrg(obj.Key.Organization))
	if err != nil {
		return err
	}
	cloudlet := edgeproto.Cloudlet{
		Key: obj.CloudletKey,
	}
	if authzOk, _ := authzCloudlet.Ok(&cloudlet); !authzOk {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("No permissions for Cloudlet %s", cloudlet.Key.GetKeyString()))
	}
	return nil
}

func newShowCloudletAuthz(ctx context.Context, region, username, resource, action string) (ShowCloudletAuthz, error) {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, "", resource, action)
	if err != nil {
		return nil, err
	}
	return &authzCloudlet, nil
}

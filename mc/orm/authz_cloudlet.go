package orm

import (
	"context"
	fmt "fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// AuthzCloudlet provides an efficient way to check if the user
// can see the Cloudlet and Create/Delete ClusterInsts/AppInsts on the Cloudlet,
// based on an Organization's cloudlet pool associations.
type AuthzCloudlet struct {
	orgs             map[string]struct{}
	noPoolOrgs       map[string]struct{}
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

	s.noPoolOrgs = make(map[string]struct{})
	for k, _ := range s.orgs {
		s.noPoolOrgs[k] = struct{}{}
	}

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
	mypools := make(map[string]struct{})
	for _, op := range ops {
		if _, found := orgs[op.Org]; !found {
			// no perms for org
			continue
		}
		mypools[op.CloudletPool] = struct{}{}
		// org has pools associated with it, remove it from orgs map
		delete(s.noPoolOrgs, op.Org)
	}

	// get pools membership
	rc := RegionContext{
		region:    region,
		username:  username,
		skipAuthz: true,
	}
	// build map of cloudlets associated with all cloudlet pools
	s.cloudletPoolSide = make(map[edgeproto.CloudletKey]int)
	err = ShowCloudletPoolMemberStream(ctx, &rc, &edgeproto.CloudletPoolMember{}, func(member *edgeproto.CloudletPoolMember) {
		// cloudlet may belong to multiple pools, if any pool
		// is ours, allow access.
		side, found := s.cloudletPoolSide[member.CloudletKey]
		if !found {
			side = notMyPool
		}
		if _, found := mypools[member.PoolKey.Name]; found {
			side = myPool
		}
		s.cloudletPoolSide[member.CloudletKey] = side
	})
	return err
}

// Ok checks if user is authorized to perform resource+action on cloudlet.
// Ok may be called many times, once for each cloudlet in a show command,
// so operates on the cached database data, rather than having to call into
// the database/regional controller each time.
func (s *AuthzCloudlet) Ok(obj *edgeproto.Cloudlet) bool {
	if s.allowAll {
		return true
	}
	if _, found := s.orgs[obj.Key.Organization]; found {
		// operator has access to cloudlets created by their org,
		// regardless of whether that cloudlet belongs to
		// developer pools or not.
		return true
	}

	// First determine if cloudlet is "public" or "private".
	// "Public" cloudlets do not belong to any cloudlet pool.
	// "Private" cloudlets belong to one or more cloudlet pools.
	poolSide, found := s.cloudletPoolSide[obj.Key]
	if found {
		// "Private" cloudlet, accessible if it belongs to one
		// of our pools
		return poolSide == myPool
	} else {
		// "Public" cloudlet, accessible by orgs not associated
		// with any pools.
		if len(s.noPoolOrgs) > 0 {
			return true
		}
		return false
	}
}

func authzCreateClusterInst(ctx context.Context, region, username string, obj *edgeproto.ClusterInst, resource, action string) error {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, obj.Key.Organization, resource, action, withRequiresOrg(obj.Key.Organization))
	if err != nil {
		return err
	}
	cloudlet := edgeproto.Cloudlet{
		Key: obj.Key.CloudletKey,
	}
	if !authzCloudlet.Ok(&cloudlet) {
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
	if !authzCloudlet.Ok(&cloudlet) {
		return echo.ErrForbidden
	}
	// Enforce that target ClusterInst org is the same as AppInst org.
	// This prevents Developers from using reservable ClusterInsts directly.
	// Only auto-provisioning service (which goes direct to controller API)
	// can instantiate AppInsts with mismatched orgs.
	if !authzCloudlet.admin && obj.Key.ClusterInstKey.Organization != "" && obj.Key.ClusterInstKey.Organization != obj.Key.AppKey.Organization {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("AppInst organization must match ClusterInst organization"))
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

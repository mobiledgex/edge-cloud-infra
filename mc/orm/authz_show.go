package orm

import (
	"context"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type PoolOrgPair struct {
	DeveloperOrg    string
	CloudletPoolOrg string
}

type AuthzShow struct {
	allowedOrgs  map[string]struct{}
	allowAll     bool
	poolOrgPairs map[PoolOrgPair]struct{}
}

func newShowAuthz(ctx context.Context, region, username, resource, action string) (*AuthzShow, error) {
	orgs, err := enforcer.GetAuthorizedOrgs(ctx, username, resource, action)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		// no access to any orgs for given resource/action
		return nil, echo.ErrForbidden
	}
	authz := AuthzShow{
		allowedOrgs: orgs,
	}
	if _, found := orgs[""]; found {
		// user is an admin.
		authz.allowAll = true
	}
	return &authz, nil
}

func (s *AuthzShow) Ok(org string) bool {
	if s.allowAll {
		return true
	}
	_, found := s.allowedOrgs[org]
	return found
}

func (s *AuthzShow) setCloudletPoolOrgs(ctx context.Context, region, username string) error {
	// get pools associated with orgs
	db := loggedDB(ctx)
	op := ormapi.OrgCloudletPool{}
	op.Region = region
	ops := []ormapi.OrgCloudletPool{}
	err := db.Where(&op).Find(&ops).Error
	if err != nil {
		return err
	}
	ops = getAccessGranted(ops)

	rc := RegionContext{
		region:    region,
		username:  username,
		skipAuthz: true,
	}
	// Validate region
	err = ShowCloudletPoolStream(ctx, &rc, &edgeproto.CloudletPool{}, func(pool *edgeproto.CloudletPool) {})
	if err != nil {
		return err
	}

	orgs := []string{}
	s.poolOrgPairs = make(map[PoolOrgPair]struct{})
	for _, op := range ops {
		orgs = append(orgs, op.Org)
		pair := PoolOrgPair{
			DeveloperOrg:    op.Org,
			CloudletPoolOrg: op.CloudletPoolOrg,
		}
		s.poolOrgPairs[pair] = struct{}{}
	}
	return nil
}

func newShowPoolAuthz(ctx context.Context, region, username string, resource, action string) (*AuthzShow, error) {
	orgs, err := enforcer.GetAuthorizedOrgs(ctx, username, resource, action)
	if err != nil {
		return nil, err
	}
	authz := AuthzShow{
		allowedOrgs: orgs,
	}
	if _, found := orgs[""]; found {
		// user is an admin.
		authz.allowAll = true
		return &authz, nil
	}

	// get pools associated with orgs
	err = authz.setCloudletPoolOrgs(ctx, region, username)
	if err != nil {
		return nil, err
	}
	if len(authz.allowedOrgs) == 0 && len(authz.poolOrgPairs) == 0 {
		// no access to any orgs for given resource/action
		return nil, echo.ErrForbidden
	}
	return &authz, nil
}

type AuthzClusterInstShow struct {
	*AuthzShow
}

func newShowClusterInstAuthz(ctx context.Context, region, username string, resource, action string) (ShowClusterInstAuthz, error) {
	authz, err := newShowPoolAuthz(ctx, region, username, resource, action)
	if err != nil {
		return nil, err
	}
	return &AuthzClusterInstShow{authz}, nil
}

func (s *AuthzClusterInstShow) Ok(obj *edgeproto.ClusterInst) (bool, bool) {
	filterOutput := false
	allow := s.AuthzShow.Ok(obj.Key.Organization)
	if allow {
		return allow, filterOutput
	}
	poolPair := PoolOrgPair{
		DeveloperOrg:    obj.Key.Organization,
		CloudletPoolOrg: obj.Key.CloudletKey.Organization,
	}
	_, allow = s.AuthzShow.poolOrgPairs[poolPair]
	return allow, filterOutput
}

func (s *AuthzClusterInstShow) Filter(obj *edgeproto.ClusterInst) {
	// nothing to filter for Operator, show all objects for Developer & Operator
}

type AuthzAppInstShow struct {
	*AuthzShow
}

func newShowAppInstAuthz(ctx context.Context, region, username string, resource, action string) (ShowAppInstAuthz, error) {
	authz, err := newShowPoolAuthz(ctx, region, username, resource, action)
	if err != nil {
		return nil, err
	}
	return &AuthzAppInstShow{authz}, nil
}

func (s *AuthzAppInstShow) Ok(obj *edgeproto.AppInst) (bool, bool) {
	filterOutput := false
	allow := s.AuthzShow.Ok(obj.Key.AppKey.Organization)
	if allow {
		return allow, filterOutput
	}
	poolPair := PoolOrgPair{
		DeveloperOrg:    obj.Key.AppKey.Organization,
		CloudletPoolOrg: obj.Key.ClusterInstKey.CloudletKey.Organization,
	}
	_, allow = s.AuthzShow.poolOrgPairs[poolPair]
	return allow, filterOutput
}

func (s *AuthzAppInstShow) Filter(obj *edgeproto.AppInst) {
	// nothing to filter for Operator, show all objects for Developer & Operator
}

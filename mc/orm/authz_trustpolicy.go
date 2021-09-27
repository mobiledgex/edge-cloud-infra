package orm

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// AuthzTrustPolicy allows a user to see a trust policy only if
// 1) user is an admin (AuthzCloudlet allowAll)
// 2) user is part of the org for that policy
// 3) there at least one cloudlet using that policy that they can see, based
//    on AuthzCloudlet pool checking
type AuthzTrustPolicy struct {
	authzCloudlet        *AuthzCloudlet
	allowedTrustPolicies map[edgeproto.PolicyKey]struct{}
}

func (s *AuthzTrustPolicy) Ok(obj *edgeproto.TrustPolicy) (bool, bool) {
	filterOutput := false
	if s.authzCloudlet.allowAll {
		return true, filterOutput
	}
	if _, found := s.authzCloudlet.orgs[obj.Key.Organization]; found {
		// operator has access to policies created by their org
		return true, filterOutput
	}
	if _, found := s.allowedTrustPolicies[obj.Key]; found {
		return true, filterOutput
	}
	return false, filterOutput
}

func (s *AuthzTrustPolicy) Filter(obj *edgeproto.TrustPolicy) {
}

func (s *AuthzTrustPolicy) populate(ctx context.Context, region, username string) error {
	rc := ormutil.RegionContext{
		Region:    region,
		Username:  username,
		SkipAuthz: true, // skip since we already have the cloudlet authz
		Database:  database,
	}
	// allow policies associated with cloudlets that the user can see
	err := ctrlclient.ShowCloudletStream(ctx, &rc, &edgeproto.Cloudlet{}, connCache, nil, func(cloudlet *edgeproto.Cloudlet) error {
		if authzOk, _ := s.authzCloudlet.Ok(cloudlet); !authzOk || cloudlet.TrustPolicy == "" {
			return nil
		}
		key := edgeproto.PolicyKey{
			Organization: cloudlet.Key.Organization,
			Name:         cloudlet.TrustPolicy,
		}
		s.allowedTrustPolicies[key] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func newShowTrustPolicyAuthz(ctx context.Context, region, username, resource, action string) (ctrlclient.ShowTrustPolicyAuthz, error) {
	authzCloudlet := AuthzCloudlet{}
	err := authzCloudlet.populate(ctx, region, username, "", resource, action)
	if err != nil {
		return nil, err
	}
	authzTrustPolicy := AuthzTrustPolicy{
		authzCloudlet:        &authzCloudlet,
		allowedTrustPolicies: make(map[edgeproto.PolicyKey]struct{}),
	}
	err = authzTrustPolicy.populate(ctx, region, username)
	if err != nil {
		return nil, err
	}
	return &authzTrustPolicy, nil
}

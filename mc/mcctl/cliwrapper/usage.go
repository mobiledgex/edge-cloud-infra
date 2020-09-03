package cliwrapper

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

func (s *Client) ShowAppUsage(uri, token string, query *ormapi.RegionAppInstUsage) (*ormapi.AllUsage, int, error) {
	args := []string{"usage", "app"}
	usage := ormapi.AllUsage{}
	st, err := s.runObjs(uri, token, args, query, &usage)
	return &usage, st, err
}

func (s *Client) ShowClusterUsage(uri, token string, query *ormapi.RegionClusterInstUsage) (*ormapi.AllUsage, int, error) {
	args := []string{"usage", "cluster"}
	usage := ormapi.AllUsage{}
	st, err := s.runObjs(uri, token, args, query, &usage)
	return &usage, st, err
}

func (s *Client) ShowCloudletPoolUsage(uri, token string, query *ormapi.RegionCloudletPoolUsage) (*ormapi.CloudletPoolUsage, int, error) {
	args := []string{"usage", "cloudletpool"}
	usage := ormapi.CloudletPoolUsage{}
	st, err := s.runObjs(uri, token, args, query, &usage)
	return &usage, st, err
}

package shepherd_unittest

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "unit test"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	return nil
}

func (s *Platform) GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error) {
	return "localhost", nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &UTClient{}, nil
}

// UTClient hijacks a set of commands and returns predetermined output
// For all other commands it just calls pc.LocalClient equivalents
type UTClient struct {
	pc.LocalClient
}

func (s *UTClient) Output(command string) (string, error) {
	out, err := GetUTData(command)
	if err != nil {
		return s.Output(command)
	}
	return out, nil
}

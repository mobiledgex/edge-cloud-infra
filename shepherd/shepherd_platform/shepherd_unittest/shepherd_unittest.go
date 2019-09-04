package shepherd_unittest

import (
	"io"
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
	c pc.LocalClient
}


func (s *UTClient) Output(command string) (string, error) {
	out, err := GetUTData(command)
	if err != nil {
		return s.c.Output(command)
	}
	return out, nil
}

func (s *UTClient) Shell(sin io.Reader, sout, serr io.Writer, args ...string) error {
	return s.c.Shell(sin,sout,serr,args...)
}

func (s *UTClient) Start(command string) (io.ReadCloser, io.ReadCloser, io.WriteCloser, error) {
	return s.c.Start(command)
}

func (s *UTClient) Wait() error {
	return s.c.Wait()
}

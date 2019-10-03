package shepherd_fake

import (
	"context"
	"fmt"
	"time"

	exporter "github.com/mobiledgex/edge-cloud-infra/shepherd/fakePromExporter"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
	promStarted bool
}

func (s *Platform) GetType() string {
	return "fake"
}

func (s *Platform) Init(ctx context.Context, key *edgeproto.CloudletKey, physicalName, vaultAddr string) error {
	s.promStarted = false
	// maybe move this to the e2e-test program itself to start it when controller and everything else gets started?
	go exporter.StartExporter(ctx)
	return nil
}

func (s *Platform) GetClusterIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, error) {
	// start a docker container for prometheus
	// only start one, as there is only one port 9090 we can use
	if !s.promStarted {
		err := exporter.StartPromContainer(ctx)
		if err == nil {
			s.promStarted = true
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "Prometheus not started", "error", err)
		}
	}
	// we dont want every single shepherd hitting the one prometheus we have up
	if s.promStarted {
		return "localhost", nil
	} else {
		return "", fmt.Errorf("No Prometheus up or another shepherd is already on it")
	}
}

func (s *Platform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) GetPlatformStats(ctx context.Context) (shepherd_common.CloudletMetrics, error) {
	return shepherd_common.CloudletMetrics{}, nil
}

func (s *Platform) GetVmStats(ctx context.Context, key *edgeproto.AppInstKey) (shepherd_common.AppMetrics, error) {
	return shepherd_common.AppMetrics{}, nil
}

func (s *Platform) GetMetricsCollectInterval() time.Duration {
	return 0
}

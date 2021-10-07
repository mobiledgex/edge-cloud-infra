package orm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

type ConnCache struct {
	sync.Mutex
	cache          map[string]*grpc.ClientConn
	used           map[string]bool
	notifyRootConn *grpc.ClientConn
	stopCleanup    chan struct{}
}

var connCacheCleanupInterval = 30 * time.Minute

func NewConnCache() *ConnCache {
	rcc := &ConnCache{}
	rcc.cache = make(map[string]*grpc.ClientConn)
	rcc.used = make(map[string]bool)
	return rcc
}

func (s *ConnCache) GetRegionConn(ctx context.Context, region string) (*grpc.ClientConn, error) {
	// Although we hold the lock while doing the connect, the
	// connect is non-blocking, so will not actually block us.
	s.Lock()
	defer s.Unlock()
	conn, found := s.cache[region]
	var err error
	if !found {
		conn, err = connectController(ctx, region)
		if err != nil {
			return nil, err
		}
		s.cache[region] = conn
	}
	s.used[region] = true
	return conn, nil
}

func (s *ConnCache) GetNotifyRootConn(ctx context.Context) (*grpc.ClientConn, error) {
	s.Lock()
	defer s.Unlock()
	if s.notifyRootConn == nil {
		conn, err := connectNotifyRoot(ctx)
		if err != nil {
			return nil, err
		}
		s.notifyRootConn = conn
	}
	return s.notifyRootConn, nil
}

func (s *ConnCache) Cleanup() {
	s.Lock()
	defer s.Unlock()
	for region, conn := range s.cache {
		used := s.used[region]
		if used {
			s.used[region] = false
		} else {
			// cleanup
			conn.Close()
			delete(s.cache, region)
			delete(s.used, region)
		}
	}
}

func (s *ConnCache) Start() {
	s.stopCleanup = make(chan struct{})
	go func() {
		done := false
		for !done {
			select {
			case <-time.After(connCacheCleanupInterval):
				s.Cleanup()
			case <-s.stopCleanup:
				done = true
			}
		}
	}()
}

func (s *ConnCache) Finish() {
	close(s.stopCleanup)
	s.Lock()
	for region, conn := range s.cache {
		conn.Close()
		delete(s.cache, region)
		delete(s.used, region)
	}
	if s.notifyRootConn != nil {
		s.notifyRootConn.Close()
		s.notifyRootConn = nil
	}
	s.Unlock()
}

func connectController(ctx context.Context, region string) (*grpc.ClientConn, error) {
	addr, err := getControllerAddrForRegion(ctx, region)
	if err != nil {
		return nil, err
	}
	return connectGrpcAddr(ctx, addr, []node.MatchCA{node.AnyRegionalMatchCA()})
}

func connectNotifyRoot(ctx context.Context) (*grpc.ClientConn, error) {

	if serverConfig.NotifyAddrs == "" {
		return nil, fmt.Errorf("No parent notify address specified, cannot connect to notify root")
	}
	addrs := strings.Split(serverConfig.NotifyAddrs, ",")
	return connectGrpcAddr(ctx, addrs[0], []node.MatchCA{node.GlobalMatchCA()})
}

func connectGrpcAddr(ctx context.Context, addr string, serverIssuers []node.MatchCA) (*grpc.ClientConn, error) {
	tlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerGlobal,
		serverIssuers)
	if err != nil {
		return nil, err
	}
	dialOption := tls.GetGrpcDialOption(tlsConfig)
	return grpc.Dial(addr, dialOption,
		grpc.WithUnaryInterceptor(log.UnaryClientTraceGrpc),
		grpc.WithStreamInterceptor(log.StreamClientTraceGrpc),
	)
}

func getControllerAddrForRegion(ctx context.Context, region string) (string, error) {
	ctrl, err := getControllerObj(ctx, region)
	if err != nil {
		return "", err
	}
	return ctrl.Address, nil
}

func getControllerObj(ctx context.Context, region string) (*ormapi.Controller, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	ctrl := ormapi.Controller{
		Region: region,
	}
	db := loggedDB(ctx)
	res := db.Where(&ctrl).First(&ctrl)
	if res.Error != nil {
		if res.RecordNotFound() {
			return nil, fmt.Errorf("region \"%s\" not found", region)
		}
		return nil, res.Error
	}
	return &ctrl, nil
}

func CreateController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := ormutil.GetContext(c)

	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return ormutil.BindErr(err)
	}
	err = CreateControllerObj(ctx, claims, &ctrl)
	if err != nil {
		return err
	}
	return ormutil.SetReply(c, ormutil.Msg("Controller registered"))
}

func CreateControllerObj(ctx context.Context, claims *UserClaims, ctrl *ormapi.Controller) error {
	if ctrl.Region == "" {
		return fmt.Errorf("Controller Region not specified")
	}
	if ctrl.Address == "" {
		return fmt.Errorf("Controller Address not specified")
	}
	if err := authorized(ctx, claims.Username, "", ResourceControllers, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	err := db.Create(ctrl).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	return nil
}

func DeleteController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := ormutil.GetContext(c)

	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return ormutil.BindErr(err)
	}
	err = DeleteControllerObj(ctx, claims, &ctrl)
	if err != nil {
		return err
	}
	// Close regional influxDB connection when controller is deleted
	influxDbConnCache.DeleteClient(ctrl.Region)
	return ormutil.SetReply(c, ormutil.Msg("Controller deregistered"))
}

func DeleteControllerObj(ctx context.Context, claims *UserClaims, ctrl *ormapi.Controller) error {
	if ctrl.Region == "" {
		return fmt.Errorf("Controller Region not specified")
	}
	if err := authorized(ctx, claims.Username, "", ResourceControllers, ActionManage); err != nil {
		return err
	}
	db := loggedDB(ctx)
	err := db.Delete(ctrl).Error
	if err != nil {
		return ormutil.DbErr(err)
	}
	return nil
}

func ShowController(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	filter, err := bindDbFilter(c, &ormapi.Controller{})
	if err != nil {
		return err
	}
	ctrls, err := ShowControllerObj(ctx, claims, filter)
	if err != nil {
		return err
	}
	return ormutil.SetReply(c, ctrls)
}

func ShowControllerObj(ctx context.Context, claims *UserClaims, filter map[string]interface{}) ([]ormapi.Controller, error) {
	ctrls := []ormapi.Controller{}
	db := loggedDB(ctx)
	err := db.Where(filter).Find(&ctrls).Error
	if err != nil {
		return nil, ormutil.DbErr(err)
	}
	return ctrls, nil
}

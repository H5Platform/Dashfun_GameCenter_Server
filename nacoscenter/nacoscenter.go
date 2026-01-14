package nacoscenter

import (
	"context"
	"dashfun_gamecenter/config"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	asv1 "github.com/dashfun_web3/api_proto/gen/accountservice/v1"
	hsv1 "github.com/dashfun_web3/api_proto/gen/healthservice/v1"
	lbv1 "github.com/dashfun_web3/api_proto/gen/leaderboardservice/v1"
	usv1 "github.com/dashfun_web3/api_proto/gen/userservice/v1"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var onceNacosCenter sync.Once
var instNacosCenter *NacosCenter

const RefreshNacosServerInterval = 60 * time.Second

type NacosService string

const (
	NacosUserService NacosService = "user_center.service"
)

func NacosServiceName(service NacosService, name string) NacosService {
	return NacosService(fmt.Sprintf("%s.%s", service, name))
}

type grpcClientConn struct {
	conn     *grpc.ClientConn
	connTime int64  // 连接时间
	connName string // 连接名称,用服务器的ip和端口组成，以便区分不同的服务
}

type NacosCenter struct {
	client               naming_client.INamingClient
	userServiceClient    usv1.UserServiceClient
	lbServiceClient      lbv1.LeaderboardServiceClient
	accountServiceClient asv1.AccountServiceClient
	serviceConnection    map[NacosService]*grpcClientConn
	sync.RWMutex
}

func NewAuthDataOutgoingContext(method, token string, timeout time.Duration) (context.Context, func()) {
	md := metadata.Pairs("method", method, "token", token)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return metadata.NewOutgoingContext(ctx, md), cancel
}

func Get() *NacosCenter {
	onceNacosCenter.Do(func() {
		instNacosCenter = &NacosCenter{}
		instNacosCenter.init()
	})
	return instNacosCenter
}

func (n *NacosCenter) init() {
	n.client = n.newNacosClient()
	n.serviceConnection = make(map[NacosService]*grpcClientConn)
}

func (n *NacosCenter) newNacosClient() naming_client.INamingClient {
	nacosCfg := config.GetConfig().NacosCfg
	serverConfig := []constant.ServerConfig{{
		IpAddr: nacosCfg.IpAddr, Port: nacosCfg.Port,
	}}

	logLevel := "debug"
	if config.IsProd() {
		logLevel = "fatal"
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:          config.NacosNamespace(),
		TimeoutMs:            5000,
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
		LogLevel:             logLevel,
	}
	client, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  clientConfig,
	})

	if err != nil {
		log.Panicf("create Nacos naming client error: %v\n", err)
	}

	return client
}

func (n *NacosCenter) GetAccountServiceClient() (asv1.AccountServiceClient, error) {
	n.Lock()
	defer n.Unlock()
	conn, changed, err := n.getRpcConnection(NacosServiceName(NacosUserService, "account"))
	if err != nil {
		zap.S().Errorw("get grpc connection error", "err", err)
		return nil, err
	}
	if n.accountServiceClient == nil || changed {
		n.accountServiceClient = asv1.NewAccountServiceClient(conn.conn)
	}
	return n.accountServiceClient, nil
}

func (n *NacosCenter) GetUserServiceClient() (usv1.UserServiceClient, error) {
	n.Lock()
	defer n.Unlock()
	conn, changed, err := n.getRpcConnection(NacosServiceName(NacosUserService, "user"))
	if err != nil {
		zap.S().Errorw("get grpc connection error", "err", err)
		return nil, err
	}
	if n.userServiceClient == nil || changed {
		n.userServiceClient = usv1.NewUserServiceClient(conn.conn)
	}
	return n.userServiceClient, nil
}

func (n *NacosCenter) GetLeaderboardServiceClient() (lbv1.LeaderboardServiceClient, error) {
	n.Lock()
	defer n.Unlock()
	conn, changed, err := n.getRpcConnection(NacosServiceName(NacosUserService, "leaderboard"))
	if err != nil {
		zap.S().Errorw("get grpc connection error", "err", err)
		return nil, err
	}
	if n.lbServiceClient == nil || changed {
		n.lbServiceClient = lbv1.NewLeaderboardServiceClient(conn.conn)
	}
	return n.lbServiceClient, nil
}

func (n *NacosCenter) getRpcConnection(serviceName NacosService) (*grpcClientConn, bool, error) {
	if n.client == nil {
		zap.S().Errorw("Nacos client is nil")
		return nil, false, errors.New("nacos client is nil")
	}

	// 当前有链接且可用
	currConn, ok := n.serviceConnection[serviceName]
	if ok {
		if time.Duration(time.Now().UnixMilli()-currConn.connTime) < RefreshNacosServerInterval && currConn.conn.GetState() == connectivity.Ready {
			return currConn, false, nil
		}
		if currConn.conn.GetState() != connectivity.Ready {
			// 当前链接不可用，新建一个nacos client，获取最新的服务地址
			n.client = n.newNacosClient()
		}
	}

	instance, err := n.client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: "user_center.service",
	})

	if err != nil {
		n.client = n.newNacosClient()
		return nil, false, err
	}

	if instance == nil {
		n.client = n.newNacosClient()
		return nil, false, errors.New("no instance found for service: " + string(serviceName))
	}

	connName := fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	if currConn != nil && currConn.connName == connName {
		return currConn, false, nil
	}

	conn, err := grpc.NewClient(connName, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Errorw("grpc NewClient error", "err", err)
		return nil, false, err
	}

	zap.S().Infow("grpc service connecting...", "serviceName", serviceName, "connName", connName)
	hsClient := hsv1.NewHealthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = hsClient.Ping(ctx, &hsv1.PingRequest{})
	if err != nil {
		return nil, false, err
	}

	n.serviceConnection[serviceName] = &grpcClientConn{
		conn:     conn,
		connTime: time.Now().Unix(),
		connName: fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
	}
	return n.serviceConnection[serviceName], true, nil
}

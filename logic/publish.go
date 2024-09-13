package logic

import (
	"github.com/go-redis/redis"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"gochat_my/config"
	"gochat_my/tools"
	"strings"
	"time"
)

var (
	RedisClient     *redis.Client
	RedisSessClient *redis.Client
)

func (logic *Logic) InitPublishRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.DB,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisClient Ping Result pong : %s, err : %s ", pong, err.Error())
		return err
	}
	RedisSessClient = RedisClient
	return nil
}

func (logic *Logic) InitRpcServer() (err error) {
	var network, addr string
	rpcAddrList := strings.Split(config.Conf.Logic.LogicBase.RpcAddress, ",")
	for _, bind := range rpcAddrList {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogiRpc ParseNetwork err ： %s", err.Error())
		}
		logrus.Infof("logic start run at--> %s:%s", network, addr)
		go logic.createRpcServer(network, addr)
	}
	return
}

func (logic *Logic) createRpcServer(network, addr string) {
	s := server.NewServer()
	logic.addRegistryPlugin(s, network, addr)
	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), logic.ServerId)
	if err != nil {
		logrus.Errorf("register error : %s ", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}

func (logic *Logic) addRegistryPlugin(s *server.Server, network, addr string) {
	// 将 服务信息注册到 etcd
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	s.Plugins.Add(r)
}

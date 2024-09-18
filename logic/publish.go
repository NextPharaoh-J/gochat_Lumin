package logic

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"gochat_my/config"
	"gochat_my/proto"
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

func (logic *Logic) getUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
func (logic *Logic) getRoomUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
func (logic *Logic) getRoomOnlineCountKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) RedisPublishChannel(serverId string, toUserId int, msg []byte) (err error) {
	redisMsg := proto.RedisMsg{
		Op:       config.OpSingleSend,
		ServerId: serverId,
		UserId:   toUserId,
		Msg:      msg,
	}
	redisMsgStr, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("RedisPublishChannel json.Marshal err : %s", err.Error())
		return
	}
	redisChannel := config.QueueName
	if err = RedisClient.RPush(redisChannel, redisMsgStr).Err(); err != nil {
		logrus.Errorf("RedisPublishChannel RPush err : %s", err.Error())
		return
	}
	return
}
func (logic *Logic) RedisPublishRoomInfo(roomId int, count int, roomUserInfo map[string]string, msg []byte) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomSend,
		RoomId:       roomId,
		Count:        count,
		RoomUserInfo: roomUserInfo,
		Msg:          msg,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo json.Marshal err : %s", err.Error())
	}
	err = RedisClient.LPush(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo LPush err : %s", err.Error())
	}
	return
}
func (logic *Logic) RedisPublishRoomCount(roomId int, count int) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:     config.OpRoomSend,
		RoomId: roomId,
		Count:  count,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo json.Marshal err : %s", err.Error())
	}
	err = RedisClient.LPush(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo LPush err : %s", err.Error())
	}
	return
}
func (logic *Logic) RedisPushRoomInfo(roomId int, count int, RoomUserInfo map[string]string) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomSend,
		RoomId:       roomId,
		Count:        count,
		RoomUserInfo: RoomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo json.Marshal err : %s", err.Error())
	}
	err = RedisClient.LPush(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("RedisPublishRoomInfo LPush err : %s", err.Error())
	}
	return
}

package connect

import (
	"context"
	"errors"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"gochat_my/config"
	"gochat_my/proto"
	"gochat_my/tools"
	"strings"
	"sync"
	"time"
)

var logicRpcClient client.XClient
var once sync.Once

type RpcConnect struct{}

func (rpc *RpcConnect) Connect(connReq *proto.ConnectRequest) (uid int, err error) {
	reply := &proto.ConnectReply{}
	err = logicRpcClient.Call(context.Background(), "Connect", connReq, reply)
	if err != nil {
		logrus.Fatalf("fail to call : %v", err)
	}
	uid = reply.UserId
	logrus.Infof("connect logic UserId:%d", uid)
	return
}
func (rpc *RpcConnect) DisConnect(disConnReq *proto.DisConnectRequest) (err error) {
	reply := &proto.DisConnectReply{}
	if err = logicRpcClient.Call(context.Background(), "DisConnect", disConnReq, reply); err != nil {
		logrus.Fatalf("fail to call : %v", err)
	}
	return
}

func (c *Connect) InitLogicRpcClient() (err error) {
	etcdConfigOpt := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
		Bucket:            "",
		PersistConnection: true,
		Username:          config.Conf.Common.CommonEtcd.UserName,
		Password:          config.Conf.Common.CommonEtcd.Password,
	}
	once.Do(func() {
		d, e := etcdV3.NewEtcdDiscovery(
			config.Conf.Common.CommonEtcd.BasePath,
			config.Conf.Common.CommonEtcd.ServerPathLogic,
			[]string{config.Conf.Common.CommonEtcd.Host},
			true,
			etcdConfigOpt)
		if e != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail : %s")
		}
		logicRpcClient = client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathLogic, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	})
	if logicRpcClient == nil {
		return errors.New("get rpc client nil")
	}
	return
}
func (c *Connect) InitConnectWebSocketRpcServer() (err error) {
	var network, addr string
	connectRpcAddress := strings.Split(config.Conf.Connect.ConnectRpcAddressWebSockts.Address, ",")
	for _, bind := range connectRpcAddress {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitConnectWebSocketRpcServer ParseNetwork err : %s", err.Error())
		}
		logrus.Infof("Connect start run at --> %s:%s", network, addr)
		go c.createConnectWebSocketsRpcServer(network, addr)
	}
	return
}
func (c *Connect) createConnectWebSocketsRpcServer(network, addr string) {
	s := server.NewServer()
	c.addRegistryPlugin(s, network, addr)

	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcConnectPush), c.ServerId)
	if err != nil {
		logrus.Errorf("register RpcConnectPush error : %s ", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}
func (c *Connect) addRegistryPlugin(s *server.Server, network, addr string) {
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

// MicroServe Instance of Connect
type RpcConnectPush struct{}

func (rpc *RpcConnectPush) PushSingleMsg(ctx context.Context, pushMsgReq *proto.PushMsgRequest, successReply *proto.SuccessReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
	)

	logrus.Infof("rpc PushMsg : %v", pushMsgReq)
	if pushMsgReq == nil {
		logrus.Errorf("rpc PushMsgReq is nil")
		return
	}
	bucket = DefaultServer.Bucket(pushMsgReq.UserId)
	if channel = bucket.Channel(pushMsgReq.UserId); channel != nil {
		err = channel.Push(&pushMsgReq.Msg)
		logrus.Infof("rpc PushMsgReq channel push success : %v", pushMsgReq.Msg)
		return
	}
	successReply.Msg = config.SuccessReplyMsg
	successReply.Code = config.SuccessReplyCode
	logrus.Infof("successReply : %v", successReply)
	return
}
func (rpc *RpcConnectPush) PushRoomMsg(ctx context.Context, pushMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Msg = config.SuccessReplyMsg
	successReply.Code = config.SuccessReplyCode
	logrus.Infof("PushRoomMsg : %v", pushMsgReq)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(pushMsgReq)
	}
	return
}
func (rpc *RpcConnectPush) PushRoomCount(ctx context.Context, pushRoomMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Msg = config.SuccessReplyMsg
	successReply.Code = config.SuccessReplyCode
	logrus.Infof("PushRoomCount : %v", pushRoomMsgReq)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(pushRoomMsgReq)
	}
	return
}
func (rpc *RpcConnectPush) PushRoomInfo(ctx context.Context, pushRoomMsgReq *proto.PushRoomMsgRequest, successReply *proto.SuccessReply) (err error) {
	successReply.Msg = config.SuccessReplyMsg
	successReply.Code = config.SuccessReplyCode
	logrus.Infof("PushRoomInfo : %v", pushRoomMsgReq)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(pushRoomMsgReq)
	}
	return
}

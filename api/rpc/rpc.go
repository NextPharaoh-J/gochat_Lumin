package rpc

import (
	"context"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"gochat_my/config"
	"gochat_my/proto"
	"sync"
	"time"
)

type RpcLogic struct{}

var (
	logicRpcClient client.XClient
	once           sync.Once
	RpcLogicObj    *RpcLogic
)

func InitLogicRpcClient() {
	once.Do(func() {
		etcdConfigOpt := &store.Config{
			ClientTLS:         nil,
			TLS:               nil,
			ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
			Bucket:            "",
			PersistConnection: true,
			Username:          config.Conf.Common.CommonEtcd.UserName,
			Password:          config.Conf.Common.CommonEtcd.Password,
		}
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
		RpcLogicObj = &RpcLogic{}
	})
	if logicRpcClient == nil {
		logrus.Fatalf("get logic rpc client fail")
	}
}
func (rpc *RpcLogic) CheckAuth(req *proto.CheckAuthRequest) (code int, userId int, userName string) {
	reply := new(proto.CheckAuthResponse)
	logicRpcClient.Call(context.Background(), "CheckAuth", req, reply)
	code = reply.Code
	userId = reply.UserId
	userName = reply.UserName
	return
}
func (rpc *RpcLogic) Login(req *proto.LoginRequest) (code int, authToken string, msg string) {
	reply := new(proto.LoginResponse)
	err := logicRpcClient.Call(context.Background(), "Login", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}
func (rpc *RpcLogic) Logout(req *proto.LogoutRequest) (code int) {
	reply := new(proto.LogoutResponse)
	logicRpcClient.Call(context.Background(), "Logout", req, reply)
	code = reply.Code
	return
}
func (rpc *RpcLogic) Register(req *proto.RegisterRequest) (code int, authToken string, msg string) {
	reply := new(proto.RegisterReply)
	err := logicRpcClient.Call(context.Background(), "Register", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}
func (rpc *RpcLogic) GetUserNameByUserId(req *proto.GetUserInfoRequest) (code int, userName string) {
	reply := new(proto.GetUserInfoResponse)
	logicRpcClient.Call(context.Background(), "GetUserNameByUserId", req, reply)
	code = reply.Code
	userName = reply.UserName
	return
}

func (rpc *RpcLogic) Push(req *proto.Send) (code int, msg string) {
	reply := new(proto.SuccessReply)
	err := logicRpcClient.Call(context.Background(), "Push", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	msg = reply.Msg
	return
}
func (rpc *RpcLogic) PushRoom(req *proto.Send) (code int, msg string) {
	reply := new(proto.SuccessReply)
	err := logicRpcClient.Call(context.Background(), "PushRoom", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	msg = reply.Msg
	return
}
func (rpc *RpcLogic) Count(req *proto.Send) (code int, msg string) {
	reply := new(proto.SuccessReply)
	err := logicRpcClient.Call(context.Background(), "Count", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	msg = reply.Msg
	return
}
func (rpc *RpcLogic) GetRoomInfo(req *proto.Send) (code int, msg string) {
	reply := new(proto.SuccessReply)
	err := logicRpcClient.Call(context.Background(), "GetRoomInfo", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	msg = reply.Msg
	return
}

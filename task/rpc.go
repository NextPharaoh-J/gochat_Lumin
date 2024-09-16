package task

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"gochat_my/config"
	"gochat_my/proto"
	"gochat_my/tools"
	"strings"
	"sync"
	"time"
)

type Instance struct {
	ServerType string
	ServerId   string
	Client     client.XClient
}

type RpcConnectClient struct {
	lock         sync.Mutex
	ServerInsMap map[string][]Instance // serverId -- []instance 每个服务有多个实例
	IndexMap     map[string]int        // serverId -- index
}

var RClient = &RpcConnectClient{
	ServerInsMap: make(map[string][]Instance),
	IndexMap:     make(map[string]int),
}

func (rc *RpcConnectClient) GetClientByServerId(serverId string) (c client.XClient, err error) {
	// 轮询[]Instance 实现负载均衡
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.ServerInsMap[serverId]; !ok || len(rc.ServerInsMap[serverId]) <= 0 {
		return nil, errors.New("no connect layer ip :" + serverId)
	}
	if _, ok := rc.IndexMap[serverId]; !ok {
		rc.IndexMap = map[string]int{
			serverId: 0,
		}
	}
	idx := rc.IndexMap[serverId] % len(rc.ServerInsMap[serverId])
	ins := rc.ServerInsMap[serverId][idx]
	// 下次调用serverId的IndexMap时会自动到下一个实例，实现轮训服务的多个实例
	rc.IndexMap[serverId] = (rc.IndexMap[serverId] + 1) % len(rc.ServerInsMap[serverId])
	return ins.Client, nil
}
func (rc *RpcConnectClient) GetAllConnectTypeRpcClient() (rpcClientList []client.XClient) {
	for serverId := range rc.ServerInsMap {
		c, err := rc.GetClientByServerId(serverId)
		if err != nil {
			logrus.Infof("GetAllConnectTypeRpcClient err :%s", err.Error())
			continue
		}
		rpcClientList = append(rpcClientList, c)
	}
	return
}

func (task *Task) InitConnectRpcClient() (err error) {
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
		logrus.Fatalf("init task rpc etcd discovery client fail : %s", e.Error())
	}
	if len(d.GetServices()) <= 0 {
		logrus.Panicf("no etcd server find !")
	}
	for _, connectConf := range d.GetServices() {
		logrus.Infof("key is %s,value is %s", connectConf.Key, connectConf.Value)
		// RpcConnectClients
		serverType := getParamByKey(connectConf.Value, "serverType")
		serverId := getParamByKey(connectConf.Value, "serverId")
		logrus.Infof("serverType is %s,serverId is %s", serverType, serverId)
		if serverType == "" || serverId == "" {
			continue
		}
		d, e := client.NewPeer2PeerDiscovery(connectConf.Key, "")
		if e != nil {
			logrus.Fatalf("init task rpc client fail : %s", e.Error())
			continue
		}
		c := client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathConnect, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		ins := Instance{
			ServerType: serverType,
			ServerId:   serverId,
			Client:     c,
		}
		if _, ok := RClient.ServerInsMap[serverId]; !ok {
			RClient.ServerInsMap[serverId] = []Instance{ins}
		} else {
			RClient.ServerInsMap[serverId] = append(RClient.ServerInsMap[serverId], ins)
		}
	}
	go task.watchServicesChange(d)
	return
}
func (task *Task) watchServicesChange(d client.ServiceDiscovery) { // create a service index map
	for kvChan := range d.WatchService() {
		if len(kvChan) <= 0 {
			logrus.Errorf("connect services change ,connect alarm ,no abilable ip")
		}
		logrus.Infof("connect services change trigger ... ")
		insMap := make(map[string][]Instance)
		for _, kv := range kvChan {
			logrus.Infof("connect services change ,key is %s ,value is %s ", kv.Key, kv.Value)
			serverType := getParamByKey(kv.Value, "serverType")
			serverId := getParamByKey(kv.Value, "serverId")
			logrus.Infof("serverType is %s ,serverId is %s ", serverType, serverId)
			if serverType == "" || serverId == "" {
				continue
			}
			d, e := client.NewPeer2PeerDiscovery(kv.Key, "")
			if e != nil {
				logrus.Errorf("init task client.NewPeer2PeerDiscovery watch client fail : %s", e.Error())
				continue
			}
			c := client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathConnect, client.Failtry, client.RandomSelect, d, client.DefaultOption)
			ins := Instance{
				ServerType: serverType,
				ServerId:   serverId,
				Client:     c,
			}
			if _, ok := RClient.ServerInsMap[serverId]; !ok {
				insMap[serverId] = []Instance{ins}
			} else {
				insMap[serverId] = append(insMap[serverId], ins)
			}
			RClient.lock.Lock()
			RClient.ServerInsMap = insMap
			RClient.lock.Unlock()
		}

	}
}
func getParamByKey(s, key string) string {
	params := strings.Split(s, "&")
	for _, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) == 2 && kv[0] == key {
			return kv[1]
		}
	}
	logrus.Infof("getParamByKey Failed ,k-v pair not exist")
	return ""
}

func (task *Task) pushSingleToConnect(serverId string, userId int, msg []byte) {
	logrus.Infof("pushSingleToConnect Msg : %s", string(msg))
	pushMsgReq := &proto.PushMsgRequest{
		UserId: userId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpSingleSend,
			SeqId:     tools.GetSnowflakeId(),
			Body:      msg,
		},
	}
	reply := &proto.SuccessReply{}
	connectRpc, err := RClient.GetClientByServerId(serverId)
	if err != nil {
		logrus.Infof("get rpc client err : %s", err.Error())
	}
	err = connectRpc.Call(context.Background(), "PushSingleMsg", pushMsgReq, reply)
	if err != nil {
		logrus.Infof("pushSingleToConnect call err : %v", err)
	}
	logrus.Infof("reply %s", reply.Msg)
}
func (task *Task) broadcastRoomToConnect(roomId int, msg []byte) {
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomSend,
			SeqId:     tools.GetSnowflakeId(),
			Body:      msg,
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient()
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomToConnect rpc %v", rpc)
		_ = rpc.Call(context.Background(), "PushRoomMsg", pushRoomMsgReq, reply)
		logrus.Infof("broadcastRoomToConnect reply %s", reply.Msg)
	}
}
func (task *Task) broadcastRoomCountToConnect(roomId, count int) {
	msg := &proto.RedisRoomCountMsg{Count: count, Op: config.OpRoomCountSend}
	var body []byte
	var err error
	if body, err = json.Marshal(msg); err != nil {
		logrus.Warnf("broadcastRoomCountToConnect marshal err %v", err)
		return
	}
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomSend,
			SeqId:     tools.GetSnowflakeId(),
			Body:      body,
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient()
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomCountToConnect rpc %v", rpc)
		_ = rpc.Call(context.Background(), "PushRoomCount", pushRoomMsgReq, reply)
		logrus.Infof("broadcastRoomCountToConnect reply %s", reply.Msg)
	}
}
func (task *Task) broadcastRoomInfoToConnect(roomId int, roomUserInfo map[string]string) {
	msg := &proto.RedisRoomInfo{
		Count:        len(roomUserInfo),
		Op:           config.OpRoomInfoSend,
		RoomUserInfo: roomUserInfo,
		RoomId:       roomId,
	}
	var body []byte
	var err error
	if body, err = json.Marshal(msg); err != nil {
		logrus.Warnf("broadcastRoomInfoToConnect marshal err %v", err)
		return
	}
	pushRoomMsgReq := &proto.PushRoomMsgRequest{
		RoomId: roomId,
		Msg: proto.Msg{
			Ver:       config.MsgVersion,
			Operation: config.OpRoomSend,
			SeqId:     tools.GetSnowflakeId(),
			Body:      body,
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := RClient.GetAllConnectTypeRpcClient()
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomInfoToConnect rpc %v", rpc)
		_ = rpc.Call(context.Background(), "PushRoomInfo", pushRoomMsgReq, reply)
		logrus.Infof("broadcastRoomInfoToConnect rpc reply %s", reply.Msg)
	}
}

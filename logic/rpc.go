package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"gochat_my/logic/dao"
	"gochat_my/proto"
	"gochat_my/tools"
	"strconv"
	"time"
)

// MicroServe Instance of Logic
type RpcLogic struct{}

func (rpc *RpcLogic) Register(ctx context.Context, args *proto.RegisterRequest, reply *proto.RegisterReply) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	uData := u.CheckHaveUserName(u.UserName)
	if uData.Id > 0 {
		return errors.New("userName already exist,plz login")
	}
	u.UserName = args.Name
	u.Password = args.Password
	userId, err := u.Add()
	if err != nil {
		logrus.Infof("register add err :%s", err.Error())
		return
	}
	if userId == 0 {
		return errors.New("userId register empty")
	}
	randToken := tools.GetRandomToken(32)
	sessionId := tools.CreateSeessionId(randToken)
	userData := make(map[string]interface{})
	userData["userId"] = userId
	userData["sessionId"] = sessionId
	RedisSessClient.Do("MULTI")                          //HMSet：将多个字段和值设置到 Redis 哈希表中。
	RedisSessClient.HMSet(sessionId, userData)           //MULTI 和 EXEC：用于开启和提交 Redis 事务，保证多个命令按顺序原子性地执行。
	RedisSessClient.Expire(sessionId, 86400*time.Second) //Expire：为 Redis 键设置过期时间，使其在特定时间后自动删除。
	err = RedisSessClient.Do("EXEC").Err()
	if err != nil {
		logrus.Infof("register set redis token fail !")
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}
func (rpc *RpcLogic) Login(ctx context.Context, args *proto.LoginRequest, reply *proto.LoginResponse) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	userName := args.Name
	password := args.Password
	data := u.CheckHaveUserName(userName)
	if data.Id == 0 || (password != data.Password) {
		return errors.New("userName not exist or wrong password")
	}
	// token setting
	loginSessionId := tools.GetSessionIdByUserId(data.Id)
	randToken := tools.GetRandomToken(32)
	sessionId := tools.CreateSeessionId(randToken)
	userData := make(map[string]interface{})
	userData["userId"] = data.Id
	userData["userName"] = data.UserName
	// login check
	token, _ := RedisSessClient.Get(loginSessionId).Result()
	if token != "" {
		oldSession := tools.CreateSeessionId(token)
		err := RedisSessClient.Del(oldSession).Err()
		if err != nil {
			return errors.New("logout user fail ,token is " + token)
		}
	}
	RedisSessClient.Do("MULTI")
	RedisSessClient.HMSet(sessionId, userData)
	RedisSessClient.Expire(sessionId, 86400*time.Second)
	RedisSessClient.Set(loginSessionId, randToken, 86400*time.Second)
	err = RedisSessClient.Do("EXEC").Err()
	if err != nil {
		logrus.Infof("login set redis token fail !")
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}
func (rpc *RpcLogic) Logout(ctx context.Context, args *proto.LogoutRequest, reply *proto.LogoutResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := tools.GetSessionName(authToken)
	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail,authToken is :%s !", authToken)
		return
	}
	if len(userDataMap) == 0 {
		logrus.Infof("user session emtpy,authToken is :%s !", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	sessIdMap := tools.GetSessionIdByUserId(intUserId)
	err = RedisSessClient.Del(sessIdMap).Err()
	if err != nil {
		logrus.Infof("logout del sess fail :%s !", err.Error())
		return
	}
	logic := new(Logic)
	serverIdKey := logic.getUserKey(fmt.Sprintf("%d", intUserId))
	err = RedisSessClient.Del(serverIdKey).Err()
	if err != nil {
		logrus.Infof("logout del server id error:%s !", err.Error())
		return
	}
	err = RedisSessClient.Del(sessionName).Err()
	if err != nil {
		logrus.Infof("logout fail :%s !", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}
func (rpc *RpcLogic) CheckAuth(ctx context.Context, args *proto.CheckAuthRequest, reply *proto.CheckAuthResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := tools.GetSessionName(authToken)
	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail,authToken is :%s !", authToken)
		return
	}
	if len(userDataMap) == 0 {
		logrus.Infof("user session emtpy,authToken is :%s !", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	reply.UserId = intUserId
	userName, _ := userDataMap["userName"]
	reply.Code = config.SuccessReplyCode
	reply.UserName = userName
	return
}
func (rpc *RpcLogic) GetUserInfoByUserId(ctx context.Context, args *proto.GetUserInfoRequest, reply *proto.GetUserInfoResponse) (err error) {
	reply.Code = config.FailReplyCode
	userId := args.UserId
	u := new(dao.User)
	userName := u.GetUserNameByUserId(userId)
	reply.UserId = userId
	reply.UserName = userName
	reply.Code = config.SuccessReplyCode
	return
}

// single send msg
func (rpc *RpcLogic) Push(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	var bodyBytes []byte
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic/rpc push send data error :%s !", err.Error())
		return
	}
	logic := new(Logic)
	userKey := logic.getUserKey(strconv.Itoa(sendData.ToUserId))
	serverIdStr := RedisSessClient.Get(userKey).Val()
	err = logic.RedisPublishChannel(serverIdStr, sendData.ToUserId, bodyBytes)
	if err != nil {
		logrus.Errorf("logic/rpc  RedisPublishChannel error :%s !", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// push msg to room
func (rpc *RpcLogic) PushRoom(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	roomUSerKey := logic.getUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisSessClient.HGetAll(roomUSerKey).Result()
	if err != nil {
		logrus.Errorf("logic/rpc pushroom redis.HGet fail :%s !", roomUSerKey)
	}
	var bodyBytes []byte
	sendData.RoomId = roomId
	sendData.Op = config.OpSingleSend
	sendData.Msg = args.Msg
	sendData.FromUserName = args.FromUserName
	sendData.FromUserId = args.FromUserId
	sendData.CreateTime = tools.GetNowDateTime()
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic/rpc push send data error :%s !", err.Error())
		return
	}
	err = logic.RedisPublishRoomInfo(roomId, len(roomUserInfo), roomUserInfo, bodyBytes)
	if err != nil {
		logrus.Errorf("logic/rpc  RedisPublishChannel error :%s !", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// get room online person count
func (rpc *RpcLogic) Count(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	roomId := args.RoomId
	logic := new(Logic)
	var count int
	count, err = RedisSessClient.Get(logic.getRoomOnlineCountKey(strconv.Itoa(roomId))).Int()
	if err != nil {
		logrus.Errorf("logic/rpc Room Online Count is not exist:%s", err.Error())
	}
	err = logic.RedisPublishRoomCount(roomId, count)
	if err != nil {
		logrus.Errorf("logic/rpc Room Online Count err :%s", err.Error())
	}
	reply.Code = config.SuccessReplyCode
	return
}
func (rpc *RpcLogic) GetRoomInfo(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	roomUSerKey := logic.getUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisSessClient.HGetAll(roomUSerKey).Result()
	if len(roomUserInfo) == 0 {
		return errors.New("logic/rpc getRoominfo no this user ")
	}
	err = logic.RedisPushRoomInfo(roomId, len(roomUserInfo), roomUserInfo)
	if err != nil {
		logrus.Errorf("logic/rpc  RedisPublishChannel error :%s !", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) Connect(ctx context.Context, args *proto.ConnectRequest, reply *proto.ConnectReply) (err error) {
	return
}
func (rpc *RpcLogic) DisConnect(ctx context.Context, args *proto.ConnectRequest, reply *proto.DisConnectReply) (err error) {
	return
}

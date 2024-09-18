package handler

import (
	"github.com/gin-gonic/gin"
	"gochat_my/api/rpc"
	"gochat_my/config"
	"gochat_my/proto"
	"gochat_my/tools"
	"strconv"
)

type FormPush struct {
	Msg       string `form:"msg" json:"msg" binding:"required"`
	ToUserId  string `form:"to_user_id" json:"to_user_id" binding:"required"`
	RoomId    int    `form:"room_id" json:"room_id" binding:"required"`
	AuthToken string `form:"auth_token" json:"auth_token" binding:"required"`
}

func Push(c *gin.Context) {
	var (
		form         FormPush
		code         int
		toUserId     int
		fromUserId   int
		toUserName   string
		fromUserName string
		rpcMsg       string
	)
	if err := c.ShouldBind(&form); err != nil {
		tools.FailWtihMsg(c, err.Error())
		return
	}
	toUserId, _ = strconv.Atoi(form.ToUserId)
	getUserNameReq := &proto.GetUserInfoRequest{UserId: toUserId}
	code, toUserName = rpc.RpcLogicObj.GetUserNameByUserId(getUserNameReq)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc fail get toUserName ")
		return
	}
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: form.AuthToken}
	code, fromUserId, fromUserName = rpc.RpcLogicObj.CheckAuth(checkAuthReq)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc fail checkAuth")
		return
	}
	roomId := form.RoomId
	req := &proto.Send{
		Msg:          form.Msg,
		ToUserId:     toUserId,
		ToUserName:   toUserName,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		RoomId:       roomId,
		Op:           config.OpSingleSend,
	}
	code, rpcMsg = rpc.RpcLogicObj.Push(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, rpcMsg)
		return
	}
	tools.SuccessWtihMsg(c, "ok", nil)
}

type FormRoom struct {
	AuthToken string `form:"auth_token" json:"auth_token" binding:"required"`
	RoomId    int    `form:"room_id" json:"room_id" binding:"required"`
	Msg       string `form:"msg" json:"msg" binding:"required"`
}

func PushRoom(c *gin.Context) {
	var (
		form         FormRoom
		authcode     int
		code         int
		fromUserId   int
		fromUserName string
	)
	if err := c.ShouldBind(&form); err != nil {
		tools.FailWtihMsg(c, err.Error())
		return
	}
	CheckAuthReq := &proto.CheckAuthRequest{AuthToken: form.AuthToken}
	authcode, fromUserId, fromUserName = rpc.RpcLogicObj.CheckAuth(CheckAuthReq)
	if authcode == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc fail get self info ")
		return
	}
	req := &proto.Send{
		Msg:          form.Msg,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		RoomId:       form.RoomId,
		Op:           config.OpRoomSend,
	}
	code, _ = rpc.RpcLogicObj.PushRoom(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc push room msg fail")
		return
	}
	tools.SuccessWtihMsg(c, "ok", nil)
}

type FormCount struct {
	RoomId int `form:"room_id" json:"room_id" binding:"required"`
}

func Count(c *gin.Context) {
	var (
		form FormCount
		code int
	)
	if err := c.ShouldBind(&form); err != nil {
		tools.FailWtihMsg(c, err.Error())
		return
	}
	req := &proto.Send{
		RoomId: form.RoomId,
		Op:     config.OpRoomSend,
	}
	code, _ = rpc.RpcLogicObj.Count(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc get room count fail")
		return
	}
	tools.SuccessWtihMsg(c, "ok", nil)
}

type FormRoomInfo struct {
	RoomId int `form:"room_id" json:"room_id" binding:"required"`
}

func GetRoomInfo(c *gin.Context) {
	var (
		form FormRoomInfo
		code int
		msg  string
	)
	if err := c.ShouldBind(&form); err != nil {
		tools.FailWtihMsg(c, err.Error())
		return
	}
	req := &proto.Send{
		RoomId: form.RoomId,
		Op:     config.OpRoomSend,
	}
	code, msg = rpc.RpcLogicObj.PushRoom(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "rpc get room info fail")
		return
	}
	tools.SuccessWtihMsg(c, msg, nil)
}

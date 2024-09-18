package handler

import (
	"github.com/gin-gonic/gin"
	"gochat_my/api/rpc"
	"gochat_my/proto"
	"gochat_my/tools"
)

type FormLogin struct {
	userName string `form:"username" json:"username" binding:"required"`
	password string `form:"password" json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var form FormLogin
	if err := c.ShouldBindJSON(&form); err != nil {
		//logrus.Errorf("api login formData bind err: %v", err)
		tools.FailWtihMsg(c, err.Error())
	}
	req := &proto.LoginRequest{
		Name:     form.userName,
		Password: form.password,
	}
	code, authToken, msg := rpc.RpcLogicObj.Login(req)
	if code == tools.CodeFail || authToken == "" {
		tools.FailWtihMsg(c, msg)
		return
	}
	tools.SuccessWtihMsg(c, "login success", authToken)
}

type FormRegister struct {
	userName string `form:"username" json:"username" binding:"required"`
	password string `form:"password" json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var form FormRegister
	if err := c.ShouldBindJSON(&form); err != nil {
		//logrus.Errorf("api login formData bind err: %v", err)
		tools.FailWtihMsg(c, err.Error())
	}
	req := &proto.RegisterRequest{
		Name:     form.userName,
		Password: form.password,
	}
	code, authToken, msg := rpc.RpcLogicObj.Register(req)
	if code == tools.CodeFail || authToken == "" {
		tools.FailWtihMsg(c, msg)
		return
	}
	tools.SuccessWtihMsg(c, "register success", authToken)
}

type FormCheckAuth struct {
	AuthToken string `form:"auth_token" json:"auth_token" binding:"required"`
}

func CheckAuth(c *gin.Context) {
	var form FormCheckAuth
	if err := c.ShouldBindJSON(&form); err != nil {
		//logrus.Errorf("api login formData bind err: %v", err)
		tools.FailWtihMsg(c, err.Error())
	}
	req := &proto.CheckAuthRequest{
		AuthToken: form.AuthToken,
	}
	code, userId, userName := rpc.RpcLogicObj.CheckAuth(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "auth fail")
		return
	}
	var jsonReply = map[string]interface{}{
		"userName": userName,
		"userId":   userId,
	}
	tools.SuccessWtihMsg(c, "register success", jsonReply)
}

type FormLogout struct {
	AuthToken string `form:"auth_token" json:"auth_token" binding:"required"`
}

func Logout(c *gin.Context) {
	var form FormLogout
	if err := c.ShouldBindJSON(&form); err != nil {
		//logrus.Errorf("api login formData bind err: %v", err)
		tools.FailWtihMsg(c, err.Error())
	}
	req := &proto.LogoutRequest{
		AuthToken: form.AuthToken,
	}
	code := rpc.RpcLogicObj.Logout(req)
	if code == tools.CodeFail {
		tools.FailWtihMsg(c, "logout fail")
		return
	}
	tools.SuccessWtihMsg(c, "logout ok", nil)
}

package router

import (
	"github.com/gin-gonic/gin"
	"gochat_my/api/handler"
	"gochat_my/api/rpc"
	"gochat_my/proto"
	"gochat_my/tools"
	"net/http"
)

func Register() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())
	initUserRouter(r)
	initPushRouter(r)
	r.NoRoute(func(c *gin.Context) {
		tools.FailWtihMsg(c, "please check request url !")
	})
	return r
}
func initUserRouter(r *gin.Engine) {
	userGroup := r.Group("/user")
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)
	userGroup.Use(CheckSessionId())
	{
		userGroup.POST("/logout", handler.Logout)
		userGroup.POST("/CheckAuth", handler.CheckAuth)
	}

}
func initPushRouter(r *gin.Engine) {
	pushGroup := r.Group("/push")
	pushGroup.Use(CheckSessionId())
	{
		pushGroup.POST("push", handler.Push)
		pushGroup.POST("pushRoom", handler.PushRoom)
		pushGroup.POST("/count", handler.Count)
		pushGroup.POST("/getRoomInfo", handler.GetRoomInfo)
	}

}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin,X-Requested-With,Content-Type,Accept")
			c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}

type FormCheckSessionId struct {
	AuthToken string `json:"auth_token" form:"auth_token" binding:"required"`
}

func CheckSessionId() gin.HandlerFunc {
	return func(c *gin.Context) {
		var formCheck FormCheckSessionId
		if err := c.ShouldBindBodyWithJSON(&formCheck); err != nil {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, nil, nil)
			return
		}
		authToken := formCheck.AuthToken
		req := &proto.CheckAuthRequest{
			AuthToken: authToken,
		}
		code, userId, userName := rpc.RpcLogicObj.CheckAuth(req)
		if code == tools.CodeFail || userId <= 0 || userName == "" {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, nil, nil)
			return
		}
		c.Next()
		return
	}
}

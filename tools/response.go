package tools

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	CodeSuccess      = 0
	CodeFail         = 1
	CodeUnknowError  = -1
	CodeSessionError = 4000
)

var MsgCodeMap = map[int]string{
	CodeUnknowError:  "unknow error",
	CodeSuccess:      "success",
	CodeFail:         "fail",
	CodeSessionError: "session error",
}

func ResponseWithCode(c *gin.Context, code int, msg interface{}, data interface{}) {
	if msg == nil {
		if val, ok := MsgCodeMap[code]; ok {
			msg = val
		} else {
			msg = MsgCodeMap[-1]
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})

}

func SuccessWtihMsg(c *gin.Context, msg interface{}, data interface{}) {
	ResponseWithCode(c, http.StatusOK, msg, data)
}
func FailWtihMsg(c *gin.Context, msg interface{}) {
	ResponseWithCode(c, http.StatusOK, msg, nil)
}

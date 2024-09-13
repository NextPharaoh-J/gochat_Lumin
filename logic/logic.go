package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"runtime"
)

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	logicConf := config.Conf.Logic
	runtime.GOMAXPROCS(logicConf.LogicBase.CpuNum)
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())
	if err := logic.InitPublishRedisClient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail , err : %s", err.Error())
	}
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init RpcServer fail ,err: %s", err.Error())
	}

}

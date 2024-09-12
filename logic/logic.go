package logic

import "GOgochat_my/config"

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	logicConf := config.Conf.Logic
	_ = logicConf
}

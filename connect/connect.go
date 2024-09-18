package connect

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"runtime"
	"time"
)

var DefaultServer *Server

type Connect struct {
	ServerId string
}

func New() *Connect {
	return new(Connect)
}

func (c *Connect) Run() {
	connectConf := config.Conf.Connect
	runtime.GOMAXPROCS(connectConf.ConnectBucket.CpuNum) //maximum cpus executing
	if err := c.InitLogicRpcClient(); err != nil {
		logrus.Panicf("InitLogicRpcClient fail , err : %s", err.Error())
	}
	Buckets := make([]*Bucket, config.Conf.Connect.ConnectBucket.CpuNum)
	for i := 0; i < config.Conf.Connect.ConnectBucket.CpuNum; i++ {
		Buckets[i] = NewBucket(BucketOptions{
			ChannelSize:   connectConf.ConnectBucket.Channel,
			RoomSize:      connectConf.ConnectBucket.Room,
			RoutineSize:   connectConf.ConnectBucket.RoutineSize,
			RoutineAmount: connectConf.ConnectBucket.RoutineAmount,
		})
	}
	operator := new(DefaultOperator)
	DefaultServer = NewServer(Buckets, operator, ServerOptions{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	c.ServerId = fmt.Sprintf("%s-%s", "ws", uuid.New().String())
	// init connect layer rpc server ,task layer will call this
	if err := c.InitConnectWebsocketRpcServer(); err != nil {
		logrus.Panicf("InitConnectWebSocketRpcServer Fatal error : %s \n", err.Error())
	}
	// start connect layer server handler persistent connection
	if err := c.InitWebSocket(); err != nil {
		logrus.Panicf("Connect layer InitWebSocket error : %s \n", err.Error())
	}
}

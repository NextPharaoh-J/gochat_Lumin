package task

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"gochat_my/tools"
	"time"
)

var RedisClient *redis.Client

func (task *Task) InitQueueRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.DB,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisClient Ping Result pong : %s, err : %s ", pong, err.Error())
		return err
	}
	go func() {
		for {
			var result []string
			// 10s timeout: if queue empty 10s throw info,or pop a message from RedisQueue
			result, err = RedisClient.BRPop(time.Second*10, config.QueueName).Result()
			if err != nil {
				logrus.Infof("RedisClient BRPop timeout, Err : %s ", err.Error())
			}
			if len(result) >= 2 { //data struct unzip [name,value]
				task.Push(result[1])
			}
		}
	}()

	return
}

package task

import (
	"gochat_my/config"
	"gochat_my/tools"
	"testing"
	"time"
)

func Test_TestQueue(t *testing.T) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.DB,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)

	// 向 Redis 队列推送测试数据
	err := RedisClient.LPush(config.QueueName, "test-message").Err()
	if err != nil {
		t.Fatalf("failed to push message to Redis: %v", err)
	}

	result, err := RedisClient.BRPop(time.Second*10, config.QueueName).Result()
	if err != nil {
		t.Fail()
	}
	t.Log(result, len(result))
	if len(result) >= 1 {
		t.Log(result[1])
	}
}

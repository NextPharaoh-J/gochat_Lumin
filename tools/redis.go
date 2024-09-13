package tools

import (
	"github.com/go-redis/redis"
	"sync"
	"time"
)

var syncLock sync.Mutex
var RedisClientMap = map[string]*redis.Client{}

type RedisOption struct {
	Address  string
	Password string
	Db       int
}

func GetRedisInstance(option RedisOption) *redis.Client {
	// 注册RedisClient 并放在Map中
	addr := option.Address
	pwd := option.Password
	db := option.Db
	syncLock.Lock()
	if redisClient, ok := RedisClientMap[addr]; ok {
		return redisClient
	}
	client := redis.NewClient(&redis.Options{
		Addr:       addr,
		Password:   pwd,
		DB:         db,
		MaxConnAge: 20 * time.Second,
	})
	RedisClientMap[addr] = client
	syncLock.Unlock()
	return RedisClientMap[addr]
}

package storage

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
)

func initRedis(dsn string) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(opts)

	_, pingErr := redisClient.Ping(context.Background()).Result()
	if pingErr != nil {
		panic(pingErr)
	}
}

func GetRedisClient(dsn string) *redis.Client {
	if redisClient == nil {
		initRedis(dsn)
	}
	return redisClient
}

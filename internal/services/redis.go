package services

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

var redisOnce sync.Once
var myRedis *Redis

type Redis struct {
	client *redis.Client
}

func GetRedis() *Redis {
	redisOnce.Do(func() {
		myRedis = &Redis{}
	})
	return myRedis
}

func (rd *Redis) Connect(ctx context.Context) {
	config := GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config.Notification.RedisAddr,
		DB:       config.Notification.RedisDB,
		Username: config.Notification.RedisUsername,
		Password: config.Notification.RedisPassword,
	})
	rd.client = client
}

func (rd *Redis) Close() error {
	return rd.client.Close()
}

func (rd *Redis) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return rd.client.Publish(ctx, channel, message).Result()
}

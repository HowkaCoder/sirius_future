package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(addr string) *RedisService {
	if addr == "" {
		return &RedisService{}
	}
	return &RedisService{client: redis.NewClient(&redis.Options{
		Addr:        addr,
		DialTimeout: time.Second,
		ReadTimeout: time.Second,
	})}
}

func (rs *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	if rs.client == nil {
		return nil
	}
	return rs.client.Set(context.Background(), key, value, expiration).Err()
}

func (rs *RedisService) Get(key string) (string, error) {
	if rs.client == nil {
		return "", redis.Nil
	}
	return rs.client.Get(context.Background(), key).Result()
}

func (rs *RedisService) Delete(key string) error {
	if rs.client == nil {
		return nil
	}
	return rs.client.Del(context.Background(), key).Err()
}

package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

var ctx = context.Background()

func NewRedisService(addr string) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, // адрес Redis (например, "localhost:6379")
	})
	return &RedisService{client: rdb}
}

// Set устанавливает значение в кэш
func (rs *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	return rs.client.Set(ctx, key, value, expiration).Err()
}

// Get получает значение из кэша
func (rs *RedisService) Get(key string) (string, error) {
	return rs.client.Get(ctx, key).Result()
}

// Delete удаляет значение из кэша
func (rs *RedisService) Delete(key string) error {
	return rs.client.Del(ctx, key).Err()
}

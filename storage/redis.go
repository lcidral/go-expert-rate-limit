package storage

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{client: client}
}

func (r *RedisStorage) Get(key string) (int, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (r *RedisStorage) Set(key string, value int, expiration time.Duration) error {
	ctx := context.Background()
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisStorage) Incr(key string) error {
	ctx := context.Background()
	return r.client.Incr(ctx, key).Err()
}

func (r *RedisStorage) IsBlocked(key string) bool {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key+"_blocked").Result()
	return err == nil && val == "true"
}

func (r *RedisStorage) Block(key string, duration time.Duration) error {
	ctx := context.Background()
	return r.client.Set(ctx, key+"_blocked", "true", duration).Err()
}

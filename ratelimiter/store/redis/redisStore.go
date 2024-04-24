package ratelimiter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_ADDR"),
		})
	}
	return &RedisStore{client: client}
}

func (rs *RedisStore) Get(ctx context.Context, key string) (int, error) {
	count, err := rs.client.Get(ctx, key).Int()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		fmt.Printf("Error accessing Redis: %v\n", err)
		return 0, err
	}
	return count, nil
}

func (rs *RedisStore) IsBlocked(ctx context.Context, key string) (bool, error) {
	count, err := rs.client.Get(ctx, key+":blocked").Int()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		fmt.Printf("Error accessing Redis: %v\n", err)
		return false, err
	}
	return count >= 1, nil
}

func (rs *RedisStore) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	result, err := rs.client.Incr(ctx, key).Result()
	if err != nil {
		fmt.Printf("Error incrementing count in Redis: %v\n", err)
		return 0, err
	}
	if result == 1 {
		rs.client.Expire(ctx, key, time.Second*expiration).Result()
	}
	return int(result), nil
}

func (rs *RedisStore) Block(ctx context.Context, key string, blockDuration time.Duration) error {
	_, err := rs.client.Incr(ctx, key+":blocked").Result()
	if err != nil {
		fmt.Printf("Error blocking in Redis: %v\n", err)
		return err
	}
	rs.client.Expire(ctx, key+":blocked", time.Second*blockDuration).Result()
	return nil
}

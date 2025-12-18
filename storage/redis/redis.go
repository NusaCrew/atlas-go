package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/NusaCrew/atlas-go/config"
	"github.com/NusaCrew/atlas-go/log"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Close() error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	log.Info("successfully connected to redis")
	return &redisClient{
		client: rdb,
	}, nil
}

func (r *redisClient) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisClient) Close() error {
	return r.client.Close()
}

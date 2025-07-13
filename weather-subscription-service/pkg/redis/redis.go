package redis

import (
	"context"
	"time"
	"weather-subscription/internal/config"
	"weather-subscription/internal/svc"

	redisGo "github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redisGo.Client
}

func NewClient(cfg config.RedisConfig) *Redis {
	client := redisGo.NewClient(&redisGo.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Redis{client: client}
}

func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redisGo.Nil {
		return "", svc.ErrorCacheMiss
	}
	return val, err
}

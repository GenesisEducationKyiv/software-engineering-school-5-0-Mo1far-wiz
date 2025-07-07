package redis

import (
	"context"
	"time"
	"weather/internal/config"
<<<<<<< HEAD:pkg/redis/redis.go
	"weather/internal/svc"
=======
	"weather/internal/srverrors"
>>>>>>> 633e62d (architecture image added + architecture tests and refactoring):internal/redis/redis.go

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
<<<<<<< HEAD:pkg/redis/redis.go
		return "", svc.ErrorCacheMiss
=======
		return "", srverrors.ErrCacheMiss
>>>>>>> 633e62d (architecture image added + architecture tests and refactoring):internal/redis/redis.go
	}
	return val, err
}

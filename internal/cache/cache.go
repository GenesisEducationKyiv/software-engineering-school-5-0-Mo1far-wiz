// internal/cache/cache.go
package cache

import (
	"context"
	"errors"
	"time"
)

const defaultTTL = time.Hour

var ErrCacheMiss = errors.New("cache: key not found")

type Cacher interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type CacheService struct {
	cache      Cacher
	defaultTTL time.Duration
}

func NewCacheService(cache Cacher) *CacheService {
	return &CacheService{
		cache:      cache,
		defaultTTL: defaultTTL,
	}
}

func (s *CacheService) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	return s.cache.Set(ctx, key, value, ttl)
}

func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.cache.Get(ctx, key)
}

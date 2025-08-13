package cache

import (
	"context"
	"errors"
	"time"
	"weather-subscription/internal/metrics"
	"weather-subscription/internal/svc"
)

const defaultTTL = time.Hour

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

	start := time.Now()
	err := s.cache.Set(ctx, key, value, ttl)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.CacheOpsTotal.WithLabelValues("set", status).Inc()
	metrics.CacheLatency.WithLabelValues("set").Observe(duration)

	return err
}

func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	val, err := s.cache.Get(ctx, key)
	duration := time.Since(start).Seconds()

	var status string
	switch {
	case err == nil:
		status = "success"
	case errors.Is(err, svc.ErrorCacheMiss):
		status = "miss"
	default:
		status = "error"
	}

	metrics.CacheOpsTotal.WithLabelValues("get", status).Inc()
	metrics.CacheLatency.WithLabelValues("get").Observe(duration)

	return val, err
}

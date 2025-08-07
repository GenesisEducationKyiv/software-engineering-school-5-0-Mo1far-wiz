package weather

import (
	"context"
	"encoding/json"
	"weather-subscription/internal/cache"
	"weather-subscription/internal/models"
	"weather-subscription/internal/svc"

	joinErr "errors"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type APIInterface interface {
	GetCityWeather(ctx context.Context, city string) (models.Weather, error)
	setNext(APIInterface)
}

type WeatherAPIService struct {
	head   APIInterface
	cache  *cache.CacheService
	logger Logger
}

func (rs *WeatherAPIService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	weatherStr, err := rs.cache.Get(ctx, city)
	if err == nil {
		var weather models.Weather
		if err := json.Unmarshal([]byte(weatherStr), &weather); err != nil {
			return models.Weather{}, errors.Wrap(err, "unmarshal weather cache")
		}
		return weather, nil
	}
	if !errors.Is(err, svc.ErrorCacheMiss) {
		rs.logger.Error("cache service error", zap.Error(err))
	}

	weather, err := rs.head.GetCityWeather(ctx, city)
	if err != nil {
		return models.Weather{}, joinErr.Join(svc.ErrorGetCityWeather, err)
	}

	weatherRaw, err := json.Marshal(weather)
	if err != nil {
		return models.Weather{}, err
	}
	if err := rs.cache.Set(ctx, city, string(weatherRaw), 0); err != nil {
		rs.logger.Error("cache set failed", zap.Error(err))
	}

	return weather, nil
}

func NewWeatherAPIService(
	cache *cache.CacheService,
	logger Logger,
	sources ...APIInterface,
) (*WeatherAPIService, error) {
	if len(sources) == 0 {
		return nil, errors.New("need at least one API source")
	}

	rs := &WeatherAPIService{
		cache:  cache,
		logger: logger,
	}

	for i := 0; i < len(sources)-1; i++ {
		sources[i].setNext(sources[i+1])
	}

	rs.head = sources[0]
	return rs, nil
}

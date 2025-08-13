package service

import (
	"context"
	"errors"
	"time"
	"weather-subscription/internal/metrics"
	"weather-subscription/internal/models"
	"weather-subscription/internal/svc"
	"weather-subscription/internal/weather"

	"go.uber.org/zap"
)

type Weather struct {
	weatherAPIService *weather.WeatherAPIService
	logger            Logger
}

func NewWeather(weatherAPIService *weather.WeatherAPIService, logger Logger) *Weather {
	return &Weather{
		weatherAPIService: weatherAPIService,
		logger:            logger,
	}
}

func observeGetCityWeather(
	f func() (models.Weather, error),
) (models.Weather, error) {
	start := time.Now()
	result, err := f()
	duration := time.Since(start).Seconds()

	metrics.RequestsTotal.
		WithLabelValues("weather_service_get_city", metrics.StatusLabel(err)).
		Inc()
	metrics.RequestLatency.
		WithLabelValues("weather_service_get_city").
		Observe(duration)
	if err != nil {
		metrics.ErrorsTotal.
			WithLabelValues("weather_service_get_city", metrics.TypeLabel(err)).
			Inc()
	}
	return result, err
}

func (w *Weather) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	return observeGetCityWeather(func() (models.Weather, error) {
		data, err := w.weatherAPIService.GetCityWeather(ctx, city)
		if err != nil {
			w.logger.Error("on getting city weather", zap.String("city", city), zap.Error(err))
			return models.Weather{}, errors.Join(svc.ErrorGetCityWeather, err)
		}
		return data, nil
	})
}

package service

import (
	"context"
	"errors"
	"weather/internal/models"
	"weather/internal/svc"
	"weather/internal/weather"

	"go.uber.org/zap"
)

type WeatherService struct {
	weatherAPIService *weather.WeatherAPIService
	logger            Logger
}

func NewWeatherService(WeatherAPIService *weather.WeatherAPIService, logger Logger) *WeatherService {
	return &WeatherService{
		weatherAPIService: WeatherAPIService,
		logger:            logger,
	}
}

func (w *WeatherService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	weather, err := w.weatherAPIService.GetCityWeather(ctx, city)
	if err != nil {
		w.logger.ConsoleLogError("on getting city weather",
			zap.String("error", err.Error()))
		return models.Weather{}, errors.Join(svc.ErrorGetCityWeather, err)
	}

	return weather, nil
}

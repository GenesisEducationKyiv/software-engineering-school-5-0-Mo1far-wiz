package service

import (
	"context"
	"errors"
	"weather/internal/models"
	"weather/internal/svc"
	"weather/internal/weather"

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

func (w *Weather) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	weather, err := w.weatherAPIService.GetCityWeather(ctx, city)
	if err != nil {
		w.logger.ConsoleLogError("on getting city weather",
			zap.String("error", err.Error()))
		return models.Weather{}, errors.Join(svc.ErrorGetCityWeather, err)
	}

	return weather, nil
}

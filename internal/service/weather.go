package service

import (
	"context"
	"errors"
	"weather/internal/models"
	"weather/internal/svc"
	"weather/internal/weather"

	"go.uber.org/zap"
)

<<<<<<< HEAD
type Weather struct {
=======
type WeatherService struct {
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
	weatherAPIService *weather.WeatherAPIService
	logger            Logger
}

<<<<<<< HEAD
func NewWeather(weatherAPIService *weather.WeatherAPIService, logger Logger) *Weather {
	return &Weather{
		weatherAPIService: weatherAPIService,
=======
func NewWeatherService(WeatherAPIService *weather.WeatherAPIService, logger Logger) *WeatherService {
	return &WeatherService{
		weatherAPIService: WeatherAPIService,
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
		logger:            logger,
	}
}

<<<<<<< HEAD
func (w *Weather) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
=======
func (w *WeatherService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
	weather, err := w.weatherAPIService.GetCityWeather(ctx, city)
	if err != nil {
		w.logger.ConsoleLogError("on getting city weather",
			zap.String("error", err.Error()))
		return models.Weather{}, errors.Join(svc.ErrorGetCityWeather, err)
	}

	return weather, nil
}

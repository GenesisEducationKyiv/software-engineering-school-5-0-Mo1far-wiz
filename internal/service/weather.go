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
<<<<<<< HEAD
type Weather struct {
=======
type WeatherService struct {
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
=======
type Weather struct {
>>>>>>> 8d3b469 (services implemented)
	weatherAPIService *weather.WeatherAPIService
	logger            Logger
}

<<<<<<< HEAD
<<<<<<< HEAD
func NewWeather(weatherAPIService *weather.WeatherAPIService, logger Logger) *Weather {
	return &Weather{
		weatherAPIService: weatherAPIService,
=======
func NewWeatherService(WeatherAPIService *weather.WeatherAPIService, logger Logger) *WeatherService {
	return &WeatherService{
		weatherAPIService: WeatherAPIService,
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
=======
func NewWeather(weatherAPIService *weather.WeatherAPIService, logger Logger) *Weather {
	return &Weather{
		weatherAPIService: weatherAPIService,
>>>>>>> 8d3b469 (services implemented)
		logger:            logger,
	}
}

<<<<<<< HEAD
<<<<<<< HEAD
func (w *Weather) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
=======
func (w *WeatherService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
=======
func (w *Weather) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
>>>>>>> 8d3b469 (services implemented)
	weather, err := w.weatherAPIService.GetCityWeather(ctx, city)
	if err != nil {
		w.logger.ConsoleLogError("on getting city weather",
			zap.String("error", err.Error()))
		return models.Weather{}, errors.Join(svc.ErrorGetCityWeather, err)
	}

	return weather, nil
}

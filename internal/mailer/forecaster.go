package mailer

import (
	"context"
	"log"
	"weather/internal/models"
	"weather/internal/weather"

	"golang.org/x/exp/slices"
)

type Forecaster struct {
	weather *weather.RemoteService
}

func NewForecaster(weather *weather.RemoteService) *Forecaster {
	return &Forecaster{
		weather: weather,
	}
}

func (f *Forecaster) GetForecasts(
	ctx context.Context,
	subscriptions []models.Subscription,
) []models.Forecast {
	forecasts := make([]models.Forecast, len(subscriptions))
	for idx, sub := range subscriptions {
		weatherData, err := f.weather.GetCityWeather(ctx, sub.City)
		if err != nil {
			log.Printf("weather fetch error for %q: %v\n", sub.City, err)
			continue
		}

		forecasts[idx] = models.Forecast{
			Email:   sub.Email,
			Weather: weatherData,
		}
	}

	filter := func(f models.Forecast) bool { return f.Email == "" }
	forecasts = slices.DeleteFunc(forecasts, filter)

	return forecasts
}

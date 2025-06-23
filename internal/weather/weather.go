package weather

import (
	"context"
	"weather/internal/models"

	"go.uber.org/zap"
)

type Logger interface {
	LogInfo(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}

type APIInterface interface {
	GetCityWeather(ctx context.Context, city string) (models.Weather, error)
	setNext(APIInterface)
}

type WeatherService struct {
	head APIInterface
}

func (rs *WeatherService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	return rs.head.GetCityWeather(ctx, city)
}

func NewWeatherService(sources ...APIInterface) *WeatherService {
	rs := &WeatherService{}

	for i := 0; i < len(sources)-1; i++ {
		sources[i].setNext(sources[i+1])
	}

	rs.head = sources[0]
	return rs
}

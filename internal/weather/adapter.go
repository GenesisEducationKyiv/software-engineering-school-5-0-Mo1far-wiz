package weather

import (
	"context"
	"weather/internal/models"
)

type APIInterface interface {
	GetCityWeather(ctx context.Context, city string) (models.Weather, error)
}

type RemoteService struct {
	remote APIInterface
}

func (rs *RemoteService) GetCityWeather(ctx context.Context, city string) (models.Weather, error) {
	return rs.remote.GetCityWeather(ctx, city)
}

func NewRemoteService(api APIInterface) *RemoteService {
	return &RemoteService{
		remote: api,
	}
}

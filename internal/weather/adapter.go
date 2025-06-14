package weather

import (
	"weather/internal/models"
)

type APIInterface interface {
	GetCityWeather(city string) (models.Weather, error)
}

type RemoteService struct {
	remote APIInterface
}

func (rs *RemoteService) GetCityWeather(city string) (models.Weather, error) {
	return rs.remote.GetCityWeather(city)
}

func NewRemoteService(api APIInterface) *RemoteService {
	return &RemoteService{
		remote: api,
	}
}

package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"weather/internal/models"

	"github.com/pkg/errors"
)

type WeatherApiResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		TempF     float32 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		Humidity int `json:"humidity"`
	} `json:"current"`
}

func (wa WeatherApiResponse) GetWeatherModel() models.Weather {
	return models.Weather{
		Temperature: int(wa.Current.TempC),
		Humidity:    wa.Current.Humidity,
		Description: wa.Current.Condition.Text,
	}
}

type WeatherApi struct {
	BaseURL string
	ApiKey  string
}

func (wa *WeatherApi) GetCityWeather(city string) (models.Weather, error) {
	reqURL := wa.BaseURL + "?key=" + wa.ApiKey + "&q=" + city
	resp, err := http.Get(reqURL)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to send GET request to weather api")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return models.Weather{}, errors.New(fmt.Sprintf("city not found: %s", city))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to read request body")
	}

	var weather WeatherApiResponse
	err = json.Unmarshal(body, &weather)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to unmarshal request body")
	}

	return weather.GetWeatherModel(), nil
}

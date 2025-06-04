package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"weather/internal/models"

	"github.com/pkg/errors"
)

type WeatherAPIResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		TempF     float32 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		Humidity int `json:"humidity"`
	} `json:"current"`
}

func (wa WeatherAPIResponse) GetWeatherModel() models.Weather {
	return models.Weather{
		Temperature: int(wa.Current.TempC),
		Humidity:    wa.Current.Humidity,
		Description: wa.Current.Condition.Text,
	}
}

type WeatherAPI struct {
	BaseURL string
	APIKey  string
}

func (wa *WeatherAPI) GetCityWeather(city string) (weather models.Weather, err error) {
	reqURL := wa.BaseURL + "?key=" + wa.APIKey + "&q=" + city
	resp, err := http.Get(reqURL)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to send GET request to weather api")
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "failed to close response body")
		}
	}()

	if resp.StatusCode == http.StatusBadRequest {
		return models.Weather{}, fmt.Errorf("city not found: %s", city)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to read request body")
	}

	var weatherResp WeatherAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to unmarshal request body")
	}

	return weatherResp.GetWeatherModel(), nil
}

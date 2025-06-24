package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"weather/internal/config"
	"weather/internal/models"

	joinErr "errors"

	"github.com/pkg/errors"
)

type weatherAPIResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		TempF     float32 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		Humidity int `json:"humidity"`
	} `json:"current"`
}

func (wa weatherAPIResponse) getWeatherModel() models.Weather {
	return models.Weather{
		Temperature: int(wa.Current.TempC),
		Humidity:    wa.Current.Humidity,
		Description: wa.Current.Condition.Text,
	}
}

type WeatherAPI struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewWeatherAPI(config config.WeatherAPIConfig) *WeatherAPI {
	return &WeatherAPI{
		baseURL: config.ServiceBaseURL,
		apiKey:  config.APIKey,
		client:  http.DefaultClient,
	}
}

func (wa *WeatherAPI) WithClient(client *http.Client) *WeatherAPI {
	wa.client = client
	return wa
}

func (wa *WeatherAPI) GetCityWeather(ctx context.Context, city string) (weather models.Weather, err error) {
	reqURL := wa.baseURL + "?key=" + wa.apiKey + "&q=" + city

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to create new GET request to weather api")
	}

	resp, err := wa.client.Do(req)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to send GET request to weather api")
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			closeErr = errors.Wrap(closeErr, "failed to close response body")
			if err != nil {
				err = joinErr.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	if resp.StatusCode == http.StatusBadRequest {
		return models.Weather{}, fmt.Errorf("city not found: %s", city)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to read request body")
	}

	var weatherResp weatherAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unable to unmarshal request body")
	}

	return weatherResp.getWeatherModel(), nil
}

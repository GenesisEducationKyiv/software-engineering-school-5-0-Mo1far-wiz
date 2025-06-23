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

const weatherAPIName = "WeatherAPI"

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
	next    APIInterface
	logger  Logger
}

func (wa *WeatherAPI) setNext(next APIInterface) {
	wa.next = next
}

func NewWeatherAPI(config config.WeatherAPIConfig, logger Logger) *WeatherAPI {
	return &WeatherAPI{
		baseURL: config.ServiceBaseURL,
		apiKey:  config.APIKey,
		client:  http.DefaultClient,
		logger:  logger,
	}
}

func (wa *WeatherAPI) WithClient(client *http.Client) *WeatherAPI {
	wa.client = client
	return wa
}

func (wa *WeatherAPI) GetCityWeather(ctx context.Context, city string) (weather models.Weather, err error) {
	resp, err := wa.getCityWeatherRequest(ctx, city)
	if err != nil {
		if wa.next != nil {
			return wa.next.GetCityWeather(ctx, city)
		}
		return models.Weather{}, errors.Wrap(err, weatherAPIName)
	}

	return resp, nil
}

func (wa *WeatherAPI) getCityWeatherRequest(
	ctx context.Context,
	city string,
) (weather models.Weather, err error) {
	reqURL := wa.baseURL + "?key=" + wa.apiKey + "&q=" + city

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "new request")
	}

	resp, err := wa.client.Do(req)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "client do")
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			closeErr = errors.Wrap(closeErr, "close body")
			if err != nil {
				err = joinErr.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "read body")
	}

	if resp.StatusCode == http.StatusBadRequest {
		wa.logger.LogError(
			fmt.Sprintf("%s – Response: %s", weatherAPIName, string(body)),
		)
		return models.Weather{}, fmt.Errorf("city not found: %s", city)
	}
	if resp.StatusCode != http.StatusOK {
		wa.logger.LogError(
			fmt.Sprintf("%s – Response: %s", weatherAPIName, string(body)),
		)
		return models.Weather{}, fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	var weatherResp weatherAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unmarshal")
	}

	wa.logger.LogInfo(
		fmt.Sprintf("%s – Response: %v", weatherAPIName, weatherResp),
	)

	return weatherResp.getWeatherModel(), nil
}

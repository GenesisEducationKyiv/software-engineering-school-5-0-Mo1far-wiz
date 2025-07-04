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

const (
	visualCrossingName    = "VisualCrossing"
	freezingPointOfWaterF = 32.0
	slopeF                = 5.0 / 9.0
)

type visualCrossingAPIResponse struct {
	Current struct {
		Temp       float32 `json:"temp"`
		Humidity   float32 `json:"humidity"`
		Conditions string  `json:"conditions"`
	} `json:"currentConditions"`
}

func FtoC(fahrenheit float64) float64 {
	return (fahrenheit - freezingPointOfWaterF) * slopeF
}

func (wa visualCrossingAPIResponse) getWeatherModel() models.Weather {
	return models.Weather{
		Temperature: int(FtoC(float64(wa.Current.Temp))),
		Humidity:    int(wa.Current.Humidity),
		Description: wa.Current.Conditions,
	}
}

type VisualCrossingAPI struct {
	baseURL string
	apiKey  string
	client  *http.Client
	next    APIInterface
	logger  Logger
}

func (wa *VisualCrossingAPI) setNext(next APIInterface) {
	wa.next = next
}

func NewVisualCrossingAPI(config config.WeatherAPIConfig, logger Logger) *VisualCrossingAPI {
	return &VisualCrossingAPI{
		baseURL: config.ServiceBaseURL,
		apiKey:  config.APIKey,
		client:  http.DefaultClient,
		logger:  logger,
	}
}

func (wa *VisualCrossingAPI) WithClient(client *http.Client) *VisualCrossingAPI {
	wa.client = client
	return wa
}

func (wa *VisualCrossingAPI) WithLogging() *VisualCrossingAPI {
	wa.client.Transport = &WeatherLoggingRoundTripper{
		Base:    wa.client.Transport,
		Logger:  wa.logger,
		APIName: visualCrossingName,
	}
	return wa
}

func (wa *VisualCrossingAPI) GetCityWeather(
	ctx context.Context,
	city string,
) (weather models.Weather, err error) {
	resp, err := wa.getCityWeatherRequest(ctx, city)
	if err != nil {
		if wa.next != nil {
			return wa.next.GetCityWeather(ctx, city)
		}
		return models.Weather{}, errors.Wrap(err, visualCrossingName)
	}

	return resp, nil
}

func (wa *VisualCrossingAPI) getCityWeatherRequest(
	ctx context.Context,
	city string,
) (weather models.Weather, err error) {
	reqURL := wa.baseURL + city + "?key=" + wa.apiKey

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
		return models.Weather{}, fmt.Errorf("city not found: %s", city)
	}

	if resp.StatusCode != http.StatusOK {
		return models.Weather{}, fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	var weatherResp visualCrossingAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return models.Weather{}, errors.Wrap(err, "unmarshal")
	}

	return weatherResp.getWeatherModel(), nil
}

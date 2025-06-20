package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather/internal/api/handlers"
	"weather/internal/config"
	"weather/internal/models"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCityWeather_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	expected := models.Weather{
		Temperature: 13,
		Humidity:    25,
		Description: "test",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]interface{}{
			"current": map[string]interface{}{
				"temp_c":   expected.Temperature,
				"humidity": expected.Humidity,
				"condition": map[string]string{
					"text": expected.Description,
				},
			},
		}

		b, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		_, err = w.Write(b)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
	}))
	defer ts.Close()

	api := weather.NewWeatherAPI(config.WeatherAPIConfig{
		ServiceBaseURL: ts.URL,
		APIKey:         "unused",
	}).WithClient(ts.Client())

	svc := weather.NewRemoteService(api)
	h := handlers.NewWeatherHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/weather", nil)
	c.Request = req
	c.Set("city", "Kyiv")

	h.CityWeather(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"temperature":13`)
	assert.Contains(t, w.Body.String(), `"humidity":25`)
	assert.Contains(t, w.Body.String(), `"description":"test"`)
}

func TestCityWeather_BadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &weather.RemoteService{}
	h := handlers.NewWeatherHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/weather", nil)
	c.Request = req

	h.CityWeather(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
}

func TestCityWeather_NotFound(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ts := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(ts.Close)

	api := weather.NewWeatherAPI(config.WeatherAPIConfig{
		ServiceBaseURL: ts.URL,
		APIKey:         "unused",
	}).WithClient(ts.Client())
	svc := weather.NewRemoteService(api)
	h := handlers.NewWeatherHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/weather", nil)
	c.Request = req
	c.Set("city", "Gotham")

	h.CityWeather(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "City not found")
}

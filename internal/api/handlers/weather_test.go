package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"weather/internal/api/handlers"
	"weather/internal/cache"
	mock_cache "weather/internal/cache/mock"
	"weather/internal/config"
	"weather/internal/models"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCityWeather_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	logger := &noopLogger{}

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
		err := json.NewEncoder(w).Encode(resp)
		assert.Nil(t, err)
	}))
	defer ts.Close()

	api := weather.NewWeatherAPI(config.WeatherAPIConfig{
		ServiceBaseURL: ts.URL,
		APIKey:         "unused",
	}, logger).WithClient(ts.Client())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacher := mock_cache.NewMockCacher(ctrl)
	cacheSvc := cache.NewCacheService(mockCacher)

	mockCacher.
		EXPECT().
		Get(gomock.Any(), "Kyiv").
		Return("", cache.ErrCacheMiss)

	payload, err := json.Marshal(expected)
	assert.Nil(t, err)

	mockCacher.
		EXPECT().
		Set(gomock.Any(), "Kyiv", string(payload), time.Hour).
		Return(nil)

	svc, err := weather.NewWeatherService(cacheSvc, logger, api)
	if err != nil {
		t.Fatalf("failed to create WeatherService: %v", err)
	}
	h := handlers.NewWeatherHandler(svc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/weather", nil)
	c.Request = req
	c.Set("city", "Kyiv")

	h.CityWeather(c)

	assert.Equal(t, http.StatusOK, w.Code, "should return 200 OK")

	var got models.Weather
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	assert.Equal(t, expected, got, "response body must match expected weather")
}

func TestCityWeather_BadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	logger := &noopLogger{}

	svc := &weather.WeatherService{}
	h := handlers.NewWeatherHandler(svc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/weather", nil)
	c.Request = req

	h.CityWeather(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
}

func TestCityWeather_NotFound(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	logger := &noopLogger{}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	api := weather.NewWeatherAPI(config.WeatherAPIConfig{
		ServiceBaseURL: ts.URL,
		APIKey:         "unused",
	}, logger).WithClient(ts.Client())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacher := mock_cache.NewMockCacher(ctrl)
	cacheSvc := cache.NewCacheService(mockCacher)

	mockCacher.
		EXPECT().
		Get(gomock.Any(), "Gotham").
		Return("", cache.ErrCacheMiss)

	svc, err := weather.NewWeatherService(cacheSvc, logger, api)
	assert.NoError(t, err)
	h := handlers.NewWeatherHandler(svc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/weather", nil)
	c.Request = req
	c.Set("city", "Gotham")

	h.CityWeather(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "City not found")
}

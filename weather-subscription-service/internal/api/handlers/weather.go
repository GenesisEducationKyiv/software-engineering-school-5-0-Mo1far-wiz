package handlers

import (
	"context"
	"errors"
	"net/http"
	"weather-subscription/internal/models"
	"weather-subscription/internal/svc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WeatherService interface {
	GetCityWeather(ctx context.Context, city string) (models.Weather, error)
}

type WeatherHandler struct {
	weather WeatherService
	logger  Logger
}

func NewWeatherHandler(weather WeatherService, logger Logger) *WeatherHandler {
	return &WeatherHandler{
		weather: weather,
		logger:  logger,
	}
}

func (h *WeatherHandler) CityWeather(c *gin.Context) {
	city := c.GetString("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	weather, err := h.weather.GetCityWeather(c.Request.Context(), city)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrorGetCityWeather):
			c.JSON(http.StatusNotFound, "City not found")
		default:
			c.JSON(http.StatusInternalServerError, "Internal service error")
			h.logger.Error("Uncaught error, sending StatusInternalServerError",
				zap.String("error", err.Error()))
		}
	}

	c.JSON(http.StatusOK, weather)
}

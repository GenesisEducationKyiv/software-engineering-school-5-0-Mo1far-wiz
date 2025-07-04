package handlers

import (
	"net/http"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WeatherHandler struct {
	weatherService *weather.WeatherService
	logger         Logger
}

func NewWeatherHandler(weatherService *weather.WeatherService, logger Logger) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
		logger:         logger,
	}
}

func (h *WeatherHandler) CityWeather(c *gin.Context) {
	city := c.GetString("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	weather, err := h.weatherService.GetCityWeather(c.Request.Context(), city)
	if err != nil {
		h.logger.ConsoleLogError("on getting city weather",
			zap.String("error", err.Error()))
		c.JSON(http.StatusNotFound, "City not found")
		return
	}

	c.JSON(http.StatusOK, weather)
}

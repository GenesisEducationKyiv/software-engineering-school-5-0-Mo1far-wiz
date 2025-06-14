package handlers

import (
	"net/http"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	weatherService *weather.RemoteService
}

func NewWeatherHandler(weatherService *weather.RemoteService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

func (h *WeatherHandler) CityWeather(c *gin.Context) {
	city := c.GetString("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	weather, err := h.weatherService.GetCityWeather(city)
	if err != nil {
		logErrorF(err, "on getting city weather")
		c.JSON(http.StatusNotFound, "City not found")
		return
	}

	c.JSON(http.StatusOK, weather)
}

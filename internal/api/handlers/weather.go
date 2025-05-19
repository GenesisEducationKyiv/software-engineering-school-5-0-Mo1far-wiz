package handlers

import (
	"net/http"
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	store          store.Storage
	weatherService *weather.RemoteService
}

func NewWeatherHandler(store store.Storage, weatherService *weather.RemoteService) *WeatherHandler {
	return &WeatherHandler{
		store:          store,
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
		logError(err, "on getting city weather")
		c.JSON(http.StatusBadRequest, "City not found")
		return
	}

	c.JSON(http.StatusOK, weather)
}

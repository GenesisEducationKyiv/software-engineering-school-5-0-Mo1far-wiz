package api

import (
	"weather/internal/api/handlers"
	"weather/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func Mount(
	router *gin.Engine,
	weatherService handlers.WeatherService,
	subscriptionService handlers.SubscriptionService,
	logger handlers.Logger,
) {
	gin.SetMode(gin.ReleaseMode)

	weatherHandler := handlers.NewWeatherHandler(weatherService, logger)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, logger)

	api := router.Group("/api")

	weatherGroup := api.Group("/weather")
	weatherGroup.Use(middleware.ExtractQuery("city"))
	{
		weatherGroup.GET("/", weatherHandler.CityWeather)
	}

	subscriptionGroup := api.Group("/")
	subscriptionGroup.Use(middleware.ExtractParam("token"))
	{
		subscriptionGroup.POST("/subscribe", subscriptionHandler.Subscribe)
		subscriptionGroup.GET("/confirm/:token", subscriptionHandler.Confirm)
		subscriptionGroup.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	}
}

package api

import (
	"weather-subscription/internal/api/handlers"
	"weather-subscription/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Mount(
	router *gin.Engine,
	weatherService handlers.WeatherService,
	subscriptionService handlers.SubscriptionService,
	logger handlers.Logger,
) {
	gin.SetMode(gin.ReleaseMode)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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

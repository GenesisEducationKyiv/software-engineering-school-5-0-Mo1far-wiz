package api

import (
	"weather/internal/api/handlers"
	"weather/internal/api/middleware"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

func Mount(
	router *gin.Engine,
	storage handlers.SubscriptionStore,
	weatherService *weather.WeatherService,
	emailSender handlers.EmailSender,
	targetManager handlers.SubscriptionTargetManager,
	logger handlers.Logger,
) {
	gin.SetMode(gin.ReleaseMode)

	weatherHandler := handlers.NewWeatherHandler(weatherService, logger)
	subscriptionHandler := handlers.NewSubscriptionHandler(storage, emailSender, targetManager, logger)

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

package api

import (
	"weather/internal/api/handlers"
	"weather/internal/api/middleware"
	"weather/internal/mailer"
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

func Mount(router *gin.Engine, storage store.Storage,
	weatherService *weather.RemoteService, mailerService *mailer.SMTPMailer) {
	weatherHandler := handlers.NewWeatherHandler(storage, weatherService)
	subscriptionHandler := handlers.NewSubscriptionHandler(storage, mailerService)

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

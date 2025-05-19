package api

import (
	"weather/internal/api/handlers"
	"weather/internal/api/middleware"
	"weather/internal/mailer"
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

func Mount(router *gin.Engine, storage store.Storage, weatherService *weather.RemoteService, mailerService *mailer.SmtpMailer) {
	weatherHandler := handlers.NewWeatherHandler(storage, weatherService)
	subscriptionHandler := handlers.NewSubscriptionHandler(storage, mailerService)

	api := router.Group("/api")

	weather := api.Group("/weather")
	weather.Use(middleware.ExtractQuery("city"))
	weather.GET("/", weatherHandler.CityWeather)

	subscription := api.Group("/")
	subscription.Use(middleware.ExtractParam("token"))
	subscription.POST("/subscribe", subscriptionHandler.Subscribe)
	subscription.GET("/confirm/:token", subscriptionHandler.Confirm)
	subscription.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
}

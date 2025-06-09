package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"weather/internal/application"
	"weather/internal/config"
	"weather/internal/database"
	"weather/internal/env"
	"weather/internal/mailer"
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

func main() {
	dbName := env.GetString("DB_NAME", "weather")
	dbPassword := env.GetString("DB_PASSWORD", "")
	dbUser := env.GetString("DB_USER", "postgres")
	dbHost := env.GetString("DB_HOST", "localhost")
	dbPort := env.GetInt("DB_PORT", 5432)
	dbSSL := env.GetString("DB_SSL_MODE", "")

	dbAddr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
		dbSSL,
	)

	dbCfg := config.DBConfig{
		Addr:         dbAddr,
		MaxOpenConns: env.GetInt("MAX_OPEN_CONNS", 30),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	appPort := env.GetInt("APP_PORT", 8080)
	readTimeoutDuration := time.Duration(env.GetInt("READ_TIMEOUT", 5)) * time.Second
	writeTimeoutDuration := time.Duration(env.GetInt("WRITE_TIMEOUT", 5)) * time.Second
	idleTimeoutDuration := time.Duration(env.GetInt("IDLE_TIMEOUT", 5)) * time.Second
	cfg := config.Config{
		Addr:         fmt.Sprintf(":%d", appPort),
		ReadTimeout:  readTimeoutDuration,
		WriteTimeout: writeTimeoutDuration,
		IdleTimeout:  idleTimeoutDuration,
		DB:           dbCfg,
	}

	db, err := database.New(dbCfg)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	weatherServiceURL := env.GetString("WEATHER_SERVICE_URL", "http://api.weatherapi.com/v1/current.json")
	weatherAPIKey := env.GetString("WEATHER_API_KEY", "fake-api-key")
	weatherService := weather.NewRemoteService(&weather.WeatherAPI{
		BaseURL: weatherServiceURL,
		APIKey:  weatherAPIKey,
	})

	smtpUser := env.GetString("SMTP_USER", "email")
	smtpPassword := env.GetString("SMTP_PASS", "smash")
	smtpHost := env.GetString("SMTP_HOST", "host")
	smtpPort := env.GetString("SMTP_PORT", "port")

	store := store.NewStorage(db)
	subscriptions, err := store.Subscription.GetSubscribed(context.Background())
	if err != nil {
		log.Panic(err)
	}

	mailerSvc := mailer.New(smtpUser, smtpPassword, smtpHost, smtpPort, subscriptions, weatherService)

	gin.SetMode(gin.ReleaseMode)
	app := application.Application{
		Config:         cfg,
		Store:          store,
		Router:         gin.Default(),
		WeatherService: weatherService,
		MailerService:  mailerSvc,
	}

	app.Run()
}

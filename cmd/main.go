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

func getDatabaseConfig() config.DBConfig {
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

	return config.DBConfig{
		Addr:         dbAddr,
		MaxOpenConns: env.GetInt("MAX_OPEN_CONNS", 30),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}
}

func getApplicationConfig() config.ApplicationConfig {
	appPort := env.GetInt("APP_PORT", 8080)
	readTimeoutDuration := time.Duration(env.GetInt("READ_TIMEOUT", 5)) * time.Second
	writeTimeoutDuration := time.Duration(env.GetInt("WRITE_TIMEOUT", 5)) * time.Second
	idleTimeoutDuration := time.Duration(env.GetInt("IDLE_TIMEOUT", 5)) * time.Second
	return config.ApplicationConfig{
		Addr:         fmt.Sprintf(":%d", appPort),
		ReadTimeout:  readTimeoutDuration,
		WriteTimeout: writeTimeoutDuration,
		IdleTimeout:  idleTimeoutDuration,
	}
}

func getWeatherAPIConfig() config.WeatherAPIConfig {
	weatherServiceURL := env.GetString("WEATHER_SERVICE_URL", "http://api.weatherapi.com/v1/current.json")
	weatherAPIKey := env.GetString("WEATHER_API_KEY", "fake-api-key")

	return config.WeatherAPIConfig{
		ServiceBaseURL: weatherServiceURL,
		APIKey:         weatherAPIKey,
	}
}

func getSMTPConfig() config.SMTPConfig {
	smtpUser := env.GetString("SMTP_USER", "email")
	smtpPassword := env.GetString("SMTP_PASS", "smash")
	smtpHost := env.GetString("SMTP_HOST", "host")
	smtpPort := env.GetString("SMTP_PORT", "port")

	return config.SMTPConfig{
		SMTPUser:     smtpUser,
		SMTPPassword: smtpPassword,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
	}
}

func main() {
	dbConfig := getDatabaseConfig()
	appConfig := getApplicationConfig()

	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = database.ValidateConnection(db)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStorage(db)

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	weatherServiceConfig := getWeatherAPIConfig()
	weatherService := weather.NewRemoteService(weather.NewWeatherAPI(weatherServiceConfig))

	smtpConfig := getSMTPConfig()
	mailerService := mailer.New(smtpConfig, weatherService)

	ctx, cancel := context.WithTimeout(context.Background(), mailer.LoadTimeoutDuration)
	err = mailerService.LoadTargets(ctx, store.Mailer)
	cancel()
	if err != nil {
		log.Panic(err)
	}

	app := application.Application{
		Config:         appConfig,
		Store:          store,
		Router:         gin.Default(),
		WeatherService: weatherService,
		MailerService:  mailerService,
	}

	app.Run()
}

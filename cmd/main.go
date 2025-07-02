package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"weather/internal/application"
	"weather/internal/cache"
	"weather/internal/config"
	"weather/internal/database"
	"weather/internal/env"
	"weather/internal/mailer"
	"weather/internal/redis"
	"weather/internal/store"
	"weather/internal/weather"
	"weather/pkg/logger"

	"github.com/gin-gonic/gin"
)

const logFile = "weather.log"

func getDatabaseConfig() config.DBConfig {
	dbName := env.GetString("DB_NAME", "weather")
	dbPassword := env.GetString("DB_PASSWORD", "")
	dbUser := env.GetString("DB_USER", "postgres")
	dbHost := env.GetString("DB_HOST", "localhost")
	dbPort := env.GetInt("DB_PORT", 5432)
	dbSSL := env.GetString("DB_SSL_MODE", "")

	dbAddr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
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

func getVisualCrossingAPIConfig() config.WeatherAPIConfig {
	weatherServiceURL := env.GetString("VISUALCROSSING_SERVICE_URL",
		"https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/")
	weatherAPIKey := env.GetString("VISUALCROSSING_API_KEY", "fake-api-key")

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

func getRedisConfig() config.RedisConfig {
	return config.RedisConfig{
		Addr:     "",
		Password: "",
		DB:       0,
	}
}

func main() {
	logger, err := logger.NewLogger(logFile)
	if err != nil {
		log.Fatal(err)
	}

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

	migratePath := env.GetString("MIGRATION_PATH", "not-found")
	err = database.MigrateUp(dbConfig.Addr, migratePath)
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

	weatherAPIServiceConfig := getWeatherAPIConfig()
	visualCrossingServiceConfig := getVisualCrossingAPIConfig()

	defaultRT := http.DefaultTransport
	t, ok := defaultRT.(*http.Transport)
	if !ok {
		log.Panicf(
			"expected http.DefaultTransport to be *http.Transport, got %T",
			defaultRT,
		)
	}
	baseT := t.Clone()

	weatherClient := &http.Client{
		Transport: &weather.WeatherLoggingRoundTripper{
			Base:    baseT,
			Logger:  logger,
			APIName: "WeatherAPI",
			URL:     weatherAPIServiceConfig.ServiceBaseURL,
		},
	}

	vcClient := &http.Client{
		Transport: &weather.WeatherLoggingRoundTripper{
			Base:    baseT,
			Logger:  logger,
			APIName: "VisualCrossing",
			URL:     visualCrossingServiceConfig.ServiceBaseURL,
		},
	}

	// primary
	weatherAPI := weather.NewWeatherAPI(weatherAPIServiceConfig, logger).WithClient(weatherClient)

	// secondary
	visualCrossing := weather.NewVisualCrossingAPI(visualCrossingServiceConfig, logger).WithClient(vcClient)

	redisConfig := getRedisConfig()
	redis := redis.NewClient(redisConfig)

	cacheService := cache.NewCacheService(redis)

	weatherService, err := weather.NewWeatherService(*cacheService, logger, weatherAPI, visualCrossing)
	if err != nil {
		log.Panic(err)
	}

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
		Logger:         logger,
	}

	app.Run()
}

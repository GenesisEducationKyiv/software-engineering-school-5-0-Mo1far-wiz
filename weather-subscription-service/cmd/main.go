package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pkg/env"
	"pkg/logger"
	"time"
	"weather-subscription/internal/application"
	"weather-subscription/internal/cache"
	"weather-subscription/internal/config"
	"weather-subscription/internal/database"
	"weather-subscription/internal/mailer"
	"weather-subscription/internal/metrics"
	"weather-subscription/internal/store"
	"weather-subscription/internal/weather"
	"weather-subscription/pkg/redis"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
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
	weatherAPIServiceURL := env.GetString("WEATHER_SERVICE_URL", "http://api.weatherapi.com/v1/current.json")
	weatherAPIKey := env.GetString("WEATHER_API_KEY", "fake-api-key")

	return config.WeatherAPIConfig{
		ServiceBaseURL: weatherAPIServiceURL,
		APIKey:         weatherAPIKey,
	}
}

func getVisualCrossingAPIConfig() config.WeatherAPIConfig {
	weatherAPIServiceURL := env.GetString("VISUALCROSSING_SERVICE_URL",
		"https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/")
	weatherAPIKey := env.GetString("VISUALCROSSING_API_KEY", "fake-api-key")

	return config.WeatherAPIConfig{
		ServiceBaseURL: weatherAPIServiceURL,
		APIKey:         weatherAPIKey,
	}
}

func getPublisherConfig() config.PublishConfig {
	addr := env.GetString("RABBIT_ADDR", "addr")
	routingKey := env.GetString("ROUTING_KEY", "key")
	exchangeName := env.GetString("EXCHANGE_NAME", "labubu")

	return config.PublishConfig{
		Addr:         addr,
		RoutingKey:   routingKey,
		ExchangeName: exchangeName,
	}
}

func getRedisConfig() config.RedisConfig {
	redisAddr := env.GetString("REDIS_ADDR", "127.0.0.1:6378")
	redisPassword := env.GetString("REDIS_PASSWORD", "password")
	redisDB := env.GetInt("REDIS_DB", 0)

	return config.RedisConfig{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}
}

func main() {
	metrics.Init()

	lvl := zapcore.InfoLevel
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		lvl = zapcore.DebugLevel
	}

	logger, err := logger.NewLogger(logFile, lvl)
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

	weatherAPIService, err := weather.NewWeatherAPIService(cacheService, logger, weatherAPI, visualCrossing)
	if err != nil {
		log.Panic(err)
	}

	publisherConfig := getPublisherConfig()
	mailerService := mailer.New(publisherConfig, weatherAPIService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), mailer.LoadTimeoutDuration)
	err = mailerService.LoadTargets(ctx, store.Mailer)
	cancel()
	if err != nil {
		log.Panic(err)
	}

	router := gin.New()
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	app := application.Application{
		Config:            appConfig,
		Store:             store,
		Router:            gin.Default(),
		WeatherAPIService: weatherAPIService,
		MailerService:     mailerService,
		Logger:            logger,
	}

	app.Run()
}

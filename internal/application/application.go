package application

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather/internal/api"
	"weather/internal/config"
	"weather/internal/mailer"
	"weather/internal/service"
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const shutdownTimeout = 5 * time.Second

type Logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	Sync() error
}

type Application struct {
	Config            config.ApplicationConfig
	Store             store.Storage
	Router            *gin.Engine
	server            *http.Server
	WeatherAPIService *weather.WeatherAPIService
	MailerService     *mailer.Manager
	Logger            Logger
}

func (a *Application) Initialize() {
	a.server = &http.Server{
		Addr:         a.Config.Addr,
		Handler:      a.Router,
		ReadTimeout:  a.Config.ReadTimeout,
		WriteTimeout: a.Config.WriteTimeout,
		IdleTimeout:  a.Config.IdleTimeout,
	}

	weatherService := service.NewWeather(a.WeatherAPIService, a.Logger)
	subscriptionService := service.NewSubscription(
		a.Store.Subscription,
		a.MailerService.Mailer,
		a.MailerService.Targets,
		a.Logger,
	)

	api.Mount(
		a.Router,
		weatherService,
		subscriptionService,
		a.Logger,
	)
}

// Run starts the HTTP server and handles graceful shutdown upon receiving termination signals.
func (a *Application) Run() {
	a.Initialize()

	a.Logger.ConsoleLogInfo("application initialized")

	a.MailerService.Start()

	go func() {
		log.Printf("Starting server on %s", a.Config.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	a.MailerService.Stop()

	err := a.Logger.Sync()
	if err != nil {
		log.Panicf("Logger sync error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Panicf("Server shutdown error: %v", err)
	}

	log.Println("Server exited properly")
}

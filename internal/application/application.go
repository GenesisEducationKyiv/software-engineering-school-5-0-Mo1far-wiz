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
	"weather/internal/store"
	"weather/internal/weather"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Config         config.Config
	Store          store.Storage
	Router         *gin.Engine
	server         *http.Server
	WeatherService *weather.RemoteService
	MailerService  *mailer.SmtpMailer
}

func (a *Application) Initialize() {
	a.server = &http.Server{
		Addr:         a.Config.Addr,
		Handler:      a.Router,
		ReadTimeout:  a.Config.ReadTimeout,
		WriteTimeout: a.Config.WriteTimeout,
		IdleTimeout:  a.Config.IdleTimeout,
	}

	api.Mount(a.Router, a.Store, a.WeatherService, a.MailerService)
}

// very graceful very mindful
func (a *Application) Run() {
	a.Initialize()

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server exited properly")
}

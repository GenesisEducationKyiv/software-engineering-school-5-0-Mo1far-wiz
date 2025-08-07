package application

import (
	"context"
	"mailer/internal/config"
	"mailer/internal/consumer"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Sync() error
}

type Application struct {
	Config   config.ApplicationConfig
	Logger   Logger
	Consumer *consumer.RabbitConsumer
}

func (a *Application) Run() {
	a.Logger.Info("application initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.Consumer.Start(ctx); err != nil {
		a.Logger.Error("Failed to start consumer", zap.Error(err))
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.Logger.Debug("Shutdown signal received, stopping consumer...")

	a.Consumer.Stop()

	if err := a.Logger.Sync(); err != nil {
		a.Logger.Error("Logger sync error", zap.Error(err))
	}

	a.Logger.Info("Application exited properly")
}

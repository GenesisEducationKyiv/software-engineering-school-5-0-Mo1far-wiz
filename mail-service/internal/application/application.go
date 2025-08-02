package application

import (
	"context"
	"mailer/internal/config"
	"mailer/internal/consumer"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const shutdownTimeout = 5 * time.Second

type Logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	Sync() error
}

type Application struct {
	Config   config.ApplicationConfig
	Logger   Logger
	Consumer *consumer.RabbitConsumer
}

func (a *Application) Run() {
	a.Logger.ConsoleLogInfo("application initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.Consumer.Start(ctx); err != nil {
		a.Logger.ConsoleLogError("Failed to start consumer", zap.Error(err))
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.Logger.ConsoleLogInfo("Shutdown signal received, stopping consumer...")

	a.Consumer.Stop()

	if err := a.Logger.Sync(); err != nil {
		a.Logger.ConsoleLogError("Logger sync error", zap.Error(err))
	}

	a.Logger.ConsoleLogInfo("Application exited properly")
}

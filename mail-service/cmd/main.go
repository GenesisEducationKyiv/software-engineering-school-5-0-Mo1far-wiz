package main

import (
	"fmt"
	"log"
	"mailer/internal/application"
	"mailer/internal/config"
	"mailer/internal/consumer"
	"mailer/internal/smtp"
	"os"
	"pkg/env"
	"pkg/logger"

	"go.uber.org/zap/zapcore"
)

const logFile = "mailer.log"

func getApplicationConfig() config.ApplicationConfig {
	appPort := env.GetInt("MAILER_PORT", 8081)
	return config.ApplicationConfig{
		Addr: fmt.Sprintf(":%d", appPort),
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

func getRabbitConfig() config.RabbitConfig {
	addr := env.GetString("RABBIT_ADDR", "addr")
	queueName := env.GetString("QUEUE_NAME", "queue")
	routingKey := env.GetString("ROUTING_KEY", "key")
	exchangeName := env.GetString("EXCHANGE_NAME", "labubu")

	return config.RabbitConfig{
		Addr:         addr,
		QueueName:    queueName,
		RoutingKey:   routingKey,
		ExchangeName: exchangeName,
	}
}

func main() {
	lvl := zapcore.InfoLevel
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		lvl = zapcore.DebugLevel
	}

	logger, err := logger.NewLogger(logFile, lvl)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Logger initialized.")

	smtpConfig := getSMTPConfig()
	smtpMailer := smtp.NewSMTPMailer(
		smtpConfig, logger,
	)

	rabbitConfig := getRabbitConfig()
	rabbitConsumer := consumer.New(rabbitConfig, smtpMailer, logger)
	if rabbitConsumer == nil {
		log.Fatal("Failed to create RabbitMQ consumer")
	}
	defer rabbitConsumer.Stop()

	applicationConfig := getApplicationConfig()
	app := &application.Application{
		Config:   applicationConfig,
		Logger:   logger,
		Consumer: rabbitConsumer,
	}
	app.Run()
}

package main

import (
	"fmt"
	"log"
	"mailer/internal/application"
	"mailer/internal/config"
	"mailer/internal/service"
	"mailer/internal/smtp"
	"pkg/env"
	"pkg/logger"
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

func main() {
	logger, err := logger.NewLogger(logFile)
	if err != nil {
		log.Fatal(err)
	}
	logger.ConsoleLogInfo("Logger initialized.")

	smtpConfig := getSMTPConfig()
	smtpMailer := smtp.NewSMTPMailer(
		smtpConfig, logger,
	)

	mailer := service.NewMailer(smtpMailer, logger)
	applicationConfig := getApplicationConfig()
	app := &application.Application{
		Config: applicationConfig,
		Logger: logger,
		Mailer: mailer,
	}
	app.Run()
}

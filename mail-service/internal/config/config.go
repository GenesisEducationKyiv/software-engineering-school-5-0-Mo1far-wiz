package config

import "time"

type ApplicationConfig struct {
	Addr         string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration
}

type SMTPConfig struct {
	SMTPUser     string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     string
}

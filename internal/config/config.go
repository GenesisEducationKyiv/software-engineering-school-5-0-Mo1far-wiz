package config

import (
	"time"
)

type ApplicationConfig struct {
	Addr         string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type WeatherAPIConfig struct {
	ServiceBaseURL string
	APIKey         string
}

type SMTPConfig struct {
	SMTPUser     string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     string
}

package config

type ApplicationConfig struct {
	Addr string
}

type SMTPConfig struct {
	SMTPUser     string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     string
}

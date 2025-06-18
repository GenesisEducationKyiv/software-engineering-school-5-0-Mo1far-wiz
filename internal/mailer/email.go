package mailer

import (
	"context"
	"fmt"
	"time"
	"weather/internal/models"
)

type EmailBuilder struct {
	subject string
	body    string
}

func NewEmailBuilder() *EmailBuilder {
	return &EmailBuilder{
		subject: "Hello %s,\n\nCurrent weather in %s:\n" +
			"- %s\n- Temperature: %d°C\n- Humidity: %d%%\n",
		body: " for %s – %s",
	}
}

func (e *EmailBuilder) BuildWeatherForecastEmail(ctx context.Context,
	email string,
	city string,
	weatherData models.Weather,
) (string, string) {
	subject := fmt.Sprintf(e.subject, city, time.Now().Format("2006-01-02"))
	body := fmt.Sprintf(e.body,
		email, city,
		weatherData.Description,
		weatherData.Temperature,
		weatherData.Humidity,
	)

	return subject, body
}

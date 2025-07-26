package mailer

import (
	"context"
	"pkg/protos/mailer"
	"sync"
	"time"
	"weather-subscription/internal/config"
	"weather-subscription/internal/models"
	"weather-subscription/internal/weather"
	"weather-subscription/pkg/mail"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	Day                    = 24 * time.Hour
	SendEmailDailyTimeout  = time.Minute * 15
	SendEmailHourlyTimeout = time.Minute * 15
	SendEmailTimeout       = time.Minute * 16
	LoadTimeoutDuration    = time.Second * 5
	MaxRetries             = 3
	RetryDelay             = time.Second * 2
)

type MailerStore interface {
	GetSubscribed(ctx context.Context) ([]models.Subscription, error)
}

type Manager struct {
	MailService  *mail.MailService
	Targets      *TargetManager
	Forecasts    *Forecaster
	Logger       logger
	emailBuilder *EmailBuilder

	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

type logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	LogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}

func New(mailServiceConfig config.MailServiceConfig, weatherAPIService *weather.WeatherAPIService, logger logger) *Manager {
	forecaster := NewForecaster(weatherAPIService)
	mailService := mail.NewMailService(mailServiceConfig, logger)
	emailBuilder := NewEmailBuilder()

	return &Manager{
		MailService:  mailService,
		Targets:      &TargetManager{},
		Forecasts:    forecaster,
		Logger:       logger,
		emailBuilder: emailBuilder,
		stopChan:     make(chan struct{}),
	}
}

func (m *Manager) LoadTargets(ctx context.Context, store MailerStore) error {
	return m.Targets.LoadTargets(ctx, store)
}

func (m *Manager) AddTarget(sub models.Subscription) {
	m.Targets.AddTarget(sub)
}

func (m *Manager) RemoveTarget(email string, frequency string) {
	m.Targets.RemoveTarget(email, frequency)
}

func (m *Manager) Shutdown() error {
	return m.MailService.Disconnect()
}

func (m *Manager) Start() error {
	m.running = true
	m.stopChan = make(chan struct{})

	err := m.MailService.Connect(context.Background())
	if err != nil {
		return errors.Wrap(err, "connection to mailer-service")
	}

	// Daily scheduler
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runScheduler(models.Daily, Day, SendEmailDailyTimeout, m.getNextMidnight)
	}()

	// Hourly scheduler
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runScheduler(models.Hourly, time.Hour, SendEmailHourlyTimeout, m.getNextHour)
	}()

	return nil
}

func (m *Manager) runScheduler(
	frequency string,
	interval time.Duration,
	timeout time.Duration,
	getNextTime func() time.Time,
) {
	nextTime := getNextTime()
	select {
	case <-time.After(time.Until(nextTime)):
	case <-m.stopChan:
		return
	}

	m.sendEmailBatch(frequency, timeout)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.sendEmailBatch(frequency, timeout)
		case <-m.stopChan:
			return
		}
	}
}

func (m *Manager) sendEmailBatch(frequency string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	targets := m.Targets.GetTargets(frequency)
	if len(targets) == 0 {
		m.Logger.LogInfo("No targets found", zap.String("frequency", frequency))
		return
	}

	forecasts := m.Forecasts.GetForecasts(ctx, targets)
	if len(forecasts) == 0 {
		m.Logger.LogInfo("No forecasts available", zap.String("frequency", frequency))
		return
	}

	emails := m.buildEmails(ctx, forecasts, m.getSubjectPrefix(frequency))

	result, err := m.MailService.SendEmailBatch(ctx, emails)
	if err != nil {
		m.Logger.LogError("Failed to send email batch",
			zap.String("frequency", frequency),
			zap.Error(err),
		)
		return
	}

	m.Logger.LogInfo("Email batch sent",
		zap.String("frequency", frequency),
		zap.Uint64("success", *result.Success),
		zap.Uint64("failed", *result.Failed),
		zap.Int("total", len(emails)),
	)
}

func (m *Manager) buildEmails(
	ctx context.Context,
	forecasts []models.Forecast,
	subjectPrefix string,
) []*mailer.Email {
	emails := make([]*mailer.Email, 0, len(forecasts))

	for _, f := range forecasts {
		subj, body := m.emailBuilder.BuildWeatherForecastEmail(ctx, f.Email, f.City, f.Weather)
		subj = subjectPrefix + subj

		email := &mailer.Email{
			ToEmail: &f.Email,
			Subject: &subj,
			Body:    &body,
		}

		emails = append(emails, email)
	}

	return emails
}

func (m *Manager) getSubjectPrefix(frequency string) string {
	switch frequency {
	case models.Daily:
		return "Daily Weather - "
	case models.Hourly:
		return "Hourly Weather - "
	default:
		return "Weather Update - "
	}
}

func (m *Manager) getNextMidnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func (m *Manager) getNextHour() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
}

func (m *Manager) Stop() {
	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)
	m.wg.Wait()

	if err := m.Shutdown(); err != nil {
		m.Logger.LogError("Failed to shutdown mail service", zap.Error(err))
	}
}

func (m *Manager) SendEmail(to, subject, body string) error {
	email := &mailer.Email{
		ToEmail: &to,
		Subject: &subject,
		Body:    &body,
	}

	result, err := m.MailService.SendEmail(context.Background(), email)

	if err != nil || *result.Failed == 1 {
		m.Logger.LogError("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return errors.Wrap(err, "failed to send email")
	}
	return nil
}

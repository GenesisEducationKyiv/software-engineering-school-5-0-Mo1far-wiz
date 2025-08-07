package mailer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"weather-subscription/internal/config"
	"weather-subscription/internal/metrics"
	"weather-subscription/internal/models"
	"weather-subscription/internal/publisher"
	"weather-subscription/internal/weather"

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
	Publisher    *publisher.EmailPublisher
	Targets      *TargetManager
	Forecasts    *Forecaster
	Logger       logger
	emailBuilder *EmailBuilder

	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

type logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

func New(
	publisherConfig config.PublishConfig,
	weatherAPIService *weather.WeatherAPIService,
	logger logger,
) *Manager {
	forecaster := NewForecaster(weatherAPIService)
	publisher, err := publisher.New(publisherConfig, logger)
	if err != nil {
		log.Panic(err)
	}

	emailBuilder := NewEmailBuilder()

	return &Manager{
		Publisher:    publisher,
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

func (m *Manager) Shutdown() {
	m.Publisher.Close()
}

func (m *Manager) Start() error {
	m.running = true
	m.stopChan = make(chan struct{})

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
	err := metrics.ObserveRequest(
		fmt.Sprintf("email_batch_%s", frequency),
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			targets := m.Targets.GetTargets(frequency)
			if len(targets) == 0 {
				m.Logger.Warn("No targets found", zap.String("frequency", frequency))
				return nil
			}

			metrics.CacheSizeGauge.WithLabelValues("subscribers", frequency).
				Set(float64(len(targets)))

			forecasts := m.Forecasts.GetForecasts(ctx, targets)
			if len(forecasts) == 0 {
				m.Logger.Warn("No forecasts available", zap.String("frequency", frequency))
				return nil
			}

			metrics.CacheSizeGauge.WithLabelValues("forecasts", frequency).
				Set(float64(len(forecasts)))

			emails := m.buildEmails(ctx, forecasts, m.getSubjectPrefix(frequency))

			if err := m.Publisher.BatchSendEmails(ctx, emails); err != nil {
				return err
			}
			m.Logger.Info("Email batch sent",
				zap.String("frequency", frequency),
				zap.Int("total", len(emails)),
			)
			return nil
		},
	)
	if err != nil {
		m.Logger.Error("Batch job failed", zap.String("frequency", frequency), zap.Error(err))
	}
}

func (m *Manager) buildEmails(
	ctx context.Context,
	forecasts []models.Forecast,
	subjectPrefix string,
) []models.Email {
	emails := make([]models.Email, 0, len(forecasts))

	for _, f := range forecasts {
		subj, body := m.emailBuilder.BuildWeatherForecastEmail(ctx, f.Email, f.City, f.Weather)
		subj = subjectPrefix + subj

		email := models.Email{
			ToEmail: f.Email,
			Subject: subj,
			Body:    body,
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

	m.Shutdown()
}

func (m *Manager) SendEmail(to, subject, body string) error {
	return metrics.ObserveRequest("email_send_manual", func() error {
		email := models.Email{ToEmail: to, Subject: subject, Body: body}
		if err := m.Publisher.SendEmail(context.Background(), email); err != nil {
			m.Logger.Error("Failed to send email",
				zap.String("to", to),
				zap.String("subject", subject),
				zap.Error(err),
			)
			return errors.Wrap(err, "failed to send email")
		}
		return nil
	})
}

package mailer

import (
	"context"
	"sync"
	"time"

	"weather/internal/config"
	"weather/internal/models"
	"weather/internal/weather"
)

const (
	Day                    = 24 * time.Hour
	SendEmailDailyTimeout  = time.Minute * 15
	SendEmailHourlyTimeout = time.Minute * 15
	LoadTimeoutDuration    = time.Second * 5
)

type MailerStore interface {
	GetSubscribed(ctx context.Context) ([]models.Subscription, error)
}

type Manager struct {
	Mailer    *SMTPMailer
	Targets   *TargetManager
	Forecasts *Forecaster

	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

func New(config config.SMTPConfig, weatherService *weather.RemoteService) *Manager {
	forecaster := NewForecaster(weatherService)
	mailer := NewSMTPMailer(config, NewEmailBuilder())

	return &Manager{
		Mailer:    mailer,
		Targets:   &TargetManager{},
		Forecasts: forecaster,
		stopChan:  make(chan struct{}),
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

func (m *Manager) Start() {
	m.running = true
	m.stopChan = make(chan struct{})

	// Daily
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		select {
		case <-time.After(nextMidnight.Sub(now)):
		case <-m.stopChan:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), SendEmailDailyTimeout)
		targets := m.Targets.GetTargets(models.Daily)
		forecasts := m.Forecasts.GetForecasts(ctx, targets)
		m.Mailer.sendEmails(ctx, forecasts, "Daily Weather")
		cancel()
		ticker := time.NewTicker(Day)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), SendEmailDailyTimeout)
				targets := m.Targets.GetTargets(models.Daily)
				forecasts := m.Forecasts.GetForecasts(ctx, targets)
				m.Mailer.sendEmails(ctx, forecasts, "Daily Weather")
				cancel()
			case <-m.stopChan:
				return
			}
		}
	}()

	// Hourly
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		now := time.Now()
		nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		select {
		case <-time.After(nextHour.Sub(now)):
		case <-m.stopChan:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), SendEmailHourlyTimeout)
		targets := m.Targets.GetTargets(models.Hourly)
		forecasts := m.Forecasts.GetForecasts(ctx, targets)
		m.Mailer.sendEmails(ctx, forecasts, "Hourly Weather")
		cancel()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), SendEmailHourlyTimeout)
				targets := m.Targets.GetTargets(models.Hourly)
				forecasts := m.Forecasts.GetForecasts(ctx, targets)
				m.Mailer.sendEmails(ctx, forecasts, "Hourly Weather")
				cancel()
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	m.running = false
	close(m.stopChan)
	m.wg.Wait()
}

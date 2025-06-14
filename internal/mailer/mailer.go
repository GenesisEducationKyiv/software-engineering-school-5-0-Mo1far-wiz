package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"weather/internal/config"
	"weather/internal/models"
	"weather/internal/weather"

	joinErr "errors"

	"github.com/pkg/errors"
)

const (
	Day = 24 * time.Hour
)

type SubscribedStore interface {
	GetSubscribed(ctx context.Context) ([]models.Subscription, error)
}

type SMTPMailer struct {
	User           string
	Password       string
	Host           string
	Port           string
	WeatherService *weather.RemoteService

	mx      sync.RWMutex
	targets map[string][]models.Subscription

	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

func New(config config.SMTPConfig,
	store SubscribedStore, weatherService *weather.RemoteService) *SMTPMailer {
	subscriptions, err := store.GetSubscribed(context.Background())
	if err != nil {
		log.Panic(err)
	}

	targets := make(map[string][]models.Subscription)

	for _, sub := range subscriptions {
		targets[sub.Frequency] = append(targets[sub.Frequency], sub)
	}

	return &SMTPMailer{
		User:           config.SMTPUser,
		Password:       config.SMTPPassword,
		Host:           config.SMTPHost,
		Port:           config.SMTPPort,
		WeatherService: weatherService,
		targets:        targets,
		stopChan:       make(chan struct{}),
	}
}

func (m *SMTPMailer) AddDailyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets[models.Daily] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets[models.Daily] = append(m.targets[models.Daily], sub)
}

func (m *SMTPMailer) AddHourlyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets[models.Hourly] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets[models.Hourly] = append(m.targets[models.Hourly], sub)
}

func (m *SMTPMailer) RemoveDailyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets[models.Daily]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets[models.Daily] = subs[:len(subs)-1]
			return
		}
	}
}

func (m *SMTPMailer) RemoveHourlyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets[models.Hourly]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets[models.Hourly] = subs[:len(subs)-1]
			return
		}
	}
}

func (m *SMTPMailer) Start() {
	m.mx.Lock()
	if m.running {
		m.mx.Unlock()
		return
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mx.Unlock()

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
		m.sendDailyEmails()
		ticker := time.NewTicker(Day)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.sendDailyEmails()
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
		m.sendHourlyEmails()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.sendHourlyEmails()
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *SMTPMailer) Stop() {
	m.mx.Lock()
	if !m.running {
		m.mx.Unlock()
		return
	}
	m.running = false
	close(m.stopChan)
	m.mx.Unlock()
	m.wg.Wait()
}

func (m *SMTPMailer) sendDailyEmails() {
	m.mx.RLock()
	subs := append([]models.Subscription(nil), m.targets[models.Daily]...)
	m.mx.RUnlock()

	for _, sub := range subs {
		weatherData, err := m.WeatherService.GetCityWeather(sub.City)
		if err != nil {
			log.Printf("weather fetch error for %q: %v\n", sub.City, err)
			continue
		}
		subject := fmt.Sprintf("Daily Weather for %s – %s", sub.City, time.Now().Format("2006-01-02"))
		body := fmt.Sprintf(
			"Hello %s,\n\nCurrent weather in %s:\n"+
				"- %s\n- Temperature: %d°C\n- Humidity: %d%%\n",
			sub.Email, sub.City,
			weatherData.Description,
			weatherData.Temperature,
			weatherData.Humidity,
		)

		go func(email, subj, msg string) {
			if err := m.SendEmail(email, subj, msg); err != nil {
				log.Printf("daily email error to %s: %v\n", email, err)
			}
		}(sub.Email, subject, body)
	}
}

func (m *SMTPMailer) sendHourlyEmails() {
	m.mx.RLock()
	subs := append([]models.Subscription(nil), m.targets[models.Hourly]...)
	m.mx.RUnlock()

	for _, sub := range subs {
		weatherData, err := m.WeatherService.GetCityWeather(sub.City)
		if err != nil {
			log.Printf("weather fetch error for %q: %v\n", sub.City, err)
			continue
		}
		subject := fmt.Sprintf("Hourly Weather for %s – %s", sub.City, time.Now().Format("2006-01-02 15:04"))
		body := fmt.Sprintf(
			"Hello %s,\n\nCurrent weather in %s:\n"+
				"- %s\n- Temperature: %d°C\n- Humidity: %d%%\n",
			sub.Email, sub.City,
			weatherData.Description,
			weatherData.Temperature,
			weatherData.Humidity,
		)

		go func(email, subj, msg string) {
			if err := m.SendEmail(email, subj, msg); err != nil {
				log.Printf("hourly email error to %s: %v\n", email, err)
			}
		}(sub.Email, subject, body)
	}
}

func (m *SMTPMailer) SendEmail(to, subject, body string) (err error) {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", m.User))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("\r\n")
	msg.WriteString(body)

	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)
	tlsConf := &tls.Config{InsecureSkipVerify: false, ServerName: m.Host, MinVersion: tls.VersionTLS12}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", m.Host, m.Port), tlsConf)
	if err != nil {
		return errors.Wrap(err, "connect SMTP")
	}

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		err = errors.Wrap(err, "new SMTP client")
		if closeErr := conn.Close(); closeErr != nil {
			err = joinErr.Join(err, closeErr)
		}
		return err
	}

	defer func() {
		quitErr := client.Quit()
		if quitErr != nil {
			quitErr = errors.Wrap(quitErr, "failed to quit client")
			if err != nil {
				err = joinErr.Join(err, quitErr)
			} else {
				err = quitErr
			}
		}
	}()

	if err := client.Auth(auth); err != nil {
		return errors.Wrap(err, "SMTP auth")
	}
	if err := client.Mail(m.User); err != nil {
		return errors.Wrap(err, "set sender")
	}
	if err := client.Rcpt(to); err != nil {
		return errors.Wrap(err, "set recipient")
	}

	wc, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "get data writer")
	}

	defer func() {
		closeErr := wc.Close()
		if closeErr != nil {
			closeErr = errors.Wrap(closeErr, "failed to close write closer")
			if err != nil {
				err = joinErr.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	if _, err := wc.Write([]byte(msg.String())); err != nil {
		return errors.Wrap(err, "write email body")
	}
	return nil
}

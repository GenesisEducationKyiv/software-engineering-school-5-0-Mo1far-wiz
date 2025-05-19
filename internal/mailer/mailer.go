package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"weather/internal/models"
	"weather/internal/weather"
)

type SmtpMailer struct {
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

func New(user, password, host, port string, weatherService *weather.RemoteService) *SmtpMailer {
	return &SmtpMailer{
		User:           user,
		Password:       password,
		Host:           host,
		Port:           port,
		WeatherService: weatherService,
		targets:        make(map[string][]models.Subscription),
		stopChan:       make(chan struct{}),
	}
}

func (m *SmtpMailer) AddDailyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets["daily"] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets["daily"] = append(m.targets["daily"], sub)
}

func (m *SmtpMailer) AddHourlyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets["hourly"] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets["hourly"] = append(m.targets["hourly"], sub)
}

func (m *SmtpMailer) RemoveDailyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets["daily"]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets["daily"] = subs[:len(subs)-1]
			return
		}
	}
}

func (m *SmtpMailer) RemoveHourlyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets["hourly"]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets["hourly"] = subs[:len(subs)-1]
			return
		}
	}
}

func (m *SmtpMailer) Start() {
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
		ticker := time.NewTicker(24 * time.Hour)
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

func (m *SmtpMailer) Stop() {
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

func (m *SmtpMailer) sendDailyEmails() {
	m.mx.RLock()
	subs := append([]models.Subscription(nil), m.targets["daily"]...)
	m.mx.RUnlock()

	for _, sub := range subs {
		weatherData, err := m.WeatherService.GetCityWeather(sub.City)
		if err != nil {
			fmt.Printf("weather fetch error for %q: %v\n", sub.City, err)
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
				fmt.Printf("daily email error to %s: %v\n", email, err)
			}
		}(sub.Email, subject, body)
	}
}

func (m *SmtpMailer) sendHourlyEmails() {
	m.mx.RLock()
	subs := append([]models.Subscription(nil), m.targets["hourly"]...)
	m.mx.RUnlock()

	for _, sub := range subs {
		weatherData, err := m.WeatherService.GetCityWeather(sub.City)
		if err != nil {
			fmt.Printf("weather fetch error for %q: %v\n", sub.City, err)
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
				fmt.Printf("hourly email error to %s: %v\n", email, err)
			}
		}(sub.Email, subject, body)
	}
}

func (m *SmtpMailer) SendEmail(to, subject, body string) error {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", m.User))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("\r\n")
	msg.WriteString(body)

	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)
	tlsConf := &tls.Config{InsecureSkipVerify: true, ServerName: m.Host}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", m.Host, m.Port), tlsConf)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return fmt.Errorf("new SMTP client: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}
	if err := client.Mail(m.User); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set recipient: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("get data writer: %w", err)
	}
	defer wc.Close()

	if _, err := wc.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("write email body: %w", err)
	}
	return nil
}

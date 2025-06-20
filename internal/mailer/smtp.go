package mailer

import (
	"context"
	"crypto/tls"
	joinErr "errors"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"weather/internal/config"
	"weather/internal/models"

	"github.com/pkg/errors"
)

type SMTPMailer struct {
	User         string
	Password     string
	Host         string
	Port         string
	emailBuilder *EmailBuilder
}

func NewSMTPMailer(config config.SMTPConfig, emailBuilder *EmailBuilder) *SMTPMailer {
	return &SMTPMailer{
		User:         config.SMTPUser,
		Password:     config.SMTPPassword,
		Host:         config.SMTPHost,
		Port:         config.SMTPPort,
		emailBuilder: emailBuilder,
	}
}

func (m *SMTPMailer) sendEmails(
	ctx context.Context,
	forecasts []models.Forecast,
	subjectPrefix string,
) {
	for _, f := range forecasts {
		subj, body := m.emailBuilder.BuildWeatherForecastEmail(ctx, f.Email, f.City, f.Weather)

		go func(email, subj, msg string) {
			if err := m.SendEmail(email, subj, msg); err != nil {
				log.Printf("email error to %s: %v\n", email, err)
			}
		}(f.Email, subjectPrefix+subj, body)
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

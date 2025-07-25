package smtp

import (
	"context"
	"crypto/tls"
	joinErr "errors"
	"fmt"
	"mailer/internal/config"
	"mailer/internal/models"
	"net/smtp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type logger interface {
	LogInfo(msg string, fields ...zap.Field)
	FileLogInfo(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}

type SMTPMailer struct {
	User     string
	Password string
	Host     string
	Port     string
	logger   logger
}

func NewSMTPMailer(config config.SMTPConfig, logger logger) *SMTPMailer {
	return &SMTPMailer{
		User:     config.SMTPUser,
		Password: config.SMTPPassword,
		Host:     config.SMTPHost,
		Port:     config.SMTPPort,
		logger:   logger,
	}
}

func (m *SMTPMailer) SendEmail(email models.Email) (err error) {
	m.logger.FileLogInfo("sending email",
		zap.String("to", email.ToEmail),
		zap.String("subject", email.Subject),
		zap.String("body", email.Body),
	)

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", m.User))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", email.ToEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)
	tlsConf := &tls.Config{InsecureSkipVerify: false, ServerName: m.Host, MinVersion: tls.VersionTLS12}

	dialer := &tls.Dialer{
		Config: tlsConf,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%s", m.Host, m.Port))
	if err != nil {
		m.logger.LogError("creating SMTP client failed", zap.Error(err))
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
		m.logger.LogError("SMTP auth failed", zap.Error(err))
		return errors.Wrap(err, "SMTP auth")
	}
	if err := client.Mail(m.User); err != nil {
		m.logger.LogError("setting MAIL FROM failed", zap.Error(err))
		return errors.Wrap(err, "set sender")
	}
	if err := client.Rcpt(email.ToEmail); err != nil {
		m.logger.LogError("setting RCPT TO failed", zap.Error(err))
		return errors.Wrap(err, "set recipient")
	}

	wc, err := client.Data()
	if err != nil {
		m.logger.LogError("getting DATA writer failed", zap.Error(err))
		return errors.Wrap(err, "get data writer")
	}

	defer func() {
		closeErr := wc.Close()
		if closeErr != nil {
			m.logger.LogError("closing DATA writer failed", zap.Error(closeErr))
			closeErr = errors.Wrap(closeErr, "failed to close write closer")
			if err != nil {
				err = joinErr.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	if _, err := wc.Write([]byte(msg.String())); err != nil {
		m.logger.LogError("writing email body failed", zap.Error(err))
		return errors.Wrap(err, "write email body")
	}

	m.logger.LogInfo("Email sent successfully",
		zap.String("to", email.ToEmail),
	)
	return nil
}

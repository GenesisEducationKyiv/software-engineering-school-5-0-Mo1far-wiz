package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"weather-subscription/internal/config"
	"weather-subscription/internal/models"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

const DefaultMessagePriority = 5

type Logger interface {
	LogInfo(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}

type EmailPublisher struct {
	publisher    *rabbitmq.Publisher
	logger       Logger
	routingKey   string
	exchangeName string
}

func New(cfg config.PublishConfig, logger Logger) (*EmailPublisher, error) {
	conn, err := rabbitmq.NewConn(
		cfg.Addr,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ connection: %w", err)
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(cfg.ExchangeName),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		closeErr := conn.Close()
		return nil, errors.Join(err, closeErr)
	}

	// Log the configuration
	logger.LogInfo("Publisher created successfully",
		zap.String("exchangeName", cfg.ExchangeName),
		zap.String("exchangeKind", "topic"),
		zap.String("routingKey", cfg.RoutingKey),
	)

	return &EmailPublisher{
		publisher:    publisher,
		logger:       logger,
		routingKey:   cfg.RoutingKey,
		exchangeName: cfg.ExchangeName,
	}, nil
}

func (p *EmailPublisher) SendEmail(ctx context.Context, email models.Email) error {
	messageBytes, err := json.Marshal(email)
	if err != nil {
		p.logger.LogError("Failed to marshal email", zap.Error(err))
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	messageID := generateMessageID()

	headers := map[string]interface{}{
		"message_type": "email",
		"recipient":    email.ToEmail,
		"published_at": time.Now().UTC().Format(time.RFC3339),
		"content_type": "application/json",
		"exchange":     p.exchangeName,
	}

	publishOptions := []func(*rabbitmq.PublishOptions){
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsHeaders(headers),
		rabbitmq.WithPublishOptionsMessageID(messageID),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsPriority(DefaultMessagePriority),
		rabbitmq.WithPublishOptionsExchange(p.exchangeName),
	}

	p.logger.LogInfo("Publishing email message",
		zap.String("messageID", messageID),
		zap.String("exchangeName", p.exchangeName),
		zap.String("routingKey", p.routingKey),
		zap.String("recipient", email.ToEmail),
	)

	err = p.publisher.PublishWithContext(
		ctx,
		messageBytes,
		[]string{p.routingKey},
		publishOptions...,
	)

	if err != nil {
		p.logger.LogError("Failed to publish email",
			zap.Error(err),
			zap.String("messageID", messageID),
			zap.String("exchangeName", p.exchangeName),
			zap.String("routingKey", p.routingKey),
			zap.String("recipient", email.ToEmail),
			zap.String("subject", email.Subject),
		)
		return fmt.Errorf("failed to publish email: %w", err)
	}

	p.logger.LogInfo("Email published successfully",
		zap.String("messageID", messageID),
		zap.String("exchangeName", p.exchangeName),
		zap.String("routingKey", p.routingKey),
		zap.String("recipient", email.ToEmail),
		zap.String("subject", email.Subject),
		zap.Int("messageSize", len(messageBytes)),
	)

	return nil
}

func (p *EmailPublisher) BatchSendEmails(ctx context.Context, emails []models.Email) error {
	if len(emails) == 0 {
		return errors.New("no emails provided for batch sending")
	}

	successCount := 0
	var lastError error

	for i, email := range emails {
		if err := p.SendEmail(ctx, email); err != nil {
			p.logger.LogError("Failed to send email in batch",
				zap.Error(err),
				zap.Int("emailIndex", i),
				zap.String("recipient", email.ToEmail),
			)
			lastError = err
			continue
		}
		successCount++
	}

	p.logger.LogInfo("Batch email sending completed",
		zap.Int("totalEmails", len(emails)),
		zap.Int("successfulEmails", successCount),
		zap.Int("failedEmails", len(emails)-successCount),
	)

	if successCount == 0 {
		return fmt.Errorf("all emails in batch failed to send, last error: %w", lastError)
	}

	if successCount < len(emails) {
		return fmt.Errorf("partial batch failure: %d/%d emails sent successfully, last error: %w",
			successCount, len(emails), lastError)
	}

	return nil
}

func (p *EmailPublisher) Close() {
	if p.publisher != nil {
		p.publisher.Close()
	}
}

func generateMessageID() string {
	return fmt.Sprintf("email_%d", time.Now().UnixNano())
}

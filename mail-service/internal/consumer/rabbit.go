package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mailer/internal/config"
	"mailer/internal/models"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type smtpMailer interface {
	SendEmail(email models.Email) (err error)
}

type logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
}

type RabbitConsumer struct {
	consumer *rabbitmq.Consumer
	running  bool

	smtp   smtpMailer
	logger logger
}

func New(cfg config.RabbitConfig, smtp smtpMailer, logger logger) *RabbitConsumer {
	conn, err := rabbitmq.NewConn(
		cfg.Addr,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ connection: %v", err)
	}

	c, err := rabbitmq.NewConsumer(
		conn,
		cfg.QueueName,
		rabbitmq.WithConsumerOptionsRoutingKey(cfg.RoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(cfg.ExchangeName),
		rabbitmq.WithConsumerOptionsQueueDurable,
	)

	if err != nil {
		closeErr := conn.Close()
		log.Fatalf("Failed to create RabbitMQ consumer: %v", errors.Join(err, closeErr))
	}

	return &RabbitConsumer{
		consumer: c,
		smtp:     smtp,
		logger:   logger,
		running:  false,
	}
}
func (r *RabbitConsumer) Start(ctx context.Context) error {
	if r.running {
		return errors.New("consumer is already running")
	}

	r.running = true
	r.logger.ConsoleLogInfo("Starting RabbitMQ consumer...")

	go func() {
		defer func() { r.running = false }()

		err := r.consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			if err := r.processMessage(d); err != nil {
				r.logger.ConsoleLogError("Failed to process message", zap.Error(err))
				return r.handleError(err)
			}

			r.logger.ConsoleLogInfo("Message processed successfully", zap.String("messageID", d.MessageId))
			return rabbitmq.Ack
		})

		if err != nil {
			r.logger.ConsoleLogError("Consumer stopped with error", zap.Error(err))
		} else {
			r.logger.ConsoleLogInfo("Consumer stopped gracefully")
		}
	}()

	r.logger.ConsoleLogInfo("Consumer started and running persistently...")
	return nil
}

func (r *RabbitConsumer) processMessage(d rabbitmq.Delivery) error {
	r.logger.ConsoleLogInfo(
		"Received message",
		zap.String("messageID", d.MessageId),
		zap.String("routingKey", d.RoutingKey),
		zap.String("body", string(d.Body)),
	)

	var email models.Email
	if err := json.Unmarshal(d.Body, &email); err != nil {
		r.logger.ConsoleLogError("Failed to unmarshal message", zap.Error(err))
		return &ValidationError{Err: err}
	}

	return r.smtp.SendEmail(email)
}

func (r *RabbitConsumer) handleError(err error) rabbitmq.Action {
	switch err.(type) {
	case *ValidationError:
		r.logger.ConsoleLogError("Validation error, discarding message", zap.Error(err))
		return rabbitmq.NackDiscard
	default:
		r.logger.ConsoleLogError("Unknown error, discarding message", zap.Error(err))
		return rabbitmq.NackDiscard
	}
}

func (r *RabbitConsumer) Stop() {
	if !r.running {
		r.logger.ConsoleLogInfo("Consumer is not running")
		return
	}

	r.logger.ConsoleLogInfo("Stopping RabbitMQ consumer...")

	if r.consumer != nil {
		r.consumer.Close()
	}
}

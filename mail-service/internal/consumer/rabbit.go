package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mailer/internal/config"
	"mailer/internal/models"
	"time"

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
	stopCh   chan struct{}
	doneCh   chan struct{}

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
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsQueueDurable,   // Make queue persistent
		rabbitmq.WithConsumerOptionsConcurrency(5), // Process up to 5 messages concurrently
	)
	if err != nil {
		conn.Close()
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}

	return &RabbitConsumer{
		consumer: c,
		smtp:     smtp,
		logger:   logger,
		running:  false,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}
func (r *RabbitConsumer) Start(ctx context.Context) error {
	if r.running {
		return errors.New("consumer is already running")
	}

	r.running = true
	r.logger.ConsoleLogInfo("Starting RabbitMQ consumer...")

	go func() {
		defer close(r.doneCh)
		defer func() { r.running = false }()

		err := r.consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			select {
			case <-r.stopCh:
				r.logger.ConsoleLogInfo("Stop signal received, finishing current message...")
				return rabbitmq.Ack
			default:
			}

			if err := r.processMessage(ctx, d); err != nil {
				r.logger.ConsoleLogError("Failed to process message", zap.Error(err))
				return r.handleError(err, d)
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

func (r *RabbitConsumer) processMessage(ctx context.Context, d rabbitmq.Delivery) error {
	r.logger.ConsoleLogInfo("Received message", zap.String("messageID", d.MessageId), zap.String("routingKey", d.RoutingKey))

	var email models.Email
	if err := json.Unmarshal(d.Body, &email); err != nil {
		r.logger.ConsoleLogError("Failed to unmarshal message", zap.Error(err))
		return &ValidationError{Err: err}
	}

	return r.smtp.SendEmail(email)
}

func (r *RabbitConsumer) handleError(err error, d rabbitmq.Delivery) rabbitmq.Action {
	switch err.(type) {
	case *ValidationError:
		// Don't requeue validation errors - they'll keep failing
		r.logger.ConsoleLogError("Validation error, discarding message", zap.Error(err))
		return rabbitmq.NackDiscard
	case *TemporaryError:
		// Requeue temporary errors for retry
		r.logger.ConsoleLogError("Temporary error, requeuing message", zap.Error(err))
		return rabbitmq.NackRequeue
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

	close(r.stopCh)

	select {
	case <-r.doneCh:
		r.logger.ConsoleLogInfo("Consumer stopped gracefully")
	case <-time.After(30 * time.Second):
		r.logger.ConsoleLogError("Timeout waiting for consumer to stop, forcing close")
	}

	if r.consumer != nil {
		r.consumer.Close()
	}
}

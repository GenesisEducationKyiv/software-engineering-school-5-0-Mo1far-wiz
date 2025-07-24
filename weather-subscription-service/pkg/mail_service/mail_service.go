package mail_service

import (
	"context"
	"fmt"
	"pkg/protos/mailer"
	"sync"
	"time"
	"weather-subscription/internal/config"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	LogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}

type MailService struct {
	config config.GRPCConfig
	logger logger

	conn   *grpc.ClientConn
	client mailer.MailServiceClient
	stream mailer.MailService_SendEmailClient
	mu     sync.RWMutex
}

func NewMailService(cfg config.GRPCConfig, logger logger) *MailService {
	return &MailService{
		config: cfg,
		logger: logger,
	}
}

func (ms *MailService) Connect(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(ms.config.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		ms.logger.LogError("Failed to connect to mail service", zap.Error(err))
		return errors.Wrap(err, "failed to connect to mail service")
	}

	ms.conn = conn
	ms.client = mailer.NewMailServiceClient(conn)
	ms.logger.LogInfo("Connected to mail service", zap.String("address", ms.config.Addr))

	return nil
}

func (ms *MailService) Disconnect() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.stream != nil {
		if err := ms.stream.CloseSend(); err != nil {
			ms.logger.LogError("Failed to close stream", zap.Error(err))
		}
		ms.stream = nil
	}

	if ms.conn != nil {
		if err := ms.conn.Close(); err != nil {
			ms.logger.LogError("Failed to close connection", zap.Error(err))
			return err
		}
		ms.conn = nil
		ms.client = nil
	}

	ms.logger.LogInfo("Disconnected from mail service")
	return nil
}

func (ms *MailService) InitStream(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.client == nil {
		return fmt.Errorf("not connected to mail service")
	}

	if ms.stream != nil {
		if err := ms.stream.CloseSend(); err != nil {
			ms.logger.LogError("Failed to close existing stream", zap.Error(err))
		}
	}

	stream, err := ms.client.SendEmail(ctx)
	if err != nil {
		ms.logger.LogError("Failed to create email stream", zap.Error(err))
		return errors.Wrap(err, "failed to create email stream")
	}

	ms.stream = stream
	ms.logger.LogInfo("Email stream initialized")
	return nil
}

func (ms *MailService) SendEmail(email *mailer.Email) error {
	ms.mu.RLock()
	stream := ms.stream
	ms.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("stream not initialized")
	}

	if err := stream.Send(email); err != nil {
		ms.logger.LogError("Failed to send email",
			zap.Error(err),
			zap.String("to", *email.ToEmail),
			zap.String("subject", *email.Subject),
		)
		return errors.Wrap(err, "failed to send email")
	}

	ms.logger.LogInfo("Email sent to stream",
		zap.String("to", *email.ToEmail),
		zap.String("subject", *email.Subject),
	)
	return nil
}

func (ms *MailService) SendEmails(emails []*mailer.Email) error {
	for _, email := range emails {
		if err := ms.SendEmail(email); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MailService) FinishAndGetResult() (*mailer.Result, error) {
	ms.mu.Lock()
	stream := ms.stream
	ms.stream = nil
	ms.mu.Unlock()

	if stream == nil {
		return nil, fmt.Errorf("no active stream")
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		ms.logger.LogError("Failed to close stream and receive result", zap.Error(err))
		return nil, errors.Wrap(err, "failed to close stream and receive result")
	}

	if result != nil {
		ms.logger.LogInfo("Email batch completed",
			zap.Uint64("success", *result.Success),
			zap.Uint64("failed", *result.Failed),
		)
	}

	return result, nil
}

func (ms *MailService) SendEmailBatch(ctx context.Context, emails []*mailer.Email) (*mailer.Result, error) {
	if err := ms.InitStream(ctx); err != nil {
		return nil, err
	}

	if err := ms.SendEmails(emails); err != nil {
		return nil, err
	}

	return ms.FinishAndGetResult()
}

func (ms *MailService) IsConnected() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.conn != nil
}

func (ms *MailService) IsStreamActive() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.stream != nil
}

func (ms *MailService) ConnectWithRetry(ctx context.Context, maxRetries int, retryDelay time.Duration) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := ms.Connect(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			ms.logger.LogError("Connection attempt failed",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Error(err),
			)
		}

		if i < maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("failed to connect after %d retries: %w", maxRetries, lastErr)
}

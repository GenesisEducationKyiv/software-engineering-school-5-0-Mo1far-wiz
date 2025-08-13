package service

import (
	"io"
	"mailer/internal/models"

	"go.uber.org/zap"

	pb "pkg/protos/mailer"
)

type smtpMailer interface {
	SendEmail(email models.Email) (err error)
}
type logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	LogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	LogError(msg string, fields ...zap.Field)
}
type Mailer struct {
	pb.UnimplementedMailServiceServer

	smtp   smtpMailer
	logger logger
}

func NewMailer(smtp smtpMailer, logger logger) *Mailer {
	return &Mailer{
		smtp:   smtp,
		logger: logger,
	}
}

func (s *Mailer) SendEmail(stream pb.MailService_SendEmailServer) error {
	s.logger.LogInfo("Email stream opened")

	var success, fail uint64

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			res := &pb.Result{
				Success: &success,
				Failed:  &fail,
			}
			if serr := stream.SendAndClose(res); serr != nil {
				s.logger.ConsoleLogError("closing stream failed", zap.Error(serr))
				return serr
			}
			s.logger.LogInfo("stream is closed",
				zap.Uint64("success", success),
				zap.Uint64("failed", fail),
			)
			return nil
		}
		if err != nil {
			s.logger.LogError("failed to receive email", zap.Error(err))
			return err
		}

		email := models.Email{
			ToEmail: in.GetToEmail(),
			Subject: in.GetSubject(),
			Body:    in.GetBody(),
		}

		if err := s.smtp.SendEmail(email); err != nil {
			fail++
			s.logger.LogError(
				"smtp failed", zap.Error(err),
				zap.String("for email", email.ToEmail),
			)
		} else {
			success++
		}
	}
}

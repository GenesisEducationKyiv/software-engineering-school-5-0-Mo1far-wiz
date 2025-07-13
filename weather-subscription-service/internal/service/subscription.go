package service

import (
	"context"
	"errors"
	"weather-subscription/internal/models"
	"weather-subscription/internal/svc"

	"go.uber.org/zap"
)

type SubscriptionStore interface {
	Create(context.Context, *models.Subscription) error
	Confirm(ctx context.Context, token string) (models.Subscription, error)
	Unsubscribe(ctx context.Context, token string) (models.Subscription, error)
}

type EmailSender interface {
	SendEmail(to, subject, body string) (err error)
}

type SubscriptionTargetManager interface {
	AddTarget(sub models.Subscription)
	RemoveTarget(email string, frequency string)
}

type Subscription struct {
	store         SubscriptionStore
	targetManager SubscriptionTargetManager
	emailSender   EmailSender
	logger        Logger
}

func NewSubscription(
	store SubscriptionStore,
	emailSender EmailSender,
	targetManager SubscriptionTargetManager,
	logger Logger,
) *Subscription {
	return &Subscription{
		store:         store,
		emailSender:   emailSender,
		targetManager: targetManager,
		logger:        logger,
	}
}

func (s *Subscription) Subscribe(ctx context.Context, subscription models.Subscription) error {
	err := s.store.Create(ctx, &subscription)
	if err != nil {
		s.logger.ConsoleLogError("can't create subscription",
			zap.String("error", err.Error()))
		return errors.Join(svc.ErrorSubscriptionCreate, err)
	}

	err = s.emailSender.SendEmail(subscription.Email, "Your token", subscription.Token)
	if err != nil {
		s.logger.ConsoleLogError("failed to send confirmation email",
			zap.String("error", err.Error()))
		return errors.Join(svc.ErrorSendEmail, err)
	}
	return nil
}

func (s *Subscription) Confirm(ctx context.Context, token string) error {
	sub, err := s.store.Confirm(ctx, token)
	if err != nil {
		s.logger.ConsoleLogError("can't confirm subscription",
			zap.String("error", err.Error()))
		return errors.Join(svc.ErrorSubscriptionConfirm, err)
	}

	s.targetManager.AddTarget(sub)

	return nil
}

func (s *Subscription) Unsubscribe(ctx context.Context, token string) error {
	sub, err := s.store.Unsubscribe(ctx, token)
	if err != nil {
		s.logger.ConsoleLogError("can't cancel subscription",
			zap.String("error", err.Error()))
		return errors.Join(svc.ErrorSubscriptionUnsubscribe, err)
	}

	s.targetManager.RemoveTarget(sub.Email, sub.Frequency)

	return nil
}

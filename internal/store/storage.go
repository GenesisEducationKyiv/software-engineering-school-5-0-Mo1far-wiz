package store

import (
	"context"
	"database/sql"
	"time"
	"weather/internal/models"
)

const QueryTimeoutDuration = 1 * time.Second

type Storage struct {
	Subscription interface {
		Create(context.Context, *models.Subscription) error
		Confirm(ctx context.Context, token string) (models.Subscription, error)
		Unsubscribe(ctx context.Context, token string) (models.Subscription, error)
	}
	Mailer interface {
		GetSubscribed(ctx context.Context) ([]models.Subscription, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Subscription: &SubscriptionStore{db},
		Mailer:       &MailerStore{db},
	}
}

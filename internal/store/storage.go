package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"weather/internal/models"
)

const QueryTimeoutDuration = 1 * time.Second

var ErrorNotFound = errors.New("resource not found")
var ErrorAlreadyExists = errors.New("resource already exists")

type Storage struct {
	Subscription interface {
		Create(context.Context, *models.Subscription) error
		Confirm(ctx context.Context, token string) (models.Subscription, error)
		Unsubscribe(ctx context.Context, token string) (models.Subscription, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Subscription: &SubscriptionStore{db},
	}
}

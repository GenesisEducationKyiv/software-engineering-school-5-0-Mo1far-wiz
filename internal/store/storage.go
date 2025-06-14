package store

import (
	"database/sql"
	"time"
	"weather/internal/api/handlers"
	"weather/internal/mailer"
)

const QueryTimeoutDuration = 1 * time.Second

type Storage struct {
	Subscription handlers.SubscriptionStore
	Mailer       mailer.SubscribedStore
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Subscription: &SubscriptionStore{db},
		Mailer:       &MailerStore{db},
	}
}

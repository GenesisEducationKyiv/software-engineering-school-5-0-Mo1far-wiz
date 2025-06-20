package store

import (
	"context"
	"database/sql"
	"weather/internal/models"
	"weather/internal/srverrors"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	pgAlreadyExistsCode       = "23505"
	pgAlreadyExistsConstraint = "subscriptions_email_key"
)

var _ Subscription = (*SubscriptionStore)(nil)

type SubscriptionStore struct {
	db *sql.DB
}

func (ss *SubscriptionStore) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO weather.subscriptions (email, city, frequency, token)
		VALUES ($1, $2, $3, $4)
		RETURNING weather.subscriptions.id;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	row := ss.db.QueryRowContext(
		ctx,
		query,
		sub.Email,
		sub.City,
		sub.Frequency,
		sub.Token,
	)

	err := row.Scan(&sub.ID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == pgAlreadyExistsCode && pgErr.Constraint == pgAlreadyExistsConstraint {
				return srverrors.ErrorAlreadyExists
			}
		}
		return errors.Wrap(err, "failed to create subscription")
	}

	return nil
}

func (ss *SubscriptionStore) Confirm(ctx context.Context, token string) (models.Subscription, error) {
	const query = `
        UPDATE weather.subscriptions
        SET confirmed = true
        WHERE token = $1
        RETURNING id, email, city, frequency, token;
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var sub models.Subscription
	err := ss.db.
		QueryRowContext(ctx, query, token).
		Scan(&sub.ID, &sub.Email, &sub.City, &sub.Frequency, &sub.Token)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Subscription{}, srverrors.ErrorNotFound
		}
		return models.Subscription{}, errors.Wrap(err, "failed to confirm subscription")
	}

	return sub, nil
}

func (ss *SubscriptionStore) Unsubscribe(ctx context.Context, token string) (models.Subscription, error) {
	const query = `
        UPDATE weather.subscriptions
        SET confirmed = false
        WHERE token = $1
        RETURNING id, email, city, frequency, token;
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var sub models.Subscription
	err := ss.db.
		QueryRowContext(ctx, query, token).
		Scan(&sub.ID, &sub.Email, &sub.City, &sub.Frequency, &sub.Token)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Subscription{}, srverrors.ErrorNotFound
		}
		return models.Subscription{}, errors.Wrap(err, "failed to unsubscribe")
	}

	return sub, nil
}

package store

import (
	"context"
	"database/sql"
	"weather/internal/models"

	joinErr "errors"

	"github.com/pkg/errors"
)

var _ Mailer = (*MailerStore)(nil)

type MailerStore struct {
	db *sql.DB
}

func (ss *MailerStore) GetSubscribed(ctx context.Context) (subs []models.Subscription, err error) {
	query := `
		SELECT
			id,
			email,
			city,
			frequency,
			token
		FROM weather.subscriptions
		WHERE confirmed = true;
	`

	rows, err := ss.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get subscriptions")
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			closeErr = errors.Wrap(closeErr, "failed to close rows")
			if err != nil {
				err = joinErr.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	for rows.Next() {
		var s models.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.Email,
			&s.City,
			&s.Frequency,
			&s.Token,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan subscription row")
		}
		subs = append(subs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row iteration error")
	}

	return subs, nil
}

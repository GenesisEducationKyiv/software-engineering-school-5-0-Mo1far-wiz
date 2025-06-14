package database

import (
	"context"
	"database/sql"
	"time"
	"weather/internal/config"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	driver      = "postgres"
	pingTimeout = 5 * time.Second
)

func New(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open(driver, cfg.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s connection", driver)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse MaxIdleTime duration: %s", cfg.MaxIdleTime)
	}
	db.SetConnMaxIdleTime(duration)

	return db, nil
}

func ValidateConnection(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	return errors.Wrap(db.PingContext(ctx), "ping wasn't successful")
}

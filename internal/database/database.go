package database

import (
	"context"
	"database/sql"
	"time"
	"weather/internal/config"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const driver = "postgres"

func New(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open(driver, cfg.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s connection", driver)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, errors.Wrapf(err, "cant pass MaxIdleTime durations: %s", cfg.MaxIdleTime)
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "ping wasn't successful")
	}

	return db, nil
}

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"weather/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func MigrateUp(dbURL string, migrationPath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	if filepath.IsAbs(migrationPath) {
		migrationPath = filepath.Clean(migrationPath)
	} else {
		migrationPath = filepath.Join(wd, migrationPath)
		migrationPath = filepath.Clean(migrationPath)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

func MigrateDown(dbURL string, migrationPath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	if filepath.IsAbs(migrationPath) {
		migrationPath = filepath.Clean(migrationPath)
	} else {
		migrationPath = filepath.Join(wd, migrationPath)
		migrationPath = filepath.Clean(migrationPath)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		dbURL,
	)
	if err != nil {
		return errors.Wrap(err, "creation of migrate instructions")
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "running down migrations")
	}

	return nil
}

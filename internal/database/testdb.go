package database

import (
	"database/sql"
	"fmt"
	"testing"
	"weather/internal/config"
	"weather/internal/env"
)

func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbName := env.GetString("TEST_DB_NAME", "test_weather")
	dbPassword := env.GetString("TEST_DB_PASSWORD", "password")
	dbUser := env.GetString("TEST_DB_USER", "test")
	dbHost := env.GetString("TEST_DB_HOST", "127.0.0.1")
	dbPort := env.GetInt("TEST_DB_PORT", 5432)
	dbSSL := env.GetString("TEST_DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSL,
	)

	cfg := config.DBConfig{
		Addr:         dsn,
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxIdleTime:  "1m",
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := ValidateConnection(db); err != nil {
		db.Close()
		t.Fatalf("ping test db failed: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing test db: %v", err)
		}
	})

	return db
}

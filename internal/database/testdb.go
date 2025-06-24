package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"
	"weather/internal/config"
)

func NewTestDB(dsn string, t *testing.T) *sql.DB {
	t.Helper()

	log.Println(dsn)

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
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
		t.Fatalf("ping test db failed: %v", err)
	}

	migrationPath := getProjectRoot() + "/internal/database/migrations"
	log.Println(migrationPath)

	err = MigrateUp(dsn, migrationPath)
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing test db: %v", err)
		}
	})

	return db
}

func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return dir
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	wd, err := os.Getwd()
	if err != nil {
		return dir
	}
	return wd
}

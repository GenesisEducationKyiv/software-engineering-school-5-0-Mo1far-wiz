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

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "../../../")
	migrationPath := filepath.Join(projectRoot, "internal/database/migrations")

	migrationPath = filepath.Clean(migrationPath)

	log.Printf("Migration path: %s", migrationPath)

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

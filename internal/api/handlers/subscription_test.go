package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"weather/internal/api/handlers"
	"weather/internal/database"
	"weather/internal/models"
	"weather/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func setupENV() {
	err := os.Setenv("TEST_DB_NAME", "test_weather")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_NAME"))
	}
	err = os.Setenv("TEST_DB_USER", "test")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_USER"))
	}
	err = os.Setenv("TEST_DB_PASSWORD", "password")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_PASSWORD"))
	}
	err = os.Setenv("TEST_DB_HOST", "127.0.0.1")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_HOST"))
	}
	err = os.Setenv("TEST_DB_PORT", "5433")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_PORT"))
	}
	err = os.Setenv("TEST_DB_SSL_MODE", "disable")
	if err != nil {
		log.Fatal(errors.Wrap(err, "test env TEST_DB_SSL_MODE"))
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	setupENV()

	db := database.NewTestDB(t)

	stmts := []string{
		`CREATE SCHEMA IF NOT EXISTS weather;`,
		`DROP TABLE IF EXISTS weather.subscriptions;`,
		`CREATE TABLE weather.subscriptions (
            id         SERIAL PRIMARY KEY,
            email      TEXT UNIQUE NOT NULL,
            city       TEXT NOT NULL,
            frequency  TEXT NOT NULL,
            token      TEXT NOT NULL,
            confirmed  BOOLEAN NOT NULL DEFAULT FALSE,
            subscribed BOOLEAN NOT NULL DEFAULT TRUE
        );`,
	}

	for _, sqlStmt := range stmts {
		if _, err := db.ExecContext(context.Background(), sqlStmt); err != nil {
			t.Fatalf("failed to prepare test schema: %v", err)
		}
	}

	if _, err := db.ExecContext(context.Background(),
		`TRUNCATE TABLE weather.subscriptions RESTART IDENTITY;`); err != nil {
		t.Logf("Warning: failed to clean test data: %v", err)
	}

	return db
}

type stubEmailSender struct {
	calls []string
}

func (s *stubEmailSender) SendEmail(to, subj, body string) error {
	s.calls = append(s.calls, fmt.Sprintf("%s|%s", to, subj))
	return nil
}

type noopTargetMgr struct{}

func (n *noopTargetMgr) AddTarget(models.Subscription)             {}
func (n *noopTargetMgr) RemoveTarget(string, string)               {}
func (n *noopTargetMgr) GetTargets(_ string) []models.Subscription { return nil }

func TestCreateAndConfirm(t *testing.T) {
	db := setupTestDB(t)
	store := store.NewStorage(db)
	ctx := context.Background()

	sub := &models.Subscription{
		Email:     "alice@example.com",
		City:      "Kyiv",
		Frequency: "daily",
		Token:     "tok123",
	}
	err := store.Subscription.Create(ctx, sub)
	assert.NoError(t, err, "should create without error")
	assert.NotZero(t, sub.ID, "ID should be set by DB")

	var (
		email, city, freq, token string
		confirmed, subscribed    bool
	)
	row := db.QueryRowContext(ctx,
		`SELECT email, city, frequency, token, confirmed, subscribed
         FROM weather.subscriptions WHERE id = $1`,
		sub.ID,
	)
	err = row.Scan(&email, &city, &freq, &token, &confirmed, &subscribed)
	assert.NoError(t, err)
	assert.Equal(t, sub.Email, email)
	assert.False(t, confirmed, "new record should be unconfirmed")
	assert.True(t, subscribed, "new record should be subscribed")

	_, err = store.Subscription.Confirm(ctx, sub.Token)
	assert.NoError(t, err)

	err = db.QueryRowContext(ctx,
		`SELECT confirmed FROM weather.subscriptions WHERE token = $1`,
		sub.Token,
	).Scan(&confirmed)
	assert.NoError(t, err)
	assert.True(t, confirmed)
}

func TestSubscribeHandler(t *testing.T) {
	db := setupTestDB(t)
	store := store.NewStorage(db)
	emailer := &stubEmailSender{}
	targetMgr := &noopTargetMgr{}

	handler := handlers.NewSubscriptionHandler(store.Subscription, emailer, targetMgr)
	router := gin.New()
	router.POST("/subscribe", handler.Subscribe)

	w := httptest.NewRecorder()
	body := `{"email":"bob@example.com","city":"Lviv","frequency":"hourly"}`
	req := httptest.NewRequest("POST", "/subscribe", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscription successful")

	var gotEmail string
	err := db.QueryRowContext(context.Background(),
		`SELECT email FROM weather.subscriptions WHERE email = $1`, "bob@example.com",
	).Scan(&gotEmail)
	assert.NoError(t, err)
	assert.Equal(t, "bob@example.com", gotEmail)

	assert.Len(t, emailer.calls, 1)
	assert.Contains(t, emailer.calls[0], "bob@example.com|")
}

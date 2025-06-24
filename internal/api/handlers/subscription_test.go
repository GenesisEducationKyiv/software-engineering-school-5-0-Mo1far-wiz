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

func TestMain(m *testing.M) {
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

	os.Exit(m.Run())
}
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_DB_SSL_MODE"),
	)

	return database.NewTestDB(dbURL, t)
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

	// Create
	err := store.Subscription.Create(ctx, sub)
	assert.NoError(t, err)
	assert.NotZero(t, sub.ID)

	// Check defaults
	var confirmed bool
	err = db.QueryRowContext(ctx,
		`SELECT confirmed FROM weather.subscriptions WHERE id = $1`,
		sub.ID,
	).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed)

	// Confirm
	_, err = store.Subscription.Confirm(ctx, sub.Token)
	assert.NoError(t, err)

	// Verify confirmed = true
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
	reqBody := `{"email":"bob@example.com","city":"Lviv","frequency":"hourly"}`
	req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscription successful")

	var gotEmail string
	err := db.QueryRowContext(context.Background(),
		`SELECT email FROM weather.subscriptions WHERE email = $1`,
		"bob@example.com",
	).Scan(&gotEmail)
	assert.NoError(t, err)
	assert.Equal(t, "bob@example.com", gotEmail)

	assert.Len(t, emailer.calls, 1)
	assert.Contains(t, emailer.calls[0], "bob@example.com|")
}

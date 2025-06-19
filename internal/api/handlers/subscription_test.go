package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather/internal/api/handlers"
	mock_handlers "weather/internal/api/handlers/mocks"
	"weather/internal/models"
	"weather/internal/srverrors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateToken(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("token", "")
	_, err := handlers.ValidateToken(c)
	assert.ErrorIs(t, err, srverrors.ErrorTokenNotFound)

	c.Set("token", "my-token")
	tok, err := handlers.ValidateToken(c)
	assert.NoError(t, err)
	assert.Equal(t, "my-token", tok)
}

func TestSHA256Token(t *testing.T) {
	t.Parallel()

	hashPhrase := "pes patron"
	validToken := "28656ff9525f32170fc9eebb47a75b969c12599fcdd591be8555d12e212a8d3e"

	token := handlers.SHA256Token(hashPhrase)

	assert.Equal(t, validToken, token)
}

func TestSubscription_Success(t *testing.T) {
	t.Parallel()

	var (
		body       = `{"email":"cringe@gmail.com","city":"Kyiv","frequency":"daily"}`
		setupMocks = func(store *mock_handlers.MockSubscriptionStore, email *mock_handlers.MockEmailSender) {
			store.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, sub *models.Subscription) error {
					return nil
				})
			email.EXPECT().
				SendEmail("cringe@gmail.com", gomock.Any(), gomock.Any()).
				Return(nil)
		}
		wantCode     = http.StatusOK
		wantContains = "Subscription successful"
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_handlers.NewMockSubscriptionStore(ctrl)
	email := mock_handlers.NewMockEmailSender(ctrl)
	target := mock_handlers.NewMockSubscriptionTargetManager(ctrl)

	setupMocks(store, email)

	handler := handlers.NewSubscriptionHandler(store, email, target)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request = req
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, wantCode, w.Code)
	assert.Contains(t, w.Body.String(), wantContains)
}

func TestSubscription_ErrorAlreadyExists(t *testing.T) {
	t.Parallel()

	body := `{"email":"cringe@gmail.com","city":"Kyiv","frequency":"daily"}`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_handlers.NewMockSubscriptionStore(ctrl)
	email := mock_handlers.NewMockEmailSender(ctrl)
	target := mock_handlers.NewMockSubscriptionTargetManager(ctrl)

	store.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(srverrors.ErrorAlreadyExists)

	email.EXPECT().
		SendEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0) // like if we got previous error we shouldn't send any emails

	handler := handlers.NewSubscriptionHandler(store, email, target)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request = req
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Email already subscribed")
}

func TestConfirm_Success(t *testing.T) {
	t.Parallel()

	var (
		sub = models.Subscription{
			Email:     "cringe@gmail.com",
			City:      "Kyiv",
			Frequency: "daily",
			Token:     "sub-token",
		}
		setupMocks = func(store *mock_handlers.MockSubscriptionStore,
			target *mock_handlers.MockSubscriptionTargetManager,
		) {
			store.EXPECT().
				Confirm(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string) (models.Subscription, error) {
					return sub, nil
				})
			target.EXPECT().
				AddTarget(sub)
		}
		wantCode     = http.StatusOK
		wantContains = "\"Subscription confirmed successfully\""
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_handlers.NewMockSubscriptionStore(ctrl)
	email := mock_handlers.NewMockEmailSender(ctrl)
	target := mock_handlers.NewMockSubscriptionTargetManager(ctrl)

	setupMocks(store, target)

	handler := handlers.NewSubscriptionHandler(store, email, target)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/confirm", nil)
	c.Request = req
	c.Set("token", sub.Token)

	handler.Confirm(c)

	assert.NoError(t, err)
	assert.Equal(t, wantCode, w.Code)
	assert.Contains(t, wantContains, w.Body.String())
}

func TestConfirm_InvalidToken(t *testing.T) {
	t.Parallel()

	var (
		wantCode     = http.StatusNotFound
		wantContains = "\"token not found\""
	)

	handler := handlers.NewSubscriptionHandler(
		mock_handlers.NewMockSubscriptionStore(gomock.NewController(t)),
		mock_handlers.NewMockEmailSender(gomock.NewController(t)),
		mock_handlers.NewMockSubscriptionTargetManager(gomock.NewController(t)),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/confirm", nil)
	c.Request = req

	handler.Confirm(c)

	assert.NoError(t, err)
	assert.Equal(t, wantCode, w.Code)
	assert.Contains(t, wantContains, w.Body.String())
}

func TestUnsubscribe_Success(t *testing.T) {
	t.Parallel()

	var (
		token = "token"
		sub   = models.Subscription{
			Email:     "cringe@gmail.com",
			City:      "Kyiv",
			Frequency: "daily",
			Token:     token,
		}
		setupMocks = func(store *mock_handlers.MockSubscriptionStore,
			target *mock_handlers.MockSubscriptionTargetManager,
		) {
			store.EXPECT().
				Unsubscribe(gomock.Any(), token).
				DoAndReturn(func(_ context.Context, _ string) (models.Subscription, error) {
					return sub, nil
				})
			target.EXPECT().
				RemoveTarget(sub.Email, sub.Frequency)
		}
		wantCode     = http.StatusOK
		wantContains = "\"Unsubscribed successfully\""
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_handlers.NewMockSubscriptionStore(ctrl)
	email := mock_handlers.NewMockEmailSender(ctrl)
	target := mock_handlers.NewMockSubscriptionTargetManager(ctrl)

	setupMocks(store, target)

	handler := handlers.NewSubscriptionHandler(store, email, target)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/unsubscribe", nil)
	c.Request = req
	c.Set("token", sub.Token)

	handler.Unsubscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, wantCode, w.Code)
	assert.Contains(t, wantContains, w.Body.String())
}

func TestUnsubscribe_InvalidToken(t *testing.T) {
	t.Parallel()

	var (
		wantCode     = http.StatusNotFound
		wantContains = "\"token not found\""
	)

	handler := handlers.NewSubscriptionHandler(
		mock_handlers.NewMockSubscriptionStore(gomock.NewController(t)),
		mock_handlers.NewMockEmailSender(gomock.NewController(t)),
		mock_handlers.NewMockSubscriptionTargetManager(gomock.NewController(t)),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/unsubscribe", nil)
	c.Request = req

	handler.Confirm(c)

	assert.NoError(t, err)
	assert.Equal(t, wantCode, w.Code)
	assert.Contains(t, wantContains, w.Body.String())
}

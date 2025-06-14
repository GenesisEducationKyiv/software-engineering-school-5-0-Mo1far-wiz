package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"weather/internal/models"
	"weather/internal/srverrors"

	"github.com/gin-gonic/gin"
)

type SubscriptionStore interface {
	Create(context.Context, *models.Subscription) error
	Confirm(ctx context.Context, token string) (models.Subscription, error)
	Unsubscribe(ctx context.Context, token string) (models.Subscription, error)
}

type MailerService interface {
	SendEmail(to, subject, body string) (err error)

	AddDailyTarget(sub models.Subscription)
	AddHourlyTarget(sub models.Subscription)

	RemoveDailyTarget(email string)
	RemoveHourlyTarget(email string)
}

type SubscriptionHandler struct {
	store         SubscriptionStore
	mailerService MailerService
}

type subscribeRequest struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

func sha256Token(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func NewSubscriptionHandler(store SubscriptionStore, mailerService MailerService) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:         store,
		mailerService: mailerService,
	}
}

func (s *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logErrorF(err, "cant bind request to json")
		c.JSON(http.StatusBadRequest, "Invalid input")
		return
	}

	subscription := models.Subscription{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
		Token:     sha256Token(req.Email + req.City + req.Frequency),
	}

	err := s.store.Create(c.Request.Context(), &subscription)
	if err != nil {
		logErrorF(err, "cant create subscription")
		if errors.Is(err, srverrors.ErrorAlreadyExists) {
			c.JSON(http.StatusConflict, "Email already subscribed")
		} else {
			c.JSON(http.StatusInternalServerError, "Can't create subscription")
		}
		return
	}

	err = s.mailerService.SendEmail(subscription.Email, "Your token", subscription.Token)
	if err != nil {
		logErrorF(err, "failed to send confirmation email")
		c.JSON(http.StatusInternalServerError, "Can't send email")
		return
	}

	c.JSON(http.StatusOK, "Subscription successful. Confirmation email sent.")
}

func (s *SubscriptionHandler) Confirm(c *gin.Context) {
	token := c.GetString("token")
	if token == "" || token == ":token" {
		c.JSON(http.StatusNotFound, "Token not found")
		return
	}

	sub, err := s.store.Confirm(c.Request.Context(), token)
	if err != nil {
		logErrorF(err, "cant confirm subscription")
		c.JSON(http.StatusNotFound, "Cant confirm subscription")
		return
	}

	switch sub.Frequency {
	case models.Hourly:
		s.mailerService.AddHourlyTarget(sub)
	case models.Daily:
		s.mailerService.AddDailyTarget(sub)
	}

	c.JSON(http.StatusOK, "Subscription confirmed successfully")
}

func (s *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.GetString("token")
	if token == "" || token == ":token" {
		c.JSON(http.StatusNotFound, "Token not found")
		return
	}

	sub, err := s.store.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		logErrorF(err, "cant cancel subscription")
		c.JSON(http.StatusNotFound, "Cant cancel subscription")
		return
	}

	switch sub.Frequency {
	case models.Hourly:
		s.mailerService.RemoveHourlyTarget(sub.Email)
	case models.Daily:
		s.mailerService.RemoveDailyTarget(sub.Email)
	}

	c.JSON(http.StatusOK, "Unsubscribed successfully")
}

package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"weather/internal/mailer"
	"weather/internal/models"
	"weather/internal/store"

	"github.com/gin-gonic/gin"
)

type subscribeRequest struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

func SHA256Token(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

type SubscriptionHandler struct {
	store         store.Storage
	mailerService *mailer.SmtpMailer
}

func NewSubscriptionHandler(store store.Storage, mailerService *mailer.SmtpMailer) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:         store,
		mailerService: mailerService,
	}
}

func (s *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logError(err, "cant bind request to json")
		c.JSON(http.StatusUnprocessableEntity, "Invalid input")
		return
	}

	subscription := models.Subscription{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
		Token:     SHA256Token(req.Email),
	}

	err := s.store.Subscription.Create(c.Request.Context(), &subscription)
	if err != nil {
		logError(err, "cant create subscription")
		if errors.Is(err, store.ErrorAlreadyExists) {
			c.JSON(http.StatusBadRequest, "Email already subscribed")

		} else {
			c.JSON(http.StatusBadRequest, "Invalid input")
		}
		return
	}

	err = s.mailerService.SendEmail(subscription.Email, "Your token", subscription.Token)
	if err != nil {
		logError(err, "cant bind request to json")
		c.JSON(http.StatusUnprocessableEntity, "Invalid input")
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

	sub, err := s.store.Subscription.Confirm(c.Request.Context(), token)
	if err != nil {
		logError(err, "cant confirm subscription")
		c.JSON(http.StatusBadRequest, "Invalid token")
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

	sub, err := s.store.Subscription.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		logError(err, "cant cancel subscription")
		c.JSON(http.StatusBadRequest, "Invalid token")
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

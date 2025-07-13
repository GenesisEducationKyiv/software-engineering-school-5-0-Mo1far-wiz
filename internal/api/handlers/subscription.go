package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"weather/internal/models"
	"weather/internal/svc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SubscriptionStore interface {
	Create(context.Context, *models.Subscription) error
	Confirm(ctx context.Context, token string) (models.Subscription, error)
	Unsubscribe(ctx context.Context, token string) (models.Subscription, error)
}

type EmailSender interface {
	SendEmail(to, subject, body string) (err error)
}

type SubscriptionTargetManager interface {
	AddTarget(sub models.Subscription)
	RemoveTarget(email string, frequency string)
}

type SubscriptionHandler struct {
	store         SubscriptionStore
	targetManager SubscriptionTargetManager
	emailSender   EmailSender
	logger        Logger
}

type subscribeRequest struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

func SHA256Token(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func ValidateToken(c *gin.Context) (string, error) {
	token := c.GetString("token")
	if token == "" || token == ":token" {
		return "", svc.ErrorTokenNotFound
	}

	return token, nil
}

func NewSubscriptionHandler(
	store SubscriptionStore,
	emailSender EmailSender,
	targetManager SubscriptionTargetManager,
	logger Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:         store,
		emailSender:   emailSender,
		targetManager: targetManager,
		logger:        logger,
	}
}

func (s *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.ConsoleLogError("cant bind request to json",
			zap.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, "Invalid input")
		return
	}

	subscription := models.Subscription{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
		Token:     SHA256Token(req.Email + req.City + req.Frequency),
	}

	err := s.store.Create(c.Request.Context(), &subscription)
	if err != nil {
		s.logger.ConsoleLogError("can't create subscription",
			zap.String("error", err.Error()))
		if errors.Is(err, svc.ErrorAlreadyExists) {
			c.JSON(http.StatusConflict, "Email already subscribed")
		} else {
			c.JSON(http.StatusInternalServerError, "Can't create subscription")
		}
		return
	}

	err = s.emailSender.SendEmail(subscription.Email, "Your token", subscription.Token)
	if err != nil {
		s.logger.ConsoleLogError("failed to send confirmation email",
			zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, "Can't send email")
		return
	}

	c.JSON(http.StatusOK, "Subscription successful. Confirmation email sent.")
}

func (s *SubscriptionHandler) Confirm(c *gin.Context) {
	token, err := ValidateToken(c)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	sub, err := s.store.Confirm(c.Request.Context(), token)
	if err != nil {
		s.logger.ConsoleLogError("can't confirm subscription",
			zap.String("error", err.Error()))
		c.JSON(http.StatusNotFound, "Can't confirm subscription")
		return
	}

	s.targetManager.AddTarget(sub)

	c.JSON(http.StatusOK, "Subscription confirmed successfully")
}

func (s *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token, err := ValidateToken(c)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	sub, err := s.store.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		s.logger.ConsoleLogError("can't cancel subscription",
			zap.String("error", err.Error()))
		c.JSON(http.StatusNotFound, "Can't cancel subscription")
		return
	}

	s.targetManager.RemoveTarget(sub.Email, sub.Frequency)

	c.JSON(http.StatusOK, "Unsubscribed successfully")
}

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

type SubscriptionService interface {
	Subscribe(ctx context.Context, subscription models.Subscription) error
	Confirm(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
}

type SubscriptionHandler struct {
	subscription SubscriptionService
	logger       Logger
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
	service SubscriptionService,
	logger Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscription: service,
		logger:       logger,
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

	err := s.subscription.Subscribe(c.Request.Context(), subscription)
	if err != nil {
<<<<<<< HEAD
		switch {
		case errors.Is(err, svc.ErrorAlreadyExists):
=======
		s.logger.ConsoleLogError("can't create subscription",
			zap.String("error", err.Error()))
		if errors.Is(err, svc.ErrorAlreadyExists) {
>>>>>>> ec49c67 (refactoring + moving logic to service from handler)
			c.JSON(http.StatusConflict, "Email already subscribed")
		case errors.Is(err, svc.ErrorSubscriptionCreate):
			c.JSON(http.StatusInternalServerError, "Can't create subscription")
		case errors.Is(err, svc.ErrorSendEmail):
			c.JSON(http.StatusInternalServerError, "Can't send email")
		default:
			c.JSON(http.StatusInternalServerError, "Internal service error")
			s.logger.ConsoleLogError("Uncaught error, sending StatusInternalServerError",
				zap.String("error", err.Error()))
		}
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

	err = s.subscription.Confirm(c.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrorSubscriptionConfirm):
			c.JSON(http.StatusNotFound, "Can't confirm subscription")
		default:
			c.JSON(http.StatusInternalServerError, "Internal service error")
			s.logger.ConsoleLogError("Uncaught error, sending StatusInternalServerError",
				zap.String("error", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, "Subscription confirmed successfully")
}

func (s *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token, err := ValidateToken(c)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	err = s.subscription.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrorSubscriptionUnsubscribe):
			c.JSON(http.StatusNotFound, "Can't cancel subscription")
		default:
			c.JSON(http.StatusInternalServerError, "Internal service error")
			s.logger.ConsoleLogError("Uncaught error, sending StatusInternalServerError",
				zap.String("error", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, "Unsubscribed successfully")
}

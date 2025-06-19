// internal/mailer/manager_test.go
package mailer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"weather/internal/config"
	"weather/internal/mailer"
	mock_mailer "weather/internal/mailer/mocks"
	"weather/internal/models"
)

func TestManager_LoadTargets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_mailer.NewMockMailerStore(ctrl)
	subs := []models.Subscription{
		{Email: "a@x", Frequency: "daily"},
		{Email: "b@x", Frequency: "hourly"},
		{Email: "c@x", Frequency: "daily"},
	}
	store.
		EXPECT().
		GetSubscribed(gomock.Any()).
		Return(subs, nil)

	m := mailer.New(config.SMTPConfig{}, nil)
	err := m.LoadTargets(context.Background(), store)
	assert.NoError(t, err)

	daily := m.Targets.GetTargets(models.Daily)
	assert.Len(t, daily, 2)
	assert.ElementsMatch(t,
		[]string{"a@x", "c@x"},
		[]string{daily[0].Email, daily[1].Email},
	)

	hourly := m.Targets.GetTargets(models.Hourly)
	assert.Len(t, hourly, 1)
	assert.Equal(t, "b@x", hourly[0].Email)
}

func TestManager_LoadTargets_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_mailer.NewMockMailerStore(ctrl)
	store.
		EXPECT().
		GetSubscribed(gomock.Any()).
		Return(nil, errors.New("db down"))

	m := mailer.New(config.SMTPConfig{}, nil)
	err := m.LoadTargets(context.Background(), store)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load Mailer targets")
}

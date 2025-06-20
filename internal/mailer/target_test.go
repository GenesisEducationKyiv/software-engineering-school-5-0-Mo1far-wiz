package mailer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"weather/internal/mailer"
	mock_mailer "weather/internal/mailer/mocks"
	"weather/internal/models"
)

func TestLoadTargets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_mailer.NewMockTargetStore(ctrl)

	subs := []models.Subscription{
		{Email: "a@a.com", Frequency: "daily"},
		{Email: "b@b.com", Frequency: "hourly"},
		{Email: "c@c.com", Frequency: "daily"},
	}

	mockStore.
		EXPECT().
		GetSubscribed(gomock.Any()).
		Return(subs, nil)

	tm := &mailer.TargetManager{}
	err := tm.LoadTargets(context.Background(), mockStore)
	assert.NoError(t, err)

	daily := tm.GetTargets("daily")
	assert.Len(t, daily, 2)
	assert.ElementsMatch(t,
		[]string{"a@a.com", "c@c.com"},
		[]string{daily[0].Email, daily[1].Email},
	)

	hourly := tm.GetTargets("hourly")
	assert.Len(t, hourly, 1)
	assert.Equal(t, "b@b.com", hourly[0].Email)
}

func TestLoadTargets_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_mailer.NewMockTargetStore(ctrl)

	mockStore.
		EXPECT().
		GetSubscribed(gomock.Any()).
		Return(nil, errors.New("db down"))

	tm := &mailer.TargetManager{}
	err := tm.LoadTargets(context.Background(), mockStore)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load Mailer targets")
}

func TestAddTarget_NewAndDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_mailer.NewMockTargetStore(ctrl)

	mockStore.EXPECT().
		GetSubscribed(gomock.Any())

	tm := &mailer.TargetManager{}
	err := tm.LoadTargets(context.Background(), mockStore)
	assert.NoError(t, err)

	sub := models.Subscription{Email: "new@e", Frequency: "daily"}

	tm.AddTarget(sub)
	got := tm.GetTargets("daily")
	assert.Len(t, got, 1)
	assert.Equal(t, "new@e", got[0].Email)

	tm.AddTarget(sub)
	got2 := tm.GetTargets("daily")
	assert.Len(t, got2, 1)
}

func TestRemoveTarget_FoundAndNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_mailer.NewMockTargetStore(ctrl)

	initial := []models.Subscription{
		{Email: "a@a", Frequency: "daily"},
		{Email: "b@b", Frequency: "daily"},
	}

	mockStore.EXPECT().
		GetSubscribed(gomock.Any()).
		Return(initial, nil)

	tm := &mailer.TargetManager{}
	err := tm.LoadTargets(context.Background(), mockStore)
	assert.NoError(t, err)

	tm.RemoveTarget("a@a", "daily")
	remaining := tm.GetTargets("daily")
	assert.Len(t, remaining, 1)
	assert.Equal(t, "b@b", remaining[0].Email)

	tm.RemoveTarget("no@one", "daily")
	still := tm.GetTargets("daily")
	assert.Len(t, still, 1)
	assert.Equal(t, "b@b", still[0].Email)
}

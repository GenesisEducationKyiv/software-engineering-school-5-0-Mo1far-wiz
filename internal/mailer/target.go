package mailer

import (
	"context"
	"sync"
	"weather/internal/models"

	"github.com/pkg/errors"
)

type TargetStore interface {
	GetSubscribed(ctx context.Context) ([]models.Subscription, error)
}

type TargetManager struct {
	mx      sync.RWMutex
	targets map[string][]models.Subscription
}

func (m *TargetManager) LoadTargets(ctx context.Context, store TargetStore) error {
	subscriptions, err := store.GetSubscribed(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to load Mailer targets")
	}

	targets := make(map[string][]models.Subscription)

	for _, sub := range subscriptions {
		targets[sub.Frequency] = append(targets[sub.Frequency], sub)
	}

	m.targets = targets

	return nil
}

func (m *TargetManager) GetTargets(subscriptionType string) []models.Subscription {
	m.mx.RLock()
	defer m.mx.RUnlock()

	original := m.targets[subscriptionType]
	copied := make([]models.Subscription, len(original))
	copy(copied, original)
	return copied
}

func (m *TargetManager) AddDailyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets[models.Daily] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets[models.Daily] = append(m.targets[models.Daily], sub)
}

func (m *TargetManager) AddHourlyTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets[models.Hourly] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets[models.Hourly] = append(m.targets[models.Hourly], sub)
}

func (m *TargetManager) RemoveDailyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets[models.Daily]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets[models.Daily] = subs[:len(subs)-1]
			return
		}
	}
}

func (m *TargetManager) RemoveHourlyTarget(email string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets[models.Hourly]
	for i, sub := range subs {
		if sub.Email == email {
			subs[i] = subs[len(subs)-1]
			m.targets[models.Hourly] = subs[:len(subs)-1]
			return
		}
	}
}

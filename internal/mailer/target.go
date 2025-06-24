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

	m.mx.Lock()
	m.targets = targets
	m.mx.Unlock()

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

func (m *TargetManager) AddTarget(sub models.Subscription) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, existing := range m.targets[sub.Frequency] {
		if existing.Email == sub.Email {
			return
		}
	}
	m.targets[sub.Frequency] = append(m.targets[sub.Frequency], sub)
}

func (m *TargetManager) RemoveTarget(email string, frequency string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	subs := m.targets[frequency]
	filteredSubs := make([]models.Subscription, 0, len(subs))
	for _, sub := range subs {
		if sub.Email != email {
			filteredSubs = append(filteredSubs, sub)
		}
	}
	m.targets[frequency] = filteredSubs
}

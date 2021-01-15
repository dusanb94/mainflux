package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/consumers/notify"
)

var _ notify.SubscriptionsRepository = (*subRepoMock)(nil)

type subRepoMock struct {
	mu   sync.Mutex
	subs map[string]notify.Subscription
}

// New returns a new Subscriptions repository mock.
func New(subs map[string]notify.Subscription) notify.SubscriptionsRepository {
	return subRepoMock{
		subs: subs,
	}
}

func (srm *subRepoMock) Save(ctx context.Context, sub notify.Subscription) (string, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	if s, ok := srm.subs[sub.ID]; ok || (s.Contact == sub.Contact && s.Topic == sub.Topic) {
		return "", notify.ErrConflict
	}
	return sub.ID, nil
}

func (srm *subRepoMock) Retrieve(ctx context.Context, id string) (notify.Subscription, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	ret, ok := srm.subs[id]
	if !ok {
		return notify.Subscription{}, notify.ErrNotFound
	}
	return ret, nil
}

func (srm *subRepoMock) RetrieveAll(ctx context.Context, topic, contact string) ([]notify.Subscription, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	var ret []notify.Subscription
	for _, v := range srm.subs {
		if topic == "" {
			if contact == "" {
				ret = append(ret, v)
				continue
			}
			if contact == v.Contact {
				ret = append(ret, v)
				continue
			}
		}
		if topic == v.Topic {
			if contact == "" || contact == v.Contact {
				ret = append(ret, v)
			}
		}
	}

	if len(ret) == 0 {
		return ret, notify.ErrNotFound
	}

	return ret, nil
}

func (srm *subRepoMock) Remove(ctx context.Context, id string) error {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	delete(srm.subs, id)
	return nil
}

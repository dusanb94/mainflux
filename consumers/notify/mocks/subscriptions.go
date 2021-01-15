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

// NewRepo returns a new Subscriptions repository mock.
func NewRepo(subs map[string]notify.Subscription) notify.SubscriptionsRepository {
	return &subRepoMock{
		subs: subs,
	}
}

func (srm *subRepoMock) Save(_ context.Context, sub notify.Subscription) (string, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	if _, ok := srm.subs[sub.ID]; ok {
		return "", notify.ErrConflict
	}
	for _, s := range srm.subs {
		if s.Contact == sub.Contact && s.Topic == sub.Topic {
			return "", notify.ErrConflict
		}
	}

	srm.subs[sub.ID] = sub
	return sub.ID, nil
}

func (srm *subRepoMock) Retrieve(_ context.Context, id string) (notify.Subscription, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	ret, ok := srm.subs[id]
	if !ok {
		return notify.Subscription{}, notify.ErrNotFound
	}
	return ret, nil
}

func (srm *subRepoMock) RetrieveAll(_ context.Context, pm notify.PageMetadata) ([]notify.Subscription, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	var ret []notify.Subscription
	var ind int
	offset := int(pm.Offset)
	for _, v := range srm.subs {
		if pm.Topic == "" {
			if pm.Contact == "" {
				if ind < offset {
					ind++
					continue
				}
				ret = append(ret, v)
				continue
			}
			if pm.Contact == v.Contact {
				if ind < offset {
					ind++
					continue
				}
				ret = append(ret, v)
				continue
			}
		}
		if pm.Topic == v.Topic {
			if pm.Contact == "" || pm.Contact == v.Contact {
				if ind < offset {
					ind++
					continue
				}
				ret = append(ret, v)
			}
		}
		if len(ret) == int(pm.Limit) {
			break
		}
	}

	if len(ret) == 0 {
		return ret, notify.ErrNotFound
	}

	return ret, nil
}

func (srm *subRepoMock) Remove(_ context.Context, id string) error {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	delete(srm.subs, id)
	return nil
}

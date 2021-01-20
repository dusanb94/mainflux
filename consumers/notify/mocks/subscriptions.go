// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sort"
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

func (srm *subRepoMock) RetrieveAll(_ context.Context, pm notify.PageMetadata) (notify.Page, error) {
	srm.mu.Lock()
	defer srm.mu.Unlock()

	// Sort keys
	keys := make([]string, 0)
	for k := range srm.subs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var subs []notify.Subscription
	var total int
	offset := int(pm.Offset)
	for _, k := range keys {
		v := srm.subs[k]
		if pm.Topic == "" {
			if pm.Contact == "" {
				if total < offset {
					total++
					continue
				}
				total++
				subs = appendSubs(subs, v, pm.Limit)
				continue
			}
			if pm.Contact == v.Contact {
				if total < offset {
					total++
					continue
				}
				total++
				subs = appendSubs(subs, v, pm.Limit)
				continue
			}
		}
		if pm.Topic == v.Topic {
			if pm.Contact == "" || pm.Contact == v.Contact {
				if total < offset {
					total++
					continue
				}
				total++
				subs = appendSubs(subs, v, pm.Limit)
			}
		}
	}

	if len(subs) == 0 {
		return notify.Page{}, notify.ErrNotFound
	}

	ret := notify.Page{
		PageMetadata:  pm,
		Total:         uint(total),
		Subscriptions: subs,
	}

	return ret, nil
}

func appendSubs(subs []notify.Subscription, sub notify.Subscription, max int) []notify.Subscription {
	if len(subs) < max || max == -1 {
		subs = append(subs, sub)
	}
	return subs
}

func (srm *subRepoMock) Remove(_ context.Context, id string) error {
	srm.mu.Lock()
	defer srm.mu.Unlock()
	delete(srm.subs, id)
	return nil
}

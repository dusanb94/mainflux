// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/consumers/notify/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)

	id1, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	id2, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	sub1 := notify.Subscription{
		OwnerID: "owner",
		ID:      id1,
		Contact: "ownersave@example.com",
		Topic:   "topic.subtopic",
	}

	sub2 := sub1
	sub2.ID = id2

	cases := map[string]struct {
		sub notify.Subscription
		id  string
		err error
	}{
		"save successfully": {
			sub: sub1,
			id:  id1,
			err: nil,
		},
		"save duplicate": {
			sub: sub2,
			id:  "",
			err: notify.ErrConflict,
		},
	}

	for desc, tc := range cases {
		id, err := repo.Save(context.Background(), tc.sub)
		assert.Equal(t, tc.id, id, fmt.Sprintf("%s: expected id %s got %s\n", desc, tc.id, id))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))

	}
}

func TestView(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)

	id, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got an error creating id: %s", err))

	sub := notify.Subscription{
		OwnerID: "owner",
		ID:      id,
		Contact: "ownerview@example.com",
		Topic:   "topic.subtopic",
	}

	ret, err := repo.Save(context.Background(), sub)
	require.Nil(t, err, fmt.Sprintf("creating subscription must not fail: %s", err))
	require.Equal(t, id, ret, fmt.Sprintf("provided id %s must be the same as the returned id %s", id, ret))

	cases := map[string]struct {
		sub notify.Subscription
		id  string
		err error
	}{
		"retrieve successfully": {
			sub: sub,
			id:  id,
			err: nil,
		},
		"retrieve not existing": {
			sub: notify.Subscription{},
			id:  "non-existing",
			err: notify.ErrNotFound,
		},
	}

	for desc, tc := range cases {
		sub, err := repo.Retrieve(context.Background(), tc.id)
		assert.Equal(t, tc.sub, sub, fmt.Sprintf("%s: expected sub %v got %v\n", desc, tc.sub, sub))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))

	}
}

func TestRetrieveAll(t *testing.T) {
	_, err := db.Exec("DELETE FROM subscriptions")
	require.Nil(t, err, fmt.Sprintf("cleanup must not fail: %s", err))

	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)

	const numSubs = 100

	var subs []notify.Subscription

	for i := 0; i < numSubs; i++ {
		id, err := idProvider.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		sub := notify.Subscription{
			OwnerID: "owner",
			ID:      id,
			Contact: "ownerlist@example.com",
			Topic:   fmt.Sprintf("topic.subtopic.%d", i),
		}

		ret, err := repo.Save(context.Background(), sub)
		require.Nil(t, err, fmt.Sprintf("creating subscription must not fail: %s", err))
		require.Equal(t, id, ret, fmt.Sprintf("provided id %s must be the same as the returned id %s", id, ret))
		subs = append(subs, sub)
	}

	cases := map[string]struct {
		pageMeta notify.PageMetadata
		page     notify.Page
		err      error
	}{
		"retrieve successfully": {
			pageMeta: notify.PageMetadata{
				Offset: 10,
				Limit:  2,
			},
			page: notify.Page{
				Total: numSubs,
				PageMetadata: notify.PageMetadata{
					Offset: 10,
					Limit:  2,
				},
				Subscriptions: subs[10:12],
			},
			err: nil,
		},
		"retrieve with contact": {
			pageMeta: notify.PageMetadata{
				Offset:  10,
				Limit:   2,
				Contact: "ownerlist@example.com",
			},
			page: notify.Page{
				Total: numSubs,
				PageMetadata: notify.PageMetadata{
					Offset:  10,
					Limit:   2,
					Contact: "ownerlist@example.com",
				},
				Subscriptions: subs[10:12],
			},
			err: nil,
		},
		"retrieve with topic": {
			pageMeta: notify.PageMetadata{
				Offset: 0,
				Limit:  2,
				Topic:  "topic.subtopic.11",
			},
			page: notify.Page{
				Total: 1,
				PageMetadata: notify.PageMetadata{
					Offset: 0,
					Limit:  2,
					Topic:  "topic.subtopic.11",
				},
				Subscriptions: subs[11:12],
			},
			err: nil,
		},
		"retrieve with no limit": {
			pageMeta: notify.PageMetadata{
				Offset: 0,
				Limit:  -1,
			},
			page: notify.Page{
				Total: numSubs,
				PageMetadata: notify.PageMetadata{
					Limit: -1,
				},
				Subscriptions: subs,
			},
			err: nil,
		},
	}

	for desc, tc := range cases {
		page, err := repo.RetrieveAll(context.Background(), tc.pageMeta)
		assert.Equal(t, tc.page, page, fmt.Sprintf("%s: expected page %v got %v\n", desc, tc.page, page))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestRemove(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)
	id, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got an error creating id: %s", err))
	sub := notify.Subscription{
		OwnerID: "owner",
		ID:      id,
		Contact: "ownerremove@example.com",
		Topic:   "topic.subtopic.%d",
	}

	ret, err := repo.Save(context.Background(), sub)
	require.Nil(t, err, fmt.Sprintf("creating subscription must not fail: %s", err))
	require.Equal(t, id, ret, fmt.Sprintf("provided id %s must be the same as the returned id %s", id, ret))

	cases := map[string]struct {
		id  string
		err error
	}{
		"remove successfully": {
			id:  id,
			err: nil,
		},
		"remove not existing": {
			id:  "empty",
			err: nil,
		},
	}

	for desc, tc := range cases {
		err := repo.Remove(context.Background(), tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

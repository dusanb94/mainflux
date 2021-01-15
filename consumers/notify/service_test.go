package notify_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/consumers/notify/mocks"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleUser1 = "email1@example.com"
	exampleUser2 = "email2@example.com"
)

func newService() notify.Service {
	repo := mocks.NewRepo(make(map[string]notify.Subscription))
	auth := mocks.NewAuth(map[string]string{exampleUser1: exampleUser1})
	notifier := mocks.NewNotifier()
	idp := uuid.NewMock()
	return notify.New(auth, repo, idp, notifier)
}

func TestCreateSubscription(t *testing.T) {
	svc := newService()

	cases := map[string]struct {
		token string
		sub   notify.Subscription
		id    string
		err   error
	}{
		"test success": {
			token: exampleUser1,
			sub:   notify.Subscription{Contact: exampleUser1, Topic: "valid.topic"},
			id:    uuid.Prefix + fmt.Sprintf("%012d", 1),
			err:   nil,
		},
		"test already existing": {
			token: exampleUser1,
			sub:   notify.Subscription{Contact: exampleUser1, Topic: "valid.topic"},
			id:    "",
			err:   notify.ErrConflict,
		},
		"test unauthorized access": {
			token: "",
			sub:   notify.Subscription{Contact: exampleUser1, Topic: "valid.topic"},
			id:    "",
			err:   notify.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		id, err := svc.CreateSubscription(context.Background(), tc.token, tc.sub)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		assert.Equal(t, tc.id, id, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.id, id))
	}
}

func TestViewSubscription(t *testing.T) {
	svc := newService()
	sub := notify.Subscription{Contact: exampleUser1, Topic: "valid.topic"}
	id, err := svc.CreateSubscription(context.Background(), exampleUser1, sub)
	require.Nil(t, err, fmt.Sprintf("Saving a Subscription must succeed"))
	sub.ID = id
	sub.OwnerID = exampleUser1

	cases := map[string]struct {
		token string
		id    string
		sub   notify.Subscription
		err   error
	}{
		"test success": {
			token: exampleUser1,
			id:    id,
			sub:   sub,
			err:   nil,
		},
		"test not existing": {
			token: exampleUser1,
			id:    "not_exist",
			sub:   notify.Subscription{},
			err:   notify.ErrNotFound,
		},
		"test unauthorized access": {
			token: "",
			id:    id,
			sub:   notify.Subscription{},
			err:   notify.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		sub, err := svc.ViewSubscription(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		assert.Equal(t, tc.sub, sub, fmt.Sprintf("%s: expected %v got %v\n", desc, tc.sub, sub))
	}
}

func TestListSubscriptions(t *testing.T) {
	svc := newService()
	sub := notify.Subscription{Contact: exampleUser1}
	topic := "topic.subtopic"
	var subs []notify.Subscription
	for i := 0; i < 100; i++ {
		token := exampleUser1
		if i%2 == 0 {
			token = exampleUser2
		}
		sub.Topic = fmt.Sprintf("%s.%s", topic, strconv.Itoa(i))
		id, err := svc.CreateSubscription(context.Background(), token, sub)
		require.Nil(t, err, fmt.Sprintf("Saving a Subscription must succeed"))
		sub.ID = id
		subs = append(subs, sub)
	}

	cases := map[string]struct {
		token   string
		id      string
		topic   string
		contact string
		subs    []notify.Subscription
		err     error
	}{
		"test success": {
			token: exampleUser1,
			id:    id,
			sub:   sub,
			err:   nil,
		},
		"test not existing": {
			token: exampleUser1,
			id:    "not_exist",
			sub:   notify.Subscription{},
			err:   notify.ErrNotFound,
		},
		"test unauthorized access": {
			token: "",
			id:    id,
			sub:   notify.Subscription{},
			err:   notify.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		sub, err := svc.ListSubscriptions(context.Background(), tc.token, tc.topic, tc.contact)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		assert.Equal(t, tc.sub, sub, fmt.Sprintf("%s: expected %v got %v\n", desc, tc.sub, sub))
	}
}

func TestRemoveSubscription(t *testing.T) {
	svc := newService()
	sub := notify.Subscription{Contact: exampleUser1, Topic: "valid.topic"}
	id, err := svc.CreateSubscription(context.Background(), exampleUser1, sub)
	require.Nil(t, err, fmt.Sprintf("Saving a Subscription must succeed"))
	sub.ID = id
	sub.OwnerID = exampleUser1

	cases := map[string]struct {
		token string
		id    string
		sub   notify.Subscription
		err   error
	}{
		"test success": {
			token: exampleUser1,
			id:    id,
			sub:   sub,
			err:   nil,
		},
		"test not existing": {
			token: exampleUser1,
			id:    "not_exist",
			sub:   notify.Subscription{},
			err:   notify.ErrNotFound,
		},
		"test unauthorized access": {
			token: "",
			id:    id,
			sub:   notify.Subscription{},
			err:   notify.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		sub, err := svc.RemoveSubscription(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		assert.Equal(t, tc.sub, sub, fmt.Sprintf("%s: expected %v got %v\n", desc, tc.sub, sub))
	}
}

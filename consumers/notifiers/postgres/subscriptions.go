// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
	errSaveUserDB       = errors.New("Save user to DB failed")
	errUpdateDB         = errors.New("Update user email to DB failed")
	errUpdateUserDB     = errors.New("Update user metadata to DB failed")
	errRetrieveDB       = errors.New("Retreiving from DB failed")
	errUpdatePasswordDB = errors.New("Update password to DB failed")
	errMarshal          = errors.New("Failed to marshal metadata")
	errUnmarshal        = errors.New("Failed to unmarshal metadata")
)

var _ notifiers.SubscriptionRepository = (*subscriptionsRepo)(nil)

const errDuplicate = "unique_violation"

type subscriptionsRepo struct {
	db Database
}

// NewUserRepo instantiates a PostgreSQL implementation of user
// repository.
func NewUserRepo(db Database) notifiers.SubscriptionRepository {
	return &subscriptionsRepo{
		db: db,
	}
}

func (repo subscriptionsRepo) Save(ctx context.Context, sub notifiers.Subscription) (string, error) {
	if sub.ID == "" || sub.OwnerID == "" || sub.OwnerEmail == "" {
		return "", users.ErrMalformedEntity
	}
	q := `INSERT INTO subscriptions (id, owner_id, owner_email, topic) VALUES (:id, :owner_id, :owner_email, :topic) RETURNING id`

	dbSub := toDBSub(sub)

	if _, err := repo.db.NamedQueryContext(ctx, q, dbSub); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == errDuplicate {
			return "", errors.Wrap(users.ErrConflict, err)
		}
		return "", errors.Wrap(errSaveUserDB, err)
	}

	return sub.ID, nil
}

func (subscriptionsRepo) Retrieve(ctx context.Context, ownerID, topic string) (notifiers.Subscription, error) {
	return notifiers.Subscription{}, nil
}

func (subscriptionsRepo) RetrieveAll(ctx context.Context, topic string) ([]notifiers.Subscription, error) {
	return []notifiers.Subscription{}, nil
}

func (subscriptionsRepo) Remove(ctx context.Context, ownerID, id string) error {
	return nil
}

type dbSubscription struct {
	ID         string `db:"id"`
	OwnerID    string `db:"owner_id"`
	OwnerEmail string `db:"owner_email"`
	Topic      string `db:"topic"`
}

func toDBSub(sub notifiers.Subscription) dbSubscription {
	return dbSubscription{
		ID:         sub.ID,
		OwnerID:    sub.OwnerID,
		OwnerEmail: sub.OwnerEmail,
		Topic:      sub.Topic,
	}
}

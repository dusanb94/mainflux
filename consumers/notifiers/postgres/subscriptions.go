// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things"
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

var _ notifiers.SubscriptionsRepository = (*subscriptionsRepo)(nil)

const errDuplicate = "unique_violation"

type subscriptionsRepo struct {
	db Database
}

// New instantiates a PostgreSQL implementation of Subscriptions repository.
func New(db Database) notifiers.SubscriptionsRepository {
	return &subscriptionsRepo{
		db: db,
	}
}

func (repo subscriptionsRepo) Save(ctx context.Context, sub notifiers.Subscription) (string, error) {
	if sub.ID == "" || sub.OwnerID == "" || sub.OwnerEmail == "" {
		return "", users.ErrMalformedEntity
	}
	q := `INSERT INTO subscriptions (id, owner_id, owner_email, topic) VALUES (:id, :owner_id, :owner_email, :topic) RETURNING id`

	dbSub := dbSubscription{
		ID:         sub.ID,
		OwnerID:    sub.OwnerID,
		OwnerEmail: sub.OwnerEmail,
		Topic:      sub.Topic,
	}

	if _, err := repo.db.NamedQueryContext(ctx, q, dbSub); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == errDuplicate {
			return "", errors.Wrap(users.ErrConflict, err)
		}
		return "", errors.Wrap(errSaveUserDB, err)
	}

	return sub.ID, nil
}

func (repo subscriptionsRepo) Retrieve(ctx context.Context, ownerID, topic string) (notifiers.Subscription, error) {
	q := `SELECT id, owner_id, owner_email, topic subscriptions WHERE owner_id = $1 AND topic = $2`
	sub := dbSubscription{}
	if err := repo.db.QueryRowxContext(ctx, q, ownerID, topic).StructScan(&sub); err != nil {
		if err == sql.ErrNoRows {
			return notifiers.Subscription{}, errors.Wrap(users.ErrNotFound, err)

		}
		return notifiers.Subscription{}, errors.Wrap(errRetrieveDB, err)
	}

	return fromDBSub(sub), nil
}

func (repo subscriptionsRepo) RetrieveAll(ctx context.Context, topic string) ([]notifiers.Subscription, error) {
	q := `SELECT id, owner_id, owner_email, topic subscriptions WHERE topic = $1`

	rows, err := repo.db.NamedQueryContext(ctx, q, topic)
	if err != nil {
		return []notifiers.Subscription{}, errors.Wrap(things.ErrSelectEntity, err)
	}
	defer rows.Close()

	ret := []notifiers.Subscription{}
	for rows.Next() {
		sub := dbSubscription{}
		if err := rows.StructScan(&sub); err != nil {
			return []notifiers.Subscription{}, errors.Wrap(things.ErrSelectEntity, err)
		}
		ret = append(ret, fromDBSub(sub))
	}

	return ret, nil
}

func (repo subscriptionsRepo) Remove(ctx context.Context, id string) error {
	q := `DELETE from subscriptions WHERE id = $1`

	if r := repo.db.QueryRowxContext(ctx, q, id); r.Err() != nil {
		return errors.Wrap(things.ErrRemoveEntity, r.Err())
	}
	return nil
}

type dbSubscription struct {
	ID         string `db:"id"`
	OwnerID    string `db:"owner_id"`
	OwnerEmail string `db:"owner_email"`
	Topic      string `db:"topic"`
}

func fromDBSub(sub dbSubscription) notifiers.Subscription {
	return notifiers.Subscription{
		ID:         sub.ID,
		OwnerID:    sub.OwnerID,
		OwnerEmail: sub.OwnerEmail,
		Topic:      sub.Topic,
	}
}

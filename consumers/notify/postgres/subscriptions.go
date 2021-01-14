// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things"
)

var (
	errSaveSub    = errors.New("Save sub to DB failed")
	errRetrieveDB = errors.New("Retreiving from DB failed")
)

var _ notify.SubscriptionsRepository = (*subscriptionsRepo)(nil)

const errDuplicate = "unique_violation"

type subscriptionsRepo struct {
	db Database
}

// New instantiates a PostgreSQL implementation of Subscriptions repository.
func New(db Database) notify.SubscriptionsRepository {
	return &subscriptionsRepo{
		db: db,
	}
}

func (repo subscriptionsRepo) Save(ctx context.Context, sub notify.Subscription) (string, error) {
	q := `INSERT INTO subscriptions (id, owner_id, contact, topic) VALUES (:id, :owner_id, :contact, :topic) RETURNING id`

	dbSub := dbSubscription{
		ID:      sub.ID,
		OwnerID: sub.OwnerID,
		Contact: sub.Contact,
		Topic:   sub.Topic,
	}

	if _, err := repo.db.NamedQueryContext(ctx, q, dbSub); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == errDuplicate {
			return "", errors.Wrap(notify.ErrConflict, err)
		}
		return "", errors.Wrap(errSaveSub, err)
	}

	return sub.ID, nil
}

func (repo subscriptionsRepo) Retrieve(ctx context.Context, ownerID, topic string) (notify.Subscription, error) {
	q := `SELECT id, owner_id, contact, topic subscriptions WHERE owner_id = $1 AND topic = $2`
	sub := dbSubscription{}
	if err := repo.db.QueryRowxContext(ctx, q, ownerID, topic).StructScan(&sub); err != nil {
		if err == sql.ErrNoRows {
			return notify.Subscription{}, errors.Wrap(notify.ErrNotFound, err)

		}
		return notify.Subscription{}, errors.Wrap(errRetrieveDB, err)
	}

	return fromDBSub(sub), nil
}

func (repo subscriptionsRepo) RetrieveAll(ctx context.Context, topic string) ([]notify.Subscription, error) {
	q := `SELECT id, owner_id, contact, topic FROM subscriptions WHERE topic = :topic`
	args := map[string]interface{}{"topic": topic}
	rows, err := repo.db.NamedQueryContext(ctx, q, args)
	if err != nil {
		return []notify.Subscription{}, errors.Wrap(things.ErrSelectEntity, err)
	}
	defer rows.Close()

	ret := []notify.Subscription{}
	for rows.Next() {
		sub := dbSubscription{}
		if err := rows.StructScan(&sub); err != nil {
			return []notify.Subscription{}, errors.Wrap(notify.ErrSelectEntity, err)
		}
		ret = append(ret, fromDBSub(sub))
	}

	return ret, nil
}

func (repo subscriptionsRepo) Remove(ctx context.Context, id, topic string) error {
	q := `DELETE from subscriptions WHERE owner_id = $1 AND topic = $2`

	if r := repo.db.QueryRowxContext(ctx, q, id); r.Err() != nil {
		return errors.Wrap(notify.ErrRemoveEntity, r.Err())
	}
	return nil
}

type dbSubscription struct {
	ID      string `db:"id"`
	OwnerID string `db:"owner_id"`
	Contact string `db:"contact"`
	Topic   string `db:"topic"`
}

func fromDBSub(sub dbSubscription) notify.Subscription {
	return notify.Subscription{
		ID:      sub.ID,
		OwnerID: sub.OwnerID,
		Contact: sub.Contact,
		Topic:   sub.Topic,
	}
}

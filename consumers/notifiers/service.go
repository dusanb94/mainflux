// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package notifiers

import (
	"context"

	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/mainflux/mainflux"
)

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrCreateUUID indicates error in creating uuid for entity creation
	ErrCreateUUID = errors.New("uuid creation failed")

	// ErrCreateEntity indicates error in creating entity or entities
	ErrCreateEntity = errors.New("create entity failed")

	// ErrViewEntity indicates error in viewing entity or entities
	ErrViewEntity = errors.New("view entity failed")

	// ErrRemoveEntity indicates error in removing entity
	ErrRemoveEntity = errors.New("remove entity failed")
)

// Service reprents a notification service.
type Service interface {
	// CreateSubscription persists a subscription.
	// Successful operation is indicated by non-nil error response.
	CreateSubscription(ctx context.Context, token string, sub Subscription) (string, error)

	// ViewSubscription retrieves the subscription for the given owner and topic.
	ViewSubscription(ctx context.Context, token, topic string) (Subscription, error)

	// ListSubscriptions removes the subscription having the provided identifier, that is owned
	// by the specified user.
	ListSubscriptions(ctx context.Context, token, topic string) ([]Subscription, error)

	// RemoveSubscription removes the subscription having the provided identifier.
	RemoveSubscription(ctx context.Context, token, id string) error

	consumers.Consumer
}

var _ Service = (*notifierService)(nil)

type notifierService struct {
	auth     mainflux.AuthServiceClient
	subs     SubscriptionsRepository
	idp      mainflux.IDProvider
	consumer consumers.Consumer
}

// New instantiates the things service implementation.
func New(auth mainflux.AuthServiceClient, subs SubscriptionsRepository, idp mainflux.IDProvider, consumer consumers.Consumer) Service {
	return &notifierService{
		auth:     auth,
		subs:     subs,
		idp:      idp,
		consumer: consumer,
	}
}

func (ns *notifierService) CreateSubscription(ctx context.Context, token string, sub Subscription) (string, error) {
	res, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}

	sub.ID, err = ns.idp.ID()
	if err != nil {
		return "", errors.Wrap(ErrCreateUUID, err)
	}

	sub.OwnerEmail = res.GetEmail()
	sub.OwnerID = res.GetId()

	return ns.subs.Save(ctx, sub)
}

func (ns *notifierService) ViewSubscription(ctx context.Context, token, topic string) (Subscription, error) {
	res, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Subscription{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ns.subs.Retrieve(ctx, res.GetId(), topic)
}

func (ns *notifierService) ListSubscriptions(ctx context.Context, token, topic string) ([]Subscription, error) {
	_, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return nil, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ns.subs.RetrieveAll(ctx, topic)
}

func (ns *notifierService) RemoveSubscription(ctx context.Context, token, id string) error {
	_, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ns.subs.Remove(ctx, id)
}

func (ns *notifierService) Consume(message interface{}) error {
	return ns.Consume(message)
}

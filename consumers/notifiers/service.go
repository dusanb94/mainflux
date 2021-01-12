// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package notifiers

import (
	"context"
	"fmt"

	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"

	"github.com/mainflux/mainflux"
)

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrCreateID indicates error in creating id for entity creation
	ErrCreateID = errors.New("failed to create id")

	// ErrMessage indicates an error converting a message to Mainflux message.
	ErrMessage = errors.New("failed to convert to Mainflux message")
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
	RemoveSubscription(ctx context.Context, token, ownerID, topic string) error

	consumers.Consumer
}

var _ Service = (*notifierService)(nil)

type notifierService struct {
	auth     mainflux.AuthServiceClient
	subs     SubscriptionsRepository
	idp      mainflux.IDProvider
	notifier Notifier
}

// New instantiates the things service implementation.
func New(auth mainflux.AuthServiceClient, subs SubscriptionsRepository, idp mainflux.IDProvider, notifier Notifier) Service {
	return &notifierService{
		auth:     auth,
		subs:     subs,
		idp:      idp,
		notifier: notifier,
	}
}

func (ns *notifierService) CreateSubscription(ctx context.Context, token string, sub Subscription) (string, error) {
	res, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}

	sub.ID, err = ns.idp.ID()
	if err != nil {
		return "", errors.Wrap(ErrCreateID, err)
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

func (ns *notifierService) RemoveSubscription(ctx context.Context, token, ownerID, topic string) error {
	_, err := ns.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ns.subs.Remove(ctx, ownerID, topic)
}

func (ns *notifierService) Consume(message interface{}) error {
	msg, ok := message.(messaging.Message)
	if !ok {
		return ErrMessage
	}
	topic := msg.Channel
	if msg.Subtopic != "" {
		topic = fmt.Sprintf("%s.%s", msg.Channel, msg.Subtopic)
	}

	subs, err := ns.subs.RetrieveAll(context.Background(), topic)
	if err != nil {
		return err
	}

	var to []string
	for _, sub := range subs {
		to = append(to, sub.OwnerEmail)
	}
	if len(to) > 0 {
		return ns.notifier.Notify("", to, msg)
	}

	return nil
}

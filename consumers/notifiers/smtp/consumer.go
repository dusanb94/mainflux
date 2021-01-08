// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smtp

import (
	"context"

	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
)

var errMessage = errors.New("failed to convert to Mainflux message")
var _ consumers.Consumer = (*emailer)(nil)

type emailer struct {
	agent *email.Agent
	repo  notifiers.SubscriptionsRepository
}

// New instantiates Cassandra message repository.
func New(agent *email.Agent) consumers.Consumer {
	return &emailer{agent: agent}
}

func (c emailer) Consume(message interface{}) error {
	msg, ok := message.(messaging.Message)
	if !ok {
		return errMessage
	}
	topic := msg.Channel
	if msg.Subtopic != "" {
		topic = topic + "." + msg.Subtopic
	}
	subs, err := c.repo.RetrieveAll(context.Background(), topic)
	if err != nil {
		return err
	}
	var emails []string
	for _, sub := range subs {
		emails = append(emails, sub.OwnerEmail)
	}
	return c.agent.Send(emails, "", "", "Password reset", string(msg.Payload), "")
}

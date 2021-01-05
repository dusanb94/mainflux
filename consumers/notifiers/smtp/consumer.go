// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smtp

import (
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errSaveMessage = errors.New("failed to save message to cassandra database")
	errNoTable     = errors.New("table does not exist")
)
var _ consumers.Consumer = (*emailer)(nil)

type emailer struct {
	agent *email.Agent
	repo  notifiers.SubscriptionRepository
}

// New instantiates Cassandra message repository.
func New(agent *email.Agent) consumers.Consumer {
	return &emailer{agent: agent}
}

func (c emailer) Consume(message interface{}) error {
	return c.agent.Send(nil, "", "Password reset", "", "", "")
}

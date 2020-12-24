// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smtp

import (
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errSaveMessage = errors.New("failed to save message to cassandra database")
	errNoTable     = errors.New("table does not exist")
)
var _ consumers.MessageConsumer = (*emailer)(nil)

type emailer struct {
	agent *email.Agent
}

// New instantiates Cassandra message repository.
func New(agent *email.Agent) consumers.MessageConsumer {
	return &emailer{agent: agent}
}

func (c emailer) Consume(message interface{}) error {
	return c.agent.Send(nil, "", "Password reset", "", "", "")
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smtp

import (
	"fmt"

	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	footer          = "Sent by Mainflux SMTP Notification"
	contentTemplate = `A publisher with an id %s sent the message over %s with the following values\n %s`
)

var errMessage = errors.New("failed to convert to Mainflux message")
var _ notifiers.Notifier = (*emailer)(nil)

type emailer struct {
	agent *email.Agent
}

// New instantiates SMTP message notifier.
func New(agent *email.Agent) notifiers.Notifier {
	return &emailer{agent: agent}
}

func (c *emailer) Notify(from string, to []string, msg messaging.Message) error {
	header := msg.Channel
	subject := fmt.Sprintf(`Channel %s`, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s and subtopic %s", subject, msg.Subtopic)
	}

	values := string(msg.Payload)

	content := fmt.Sprintf(contentTemplate, msg.Publisher, msg.Protocol, values)

	return c.agent.Send(to, "", subject, header, content, footer)
}

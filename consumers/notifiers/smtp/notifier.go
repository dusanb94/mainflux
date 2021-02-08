// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smtp

import (
	"fmt"

	notifiers "github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers"
	"github.com/mainflux/mainflux/pkg/transformers/json"
)

const (
	footer          = "Sent by Mainflux SMTP Notification"
	contentTemplate = "A publisher with an id %s sent the message over %s with the following values \n %s"
)

var _ notifiers.Notifier = (*notifier)(nil)

var fields = [...]string{"s_leakage", "s_blocked", "s_magnet", "s_blowout", "ALM", "magnet"}

type notifier struct {
	agent *email.Agent
	tr    transformers.Transformer
}

// New instantiates SMTP message notifier.
func New(agent *email.Agent) notifiers.Notifier {
	return &notifier{agent: agent, tr: json.New()}
}

func (n *notifier) Notify(from string, to []string, msg messaging.Message) error {
	m, err := json.New().Transform(msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf(`Notification for Channel %s`, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s and subtopic %s", subject, msg.Subtopic)
	}

	values := string(msg.Payload)
	content := fmt.Sprintf(contentTemplate, msg.Publisher, msg.Protocol, values)
	switch t := m.(type) {
	case map[string]interface{}:
		for _, k := range fields {
			if v, ok := t[k]; v != nil && ok {
				if val, ok := v.(int); ok && val != 0 {
					return n.agent.Send(to, from, subject, "", content, footer)
				}
			}
		}
	case []map[string]interface{}:
		for _, v := range t {
			for _, k := range fields {
				if v, ok := v[k]; v != nil && ok {
					if val, ok := v.(int); ok && val != 0 {
						return n.agent.Send(to, from, subject, "", content, footer)
					}
				}
			}
		}
	}

	return nil
}

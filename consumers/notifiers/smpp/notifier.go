// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smpp

import (
	"fmt"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/mainflux/mainflux/consumers/notifiers"
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
	t             *smpp.Transmitter
	tr            transformers.Transformer
	sourceAddrTON uint8
	sourceAddrNPI uint8
	destAddrTON   uint8
	destAddrNPI   uint8
}

// New instantiates SMTP message notifier.
func New(cfg Config) notifiers.Notifier {
	t := &smpp.Transmitter{
		Addr:       cfg.Address,
		User:       cfg.Username,
		Passwd:     cfg.Password,
		SystemType: cfg.SystemType,
	}
	ret := &notifier{
		t:             t,
		tr:            json.New(),
		sourceAddrTON: cfg.SourceAddrTON,
		destAddrTON:   cfg.DestAddrTON,
		sourceAddrNPI: cfg.SourceAddrNPI,
		destAddrNPI:   cfg.DestAddrNPI,
	}
	return ret
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

	send := &smpp.ShortMessage{
		Src:           from,
		DstList:       to,
		SourceAddrTON: n.sourceAddrTON,
		DestAddrTON:   n.destAddrTON,
		SourceAddrNPI: n.sourceAddrNPI,
		DestAddrNPI:   n.destAddrNPI,
		Text:          pdutext.Raw(msg.Payload),
		Register:      pdufield.FailureDeliveryReceipt,
	}

	switch t := m.(type) {
	case map[string]interface{}:
		for _, k := range fields {
			if v, ok := t[k]; v != nil && ok {
				if val, ok := v.(int); ok && val != 0 {
					_, err := n.t.Submit(send)
					return err
				}
			}
		}
	case []map[string]interface{}:
		for _, v := range t {
			for _, k := range fields {
				if v, ok := v[k]; v != nil && ok {
					if val, ok := v.(int); ok && val != 0 {
						_, err := n.t.Submit(send)
						return err
					}
				}
			}
		}
	}

	return nil
}

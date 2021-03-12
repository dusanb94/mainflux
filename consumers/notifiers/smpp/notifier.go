// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smpp

import (
	"errors"
	"fmt"
	"time"

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

var fields = [...]string{"d/s_leakage", "d/s_blocked", "d/s_magnet", "d/s_blowout", "ALM", "magnet"}
var errMessageType = errors.New("error message type")

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
	t.Bind()
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
	jm, err := json.New().Transform(msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf(`Notification for Channel %s`, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s and subtopic %s", subject, msg.Subtopic)
	}

	jsonMsg, ok := jm.(json.Messages)
	if !ok {
		return errMessageType
	}

	send := &smpp.ShortMessage{
		Src:     from,
		DstList: to,
		// Dst:           to[0],
		Validity:      10 * time.Minute,
		SourceAddrTON: n.sourceAddrTON,
		DestAddrTON:   n.destAddrTON,
		SourceAddrNPI: n.sourceAddrNPI,
		DestAddrNPI:   n.destAddrNPI,
		Text:          pdutext.Raw(msg.Payload),
		// Text:     pdutext.Raw("Lorem ipsum"),
		Register: pdufield.NoDeliveryReceipt,
	}

	for _, m := range jsonMsg.Data {
		for _, k := range fields {
			if v, ok := m.Payload[k]; v != nil && ok {
				if val, ok := v.(float64); ok && val != 0 {
					_, err := n.t.SubmitLongMsg(send)
					// fmt.Println(send)
					return err
				}
			}
		}
	}

	return nil
}

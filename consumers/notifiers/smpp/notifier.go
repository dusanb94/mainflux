// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smpp

import (
	"fmt"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	footer          = "Sent by Mainflux SMTP Notification"
	contentTemplate = "A publisher with an id %s sent the message over %s with the following values \n %s"
)

var _ notify.Notifier = (*notifier)(nil)

type notifier struct {
	t *smpp.Transmitter
}

// New instantiates SMTP message notifier.
func New(t *smpp.Transmitter) notify.Notifier {
	return &notifier{t: t}
}

func (n *notifier) Notify(from string, to []string, msg messaging.Message) error {
	subject := fmt.Sprintf(`Notification for Channel %s`, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s and subtopic %s", subject, msg.Subtopic)
	}

	_, err := n.t.Submit(&smpp.ShortMessage{
		Src:      "",
		Dst:      "",
		Text:     pdutext.Raw(msg.Payload),
		Register: pdufield.NoDeliveryReceipt,
		// 	TLVFields: pdutlv.Fields{
		// 		pdutlv.TagReceiptedMessageID: pdutlv.CString(r.FormValue("msgId")),
		// 	},
	})

	// content := fmt.Sprintf(contentTemplate, msg.Publisher, msg.Protocol, values)

	return err
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"bytes"
	"fmt"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	mux "github.com/plgd-dev/go-coap/v2/mux"
)

// Observer wraps CoAP client.
type Observer interface {
	Handle(m messaging.Message) error
	Cancel() error
	Token() string
	// Message() chan<- messaging.Message
}

type observers map[string]Observer

// Observer is used to handle CoAP subscription.
type observer struct {
	// Expired flag is used to mark that ticker sent a
	// CON message, but response is not received yet.
	// The flag changes its value once ACK message is
	// received from the client. If Expired is true
	// when ticker is triggered, Observer should be canceled
	// and removed from the Service map.
	// expired bool

	// // Message ID for notification messages.
	// msgID uint16

	// expiredLock, msgIDLock sync.Mutex

	// // Messages is used to receive messages from NATS.
	// Messages chan messaging.Message

	// // Cancel channel is used to cancel observing resource.
	// // Cancel channel should not be used to send or receive any
	// // data, it's purpose is to be closed once Observer canceled.
	// Cancel chan bool

	// Conn represents client connection.
	client   mux.Client
	token    message.Token
	messages chan messaging.Message
	// sub broker.Subscription
}

// NewObserver instantiates a new Observer.
func NewObserver(client mux.Client, token message.Token) Observer {
	return &observer{
		client: client,
		token:  token,
	}
}

func (o *observer) Cancel() error {
	return o.client.Close()
}

func (o *observer) Token() string {
	return o.token.String()
}

func (o *observer) Handle(msg messaging.Message) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   o.token,
		Context: o.client.Context(),
		// Body:    bytes.NewReader(msg.Payload),
		Body: bytes.NewReader([]byte(fmt.Sprintf("Been running for %v", "sas"))),
	}
	var opts message.Options
	var buf []byte
	opts, _, err := opts.SetContentFormat(buf, message.TextPlain)
	if err != nil {
		return err
	}
	// if obs >= 0 {
	// opts, n, err = opts.SetObserve(buf, uint32(1))
	// if err == message.ErrTooSmall {
	// 	buf = append(buf, make([]byte, n)...)
	// 	opts, n, err = opts.SetObserve(buf, uint32(1))
	// }
	// if err != nil {
	// 	return fmt.Errorf("cannot set options to response: %w", err)
	// }
	// }
	m.Options = opts
	return o.client.WriteMessage(&m)
}

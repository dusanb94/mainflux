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
	Done() <-chan struct{}
	Token() string
	Message() chan<- messaging.Message
}

type observers map[string]Observer

// Observer is used to handle CoAP subscription.
type observer struct {
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

func (o *observer) Done() <-chan struct{} {
	return o.client.Context().Done()
}

func (o *observer) Do() {
	select {
	case <-o.client.Context().Done():
		o.client.Close()
	case msg := <-o.messages:
		o.Handle(msg)
	}
}

func (o *observer) Cancel() error {
	return o.client.Close()
}

func (o *observer) Token() string {
	return o.token.String()
}

func (o *observer) Message() chan<- messaging.Message {
	return o.messages
}

func (o *observer) Handle(msg messaging.Message) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   o.token,
		Context: o.client.Context(),
		Body:    bytes.NewReader(msg.Payload),
	}
	var opts message.Options
	var buff []byte

	opts, n, err := opts.SetContentFormat(buff, message.TextPlain)
	if err == message.ErrTooSmall {
		buff = append(buff, make([]byte, n)...)
		opts, n, err = opts.SetContentFormat(buff, message.TextPlain)
	}
	if err != nil {
		return fmt.Errorf("cannot set content format to response: %w", err)
	}
	// // if obs >= 0 {
	// opts, n, err = opts.SetObserve(buff, uint32(ob))
	// if err == message.ErrTooSmall {
	// 	buff = append(buff, make([]byte, n)...)
	// 	opts, n, err = opts.SetObserve(buff, uint32(ob))
	// }
	// if err != nil {
	// 	return fmt.Errorf("cannot set options to response: %w", err)
	// }
	// // }
	m.Options = opts
	return o.client.WriteMessage(&m)
}

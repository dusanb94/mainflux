// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"bytes"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	broker "github.com/nats-io/nats.go"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	mux "github.com/plgd-dev/go-coap/v2/mux"
)

// Handler wraps CoAP client.
type Handler interface {
	Handle(m messaging.Message) error
	Cancel() error
	Done() <-chan struct{}
	Token() string
	Sub(*broker.Subscription)
}

type handlers map[string]Handler

// ErrOption indicates an error when adding an option.
var ErrOption = errors.New("unable to set option")

type handler struct {
	client   mux.Client
	token    message.Token
	messages chan messaging.Message
	sub      *broker.Subscription
}

// NewHandler instantiates a new Observer.
func NewHandler(client mux.Client, token message.Token) Handler {
	return &handler{
		client: client,
		token:  token,
	}
}

func (h *handler) Sub(s *broker.Subscription) {
	h.sub = s
}

func (h *handler) Done() <-chan struct{} {
	return h.client.Context().Done()
}

func (h *handler) Cancel() error {
	if err := h.sub.Unsubscribe(); err != nil {
		return err
	}
	return h.client.Close()
}

func (h *handler) Token() string {
	return h.token.String()
}

func (h *handler) Handle(msg messaging.Message) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   h.token,
		Context: h.client.Context(),
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
		return errors.Wrap(ErrOption, err)
	}
	m.Options = opts
	return h.client.WriteMessage(&m)
}

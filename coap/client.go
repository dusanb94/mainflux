// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"bytes"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	mux "github.com/plgd-dev/go-coap/v2/mux"
)

// Client wraps CoAP client.
type Client interface {
	SendMessage(m messaging.Message) error
	Cancel() error
	Done() <-chan struct{}
	// In CoAP terminology similar to the Session ID.
	Token() string
}

type observers map[string]Observer

// ErrOption indicates an error when adding an option.
var ErrOption = errors.New("unable to set option")

type client struct {
	client mux.Client
	token  message.Token
}

// NewClient instantiates a new Observer.
func NewClient(mc mux.Client, token message.Token) Client {
	return &client{
		client: mc,
		token:  token,
	}
}

func (c *client) Done() <-chan struct{} {
	return c.client.Context().Done()
}

func (c *client) Cancel() error {
	return c.client.Close()
}

func (c *client) Token() string {
	return c.token.String()
}

func (c *client) SendMessage(msg messaging.Message) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   c.token,
		Context: c.client.Context(),
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
	return c.client.WriteMessage(&m)
}

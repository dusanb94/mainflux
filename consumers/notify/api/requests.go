// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errInvalidTopic   = errors.New("invalid Subscritpion topic")
	errInvalidContact = errors.New("invalid Subscritpion contact")
	errNotFound       = errors.New("invalid or empty Subscription id")
)

type createSubReq struct {
	token   string
	Topic   string
	Contact string `json:"contact,omitempty"`
}

func (req createSubReq) validate() error {
	if req.token == "" {
		return notify.ErrUnauthorizedAccess
	}
	if req.Topic == "" {
		return errInvalidTopic
	}
	if req.Contact == "" {
		return errInvalidContact
	}
	return nil
}

type subReq struct {
	token string
	id    string
}

func (req subReq) validate() error {
	if req.token == "" {
		return notify.ErrUnauthorizedAccess
	}
	if req.id == "" {
		return errNotFound
	}
	return nil
}

type listSubsReq struct {
	token   string
	topic   string
	contact string
}

func (req listSubsReq) validate() error {
	if req.token == "" {
		return notify.ErrUnauthorizedAccess
	}
	return nil
}

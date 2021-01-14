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
	Contact string `json:"contact,omitempty"`
	Topic   string `json:"topic,omitempty"`
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
	token   string
	OwnerID string `json:"owner_id"`
	topic   string
}

func (req subReq) validate() error {
	if req.token == "" {
		return notify.ErrUnauthorizedAccess
	}
	if req.OwnerID == "" || req.topic == "" {
		return errNotFound
	}
	return nil
}

type listSubsReq struct {
	token string
	topic string
}

func (req listSubsReq) validate() error {
	if req.token == "" {
		return notify.ErrUnauthorizedAccess
	}
	if req.topic == "" {
		return errNotFound
	}
	return nil
}

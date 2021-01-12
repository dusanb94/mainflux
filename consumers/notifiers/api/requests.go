// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errInvalidTopic = errors.New("invalid Subscritpion topic")
	errNotFound     = errors.New("invalid or empty Subscription id")
)

type createSubReq struct {
	token string
	sub   notifiers.Subscription
}

func (req createSubReq) validate() error {
	if req.token == "" {
		return notifiers.ErrUnauthorizedAccess
	}
	if req.sub.Topic == "" {
		return errInvalidTopic
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
		return notifiers.ErrUnauthorizedAccess
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
		return notifiers.ErrUnauthorizedAccess
	}
	if req.topic == "" {
		return errNotFound
	}
	return nil
}
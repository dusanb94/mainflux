// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/consumers/notify"
)

func createSubscriptionEndpoint(svc notify.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createSubReq)
		if err := req.validate(); err != nil {
			return createSubRes{}, err
		}
		id, err := svc.CreateSubscription(ctx, req.token, req.sub)
		if err != nil {
			return createSubRes{}, err
		}
		ucr := createSubRes{
			ID: id,
		}

		return ucr, nil
	}
}

func viewSubscriptionEndpint(svc notify.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(subReq)
		if err := req.validate(); err != nil {
			return viewSubRes{}, err
		}
		sub, err := svc.ViewSubscription(ctx, req.token, req.topic)
		if err != nil {
			return viewSubRes{}, err
		}
		res := viewSubRes{
			ID:         sub.ID,
			OwnerID:    sub.OwnerID,
			OwnerEmail: sub.OwnerEmail,
			Topic:      sub.Topic,
		}
		return res, nil
	}
}

func listSubscriptionsEndpoint(svc notify.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listSubsReq)
		if err := req.validate(); err != nil {
			return listSubsRes{}, err
		}
		subs, err := svc.ListSubscriptions(ctx, req.token, req.topic)
		if err != nil {
			return listSubsRes{}, err
		}
		var res listSubsRes
		for _, sub := range subs {
			r := viewSubRes{
				ID:         sub.ID,
				OwnerID:    sub.OwnerID,
				OwnerEmail: sub.OwnerEmail,
				Topic:      sub.Topic,
			}
			res.Data = append(res.Data, r)
		}
		return res, nil
	}
}

func deleteSubscriptionEndpint(svc notify.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(subReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.RemoveSubscription(ctx, req.token, req.OwnerID, req.topic); err != nil {
			return nil, err
		}
		return removeSubRes{}, nil
	}
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/consumers/notifiers"
)

func createSubscriptionEndpoint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(subReq)
		if err := req.validate(); err != nil {
			return createSubRes{}, err
		}
		id, err := svc.CreateSubscription(ctx, req.token, req.sub)
		if err != nil {
			return createSubRes{}, err
		}
		ucr := createSubRes{
			ID:      id,
			created: true,
		}

		return ucr, nil
	}
}

func viewSubscriptionEndpint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewSubReq)
		if err := req.validate(); err != nil {
			return viewSubResp, err
		}
		sub, err := svc.ViewSubscription(ctx, req.token, req.userID)
		if err != nil {
			return viewSubResp{}, err
		}
		return viewSubResp{sub}, nil
	}
}

func listSubscriptionsEndpoint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listSubsReq)
		if err := req.validate(); err != nil {
			return listSubResp, err
		}
		subs, err := svc.ListSubscriptions(ctx, req.token, req.topic)
		if err != nil {
			return listSubsResp{}, err
		}
		return listSubsResp{subs}, nil
	}
}

func deleteSubscriptionEndpint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteSubReq)
		if err := req.validate(); err != nil {
			return err
		}
		return svc.RemoveSubscription(ctx, req.token, req.id)
}

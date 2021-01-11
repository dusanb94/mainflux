// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/pkg/errors"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	contentType = "application/json"

	offsetKey   = "offset"
	limitKey    = "limit"
	nameKey     = "name"
	emailKey    = "email"
	metadataKey = "metadata"

	defOffset = 0
	defLimit  = 10
)

var errMalformedEntity = errors.New("failed to decode request body")

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc notifiers.Service, tracer opentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/subscriptions", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_subscription")(createSubscriptionEndpoint(svc)),
		decodeCreateSubscription,
		encodeResponse,
		opts...,
	))

	mux.Get("/subscriptions/:topic", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_subscription")(viewSubscriptionEndpint(svc)),
		decodeSubscription,
		encodeResponse,
		opts...,
	))

	mux.Get("/subscriptions", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_subscriptions")(listSubscriptionsEndpoint(svc)),
		decodeListSubscriptions,
		encodeResponse,
		opts...,
	))
	mux.Delete("/subscriptions/:topic", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete_group")(deleteSubscriptionEndpint(svc)),
		decodeSubscription,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", mainflux.Version("notifier"))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeCreateSubscription(_ context.Context, r *http.Request) (interface{}, error) {
	var req createSubReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errMalformedEntity, err)
	}

	req.token = r.Header.Get("Authorization")
	return req, nil
}

func decodeSubscription(_ context.Context, r *http.Request) (interface{}, error) {
	req := subReq{
		topic: bone.GetValue(r, "topic"),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errMalformedEntity, err)
	}
	req.token = r.Header.Get("Authorization")

	return req, nil
}

func decodeListSubscriptions(_ context.Context, r *http.Request) (interface{}, error) {
	req := listSubsReq{
		token: r.Header.Get("Authorization"),
	}
	vals := bone.GetQuery(r, "topic")
	if len(vals) > 0 {
		req.topic = vals[0]
	}

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch errorVal := err.(type) {
	case errors.Error:
		w.Header().Set("Content-Type", contentType)
		switch {
		case errors.Contains(errorVal, errMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.ErrUnexpectedEOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		if errorVal.Msg() != "" {
			if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mainflux/mainflux/pkg/errors"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/users"
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

var (
	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")

	// ErrFailedDecode indicates failed to decode request body
	ErrFailedDecode = errors.New("failed to decode request body")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc users.Service, tracer opentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/subscriptions", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_subscription")(registrationEndpoint(svc)),
		decodeCreateSubscription,
		encodeResponse,
		opts...,
	))

	mux.Get("/subscriptions", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_subscriptions")(viewProfileEndpoint(svc)),
		decodeViewProfile,
		encodeResponse,
		opts...,
	))

	mux.Get("/subscriptions/:subID", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_subscription")(viewProfileEndpoint(svc)),
		decodeViewProfile,
		encodeResponse,
		opts...,
	))

	mux.Delete("/subscriptions/:subID", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete_group")(deleteGroupEndpoint(svc)),
		decodeGroupRequest,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", mainflux.Version("notifier"))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeViewUser(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserReq{
		token:  r.Header.Get("Authorization"),
		userID: bone.GetValue(r, "userID"),
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
		case errors.Contains(errorVal, users.ErrMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUnauthorizedAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Contains(errorVal, users.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Contains(errorVal, users.ErrGroupConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Contains(errorVal, ErrUnsupportedContentType):
			w.WriteHeader(http.StatusUnsupportedMediaType)
		case errors.Contains(errorVal, ErrFailedDecode):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.ErrUnexpectedEOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUserNotFound):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrRecoveryToken):
			w.WriteHeader(http.StatusNotFound)
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

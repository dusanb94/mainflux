// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/consumers/notify"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	saveOp            = "save_op"
	retrieveByEmailOp = "retrieve_by_email"
	updatePassword    = "update_password"
	members           = "members"
)

var _ notify.SubscriptionsRepository = (*subRepositoryMiddleware)(nil)

type subRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   notify.SubscriptionsRepository
}

// New instantiates a new Subscriptions repository that
// tracks request and their latency, and adds spans to context.
func New(repo notify.SubscriptionsRepository, tracer opentracing.Tracer) notify.SubscriptionsRepository {
	return subRepositoryMiddleware{
		tracer: tracer,
		repo:   repo,
	}
}

func (urm subRepositoryMiddleware) Save(ctx context.Context, sub notify.Subscription) (string, error) {
	span := createSpan(ctx, urm.tracer, saveOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Save(ctx, sub)
}

func (urm subRepositoryMiddleware) Retrieve(ctx context.Context, id string) (notify.Subscription, error) {
	span := createSpan(ctx, urm.tracer, saveOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Retrieve(ctx, id)
}

func (urm subRepositoryMiddleware) RetrieveAll(ctx context.Context, topic, contact string) ([]notify.Subscription, error) {
	span := createSpan(ctx, urm.tracer, retrieveByEmailOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.RetrieveAll(ctx, topic, contact)
}

func (urm subRepositoryMiddleware) Remove(ctx context.Context, id string) error {
	span := createSpan(ctx, urm.tracer, retrieveByEmailOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Remove(ctx, id)
}

func createSpan(ctx context.Context, tracer opentracing.Tracer, opName string) opentracing.Span {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return tracer.StartSpan(
			opName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	}
	return tracer.StartSpan(opName)
}

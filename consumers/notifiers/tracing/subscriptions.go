// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/consumers/notifiers"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	saveOp            = "save_op"
	retrieveByEmailOp = "retrieve_by_email"
	updatePassword    = "update_password"
	members           = "members"
)

var _ notifiers.SubscriptionsRepository = (*subRepositoryMiddleware)(nil)

type subRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   notifiers.SubscriptionsRepository
}

// New instantiates a new Subscriptions repository that
// tracks request and their latency, and adds spans to context.
func New(repo notifiers.SubscriptionsRepository, tracer opentracing.Tracer) notifiers.SubscriptionsRepository {
	return subRepositoryMiddleware{
		tracer: tracer,
		repo:   repo,
	}
}

func (urm subRepositoryMiddleware) Save(ctx context.Context, sub notifiers.Subscription) (string, error) {
	span := createSpan(ctx, urm.tracer, saveOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Save(ctx, sub)
}

func (urm subRepositoryMiddleware) Retrieve(ctx context.Context, ownerID, topic string) (notifiers.Subscription, error) {
	span := createSpan(ctx, urm.tracer, saveOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Retrieve(ctx, ownerID, topic)
}

func (urm subRepositoryMiddleware) RetrieveAll(ctx context.Context, topic string) ([]notifiers.Subscription, error) {
	span := createSpan(ctx, urm.tracer, retrieveByEmailOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.RetrieveAll(ctx, topic)
}

func (urm subRepositoryMiddleware) Remove(ctx context.Context, ownerID, topic string) error {
	span := createSpan(ctx, urm.tracer, retrieveByEmailOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return urm.repo.Remove(ctx, ownerID, topic)
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

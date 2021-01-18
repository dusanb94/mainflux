// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/consumers/notify"
	log "github.com/mainflux/mainflux/logger"
)

var _ notify.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    notify.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc notify.Service, logger log.Logger) notify.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreateSubscription(ctx context.Context, token string, sub notify.Subscription) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_subscription with the id %s for token %s took %s to complete", id, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateSubscription(ctx, token, sub)
}

func (lm *loggingMiddleware) ViewSubscription(ctx context.Context, token, topic string) (sub notify.Subscription, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_subscription with the topic %s for token %s took %s to complete", topic, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewSubscription(ctx, token, topic)
}

func (lm *loggingMiddleware) ListSubscriptions(ctx context.Context, token string, pm notify.PageMetadata) (res notify.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_subscription for topic %s and token %s took %s to complete", pm.Topic, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListSubscriptions(ctx, token, pm)
}

func (lm *loggingMiddleware) RemoveSubscription(ctx context.Context, token, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove_subscription for subscription %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveSubscription(ctx, token, id)
}

func (lm *loggingMiddleware) Consume(msg interface{}) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method consume took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Consume(msg)
}

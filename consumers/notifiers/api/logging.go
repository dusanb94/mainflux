// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/consumers/notifiers"
	log "github.com/mainflux/mainflux/logger"
)

var _ notifiers.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    notifiers.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc notifiers.Service, logger log.Logger) notifiers.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreateSubscription(ctx context.Context, token string, sub notifiers.Subscription) (id string, err error) {
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

func (lm *loggingMiddleware) ViewSubscription(ctx context.Context, token, topic string) (sub notifiers.Subscription, err error) {
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

func (lm *loggingMiddleware) ListSubscriptions(ctx context.Context, token, topic string) (res []notifiers.Subscription, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_subscription for topic %s and token %s took %s to complete", topic, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListSubscriptions(ctx, token, topic)
}

func (lm *loggingMiddleware) RemoveSubscription(ctx context.Context, token, ownerID, topic string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove_subscription from the owner %s for token %s and topic %s took %s to complete", ownerID, token, topic, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveSubscription(ctx, token, ownerID, topic)
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

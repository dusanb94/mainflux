// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/consumers"
	log "github.com/mainflux/mainflux/logger"
)

var _ consumers.MessageConsumer = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	c      consumers.MessageConsumer
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(c consumers.MessageConsumer, logger log.Logger) consumers.MessageConsumer {
	return &loggingMiddleware{
		logger: logger,
		c:      c,
	}
}

func (lm *loggingMiddleware) Consume(msgs interface{}) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method consume took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.c.Consume(msgs)
}

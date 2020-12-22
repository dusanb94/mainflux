// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/consumers"
)

var _ consumers.MessageConsumer = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	c       consumers.MessageConsumer
}

// MetricsMiddleware returns new message repository
// with Save method wrapped to expose metrics.
func MetricsMiddleware(c consumers.MessageConsumer, counter metrics.Counter, latency metrics.Histogram) consumers.MessageConsumer {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		c:       c,
	}
}

func (mm *metricsMiddleware) Consume(msgs interface{}) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "consume").Add(1)
		mm.latency.With("method", "consume").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mm.c.Consume(msgs)
}

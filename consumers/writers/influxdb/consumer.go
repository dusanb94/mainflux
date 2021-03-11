// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb

import (
	"math"
	"time"

	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"

	influxdata "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

const (
	senmlPoints   = "messages"
	protocolField = "protocol"
)

var errSaveMessage = errors.New("failed to save message to influxdb database")

var _ consumers.Consumer = (*influxRepo)(nil)

type influxRepo struct {
	writeAPI api.WriteAPI
}

// New returns new InfluxDB writer.
func New(writeAPI api.WriteAPI) consumers.Consumer {
	return &influxRepo{
		writeAPI: writeAPI,
	}
}

func (repo *influxRepo) Consume(message interface{}) error {
	var pts []*write.Point
	switch m := message.(type) {
	case json.Messages:
		pts = repo.jsonPoints(pts, m)
	default:
		var err error
		pts, err = repo.senmlPoints(pts, m)
		if err != nil {
			return err
		}
	}

	for _, pt := range pts {
		repo.writeAPI.WritePoint(pt)
	}

	return nil
}

func (repo *influxRepo) senmlPoints(pts []*write.Point, messages interface{}) ([]*write.Point, error) {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return nil, errSaveMessage
	}

	for _, msg := range msgs {
		tgs, flds := senmlTags(msg), senmlFields(msg)

		sec, dec := math.Modf(msg.Time)
		t := time.Unix(int64(sec), int64(dec*(1e9)))

		pt := influxdata.NewPoint(senmlPoints, tgs, flds, t)
		pts = append(pts, pt)
	}

	return pts, nil
}

func (repo *influxRepo) jsonPoints(pts []*write.Point, msgs json.Messages) []*write.Point {
	for i, m := range msgs.Data {
		t := time.Unix(0, m.Created+int64(i))

		// Copy first-level fields so that the original Payload is unchanged.
		fields := make(map[string]interface{})
		for k, v := range m.Payload {
			fields[k] = v
		}
		// At least one known field need to exist so that COUNT can be performed.
		fields[protocolField] = m.Protocol
		pt := influxdata.NewPoint(msgs.Format, jsonTags(m), fields, t)
		pts = append(pts, pt)
	}

	return pts
}

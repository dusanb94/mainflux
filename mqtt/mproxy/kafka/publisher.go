// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
)

const prefix = "channel"

var _ mainflux.MessagePublisher = (*kafkaPublisher)(nil)

type kafkaPublisher struct {
	prod sarama.AsyncProducer
}

// New instantiates Kafka message publisher.
func New(prod sarama.AsyncProducer) mainflux.MessagePublisher {
	return &kafkaPublisher{prod: prod}
}

func (pub *kafkaPublisher) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("%s.%s", prefix, msg.Channel)
	if msg.Subtopic != "" {
		topic = fmt.Sprintf("%s.%s", topic, msg.Subtopic)
	}
	prodMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
	pub.prod.Input() <- prodMsg
	return nil
}

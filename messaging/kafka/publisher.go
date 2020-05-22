// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"

	"github.com/mainflux/mainflux/messaging"
	broker "github.com/segmentio/kafka-go"
)

const prefix = "channel"

var _ messaging.Publisher = (*kafkaPublisher)(nil)

type kafkaPublisher struct {
	writer *broker.Writer
}

// New instantiates Kafka message publisher.
func New(writer *broker.Writer) messaging.Publisher {
	return &kafkaPublisher{writer: writer}
}

func (pub *kafkaPublisher) Publish(topic string, msg messaging.Message) error {
	if msg.Subtopic == "" {
		return nil
	}
	kafkaMsg := broker.Message{
		Topic: msg.Subtopic,
		Value: msg.Payload,
	}
	return pub.writer.WriteMessages(context.TODO(), kafkaMsg)
}

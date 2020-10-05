package vega

import (
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers"
)

type funcTransformer func(msg messaging.Message) (interface{}, error)

// New returns transformer service implementation for SenML messages.
func New(contentFormat string) transformers.Transformer {
	return funcTransformer(transformer)
}

func (ft funcTransformer) Transform(msg messaging.Message) (interface{}, error) {
	return ft(msg)
}

func transformer(msg messaging.Message) (interface{}, error) {
	var m Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return nil, err
	}
	return m, nil
}
